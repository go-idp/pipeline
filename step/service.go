package step

// Service represents a natively deployed service step.
//
// Instead of generating shell commands (the legacy `version: v1` hack),
// the pipeline engine deploys the definition directly with Go SDKs:
//   - docker-compose: docker compose SDK (equivalent to `docker compose up -d`)
//   - docker-swarm:   docker CLI stack deploy SDK (equivalent to `docker stack deploy`)
//   - kubernetes:     client-go server-side apply (equivalent to `kubectl apply -f -`)
//
// The SDK calls run in the pipeline process, i.e. on the agent host that
// runs `pipeline run`, which is the same host that has docker / kubectl
// access today. If the step carries a remote engine (`engine: ssh://...`),
// the Go SDK cannot drive it; pipeline instead generates the equivalent
// command (`docker compose up` / `docker stack deploy` / `kubectl apply`)
// and runs it through the engine on the remote host (see service_command.go).
type Service struct {
	// Type is the service type: docker-compose | docker-swarm | kubernetes
	Type string `json:"type" yaml:"type"`

	// Name is the compose project name / swarm stack name.
	// Required for docker-compose and docker-swarm.
	Name string `json:"name" yaml:"name"`

	// Config is the raw definition:
	//   - docker-compose / docker-swarm: docker-compose.yaml content
	//   - kubernetes:                    k8s manifest YAML (multi-doc supported)
	Config string `json:"config" yaml:"config"`

	// Timeout is the startup readiness wait in seconds, default: 120.
	Timeout int64 `json:"timeout" yaml:"timeout"`

	// ImageRegistry is the registry server (host only, no scheme) used to
	// authenticate private image pulls. Equivalent to `docker login`.
	ImageRegistry string `json:"image_registry" yaml:"image_registry"`

	// ImageRegistryUsername is the registry username.
	ImageRegistryUsername string `json:"image_registry_username" yaml:"image_registry_username"`

	// ImageRegistryPassword is the registry password.
	ImageRegistryPassword string `json:"image_registry_password" yaml:"image_registry_password"`

	// Namespace is the kubernetes namespace used for apply and readiness
	// checks. Optional, default: "default".
	Namespace string `json:"namespace" yaml:"namespace"`

	// Kubeconfig is the kubeconfig path for kubernetes.
	// Optional: falls back to $KUBECONFIG, ~/.kube/config, then in-cluster.
	Kubeconfig string `json:"kubeconfig" yaml:"kubeconfig"`
}
