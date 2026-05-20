package alcatraz

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/go-sdk/container"
	"github.com/docker/go-sdk/container/exec"
	"github.com/docker/go-sdk/container/wait"
	"github.com/docker/go-sdk/image"
	"github.com/moby/moby/client"
	"github.com/theapemachine/errnie"

	"github.com/theapemachine/alcatraz/pkg/config"
)

type Environment struct {
	ctx      context.Context
	cancel   context.CancelFunc
	err      error
	ID       string
	Name     string
	Image    string
	Command  string
	Args     []string
	ctr      errnie.Result[*container.Container]
	reader   io.Reader
	writer   io.Writer
	exitCode int
	output   errnie.Result[[]byte]
}

func NewEnvironment(
	ctx context.Context,
	id string,
) *Environment {
	ctx, cancel := context.WithCancel(ctx)

	return &Environment{
		ctx:     ctx,
		cancel:  cancel,
		ID:      id,
		Name:    id,
		Image:   config.Environment.Image,
		Command: "/bin/bash",
		Args:    []string{},
	}
}

func (environment *Environment) Start() error {
	if err := errnie.Do(func() error {
		return image.Pull(environment.ctx, environment.Image)
	}); err != nil {
		return err
	}

	if environment.ctr = errnie.Does(func() (*container.Container, error) {
		return container.Run(
			environment.ctx,
			container.WithImage(environment.Image),
			container.WithCmd(environment.Command),
			container.WithWaitStrategy(wait.ForLog(environment.Command)),
		)
	}); environment.ctr.Err() != nil {
		return environment.ctr.Err()
	}

	return nil
}

func (environment *Environment) Stop() error {
	if environment.ctr.Err() != nil {
		return environment.ctr.Err()
	}

	return environment.ctr.Value().Stop(environment.ctx)
}

func (environment *Environment) Exec(command string) (string, error) {
	if environment.ctr.Err() != nil {
		return "", environment.ctr.Err()
	}

	if environment.exitCode, environment.err = environment.attachExec(command); environment.err != nil {
		return "", environment.err
	}

	if environment.output = errnie.Does(func() ([]byte, error) {
		return io.ReadAll(environment.reader)
	}); environment.output.Err() != nil {
		return "", environment.output.Err()
	}

	return string(environment.output.Value()), nil
}

func (environment *Environment) attachExec(command string) (int, error) {
	containerRef := environment.ctr.Value()
	processOptions := exec.NewProcessOptions([]string{command})
	processOptions.ExecConfig.AttachStdin = true

	response, err := containerRef.Client().ExecCreate(
		environment.ctx,
		containerRef.ID(),
		processOptions.ExecConfig,
	)

	if err != nil {
		return 0, fmt.Errorf("container exec create: %w", err)
	}

	hijack, err := containerRef.Client().ExecAttach(
		environment.ctx,
		response.ID,
		client.ExecAttachOptions{},
	)

	if err != nil {
		return 0, fmt.Errorf("container exec attach: %w", err)
	}

	processOptions.Reader = hijack.Reader
	exec.Multiplexed().Apply(processOptions)

	environment.reader = processOptions.Reader
	environment.writer = hijack.Conn

	return environment.awaitExec(containerRef.Client(), response.ID)
}

func (environment *Environment) awaitExec(
	dockerClient client.APIClient,
	execID string,
) (int, error) {
	for {
		execResponse, err := dockerClient.ExecInspect(
			environment.ctx,
			execID,
			client.ExecInspectOptions{},
		)

		if err != nil {
			return 0, fmt.Errorf("container exec inspect: %w", err)
		}

		if !execResponse.Running {
			return execResponse.ExitCode, nil
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func (environment *Environment) Read(payload []byte) (count int, err error) {
	if environment.reader == nil {
		return 0, io.EOF
	}

	return environment.reader.Read(payload)
}

func (environment *Environment) Write(payload []byte) (count int, err error) {
	if environment.writer == nil {
		return 0, io.ErrClosedPipe
	}

	return environment.writer.Write(payload)
}

func (environment *Environment) Close() error {
	environment.cancel()

	if environment.ctr.Err() != nil {
		return environment.ctr.Err()
	}

	return environment.ctr.Value().Terminate(environment.ctx)
}
