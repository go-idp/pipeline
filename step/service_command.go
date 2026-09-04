package step

import (
	"fmt"
	"net/url"
	"strings"
)

// serviceUsesCommand reports whether this service step must be executed as a
// generated shell command routed through the configured engine (remote engine),
// instead of a native Go SDK deploy (local engine).
//
// Rule: an empty engine, or a `host` engine, executes the Go SDK locally;
// any other scheme (ssh / idp / docker / unknown) runs the command over the
// engine, reusing the pipeline's engine transport (including SSH auth). The Go
// SDK cannot drive a remote engine because `client.FromEnv` / docker's SSH
// connhelper do not accept an inline-password `ssh://` engine URI.
func (s *Step) serviceUsesCommand() bool {
	if s.Service == nil || s.Engine == "" {
		return false
	}
	return engineScheme(s.Engine) != "host"
}

// engineScheme extracts the scheme of an engine URI, falling back to the raw
// string when it is not a valid URL (e.g. "host", "docker").
func engineScheme(engine string) string {
	u, err := url.Parse(engine)
	if err != nil || u.Scheme == "" {
		return engine
	}
	return u.Scheme
}

// buildServiceCommand generates the shell command that executes this service on
// the remote engine host. `${VAR}` references in `config` are left intact and
// interpolated by the remote `docker compose` / `docker stack` / `kubectl`
// using the step environment (propagated through the engine session).
func (s *Step) buildServiceCommand() (string, error) {
	switch s.Service.Type {
	case ServiceTypeDockerCompose:
		return s.buildComposeCommand(), nil
	case ServiceTypeDockerSwarm:
		return s.buildSwarmCommand(), nil
	case ServiceTypeKubernetes, "k8s":
		return s.buildKubeCommand(), nil
	default:
		return "", fmt.Errorf("unsupported service type %s", s.Service.Type)
	}
}

const (
	// serviceConfigFile is the env var holding the path of the temp config file.
	serviceConfigFile = "PIPELINE_SERVICE_FILE"
	// serviceDockerConfig is the env var holding the path of a temp Docker
	// config dir. It redirects DOCKER_CONFIG so that `docker login` never writes
	// credentials under a read-only HOME (e.g. macOS agents whose HOME is /root),
	// while user-level CLI plugins (~/.docker/cli-plugins, e.g. Compose) are
	// symlinked into it so they stay discoverable.
	serviceDockerConfig = "PIPELINE_SERVICE_DOCKER_CONFIG"
	// serviceHeredoc is the quoted-heredoc delimiter used to write config.
	// Quoting prevents the local shell from interpolating `${VAR}`/`$`/`;`.
	serviceHeredoc = "PIPELINE_SERVICE_EOF"
)

// writeServiceConfig returns a shell snippet that writes config verbatim to a
// temp file via a quoted heredoc. The remote CLI performs interpolation.
func writeServiceConfig(config string) string {
	return fmt.Sprintf(`cat > "$%s" <<'%s'
%s
%s`, serviceConfigFile, serviceHeredoc, config, serviceHeredoc)
}

// registryLogin returns a shell snippet that logs into the registry using env
// vars (PIPELINE_SERVICE_REGISTRY / _USER / _PASS), so credentials never appear
// in the command string. It is a no-op when no registry is configured.
//
// It redirects DOCKER_CONFIG to a writable temp dir so `docker login` succeeds
// even when the pipeline's HOME is not writable (e.g. a macOS agent whose HOME is
// /root, which is read-only). Redirecting DOCKER_CONFIG alone would hide the
// user-level CLI plugins (~/.docker/cli-plugins), so the Compose plugin (Colima /
// Homebrew) is symlinked into the temp config dir's cli-plugins — otherwise
// `docker compose -f ...` would fail with "unknown shorthand flag: 'f' in -f".
// The plugin lives in the docker user's home, which may differ from $HOME, so
// both $HOME and the macOS console user's home are probed.
func registryLogin() string {
	return `if [ -n "$PIPELINE_SERVICE_REGISTRY" ]; then
  export PIPELINE_SERVICE_DOCKER_CONFIG=$(mktemp -d)
  export DOCKER_CONFIG="$PIPELINE_SERVICE_DOCKER_CONFIG"
  mkdir -p "$PIPELINE_SERVICE_DOCKER_CONFIG/cli-plugins"
  _plugin_home="${HOME:-}"
  _console_user=""
  if [ "$(uname -s)" = "Darwin" ]; then
    _console_user="$(stat -f '%Su' /dev/console 2>/dev/null || true)"
  fi
  for _home in "$_plugin_home" "/Users/$_console_user"; do
    if [ -n "$_home" ] && [ "$_home" != "/Users/" ] && [ -d "$_home/.docker/cli-plugins" ]; then
      for _p in "$_home/.docker/cli-plugins"/*; do
        [ -e "$_p" ] || continue
        ln -sf "$_p" "$PIPELINE_SERVICE_DOCKER_CONFIG/cli-plugins/$(basename "$_p")"
      done
    fi
  done
  echo "$PIPELINE_SERVICE_REGISTRY_PASS" | docker login -u "$PIPELINE_SERVICE_REGISTRY_USER" --password-stdin "$PIPELINE_SERVICE_REGISTRY"
fi`
}

