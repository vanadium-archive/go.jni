// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package naming

import (
	"v.io/v23/naming"

	jutil "v.io/x/jni/util"

	"unsafe"
)

// #include "jni.h"
import "C"

// JavaEndpoint converts the provided Go Endpoint to a Java Endpoint.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaEndpoint(jEnv interface{}, endpoint naming.Endpoint) (unsafe.Pointer, error) {
	return jutil.CallStaticObjectMethod(jEnv, jEndpointImplClass, "fromString", []jutil.Sign{jutil.StringSign}, endpointSign, endpoint.String())
}
