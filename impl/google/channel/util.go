// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package channel

import (
	"unsafe"

	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

// JavaInputChannel converts the provided Go channel of C.jobject values into a Java
// InputChannel object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaInputChannel(jEnv, chPtr, sourceChanPtr interface{}) (unsafe.Pointer, error) {
	jInputChannel, err := jutil.NewObject(jEnv, jInputChannelImplClass, []jutil.Sign{jutil.LongSign, jutil.LongSign}, int64(jutil.PtrValue(chPtr)), int64(jutil.PtrValue(sourceChanPtr)))
	if err != nil {
		return nil, err
	}
	jutil.GoRef(chPtr)         // Un-refed when the InputChannel is finalized.
	jutil.GoRef(sourceChanPtr) // Un-refed when the InputChannel is finalized.
	return jInputChannel, nil
}

// outputChannel represents the Go-side of a Java OutputChannel. Each time the
// Java side writes an object, the ReadFunc will be called. If the ReadFunc
// returns an error, an exception will be raised on the Java side.
type outputChannel struct {
	// ReadFunc is invoked when a C.jobject has been read from Java. The
	// input to the function is guaranteed to be of type C.jobject; it is
	// represented as an interface{} only because C.jobject is not
	// exported.
	//
	// The C.jobject passed to this function is globally referenced. It is
	// required that the ReadFunc implementation delete the global
	// reference via v.io/x/jni/util.DeleteGlobalRef. Failure to do so will
	// result in a reference leak.
	//
	// If the ReadFunc implementation returns an error, that error will be
	// passed back to the Java writer in the form of a VException to the
	// OutputChannel.writeValue() method.
	ReadFunc func(interface{}) error

	// CloseFunc is invoked when the Java side has closed its OutputChannel
	// and no more values will be written.
	//
	// If CloseFunc returns an error, that error will be passed back to the
	// Java side in the form of a VException OutputChannel.close() method.
	CloseFunc func() error
}

// JavaOutputChannel converts the provided Go channel of C.jobject values into
// a Java OutputChannel object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaOutputChannel(jEnv interface{}, readFunc func(interface{}) error, closeFunc func() error) (unsafe.Pointer, error) {
	outCh := outputChannel{
		ReadFunc:  readFunc,
		CloseFunc: closeFunc,
	}

	jOutputChannel, err := jutil.NewObject(jEnv, jOutputChannelImplClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&outCh)))
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&outCh) // Un-refed when the OutputChannel is finalized.
	return jOutputChannel, nil
}
