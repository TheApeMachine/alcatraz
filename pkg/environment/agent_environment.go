package environment

import (
	"context"
	"fmt"
	"time"

	sdkcontainer "github.com/docker/go-sdk/container"
	"github.com/theapemachine/errnie"
)

/*
AgentEnvironment is a hardened Linux container that agents connect to for development work.
It wraps docker/go-sdk for lifecycle management and command execution.
*/
type AgentEnvironment struct {
	ctx       context.Context
	cancel    context.CancelFunc
	cfg       AgentEnvironmentConfig
	container *sdkcontainer.Container
	running   bool
}

/*
NewAgentEnvironment allocates an environment handle without starting the container.
*/
func NewAgentEnvironment(
	ctx context.Context,
	cfg AgentEnvironmentConfig,
) *AgentEnvironment {
	environmentCtx, cancel := context.WithCancel(ctx)

	return &AgentEnvironment{
		ctx:    environmentCtx,
		cancel: cancel,
		cfg:    cfg,
	}
}

/*
ID returns the environment identifier.
*/
func (agentEnvironment *AgentEnvironment) ID() string {
	return agentEnvironment.cfg.ID
}

/*
Name returns the human-readable environment name.
*/
func (agentEnvironment *AgentEnvironment) Name() string {
	return agentEnvironment.cfg.Name
}

/*
Running reports whether the environment container is active.
*/
func (agentEnvironment *AgentEnvironment) Running() bool {
	return agentEnvironment.running
}

/*
ContainerID returns the Docker container ID when running.
*/
func (agentEnvironment *AgentEnvironment) ContainerID() string {
	if agentEnvironment.container == nil {
		return ""
	}

	return agentEnvironment.container.ID()
}

/*
Start pulls the image if needed and starts a long-lived container.
*/
func (agentEnvironment *AgentEnvironment) Start() error {
	if agentEnvironment.running {
		return domainError(
			errnie.Conflict,
			"environment.start",
			fmt.Sprintf("environment %q already running", agentEnvironment.cfg.ID),
			"environment_id", agentEnvironment.cfg.ID,
		)
	}

	startResult := errnie.Does(func() (*sdkcontainer.Container, error) {
		return sdkcontainer.Run(
			agentEnvironment.ctx,
			buildContainerOptions(agentEnvironment)...,
		)
	})

	agentEnvironment.container = startResult.Value()

	if startResult.Err() == nil {
		agentEnvironment.running = true
		return nil
	}

	return wrapError(
		errnie.IO,
		"environment.start",
		"start container",
		startResult.Err(),
		"environment_id", agentEnvironment.cfg.ID,
	)
}

/*
Exec runs a shell command inside the environment and returns combined output.
*/
func (agentEnvironment *AgentEnvironment) Exec(command string) (string, error) {
	if !agentEnvironment.running || agentEnvironment.container == nil {
		return "", domainError(
			errnie.Validation,
			"environment.exec",
			fmt.Sprintf("environment %q is not running", agentEnvironment.cfg.ID),
			"environment_id", agentEnvironment.cfg.ID,
		)
	}

	timeout := agentEnvironment.execTimeout()
	execCtx, cancel := context.WithTimeout(agentEnvironment.ctx, timeout+5*time.Second)
	defer cancel()

	execResult := errnie.Does(func() (ExecResult, error) {
		return agentEnvironment.execCommand(execCtx, command)
	})

	return execResult.Value().Output, execResult.Err()
}

/*
Stop stops the container without removing it.
*/
func (agentEnvironment *AgentEnvironment) Stop() error {
	if agentEnvironment.container == nil {
		return nil
	}

	stopResult := errnie.Does(func() (struct{}, error) {
		return struct{}{}, agentEnvironment.container.Stop(agentEnvironment.ctx)
	})

	if stopResult.Err() == nil {
		agentEnvironment.running = false
	}

	return wrapError(
		errnie.IO,
		"environment.stop",
		"stop container",
		stopResult.Err(),
		"environment_id", agentEnvironment.cfg.ID,
	)
}

/*
Terminate stops and removes the container, releasing all resources.
*/
func (agentEnvironment *AgentEnvironment) Terminate() error {
	if agentEnvironment.container == nil {
		agentEnvironment.running = false
		return nil
	}

	terminateResult := errnie.Does(func() (struct{}, error) {
		return struct{}{}, agentEnvironment.container.Terminate(agentEnvironment.ctx)
	})

	if terminateResult.Err() == nil {
		agentEnvironment.container = nil
		agentEnvironment.running = false
	}

	return wrapError(
		errnie.IO,
		"environment.terminate",
		"terminate container",
		terminateResult.Err(),
		"environment_id", agentEnvironment.cfg.ID,
	)
}

/*
Close cancels the environment context and terminates the container.
*/
func (agentEnvironment *AgentEnvironment) Close() error {
	defer agentEnvironment.cancel()

	return agentEnvironment.Terminate()
}
