package alcatraz

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/theapemachine/alcatraz/pkg/config"
)

func TestNewEnvironment(test *testing.T) {
	Convey("Given an environment identifier", test, func() {
		environment := NewEnvironment(context.Background(), "dev-1")

		Convey("It should apply configured defaults", func() {
			So(environment.ID, ShouldEqual, "dev-1")
			So(environment.Name, ShouldEqual, "dev-1")
			So(environment.Image, ShouldEqual, config.Environment.Image)
			So(environment.Command, ShouldEqual, "/bin/bash")
			So(environment.Args, ShouldBeEmpty)
		})
	})
}

func TestEnvironmentRead(test *testing.T) {
	Convey("Given an environment without an attached reader", test, func() {
		environment := NewEnvironment(context.Background(), "read-id")

		Convey("It should return EOF", func() {
			payload := make([]byte, 8)

			count, err := environment.Read(payload)

			So(count, ShouldEqual, 0)
			So(err, ShouldEqual, io.EOF)
		})
	})

	Convey("Given an environment with an attached reader", test, func() {
		environment := NewEnvironment(context.Background(), "read-id")
		environment.reader = strings.NewReader("hello")

		Convey("It should read from the stream", func() {
			payload := make([]byte, 5)

			count, err := environment.Read(payload)

			So(count, ShouldEqual, 5)
			So(err, ShouldBeNil)
			So(string(payload), ShouldEqual, "hello")
		})
	})
}

func TestEnvironmentWrite(test *testing.T) {
	Convey("Given an environment without an attached writer", test, func() {
		environment := NewEnvironment(context.Background(), "write-id")

		Convey("It should return ErrClosedPipe", func() {
			count, err := environment.Write([]byte("hello"))

			So(count, ShouldEqual, 0)
			So(err, ShouldEqual, io.ErrClosedPipe)
		})
	})

	Convey("Given an environment with an attached writer", test, func() {
		environment := NewEnvironment(context.Background(), "write-id")
		buffer := &bytes.Buffer{}
		environment.writer = buffer

		Convey("It should write to the stream", func() {
			count, err := environment.Write([]byte("hello"))

			So(count, ShouldEqual, 5)
			So(err, ShouldBeNil)
			So(buffer.String(), ShouldEqual, "hello")
		})
	})
}

func BenchmarkEnvironmentRead(benchmark *testing.B) {
	environment := NewEnvironment(context.Background(), "read-bench")
	environment.reader = strings.NewReader("hello")
	payload := make([]byte, 5)

	for benchmark.Loop() {
		environment.reader = strings.NewReader("hello")
		_, _ = environment.Read(payload)
	}
}

func BenchmarkEnvironmentWrite(benchmark *testing.B) {
	environment := NewEnvironment(context.Background(), "write-bench")
	buffer := &bytes.Buffer{}
	environment.writer = buffer
	payload := []byte("hello")

	for benchmark.Loop() {
		buffer.Reset()
		_, _ = environment.Write(payload)
	}
}
