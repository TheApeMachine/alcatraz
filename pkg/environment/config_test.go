package environment

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/theapemachine/alcatraz/pkg/config"
)

func TestAgentEnvironmentDefaults(test *testing.T) {
	Convey("Given an environment without explicit resource settings", test, func() {
		agentEnvironment := &AgentEnvironment{}

		Convey("It should apply bounded defaults", func() {
			So(agentEnvironment.execTimeout(), ShouldEqual, config.EnvironmentDefaultExecTimeout)
			So(agentEnvironment.memoryBytes(), ShouldEqual, int64(config.EnvironmentDefaultMemoryBytes))
			So(agentEnvironment.nanoCPUs(), ShouldEqual, int64(config.EnvironmentDefaultNanoCPUs))
			So(agentEnvironment.pidsLimit(), ShouldEqual, int64(config.EnvironmentDefaultPIDsLimit))
			So(agentEnvironment.workingDir(), ShouldEqual, config.EnvironmentDefaultWorkingDir)
			So(string(agentEnvironment.networkMode()), ShouldEqual, config.EnvironmentDefaultNetworkMode)
		})

		Convey("It should build a restricted Docker host config", func() {
			hostConfig := agentEnvironment.hostConfig()

			So(hostConfig.ReadonlyRootfs, ShouldBeTrue)
			So(string(hostConfig.NetworkMode), ShouldEqual, config.EnvironmentDefaultNetworkMode)
			So(hostConfig.Memory, ShouldEqual, int64(config.EnvironmentDefaultMemoryBytes))
			So(hostConfig.NanoCPUs, ShouldEqual, int64(config.EnvironmentDefaultNanoCPUs))
			So(*hostConfig.PidsLimit, ShouldEqual, int64(config.EnvironmentDefaultPIDsLimit))
			So(hostConfig.CapDrop, ShouldContain, "ALL")
			So(hostConfig.Tmpfs, ShouldContainKey, "/workspace")
		})
	})

	Convey("Given an environment with explicit resource settings", test, func() {
		agentEnvironment := &AgentEnvironment{cfg: AgentEnvironmentConfig{
			ExecTimeout: 30 * time.Second,
			MemoryBytes: 512 << 20,
			NanoCPUs:    500_000_000,
			PIDsLimit:   64,
			WorkingDir:  "/tmp/work",
			NetworkMode: "bridge",
			Name:        "dev",
			ID:          "abc123",
		}}

		Convey("It should use the configured bounds", func() {
			So(agentEnvironment.execTimeout(), ShouldEqual, 30*time.Second)
			So(agentEnvironment.memoryBytes(), ShouldEqual, int64(512<<20))
			So(agentEnvironment.nanoCPUs(), ShouldEqual, int64(500_000_000))
			So(agentEnvironment.pidsLimit(), ShouldEqual, int64(64))
			So(agentEnvironment.workingDir(), ShouldEqual, "/tmp/work")
			So(string(agentEnvironment.networkMode()), ShouldEqual, "bridge")
			So(agentEnvironment.containerName(), ShouldEqual, "alcatraz-dev")
			So(agentEnvironment.containerLabels()[agentEnvironmentLabelID], ShouldEqual, "abc123")
		})
	})
}

func TestCombinedOutput(test *testing.T) {
	Convey("Given stdout and stderr content", test, func() {
		Convey("It should preserve both streams", func() {
			output := combinedOutput("stdout", "stderr")

			So(output, ShouldEqual, "stdout\nstderr")
		})

		Convey("It should return one stream without an extra separator", func() {
			So(combinedOutput("stdout", ""), ShouldEqual, "stdout")
			So(combinedOutput("", "stderr"), ShouldEqual, "stderr")
		})
	})
}

func BenchmarkAgentEnvironmentHostConfig(benchmark *testing.B) {
	agentEnvironment := &AgentEnvironment{}

	for benchmark.Loop() {
		_ = agentEnvironment.hostConfig()
	}
}
