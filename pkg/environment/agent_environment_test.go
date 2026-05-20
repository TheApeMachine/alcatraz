package environment

import (
	"context"
	"os/exec"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/theapemachine/errnie"

	"github.com/theapemachine/alcatraz/pkg/config"
)

func dockerAvailable() bool {
	return exec.Command("docker", "info").Run() == nil
}

func TestAgentEnvironmentStart(test *testing.T) {
	if !dockerAvailable() {
		test.Skip("docker unavailable")
	}

	restoreLogging := errnie.SuppressLogging()
	defer restoreLogging()

	Convey("Given a new agent environment", test, func() {
		globalConfig := config.NewEnvironmentConfig()
		agentEnvironment := NewAgentEnvironment(
			context.Background(),
			AgentEnvironmentConfigFromGlobal("test-id", "test", globalConfig),
		)

		Convey("It should start and run a command", func() {
			defer agentEnvironment.Close()

			err := agentEnvironment.Start()
			So(err, ShouldBeNil)
			So(agentEnvironment.Running(), ShouldBeTrue)
			So(agentEnvironment.ContainerID(), ShouldNotBeEmpty)

			output, err := agentEnvironment.Exec("echo hello")
			So(err, ShouldBeNil)
			So(output, ShouldContainSubstring, "hello")
		})
	})
}

func TestAgentEnvironmentExecNotRunning(test *testing.T) {
	restoreLogging := errnie.SuppressLogging()
	defer restoreLogging()

	Convey("Given a stopped environment", test, func() {
		agentEnvironment := NewAgentEnvironment(
			context.Background(),
			AgentEnvironmentConfig{ID: "stopped"},
		)

		Convey("It should reject exec", func() {
			_, err := agentEnvironment.Exec("echo hello")

			So(err, ShouldNotBeNil)
			So(errnie.IsValidation(err), ShouldBeTrue)
		})
	})
}

func TestAgentEnvironmentTerminate(test *testing.T) {
	if !dockerAvailable() {
		test.Skip("docker unavailable")
	}

	restoreLogging := errnie.SuppressLogging()
	defer restoreLogging()

	Convey("Given a running environment", test, func() {
		globalConfig := config.NewEnvironmentConfig()
		agentEnvironment := NewAgentEnvironment(
			context.Background(),
			AgentEnvironmentConfigFromGlobal("terminate-id", "terminate", globalConfig),
		)

		Convey("It should terminate cleanly", func() {
			err := agentEnvironment.Start()
			So(err, ShouldBeNil)

			err = agentEnvironment.Terminate()
			So(err, ShouldBeNil)
			So(agentEnvironment.Running(), ShouldBeFalse)
			So(agentEnvironment.ContainerID(), ShouldBeEmpty)
		})
	})
}

func BenchmarkAgentEnvironmentExecNotRunning(benchmark *testing.B) {
	agentEnvironment := NewAgentEnvironment(
		context.Background(),
		AgentEnvironmentConfig{ID: "stopped"},
	)

	for benchmark.Loop() {
		_, _ = agentEnvironment.Exec("echo hello")
	}
}
