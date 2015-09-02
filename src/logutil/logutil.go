package logutil

import (
	"bytes"
	"fmt"
	"io"
	"runtime"
	"unicode/utf8"

	"appengine"
	"appengine_internal"
)

const (
	apiErrorStringBase  = "Code: %d, Detail: %s, Service: %s"
	callErrorStringBase = "Code: %d, Detail: %s, Timeout: %t"
	unknownErrorString  = "Unknown Error"
)

func sprintf(format string, args ...interface{}) string {
	_, file, line, _ := runtime.Caller(2)
	var fileLine = fmt.Sprintf("%s:%d]", file, line)
	return fmt.Sprintf("%s %s", fileLine, fmt.Sprintf(format, args...))
}

func Debugf(c appengine.Context, format string, args ...interface{}) {
	c.Debugf(sprintf(format, args...))
}

func Infof(c appengine.Context, format string, args ...interface{}) {
	c.Infof(sprintf(format, args...))
}

func Warningf(c appengine.Context, format string, args ...interface{}) {
	c.Warningf(sprintf(format, args...))
}

func Errorf(c appengine.Context, format string, args ...interface{}) {
	c.Errorf(sprintf(format, args...))
}

func Criticalf(c appengine.Context, format string, args ...interface{}) {
	c.Criticalf(sprintf(format, args...))
}

func ErrorStackTracef(c appengine.Context, format string, args ...interface{}) {
	c.Errorf(sprintf("%s", StackTrace(2, 2048)))
	c.Errorf(sprintf(format, args...))
}

//https://github.com/knightso/base/blob/master/errors/errors.go
func StackTrace(skip, maxBytes int) []byte {
	// this func is debug purpose and ignores errors

	buf := make([]byte, maxBytes)
	n := runtime.Stack(buf, false)
	var gotall bool
	if n < len(buf) {
		buf = buf[:n]
		gotall = true
	} else {
		for !utf8.Valid(buf) || len(buf) == 0 {
			buf = buf[:len(buf)-1]
		}
	}

	var w bytes.Buffer

	writeOrSkip := func(buf []byte, w io.Writer, line int) {
		if line == 1 || line > 1+skip*2 {
			w.Write(buf)
		}
	}

	line := 1
	for {
		lf := bytes.IndexByte(buf, '\n')
		if lf < 0 {
			writeOrSkip(buf, &w, line)
			break
		}
		writeOrSkip(buf[:lf+1], &w, line)
		buf = buf[lf+1:]
		line++
	}

	if !gotall {
		w.WriteString("\n        ... (omitted)")
	}
	w.WriteString("\n")

	return w.Bytes()
}

func AppengineErrorToString(err error) string {
	apiErr, ok := err.(*appengine_internal.APIError)
	if ok {
		return fmt.Sprintf(apiErrorStringBase, apiErr.Code, apiErr.Detail, apiErr.Service)
	}
	callErr, ok := err.(*appengine_internal.CallError)
	if ok {
		return fmt.Sprintf(callErrorStringBase, callErr.Code, callErr.Detail, callErr.Timeout)
	}
	return unknownErrorString
}
