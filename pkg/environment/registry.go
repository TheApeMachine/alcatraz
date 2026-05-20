package environment

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/theapemachine/errnie"

	"github.com/theapemachine/alcatraz/pkg/config"
)

/*
Registry tracks active agent environments in memory.
*/
type Registry struct {
	ctx          context.Context
	cancel       context.CancelFunc
	globalCfg    *config.EnvironmentConfig
	environments map[string]*AgentEnvironment
	mutex        sync.RWMutex
}

/*
NewRegistry creates a registry backed by global environment defaults.
*/
func NewRegistry(ctx context.Context, globalConfig *config.EnvironmentConfig) *Registry {
	registryCtx, cancel := context.WithCancel(ctx)

	return &Registry{
		ctx:          registryCtx,
		cancel:       cancel,
		globalCfg:    globalConfig,
		environments: make(map[string]*AgentEnvironment),
	}
}

/*
Create provisions and starts a new agent environment.
*/
func (registry *Registry) Create(name string) (*AgentEnvironment, error) {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	if len(registry.environments) >= registry.globalCfg.MaxEnvironments {
		return nil, domainError(
			errnie.Conflict,
			"registry.create",
			fmt.Sprintf("max environments (%d) reached", registry.globalCfg.MaxEnvironments),
			"max_environments", registry.globalCfg.MaxEnvironments,
		)
	}

	environmentID := uuid.NewString()
	environmentName := name

	if environmentName == "" {
		environmentName = environmentID[:8]
	}

	for _, existing := range registry.environments {
		if existing.Name() == environmentName {
			return nil, domainError(
				errnie.Conflict,
				"registry.create",
				fmt.Sprintf("environment name %q already exists", environmentName),
				"environment_name", environmentName,
			)
		}
	}

	agentEnvironment := NewAgentEnvironment(
		registry.ctx,
		AgentEnvironmentConfigFromGlobal(environmentID, environmentName, registry.globalCfg),
	)

	startResult := errnie.Does(func() (struct{}, error) {
		return struct{}{}, agentEnvironment.Start()
	})

	if startResult.Err() != nil {
		return nil, startResult.Err()
	}

	registry.environments[environmentID] = agentEnvironment

	return agentEnvironment, nil
}

/*
Get returns an environment by ID.
*/
func (registry *Registry) Get(environmentID string) (*AgentEnvironment, error) {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	agentEnvironment, ok := registry.environments[environmentID]

	if !ok {
		return nil, domainError(
			errnie.NotFound,
			"registry.get",
			fmt.Sprintf("environment %q not found", environmentID),
			"environment_id", environmentID,
		)
	}

	return agentEnvironment, nil
}

/*
List returns a snapshot of all tracked environments.
*/
func (registry *Registry) List() []*AgentEnvironment {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	environments := make([]*AgentEnvironment, 0, len(registry.environments))

	for _, agentEnvironment := range registry.environments {
		environments = append(environments, agentEnvironment)
	}

	return environments
}

/*
Destroy terminates and removes an environment from the registry.
*/
func (registry *Registry) Destroy(environmentID string) error {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	agentEnvironment, ok := registry.environments[environmentID]

	if !ok {
		return domainError(
			errnie.NotFound,
			"registry.destroy",
			fmt.Sprintf("environment %q not found", environmentID),
			"environment_id", environmentID,
		)
	}

	terminateErr := agentEnvironment.Terminate()

	if terminateErr == nil {
		delete(registry.environments, environmentID)
	}

	return terminateErr
}

/*
Close terminates all environments and cancels the registry context.
*/
func (registry *Registry) Close() error {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	terminateErrs := make([]error, 0, len(registry.environments))

	for environmentID, agentEnvironment := range registry.environments {
		terminateErrs = append(terminateErrs, agentEnvironment.Terminate())
		delete(registry.environments, environmentID)
	}

	registry.cancel()

	return errnie.Combine(terminateErrs...)
}

/*
Count returns the number of active environments.
*/
func (registry *Registry) Count() int {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	return len(registry.environments)
}
