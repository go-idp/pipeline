package step

import (
	idp "github.com/go-idp/agent/command/engine"
	"github.com/go-zoox/command/config"
	"github.com/go-zoox/command/engine"
)

func init() {
	// Register the idp engine
	engine.Register(idp.Name, func(cfg *config.Config) (engine.Engine, error) {
		engine, err := idp.New(&idp.Config{
			ID: cfg.ID,
			//
			Command:     cfg.Command,
			WorkDir:     cfg.WorkDir,
			Environment: cfg.Environment,
			User:        cfg.User,
			Shell:       cfg.Shell,
			//
			ReadOnly: cfg.ReadOnly,
			//
			Server:       cfg.Server,
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			//
			AllowedSystemEnvKeys: cfg.AllowedSystemEnvKeys,
		})
		if err != nil {
			return nil, err
		}

		return engine, nil
	})
}
