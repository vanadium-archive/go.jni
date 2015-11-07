// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package jni

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"runtime"
	"unsafe"

	"v.io/v23/logging"
)

// #include <android/log.h>
// #include <stdlib.h>
// #cgo LDFLAGS: -llog
import "C"

const (
	initialMaxStackBufSize = 128 * 1024
	maxStackBufSize = 4096 * 1024
)

var ctag *C.char = C.CString("GoLog")

func init() {
	log.SetOutput(&androidWriter{})
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	os.Stderr = w
	go lineLog(r, C.ANDROID_LOG_ERROR)

	r, w, err = os.Pipe()
	if err != nil {
		panic(err)
	}
	os.Stdout = w
	go lineLog(r, C.ANDROID_LOG_INFO)
}

func header(depth int) string {
	_, file, line, ok := runtime.Caller(3 + depth)
	if !ok {
		file = "???"
		line = 1
	}
	return fmt.Sprintf("%s:%d", file, line)
}

func alog(severity C.int, depth int, args ...interface{}) {
	writeLines(severity, header(depth) + " " + fmt.Sprint(args...))
}

func alogf(severity C.int, depth int, format string, args ...interface{}) {
	writeLines(severity, header(depth) + " " + fmt.Sprintf(format, args...))
}

func writeLines(severity C.int, msg string) {
	scanner := bufio.NewScanner(bytes.NewReader([]byte(msg)))
	for scanner.Scan() {
		write(severity, scanner.Text())
	}
}

func write(severity C.int, msg string) {
	cstr := C.CString(msg)
	C.__android_log_write(severity, ctag, cstr)
	C.free(unsafe.Pointer(cstr))
}

// NewLogger creates a new logger that uses Android logging functions.
func NewLogger(name string, level int) logging.Logger {
	return &logger{name, level}
}

type logger struct {
	name string
	level int
}

func (l *logger) String() string {
	return fmt.Sprintf("name=%s level=%d", l.name, l.level)
}

func (l *logger) Info(args ...interface{}) {
	alog(C.ANDROID_LOG_INFO, 0, args...)
}

func (l *logger) Infof(format string, args ...interface{}) {
	alogf(C.ANDROID_LOG_INFO, 0, format, args...)
}

func (l *logger) InfoDepth(depth int, args ...interface{}) {
	alog(C.ANDROID_LOG_INFO, depth, args...)
}

func (l *logger) InfoStack(all bool) {
	n := initialMaxStackBufSize
	var trace []byte
	for n <= maxStackBufSize {
		trace = make([]byte, n)
		nbytes := runtime.Stack(trace, all)
		if nbytes < len(trace) {
			alog(C.ANDROID_LOG_INFO, 0, string(trace[:nbytes]))
		}
		n *= 2
	}
	alog(C.ANDROID_LOG_INFO, 0, string(trace))
}

func (l *logger) Error(args ...interface{}) {
	alog(C.ANDROID_LOG_ERROR, 0, args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	alogf(C.ANDROID_LOG_ERROR, 0, format, args...)
}

func (l *logger) ErrorDepth(depth int, args ...interface{}) {
	alog(C.ANDROID_LOG_ERROR, depth, args...)
}

func (l *logger) Fatal(args ...interface{}) {
	alog(C.ANDROID_LOG_FATAL, 0, args...)
	l.InfoStack(true)
	os.Exit(255)
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	alogf(C.ANDROID_LOG_FATAL, 0, format, args...)
	l.InfoStack(true)
	os.Exit(255)
}

func (l *logger) FatalDepth(depth int, args ...interface{}) {
	alog(C.ANDROID_LOG_FATAL, depth, args...)
	l.InfoStack(true)
	os.Exit(255)
}

func (l *logger) Panic(args ...interface{}) {
	l.Error(args...)
	panic(fmt.Sprint(args...))
}

func (l *logger) Panicf(format string, args ...interface{}) {
	l.Errorf(format, args...)
	panic(fmt.Sprintf(format, args...))
}

func (l *logger) PanicDepth(depth int, args ...interface{}) {
	l.ErrorDepth(depth, args...)
	panic(fmt.Sprint(args...))
}

func (l *logger) V(v int) bool {
	return v <= l.level
}

func (l *logger) VDepth(_ int, v int) bool {
	return l.V(v)
}

type discardInfo struct{}

func (discardInfo) Info(...interface{})               {}
func (discardInfo) Infof(_ string, _ ...interface{})  {}
func (discardInfo) InfoDepth(_ int, _ ...interface{}) {}
func (discardInfo) InfoStack(_ bool)                  {}

func (l *logger) VI(v int) interface {
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	InfoDepth(depth int, args ...interface{})
	InfoStack(all bool)
} {
	if l.V(v) {
		return l
	}
	return discardInfo{}
}

func (l *logger) VIDepth(depth int, v int) interface {
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	InfoDepth(depth int, args ...interface{})
	InfoStack(all bool)
} {
	if l.VDepth(depth, v) {
		return l
	}
	return discardInfo{}
}

func (l *logger) FlushLog() {}

// androidWriter is used for diverting Go 'log' package output to android's INFO log.
type androidWriter struct {}

func (aw *androidWriter) Write(p []byte) (n int, err error) {
	write(C.ANDROID_LOG_INFO, string(p))
	return len(p), nil
}

// lineLog is used for diverting Go Stderr and Stdout to android's logs.
// NOTE(spetrovic): lifted from https://github.com/golang/mobile.
func lineLog(f *os.File, severity C.int) {
	const logSize = 1024 // matches android/log.h.
	r := bufio.NewReaderSize(f, logSize)
	for {
		line, _, err := r.ReadLine()
		str := string(line)
		if err != nil {
			str += " " + err.Error()
		}
		write(severity, str)
		if err != nil {
			break
		}
	}
}