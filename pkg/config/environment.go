package config

import "time"

var (
	Environment        = NewEnvironmentConfig()
	environmentRootKey = "environment"
)

const (
	EnvironmentDefaultImage         = "ubuntu:24.04"
	EnvironmentDefaultExecTimeout   = 10 * time.Minute
	EnvironmentDefaultMemoryBytes   = 4 << 30
	EnvironmentDefaultNanoCPUs      = 2_000_000_000
	EnvironmentDefaultPIDsLimit     = 512
	EnvironmentDefaultWorkingDir    = "/workspace"
	EnvironmentDefaultNetworkMode   = "none"
	EnvironmentDefaultMaxConcurrent = 8
)

/*
EnvironmentConfig holds defaults for containerized agent environments.
*/
type EnvironmentConfig struct {
	Image           string
	ExecTimeout     time.Duration
	MemoryBytes     int64
	NanoCPUs        int64
	PIDsLimit       int64
	WorkingDir      string
	NetworkMode     string
	MaxEnvironments int
}

/*
NewEnvironmentConfig reads environment settings from viper-loaded config.yml.
*/
func NewEnvironmentConfig() *EnvironmentConfig {
	return &EnvironmentConfig{
		Image:           WithDefault(environmentRootKey+".image", EnvironmentDefaultImage),
		ExecTimeout:     WithDefault(environmentRootKey+".exec_timeout", EnvironmentDefaultExecTimeout),
		MemoryBytes:     WithDefault(environmentRootKey+".memory_bytes", int64(EnvironmentDefaultMemoryBytes)),
		NanoCPUs:        WithDefault(environmentRootKey+".nano_cpus", int64(EnvironmentDefaultNanoCPUs)),
		PIDsLimit:       WithDefault(environmentRootKey+".pids_limit", int64(EnvironmentDefaultPIDsLimit)),
		WorkingDir:      WithDefault(environmentRootKey+".working_dir", EnvironmentDefaultWorkingDir),
		NetworkMode:     WithDefault(environmentRootKey+".network_mode", EnvironmentDefaultNetworkMode),
		MaxEnvironments: WithDefault(environmentRootKey+".max_environments", EnvironmentDefaultMaxConcurrent),
	}
}
