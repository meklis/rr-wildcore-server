package server

import (
	"time"

	"github.com/roadrunner-server/errors"
)

// Config All config (.rr.yaml)
// For other section use pointer to distinguish between `empty` and `not present`
type Config struct {
	// OnInit configuration
	OnInit *InitConfig `mapstructure:"on_init"`
	// AfterInit command to run after pool initialization
	AfterInit *InitConfig `mapstructure:"after_init"`
	// Command to run as application.
	Command []string `mapstructure:"command"`
	// User to run application under.
	User string `mapstructure:"user"`
	// Group to run application under.
	Group string `mapstructure:"group"`
	// Env represents application environment.
	Env map[string]string `mapstructure:"env"`
	// Relay defines connection method and factory to be used to connect to workers:
	// "pipes", "tcp://:6001", "unix://rr.sock"
	// This config section must not change on re-configuration.
	Relay string `mapstructure:"relay"`
	// RelayTimeout defines for how long socket factory will be waiting for worker connection. This config section
	// must not change on re-configuration. Defaults to 60s.
	RelayTimeout time.Duration `mapstructure:"relay_timeout"`
}

type InitConfig struct {
	// Command which is started before worker starts
	Command []string `mapstructure:"command"`
	// ExecTimeout is execute timeout for the command
	ExecTimeout time.Duration `mapstructure:"exec_timeout"`
	// Env represents application environment.
	Env map[string]string `mapstructure:"env"`
}

// RPCConfig should be in sync with rpc/config.go
// Used to set RPC address env
type RPCConfig struct {
	Listen string `mapstructure:"listen"`
}

// InitDefaults for the server config
func (cfg *Config) InitDefaults() error {
	if len(cfg.Command) == 0 {
		return errors.Str("command should not be empty")
	}

	if cfg.Relay == "" {
		cfg.Relay = "pipes"
	}

	if cfg.RelayTimeout == 0 {
		cfg.RelayTimeout = time.Second * 60
	}

	if cfg.AfterInit != nil {
		if len(cfg.AfterInit.Command) == 0 {
			return errors.Str("after_init command should not be empty")
		}

		if cfg.AfterInit.ExecTimeout == 0 {
			cfg.AfterInit.ExecTimeout = time.Minute
		}
	}

	if cfg.OnInit != nil {
		if len(cfg.OnInit.Command) == 0 {
			return errors.Str("on_init command should not be empty")
		}

		if cfg.OnInit.ExecTimeout == 0 {
			cfg.OnInit.ExecTimeout = time.Minute
		}
	}

	return nil
}
