// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package jni

import (
	"bufio"
	"log"
	"os"
	"unsafe"
)

// #include <android/log.h>
// #include <stdlib.h>
// #cgo LDFLAGS: -llog
import "C"

var ctag *C.char = C.CString("GoLog")

func init() {
	log.SetOutput(&androidWriter{})
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	os.Stderr = w
	go lineLog(r, "STDERR", C.ANDROID_LOG_ERROR)

	r, w, err = os.Pipe()
	if err != nil {
		panic(err)
	}
	os.Stdout = w
	go lineLog(r, "STDOUT", C.ANDROID_LOG_INFO)
}

// androidWriter is used for diverting Go 'log' package output to android's INFO log.
type androidWriter struct{}

func (aw *androidWriter) Write(p []byte) (n int, err error) {
	cstr := C.CString(string(p))
	C.__android_log_write(C.ANDROID_LOG_INFO, ctag, cstr)
	C.free(unsafe.Pointer(cstr))
	return len(p), nil
}

// lineLog is used for diverting Go Stderr and Stdout to android's logs.
// NOTE(spetrovic): lifted from https://github.com/golang/mobile.
func lineLog(f *os.File, tag string, severity C.int) {
	ctag := C.CString(tag)
	defer C.free(unsafe.Pointer(ctag))
	const logSize = 1024 // matches android/log.h.
	r := bufio.NewReaderSize(f, logSize)
	for {
		line, _, err := r.ReadLine()
		str := string(line)
		if err != nil {
			str += " " + err.Error()
		}
		cstr := C.CString(str)
		C.__android_log_write(severity, ctag, cstr)
		C.free(unsafe.Pointer(cstr))
		if err != nil {
			break
		}
	}
}
