package environment

import (
	"context"
	"fmt"
	"io"

	sdkcontainer "github.com/docker/go-sdk/container"
	"github.com/docker/go-sdk/container/exec"
	"github.com/theapemachine/errnie"
)

/*
ExecResult holds the output and exit status of a command run inside an environment.
*/
type ExecResult struct {
	Output   string
	ExitCode int
}

func readExecOutput(reader io.Reader) (string, error) {
	if reader == nil {
		return "", nil
	}

	readResult := errnie.Does(func() (string, error) {
		payload, readErr := io.ReadAll(reader)

		return string(payload), readErr
	})

	return readResult.Value(), wrapError(
		errnie.IO,
		"environment.exec.read",
		"read exec output",
		readResult.Err(),
	)
}

func (agentEnvironment *AgentEnvironment) execCommand(
	execCtx context.Context,
	command string,
) (ExecResult, error) {
	timeout := agentEnvironment.execTimeout()
	timeoutSeconds := fmt.Sprintf("%ds", int(timeout.Seconds()))

	execResult := errnie.Does(func() (ExecResult, error) {
		exitCode, reader, execErr := agentEnvironment.container.Exec(
			execCtx,
			[]string{
				"timeout",
				"--kill-after=5s",
				timeoutSeconds,
				"/bin/sh",
				"-c",
				command,
			},
			exec.WithWorkingDir(agentEnvironment.workingDir()),
			exec.Multiplexed(),
		)

		output, readErr := readExecOutput(reader)

		if execErr != nil {
			return ExecResult{Output: output}, execErr
		}

		if readErr != nil {
			return ExecResult{Output: output, ExitCode: exitCode}, readErr
		}

		result := ExecResult{
			Output:   output,
			ExitCode: exitCode,
		}

		if exitCode == 124 {
			return result, errnie.E(
				errnie.Timeout,
				fmt.Sprintf(
					"exec timed out after %s: %s",
					timeout,
					truncate(result.Output, 400),
				),
				nil,
			).Operation("environment.exec").With("environment_id", agentEnvironment.cfg.ID)
		}

		if exitCode != 0 {
			return result, errnie.E(
				errnie.Validation,
				fmt.Sprintf(
					"command exited %d: %s",
					exitCode,
					truncate(result.Output, 400),
				),
				nil,
			).Operation("environment.exec").With(
				"environment_id", agentEnvironment.cfg.ID,
				"exit_code", exitCode,
			)
		}

		return result, nil
	})

	return execResult.Value(), wrapError(
		errnie.Unknown,
		"environment.exec",
		"run command",
		execResult.Err(),
		"environment_id", agentEnvironment.cfg.ID,
	)
}

func combinedOutput(stdout string, stderr string) string {
	if stdout == "" {
		return stderr
	}

	if stderr == "" {
		return stdout
	}

	return stdout + "\n" + stderr
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}

	return value[:limit] + "…"
}

func buildContainerOptions(agentEnvironment *AgentEnvironment) []sdkcontainer.ContainerCustomizer {
	return []sdkcontainer.ContainerCustomizer{
		sdkcontainer.WithImage(agentEnvironment.cfg.Image),
		sdkcontainer.WithName(agentEnvironment.containerName()),
		sdkcontainer.WithAlwaysPull(),
		sdkcontainer.WithCmd("/bin/sh", "-c", "tail -f /dev/null"),
		sdkcontainer.WithHostConfigModifier(agentEnvironment.hostConfigModifier()),
		sdkcontainer.WithLabels(agentEnvironment.containerLabels()),
	}
}
