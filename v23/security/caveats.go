// +build android

package security

import (
	"unsafe"

	"v.io/v23/security"
	jutil "v.io/x/jni/util"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

func caveatValidator(call security.Call, caveat security.Caveat) error {
	jEnv, freeFunc := jutil.GetEnv()
	defer freeFunc()
	env := (*C.JNIEnv)(unsafe.Pointer(jutil.PtrValue(jEnv)))

	jCall, err := JavaCall(env, call)
	if err != nil {
		return err
	}
	jCaveat, err := JavaCaveat(env, caveat)
	if err != nil {
		return err
	}
	return jutil.CallStaticVoidMethod(env, jCaveatRegistryClass, "validate", []jutil.Sign{callSign, caveatSign}, jCall, jCaveat)
}