// pathBootstrap returns a shell snippet that prepends the common Homebrew bin
// dirs to PATH when <bin> is not already resolvable. On macOS, docker / kubectl
// are installed via Homebrew (Apple Silicon: /opt/homebrew/bin, Intel:
// /usr/local/bin), which is not on the default PATH of a non-login /bin/sh, so
// the generated command would otherwise fail with "<bin>: command not found"
// (exit 127). It is a no-op when the binary is already on PATH, so it is safe on
// Linux hosts too.
func pathBootstrap(bin string) string {
	return fmt.Sprintf(`if ! command -v %s >/dev/null 2>&1; then
  for _bin in /opt/homebrew/bin /usr/local/bin; do
    if [ -x "$_bin/%s" ]; then
      export PATH="$_bin:$PATH"
      break
    fi
  done
fi`, bin, bin)
}

// dockerHostDiscovery returns a shell snippet that points DOCKER_HOST at the
// running docker daemon and, when the pipeline's HOME is not writable (e.g. a
// macOS agent running with HOME=/root), switches HOME to the docker user's home.
//
// On macOS the daemon socket is not at /var/run/docker.sock (the Linux default);
// Docker Desktop / OrbStack / Colima place it under the user's home:
// ~/.docker/run/docker.sock, ~/.orbstack/run/docker.sock, or
// ~/.colima/<profile>/docker.sock. Without DOCKER_HOST the docker CLI falls back
// to the default context (unix:///var/run/docker.sock) and fails on macOS with
// "failed to connect to the docker API at unix:///var/run/docker.sock ...
// no such file or directory".
//
// The docker user's home is derived from a found socket (e.g. /Users/<u>/.colima/
// ...), so HOME is set to it even when the pipeline runs as a different user and
// the console login user differs from the docker user. With a writable HOME at the
// docker user's home, `docker login` writes to ~/.docker and the Compose plugin in
// ~/.docker/cli-plugins is discovered directly (no DOCKER_CONFIG redirect needed).
func dockerHostDiscovery() string {
	return `if [ -z "$DOCKER_HOST" ]; then
  _docker_home="${HOME:-}"
  _console_user=""
  if [ "$(uname -s)" = "Darwin" ]; then
    _console_user="$(stat -f '%Su' /dev/console 2>/dev/null || true)"
  fi
  for _home in "$_docker_home" "/Users/$_console_user" /Users/*; do
    if [ -n "$_home" ] && [ "$_home" != "/Users/" ] && [ -d "$_home" ]; then
      for _sock in "$_home/.docker/run/docker.sock" "$_home/.orbstack/run/docker.sock" "$_home/.colima/default/docker.sock" "$_home/.colima"/*/docker.sock; do
        if [ -S "$_sock" ]; then
          export DOCKER_HOST="unix://$_sock"
          # Derive the docker user's home from the socket path and, if the current
          # HOME is not writable (e.g. /root), switch to it so docker login /
          # plugin discovery / the socket all work under one home.
          if [ ! -w "${HOME:-/}" ]; then
            case "$_sock" in
              /Users/*) _u="${_sock#/Users/}"; export HOME="/Users/${_u%%/*}" ;;
            esac
          fi
          break 2
        fi
      done
    fi
  done
  if [ -z "$DOCKER_HOST" ] && [ -S /var/run/docker.sock ]; then
    export DOCKER_HOST="unix:///var/run/docker.sock"
  fi
fi`
}

// head returns the common prologue: fail-fast + PATH bootstrap + temp file.
func (s *Step) head() []string {
	return []string{
		"set -e",
		pathBootstrap("docker"),
		dockerHostDiscovery(),
		fmt.Sprintf("export %s=$(mktemp)", serviceConfigFile),
		// PIPELINE_SERVICE_DOCKER_CONFIG is only set when a registry is configured
		// (see registryLogin); clean it up when present, otherwise skip.
		fmt.Sprintf(`trap 'rm -f "$%s"; [ -n "$%s" ] && rm -rf "$%s"' EXIT`, serviceConfigFile, serviceDockerConfig, serviceDockerConfig),
	}
}

// composeUp returns a shell snippet that runs `docker compose up` (Compose v2,
// preferred) and falls back to `docker-compose` (v1) when the v2 plugin is not
// installed. When neither is available it fails with a helpful message pointing
// to install guidance. Compose v2 supports `--wait`/`--wait-timeout`; the v1
// standalone binary only supports a detached `up -d`.
func composeUp(project string, timeout int64) string {
	return fmt.Sprintf(`if docker compose version >/dev/null 2>&1; then
  docker compose -f "$%s" -p '%s' up -d --wait --wait-timeout %d
elif command -v docker-compose >/dev/null 2>&1; then
  echo "pipeline: 'docker compose' (Compose v2) not found, falling back to 'docker-compose' (v1)"
  docker-compose -f "$%s" -p '%s' up -d
else
  echo "pipeline: neither 'docker compose' (Compose v2) nor 'docker-compose' (v1) is available on this host." >&2
  echo "  Please confirm your docker version and install Docker Compose v2:" >&2
  echo "    macOS (Homebrew):      brew install docker-compose" >&2
  echo "    Ubuntu/Debian:         sudo apt-get install docker-compose-plugin   (old: sudo apt install docker-compose)" >&2
  echo "    Docker Desktop:        enable the Compose integration" >&2
  exit 1
fi`, serviceConfigFile, project, timeout, serviceConfigFile, project)
}

