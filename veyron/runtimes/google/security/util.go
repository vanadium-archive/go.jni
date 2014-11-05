// +build android

package security

import (
	jutil "veyron.io/jni/util"
	"veyron.io/veyron/veyron2/security"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// JavaContext converts the provided Go (security) Context into a Java Context object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaContext(jEnv interface{}, context security.Context) (C.jobject, error) {
	jContext, err := jutil.NewObject(jEnv, jContextImplClass, []jutil.Sign{jutil.LongSign}, &context)
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&context) // Un-refed when the Java Context object is finalized.
	return C.jobject(jContext), nil
}
