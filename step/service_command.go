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
func registryLogin() string {
	return `if [ -n "$PIPELINE_SERVICE_REGISTRY" ]; then
  echo "$PIPELINE_SERVICE_REGISTRY_PASS" | docker login -u "$PIPELINE_SERVICE_REGISTRY_USER" --password-stdin "$PIPELINE_SERVICE_REGISTRY"
fi`
}

// head returns the common prologue: fail-fast + temp file with cleanup.
func (s *Step) head() []string {
	return []string{
		"set -e",
		fmt.Sprintf("export %s=$(mktemp)", serviceConfigFile),
		fmt.Sprintf(`trap 'rm -f "$%s"' EXIT`, serviceConfigFile),
	}
}

func (s *Step) buildComposeCommand() string {
	return strings.Join(append(s.head(),
		writeServiceConfig(s.Service.Config),
		registryLogin(),
		fmt.Sprintf(`docker compose -f "$%s" -p '%s' up -d --wait --wait-timeout %d`, serviceConfigFile, s.Service.Name, s.Service.Timeout),
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
