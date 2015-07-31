// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package i18n

import (
	"unsafe"

	"v.io/v23/i18n"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

// Init initializes the JNI code with the given Java environment.  This method
// must be invoked before any other method in this package and must be called
// from the main Java thread (e.g., On_Load()).
// interface and then cast into the package-local environment type.
func Init(env jutil.Env) error {
	return nil
}

//export Java_io_v_v23_i18n_Catalog_nativeFormatParams
func Java_io_v_v23_i18n_Catalog_nativeFormatParams(jenv *C.JNIEnv, jCatalog C.jclass, jFormat C.jstring, jParams C.jobjectArray) C.jstring {
	env := jutil.WrapEnv(jenv)
	format := jutil.GoString(env, jutil.WrapObject(jFormat))
	strParams, err := jutil.GoStringArray(env, jutil.WrapObject(jParams))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	params := make([]interface{}, len(strParams))
	for i, strParam := range strParams {
		params[i] = strParam
	}
	result := i18n.FormatParams(format, params...)
	jRet := jutil.JString(env, result)
	return C.jstring(unsafe.Pointer(jRet))
}
