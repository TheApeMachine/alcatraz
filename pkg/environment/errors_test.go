package environment

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/theapemachine/errnie"
)

func TestDomainError(test *testing.T) {
	Convey("Given a domain error", test, func() {
		restoreLogging := errnie.SuppressLogging()
		defer restoreLogging()

		Convey("It should preserve the errnie kind", func() {
			err := domainError(
				errnie.NotFound,
				"registry.get",
				"environment missing",
				"environment_id", "env-1",
			)

			So(err, ShouldNotBeNil)
			So(errnie.IsNotFound(err), ShouldBeTrue)
		})
	})
}

func TestWrapError(test *testing.T) {
	Convey("Given a nil cause", test, func() {
		Convey("It should return nil", func() {
			So(wrapError(errnie.IO, "environment.stop", "stop container", nil), ShouldBeNil)
		})
	})
}

func BenchmarkDomainError(benchmark *testing.B) {
	restoreLogging := errnie.SuppressLogging()
	defer restoreLogging()

	for benchmark.Loop() {
		_ = domainError(
			errnie.IO,
			"environment.start",
			"start container",
			"environment_id", "env-1",
		)
	}
}

func BenchmarkWrapError(benchmark *testing.B) {
	restoreLogging := errnie.SuppressLogging()
	defer restoreLogging()

	for benchmark.Loop() {
		_ = wrapError(
			errnie.IO,
			"environment.start",
			"start container",
			errnie.E(errnie.IO, "docker unavailable", nil),
			"environment_id", "env-1",
		)
	}
}