func (s *Step) buildComposeCommand() string {
	return strings.Join(append(s.head(),
		writeServiceConfig(s.Service.Config),
		registryLogin(),
		composeUp(s.Service.Name, s.Service.Timeout),
	), "\n")
}

func (s *Step) buildSwarmCommand() string {
	return strings.Join(append(s.head(),
		writeServiceConfig(s.Service.Config),
		registryLogin(),
		swarmDeploy(s.Service.Name, s.Service.Timeout),
	), "\n")
}

// swarmDeploy deploys the swarm stack, preferring `docker stack deploy
// --detach=false` (which waits for the stack services to converge itself) when
// the executing host's docker CLI supports it. The `--detach` flag for
// `docker stack deploy` was added in Docker Engine 26.0.0; older CLIs must fall
// back to a detached deploy plus a replica convergence poll. Support is probed
// with `docker stack deploy --help` rather than a hard-coded version string, so
// any docker whose CLI lacks the flag is handled correctly (e.g. 20.10.x).
func swarmDeploy(stack string, timeout int64) string {
	body := `if docker stack deploy --help 2>/dev/null | grep -q -- '--detach'; then
  echo "docker-swarm: deploy stack '__STACK__' with --detach=false (docker stack deploy supports --detach)"
  docker stack deploy --detach=false --prune --with-registry-auth -c "$PIPELINE_SERVICE_FILE" '__STACK__'
else
  echo "docker-swarm: docker stack deploy does not support --detach, fall back to detached deploy + readiness poll"
  docker stack deploy --prune --with-registry-auth -c "$PIPELINE_SERVICE_FILE" '__STACK__'
__READINESS__
fi`
	return strings.NewReplacer(
		"__STACK__", stack,
		"__READINESS__", swarmReadiness(stack, timeout),
	).Replace(body)
}

// swarmReadiness polls the stack services' replica convergence, mirroring the
// SDK's `swarmServicesReady`: every service must have running >= desired.
func swarmReadiness(stack string, timeout int64) string {
	body := `PIPELINE_STACK='__STACK__'
PIPELINE_DEADLINE=$(( $(date +%s) + __TIMEOUT__ ))
PIPELINE_READY=0
while [ "$PIPELINE_READY" -eq 0 ]; do
  PIPELINE_SERVICES=$(docker service ls --filter "label=com.docker.stack.namespace=$PIPELINE_STACK" --format '{{.Replicas}}' 2>/dev/null || true)
  PIPELINE_READY=1
  if [ -z "$PIPELINE_SERVICES" ]; then
    PIPELINE_READY=0
  fi
  for REPLICAS in $PIPELINE_SERVICES; do
    RUNNING=${REPLICAS%%/*}
    DESIRED=${REPLICAS##*/}
    if [ "$RUNNING" -lt "$DESIRED" ]; then
      PIPELINE_READY=0
    fi
  done
  if [ "$PIPELINE_READY" -eq 1 ]; then
    echo "docker-swarm: stack ready"
    break
  fi
  if [ "$(date +%s)" -ge "$PIPELINE_DEADLINE" ]; then
    echo "docker-swarm: startup timeout" >&2
    exit 1
  fi
  sleep 6
done`
	return strings.NewReplacer(
		"__STACK__", stack,
		"__TIMEOUT__", fmt.Sprintf("%d", timeout),
	).Replace(body)
}

func (s *Step) buildKubeCommand() string {
	ns := s.Service.Namespace
	if ns == "" {
		ns = "default"
	}

	parts := s.head()
	parts = append(parts, writeServiceConfig(s.Service.Config))
	if s.Service.Kubeconfig != "" {
		parts = append(parts, fmt.Sprintf(`export KUBECONFIG="%s"`, s.Service.Kubeconfig))
	}
	parts = append(parts, pathBootstrap("kubectl"))
	parts = append(parts, `kubectl apply -f "$PIPELINE_SERVICE_FILE"`)
	parts = append(parts, kubeReadiness(ns, s.Service.Timeout))

	return strings.Join(parts, "\n")
}

// kubeReadiness waits for the namespace's Deployments to become Available,
// mirroring the SDK's Deployment readyReplicas polling. It is a no-op when the
// namespace has no Deployment (matching the SDK's "no deployment => ready").
func kubeReadiness(namespace string, timeout int64) string {
	return fmt.Sprintf(`if [ "$(kubectl get deployment -n '%s' -o name 2>/dev/null | wc -l)" -gt 0 ]; then
  kubectl wait --for=condition=Available deployment --all -n '%s' --timeout=%ds
fi`, namespace, namespace, timeout)
}
