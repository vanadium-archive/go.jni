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

// JavaInputChannel creates a new Java InputChannel object given the provided Go recv function.
//
// All objects returned by the recv function must be globally references.
//
// The recv function must return verror.ErrEndOfFile when there are no more elements
// to receive.
func JavaInputChannel(env jutil.Env, recv func() (jutil.Object, error)) (jutil.Object, error) {
	jInputChannel, err := jutil.NewObject(env, jInputChannelImplClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&recv)))
	if err != nil {
		return jutil.NullObject, err
	}
	jutil.GoRef(&recv) // Un-refed when jInputChannel is finalized.
	return jInputChannel, nil
}

// JavaOutputChannel creates a new Java OutputChannel object given the provided Go convert, send
// and close functions. Send is invoked with the result of convert, which must be non-blocking.
func JavaOutputChannel(env jutil.Env, convert func(jutil.Object) (interface{}, error), send func(interface{}) error, close func() error) (jutil.Object, error) {
	jOutputChannel, err := jutil.NewObject(env, jOutputChannelImplClass, []jutil.Sign{jutil.LongSign, jutil.LongSign, jutil.LongSign}, int64(jutil.PtrValue(&convert)), int64(jutil.PtrValue(&send)), int64(jutil.PtrValue(&close)))
	if err != nil {
		return jutil.NullObject, err
	}
	jutil.GoRef(&convert) // Un-refed when jOutputChannel is finalized.
	jutil.GoRef(&send)    // Un-refed when jOutputChannel is finalized.
	jutil.GoRef(&close)   // Un-refed when jOutputChannel is finalized.
	return jOutputChannel, nil
}
