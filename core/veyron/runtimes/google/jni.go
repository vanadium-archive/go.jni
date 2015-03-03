// +build android

package google

import (
	jchannel "v.io/x/jni/core/veyron/runtimes/google/channel"
	jipc "v.io/x/jni/core/veyron/runtimes/google/ipc"
	jnaming "v.io/x/jni/core/veyron/runtimes/google/naming"
	jrt "v.io/x/jni/core/veyron/runtimes/google/rt"
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
func Init(jEnv interface{}) error {
	if err := jipc.Init(jEnv); err != nil {
		return err
	}
	if err := jrt.Init(jEnv); err != nil {
		return err
	}
	if err := jchannel.Init(jEnv); err != nil {
		return err
	}
	if err := jnaming.Init(jEnv); err != nil {
		return err
	}
	return nil
}
