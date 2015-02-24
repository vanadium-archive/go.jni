// +build android

package google

import (
	jchannel "v.io/jni/core/veyron/runtimes/google/channel"
	jipc "v.io/jni/core/veyron/runtimes/google/ipc"
	jnaming "v.io/jni/core/veyron/runtimes/google/naming"
	jrt "v.io/jni/core/veyron/runtimes/google/rt"
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
	jipc.Init(jEnv)
	jrt.Init(jEnv)
	jchannel.Init(jEnv)
	jnaming.Init(jEnv)
}
