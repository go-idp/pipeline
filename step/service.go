package step

// Service is the service configuration of the step
type Service struct {
	// Version is the version of the service
	Version string `json:"version" yaml:"version"`

	// Type is the type of the service, e.g. "docker-compose" | "docker-swarm" | "kubernetes"
	Type string `json:"type" yaml:"type"`

	// Config is the config of the service
	// suport raw config or config file
	Config string `json:"config" yaml:"config"`

	// Name is the name of the service
	Name string `json:"name" yaml:"name"`
}
