package environment

import "github.com/theapemachine/errnie"

func domainError(
	kind errnie.Kind,
	operation string,
	message string,
	fields ...any,
) error {
	typedError := errnie.E(kind, message, nil).Operation(operation)

	if len(fields) > 0 {
		typedError = typedError.With(fields...)
	}

	return errnie.Error(typedError)
}

func wrapError(
	kind errnie.Kind,
	operation string,
	message string,
	cause error,
	fields ...any,
) error {
	if cause == nil {
		return nil
	}

	typedError := errnie.E(kind, message, cause).Operation(operation)

	if len(fields) > 0 {
		typedError = typedError.With(fields...)
	}

	return errnie.Error(typedError)
}
