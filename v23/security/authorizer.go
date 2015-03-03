// +build android

package security

import (
	"runtime"
	"unsafe"

	"v.io/v23/security"
	jutil "v.io/x/jni/util"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// GoAuthorizer converts the given Java authorizer into a Go authorizer.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoAuthorizer(jEnv, jAuthObj interface{}) (security.Authorizer, error) {
	env := (*C.JNIEnv)(unsafe.Pointer(jutil.PtrValue(jEnv)))
	jAuth := C.jobject(unsafe.Pointer(jutil.PtrValue(jAuthObj)))
	if jAuth == nil {
		return nil, nil
	}
	// Reference Java dispatcher; it will be de-referenced when the go
	// dispatcher created below is garbage-collected (through the finalizer
	// callback we setup below).
	jAuth = C.NewGlobalRef(env, jAuth)
	a := &authorizer{
		jAuth: jAuth,
	}
	runtime.SetFinalizer(a, func(a *authorizer) {
		jEnv, freeFunc := jutil.GetEnv()
		env := (*C.JNIEnv)(jEnv)
		defer freeFunc()
		C.DeleteGlobalRef(env, a.jAuth)
	})
	return a, nil
}

type authorizer struct {
	jAuth C.jobject
}

func (a *authorizer) Authorize(context security.Context) error {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	// Create a Java context.
	jContext, err := JavaContext(env, context)
	if err != nil {
		return err
	}
	// Run Java Authorizer.
	contextSign := jutil.ClassSign("io.v.v23.security.VContext")
	return jutil.CallVoidMethod(env, a.jAuth, "authorize", []jutil.Sign{contextSign}, jContext)
}
