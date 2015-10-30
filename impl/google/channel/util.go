// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package channel

import (
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

// JavaIterable converts the provided Go channel into a Java VIterable object, using the
// given convert function.
func JavaIterable(env jutil.Env, ch chan interface{}, convert func(jutil.Env, interface{}) (jutil.Object, error)) (jutil.Object, error) {
	jIterable, err := jutil.NewObject(env, jChannelIterableClass, []jutil.Sign{jutil.LongSign, jutil.LongSign}, int64(jutil.PtrValue(&ch)), int64(jutil.PtrValue(&convert)))
	if err != nil {
		return jutil.NullObject, err
	}
	jutil.GoRef(&ch)      // Un-refed when ChannelIterable is finalized.
	jutil.GoRef(&convert) // Un-refed when ChannelIterable is finalized.
	return jIterable, nil
}

// outputChannel represents the Go-side of a Java OutputChannel. Each time the
// Java side writes an object, the ReadFunc will be called. If the ReadFunc
// returns an error, an exception will be raised on the Java side.
type outputChannel struct {
	// ReadFunc is invoked when an object has been read from Java.
	//
	// The jutil.Object passed to this function is globally referenced. It is
	// required that the ReadFunc implementation delete the global
	// reference jutil.DeleteGlobalRef. Failure to do so will
	// result in a reference leak.
	//
	// If the ReadFunc implementation returns an error, that error will be
	// passed back to the Java writer in the form of a VException to the
	// OutputChannel.writeValue() method.
	ReadFunc func(jutil.Object) error

	// CloseFunc is invoked when the Java side has closed its OutputChannel
	// and no more values will be written.
	//
	// If CloseFunc returns an error, that error will be passed back to the
	// Java side in the form of a VException OutputChannel.close() method.
	CloseFunc func() error
}

// JavaOutputChannel converts the provided Go channel of jutil.Object values
// into a Java OutputChannel object.
func JavaOutputChannel(env jutil.Env, readFunc func(jutil.Object) error, closeFunc func() error) (jutil.Object, error) {
	outCh := outputChannel{
		ReadFunc:  readFunc,
		CloseFunc: closeFunc,
	}

	jOutputChannel, err := jutil.NewObject(env, jOutputChannelImplClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&outCh)))
	if err != nil {
		return jutil.NullObject, err
	}
	jutil.GoRef(&outCh) // Un-refed when the OutputChannel is finalized.
	return jOutputChannel, nil
}
