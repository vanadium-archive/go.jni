// +build android

package security

import (
	"unsafe"

	"v.io/core/veyron2/security"
	jutil "v.io/jni/util"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

func caveatValidator(context security.Context, caveat security.Caveat) error {
	jEnv, freeFunc := jutil.GetEnv(javaVM)
	defer freeFunc()
	env := (*C.JNIEnv)(unsafe.Pointer(jutil.PtrValue(jEnv)))

	jContext, err := JavaContext(env, context)
	if err != nil {
		return err
	}
	jCaveat, err := JavaCaveat(env, caveat)
	if err != nil {
		return err
	}
	return jutil.CallStaticVoidMethod(env, jCaveatRegistryClass, "validate", []jutil.Sign{contextSign, caveatSign}, jContext, jCaveat)
}
