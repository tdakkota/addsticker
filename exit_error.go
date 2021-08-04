package main

type ExitError struct {
	Msg  string
	Code int
}

func (e *ExitError) Error() string {
	return e.Msg
}
