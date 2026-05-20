package environment

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/theapemachine/errnie"

	"github.com/theapemachine/alcatraz/pkg/config"
)

func TestRegistryCreate(test *testing.T) {
	restoreLogging := errnie.SuppressLogging()
	defer restoreLogging()

	Convey("Given a registry with a max environment limit", test, func() {
		globalConfig := &config.EnvironmentConfig{
			Image:           config.EnvironmentDefaultImage,
			ExecTimeout:     config.EnvironmentDefaultExecTimeout,
			MemoryBytes:     config.EnvironmentDefaultMemoryBytes,
			NanoCPUs:        config.EnvironmentDefaultNanoCPUs,
			PIDsLimit:       config.EnvironmentDefaultPIDsLimit,
			WorkingDir:      config.EnvironmentDefaultWorkingDir,
			NetworkMode:     config.EnvironmentDefaultNetworkMode,
			MaxEnvironments: 1,
		}

		registry := NewRegistry(context.Background(), globalConfig)
		registry.environments["existing"] = NewAgentEnvironment(
			context.Background(),
			AgentEnvironmentConfig{ID: "existing", Name: "busy"},
		)

		Convey("It should reject creation when the limit is reached", func() {
			_, err := registry.Create("another")

			So(err, ShouldNotBeNil)
			So(errnie.IsConflict(err), ShouldBeTrue)
		})

		Convey("It should reject duplicate names", func() {
			registry.globalCfg.MaxEnvironments = 4

			_, err := registry.Create("busy")

			So(err, ShouldNotBeNil)
			So(errnie.IsConflict(err), ShouldBeTrue)
		})
	})
}

func TestRegistryGet(test *testing.T) {
	restoreLogging := errnie.SuppressLogging()
	defer restoreLogging()

	Convey("Given a registry with one environment", test, func() {
		registry := NewRegistry(context.Background(), config.NewEnvironmentConfig())
		agentEnvironment := NewAgentEnvironment(
			context.Background(),
			AgentEnvironmentConfig{ID: "env-1", Name: "dev"},
		)
		registry.environments["env-1"] = agentEnvironment

		Convey("It should return the environment by ID", func() {
			found, err := registry.Get("env-1")

			So(err, ShouldBeNil)
			So(found, ShouldEqual, agentEnvironment)
		})

		Convey("It should error for unknown IDs", func() {
			_, err := registry.Get("missing")

			So(err, ShouldNotBeNil)
			So(errnie.IsNotFound(err), ShouldBeTrue)
		})
	})
}

func TestRegistryList(test *testing.T) {
	Convey("Given a registry with tracked environments", test, func() {
		registry := NewRegistry(context.Background(), config.NewEnvironmentConfig())
		first := NewAgentEnvironment(context.Background(), AgentEnvironmentConfig{ID: "one"})
		second := NewAgentEnvironment(context.Background(), AgentEnvironmentConfig{ID: "two"})
		registry.environments["one"] = first
		registry.environments["two"] = second

		Convey("It should return all environments", func() {
			environments := registry.List()

			So(len(environments), ShouldEqual, 2)
		})
	})
}

func TestRegistryDestroy(test *testing.T) {
	Convey("Given a stopped environment in the registry", test, func() {
		registry := NewRegistry(context.Background(), config.NewEnvironmentConfig())
		agentEnvironment := NewAgentEnvironment(
			context.Background(),
			AgentEnvironmentConfig{ID: "env-1", Name: "dev"},
		)
		registry.environments["env-1"] = agentEnvironment

		Convey("It should remove the environment after terminate", func() {
			err := registry.Destroy("env-1")

			So(err, ShouldBeNil)
			So(registry.Count(), ShouldEqual, 0)
		})
	})
}

func BenchmarkRegistryList(benchmark *testing.B) {
	registry := NewRegistry(context.Background(), config.NewEnvironmentConfig())

	for index := range 32 {
		environmentID := "env-" + string(rune('a'+index))
		registry.environments[environmentID] = NewAgentEnvironment(
			context.Background(),
			AgentEnvironmentConfig{ID: environmentID},
		)
	}

	for benchmark.Loop() {
		_ = registry.List()
	}
}
