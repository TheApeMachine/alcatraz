package environment

import (
	"time"

	"github.com/moby/moby/api/types/container"

	"github.com/theapemachine/alcatraz/pkg/config"
)

const (
	agentEnvironmentLabelID   = "alcatraz.environment.id"
	agentEnvironmentLabelName = "alcatraz.environment.name"
)

/*
AgentEnvironmentConfig holds the runtime settings for a single agent environment.
*/
type AgentEnvironmentConfig struct {
	ID          string
	Name        string
	Image       string
	ExecTimeout time.Duration
	MemoryBytes int64
	NanoCPUs    int64
	PIDsLimit   int64
	WorkingDir  string
	NetworkMode string
}

/*
AgentEnvironmentConfigFromGlobal merges a per-environment identity with global defaults.
*/
func AgentEnvironmentConfigFromGlobal(
	environmentID string,
	environmentName string,
	globalConfig *config.EnvironmentConfig,
) AgentEnvironmentConfig {
	return AgentEnvironmentConfig{
		ID:          environmentID,
		Name:        environmentName,
		Image:       globalConfig.Image,
		ExecTimeout: globalConfig.ExecTimeout,
		MemoryBytes: globalConfig.MemoryBytes,
		NanoCPUs:    globalConfig.NanoCPUs,
		PIDsLimit:   globalConfig.PIDsLimit,
		WorkingDir:  globalConfig.WorkingDir,
		NetworkMode: globalConfig.NetworkMode,
	}
}

func (agentEnvironment *AgentEnvironment) execTimeout() time.Duration {
	if agentEnvironment.cfg.ExecTimeout > 0 {
		return agentEnvironment.cfg.ExecTimeout
	}

	return config.EnvironmentDefaultExecTimeout
}

func (agentEnvironment *AgentEnvironment) memoryBytes() int64 {
	if agentEnvironment.cfg.MemoryBytes > 0 {
		return agentEnvironment.cfg.MemoryBytes
	}

	return int64(config.EnvironmentDefaultMemoryBytes)
}

func (agentEnvironment *AgentEnvironment) nanoCPUs() int64 {
	if agentEnvironment.cfg.NanoCPUs > 0 {
		return agentEnvironment.cfg.NanoCPUs
	}

	return int64(config.EnvironmentDefaultNanoCPUs)
}

func (agentEnvironment *AgentEnvironment) pidsLimit() int64 {
	if agentEnvironment.cfg.PIDsLimit > 0 {
		return agentEnvironment.cfg.PIDsLimit
	}

	return int64(config.EnvironmentDefaultPIDsLimit)
}

func (agentEnvironment *AgentEnvironment) workingDir() string {
	if agentEnvironment.cfg.WorkingDir != "" {
		return agentEnvironment.cfg.WorkingDir
	}

	return config.EnvironmentDefaultWorkingDir
}

func (agentEnvironment *AgentEnvironment) networkMode() container.NetworkMode {
	if agentEnvironment.cfg.NetworkMode != "" {
		return container.NetworkMode(agentEnvironment.cfg.NetworkMode)
	}

	return container.NetworkMode(config.EnvironmentDefaultNetworkMode)
}

func (agentEnvironment *AgentEnvironment) containerName() string {
	if agentEnvironment.cfg.Name != "" {
		return "alcatraz-" + agentEnvironment.cfg.Name
	}

	return "alcatraz-" + agentEnvironment.cfg.ID
}

func (agentEnvironment *AgentEnvironment) containerLabels() map[string]string {
	return map[string]string{
		agentEnvironmentLabelID:   agentEnvironment.cfg.ID,
		agentEnvironmentLabelName: agentEnvironment.cfg.Name,
	}
}

func (agentEnvironment *AgentEnvironment) hostConfig() *container.HostConfig {
	pidsLimit := agentEnvironment.pidsLimit()

	return &container.HostConfig{
		AutoRemove:     false,
		CapDrop:        []string{"ALL"},
		NetworkMode:    agentEnvironment.networkMode(),
		ReadonlyRootfs: true,
		Resources: container.Resources{
			Memory:    agentEnvironment.memoryBytes(),
			NanoCPUs:  agentEnvironment.nanoCPUs(),
			PidsLimit: &pidsLimit,
		},
		SecurityOpt: []string{"no-new-privileges:true"},
		Tmpfs: map[string]string{
			"/go":        "rw,nosuid,size=2g",
			"/root":      "rw,nosuid,size=64m",
			"/tmp":       "rw,noexec,nosuid,size=512m",
			"/workspace": "rw,nosuid,size=2g",
		},
	}
}

func (agentEnvironment *AgentEnvironment) hostConfigModifier() func(hostConfig *container.HostConfig) {
	restricted := agentEnvironment.hostConfig()

	return func(hostConfig *container.HostConfig) {
		hostConfig.AutoRemove = restricted.AutoRemove
		hostConfig.CapDrop = restricted.CapDrop
		hostConfig.NetworkMode = restricted.NetworkMode
		hostConfig.ReadonlyRootfs = restricted.ReadonlyRootfs
		hostConfig.Resources = restricted.Resources
		hostConfig.SecurityOpt = restricted.SecurityOpt
		hostConfig.Tmpfs = restricted.Tmpfs
	}
}
