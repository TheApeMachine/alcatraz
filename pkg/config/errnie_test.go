package config

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewErrnieConfig(test *testing.T) {
	Convey("Given default errnie settings", test, func() {
		Convey("It should return the default log level", func() {
			errnieConfig := NewErrnieConfig()

			So(errnieConfig.Level, ShouldEqual, "info")
		})

		Convey("It should map into the errnie library config", func() {
			errnieConfig := NewErrnieConfig()
			libraryConfig := errnieConfig.ToLibraryConfig()

			So(libraryConfig.Level, ShouldEqual, "info")
			So(libraryConfig.File.Active, ShouldBeFalse)
			So(libraryConfig.Elasticsearch.Active, ShouldBeFalse)
		})
	})
}

func BenchmarkNewErrnieConfig(benchmark *testing.B) {
	for benchmark.Loop() {
		_ = NewErrnieConfig()
	}
}
