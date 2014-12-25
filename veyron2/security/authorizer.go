// +build android

package security

import (
	"fmt"
	"runtime"
	"unsafe"

	jutil "v.io/jni/util"
	"v.io/veyron/veyron2/security"
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
	// We cannot cache Java environments as they are only valid in the current
	// thread.  We can, however, cache the Java VM and obtain an environment
	// from it in whatever thread happens to be running at the time.
	var jVM *C.JavaVM
	if status := C.GetJavaVM(env, &jVM); status != 0 {
		return nil, fmt.Errorf("couldn't get Java VM from the (Java) environment")
	}
	// Reference Java dispatcher; it will be de-referenced when the go
	// dispatcher created below is garbage-collected (through the finalizer
	// callback we setup below).
	jAuth = C.NewGlobalRef(env, jAuth)
	a := &authorizer{
		jVM:   jVM,
		jAuth: jAuth,
	}
	runtime.SetFinalizer(a, func(a *authorizer) {
		jEnv, freeFunc := jutil.GetEnv(a.jVM)
		env := (*C.JNIEnv)(jEnv)
		defer freeFunc()
		C.DeleteGlobalRef(env, a.jAuth)
	})
	return a, nil
}

type authorizer struct {
	jVM   *C.JavaVM
	jAuth C.jobject
}

func (a *authorizer) Authorize(context security.Context) error {
	env, freeFunc := jutil.GetEnv(a.jVM)
	defer freeFunc()
	// Create a Java context.
	jContext, err := JavaContext(env, context)
	if err != nil {
		return err
	}
	// Run Java Authorizer.
	contextSign := jutil.ClassSign("io.veyron.veyron.veyron2.security.Context")
	return jutil.CallVoidMethod(env, a.jAuth, "authorize", []jutil.Sign{contextSign}, jContext)
}
