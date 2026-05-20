package config

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewEnvironmentConfig(test *testing.T) {
	Convey("Given default environment settings", test, func() {
		Convey("It should return hardened defaults", func() {
			environmentConfig := NewEnvironmentConfig()

			So(environmentConfig.Image, ShouldEqual, EnvironmentDefaultImage)
			So(environmentConfig.ExecTimeout, ShouldEqual, EnvironmentDefaultExecTimeout)
			So(environmentConfig.MemoryBytes, ShouldEqual, int64(EnvironmentDefaultMemoryBytes))
			So(environmentConfig.NanoCPUs, ShouldEqual, int64(EnvironmentDefaultNanoCPUs))
			So(environmentConfig.PIDsLimit, ShouldEqual, int64(EnvironmentDefaultPIDsLimit))
			So(environmentConfig.WorkingDir, ShouldEqual, EnvironmentDefaultWorkingDir)
			So(environmentConfig.NetworkMode, ShouldEqual, EnvironmentDefaultNetworkMode)
			So(environmentConfig.MaxEnvironments, ShouldEqual, EnvironmentDefaultMaxConcurrent)
		})
	})
}

func TestWithDefault(test *testing.T) {
	Convey("Given a typed default value", test, func() {
		Convey("It should return the default when the key is unset", func() {
			So(WithDefault("missing.key", 42), ShouldEqual, 42)
			So(WithDefault("missing.key", "fallback"), ShouldEqual, "fallback")
			So(WithDefault("missing.key", 10*time.Second), ShouldEqual, 10*time.Second)
		})
	})
}

func BenchmarkNewEnvironmentConfig(benchmark *testing.B) {
	for benchmark.Loop() {
		_ = NewEnvironmentConfig()
	}
}

func BenchmarkWithDefault(benchmark *testing.B) {
	for benchmark.Loop() {
		_ = WithDefault("missing.key", "fallback")
	}
}
