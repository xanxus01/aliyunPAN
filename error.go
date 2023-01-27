package aliyunDriver

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
)

type httpError struct {
	s string
	httpCode int
}

func (e httpError)Error() string {
	return e.s
}

func (e httpError)StateCode() int {
	return e.httpCode
}

func newHttpError(r *http.Response) *httpError {
	return &httpError{
		s: r.Status,
		httpCode: r.StatusCode,
	}
}

func addErrorHttpError(e error, r *http.Response) httpError {
	return httpError{
		s: e.Error(),
		httpCode: r.StatusCode,
	}
}

func AddError(e error, s string) error {
	e = newError(e.Error() + "\n" + s)
	return e
}

func newError(s string) error {
	_, file, lineNo, ok := runtime.Caller(1)
	errorString := s
	if ok {
		errorString = fmt.Sprintf("%s:%d\n", file, lineNo) + s
	}

	return errors.New(errorString)
}
