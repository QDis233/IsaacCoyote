package coyote

type InvalidMessageError struct {
	Message string
}

func (e InvalidMessageError) Error() string {
	return e.Message
}

type AlreadyRunningError struct {
	Message string
}

func (e AlreadyRunningError) Error() string {
	return e.Message
}

type SessionNotFoundError struct {
	Message string
}

func (e SessionNotFoundError) Error() string {
	return e.Message
}

type TooLongMessageError struct {
	Message string
}

func (e TooLongMessageError) Error() string {
	return e.Message
}

type InvalidPulseParamError struct {
	Message string
}

func (e InvalidPulseParamError) Error() string {
	return e.Message
}

type TooLongPulseError struct {
	Message string
}

func (e TooLongPulseError) Error() string {
	return e.Message
}

type NotBindError struct {
	Message string
}

func (e NotBindError) Error() string {
	return e.Message
}

type NoWSSessionError struct {
	Message string
}

func (e NoWSSessionError) Error() string {
	return e.Message
}

type ControllerError struct {
	Code    int
	Message string
}

func (e ControllerError) Error() string {
	return e.Message
}
