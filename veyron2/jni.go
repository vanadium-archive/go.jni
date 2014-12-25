// +build android

package veyron2

import (
	jandroid "v.io/jni/veyron2/android"
	jcontext "v.io/jni/veyron2/context"
	jsecurity "v.io/jni/veyron2/security"
	jaccess "v.io/jni/veyron2/services/security/access"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// Init initializes the JNI code with the given Java environment.  This method
// must be invoked before any other method in this package and must be called
// from the main Java thread (e.g., On_Load()).
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java environment is passed in an empty
// interface and then cast into the package-local environment type.
func Init(jEnv interface{}) {
	jcontext.Init(jEnv)
	jsecurity.Init(jEnv)
	jandroid.Init(jEnv)
	jaccess.Init(jEnv)
}
