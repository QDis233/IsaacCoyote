package isaac

type InvalidEventTypeError struct {
	Message string
}

func (e InvalidEventTypeError) Error() string {
	return e.Message
}

type InvalidCallbackError struct {
	Message string
}

func (e InvalidCallbackError) Error() string {
	return e.Message
}

type NoModDataError struct {
	Message string
}

func (e NoModDataError) Error() string {
	return e.Message
}

type TimeoutError struct {
	Message string
}

func (e TimeoutError) Error() string {
	return e.Message
}
