// +build android

package security

import (
	"runtime"
	"unsafe"

	"v.io/core/veyron2/security"
	jutil "v.io/jni/util"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

func caveatDecoder(caveat security.Caveat) (security.CaveatValidator, error) {
	jEnv, freeFunc := jutil.GetEnv(javaVM)
	defer freeFunc()
	env := (*C.JNIEnv)(unsafe.Pointer(jutil.PtrValue(jEnv)))
	jCaveat, err := JavaCaveat(env, caveat)
	if err != nil {
		return nil, err
	}
	jVal, err := jutil.CallStaticObjectMethod(env, jCaveatCoderClass, "decode", []jutil.Sign{caveatSign}, caveatValidatorSign, jCaveat)
	if err != nil {
		return nil, err
	}
	// Reference Java validator; it will be de-referenced when the Go validator
	// created below is garbage-collected (through the finalizer callback we
	// setup just below).
	jValidator := C.NewGlobalRef(env, C.jobject(jVal))
	// Create Go validator.
	validator := &jniCaveatValidator{
		jValidator: jValidator,
	}
	runtime.SetFinalizer(validator, func(v *jniCaveatValidator) {
		jEnv, freeFunc := jutil.GetEnv(javaVM)
		env := (*C.JNIEnv)(jEnv)
		defer freeFunc()
		C.DeleteGlobalRef(env, v.jValidator)
	})
	return validator, nil
}

type jniCaveatValidator struct {
	jValidator C.jobject
}

func (v *jniCaveatValidator) Validate(context security.Context) error {
	env, freeFunc := jutil.GetEnv(javaVM)
	defer freeFunc()
	jContext, err := JavaContext(env, context)
	if err != nil {
		return err
	}
	if err := jutil.CallVoidMethod(env, v.jValidator, "validate", []jutil.Sign{contextSign}, jContext); err != nil {
		return err
	}
	return nil
}
