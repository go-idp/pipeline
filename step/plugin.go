package step

// Plugin represents a plugin of the step
type Plugin struct {
	// Image is the image of the plugin, e.g. "docker.io/library/alpine:latest"
	Image string `json:"image" yaml:"image"`

	// Settings are the settings of the plugin
	// rules: PIPELINE_PLUGIN_SETTINGS_<snake case of key>=value
	// e.g. {"key": "value", "a": "b" } => PIPELINE_PLUGIN_SETTINGS_KEY=value, PIPELINE_PLUGIN_SETTINGS_A=b
	Settings map[string]string `json:"settings" yaml:"settings"`

	// Entrypoint is the entrypoint of the plugin, default is "/pipeline/plugin/run"
	Entrypoint string `json:"entrypoint" yaml:"entrypoint"`

	// inheritEnv is the flag to inherit the environment of the step
	inheritEnv bool

	// ImageRegistry is the image registry of the plugin, e.g. "docker.io"
	ImageRegistry string `json:"image_registry" yaml:"image_registry"`

	// ImageRegistryUsername is the image registry username of the plugin, e.g. "username"
	ImageRegistryUsername string `json:"image_registry_username" yaml:"image_registry_username"`

	// ImageRegistryPassword is the image registry password of the plugin, e.g. "password"
	ImageRegistryPassword string `json:"image_registry_password" yaml:"image_registry_password"`
}
