// +build android

package veyron2

import (
	jandroid "v.io/jni/v23/android"
	jcontext "v.io/jni/v23/context"
	ji18n "v.io/jni/v23/i18n"
	jsecurity "v.io/jni/v23/security"
	jaccess "v.io/jni/v23/services/security/access"
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
	if err := jcontext.Init(jEnv); err != nil {
		return err
	}
	if err := ji18n.Init(jEnv); err != nil {
		return err
	}
	if err := jsecurity.Init(jEnv); err != nil {
		return err
	}
	if err := jandroid.Init(jEnv); err != nil {
		return err
	}
	if err := jaccess.Init(jEnv); err != nil {
		return err
	}
	return nil
}
