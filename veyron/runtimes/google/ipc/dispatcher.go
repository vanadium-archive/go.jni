// +build android

package ipc

import (
	"fmt"
	"runtime"

	"veyron.io/jni/util"
	isecurity "veyron.io/jni/veyron/runtimes/google/security"
	"veyron.io/veyron/veyron2/security"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

func goDispatcher(env *C.JNIEnv, jDispatcher C.jobject) (*dispatcher, error) {
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
	jDispatcher = C.NewGlobalRef(env, jDispatcher)
	d := &dispatcher{
		jVM:         jVM,
		jDispatcher: jDispatcher,
	}
	runtime.SetFinalizer(d, func(d *dispatcher) {
		jEnv, freeFunc := util.GetEnv(d.jVM)
		env := (*C.JNIEnv)(jEnv)
		defer freeFunc()
		C.DeleteGlobalRef(env, d.jDispatcher)
	})

	return d, nil
}

type dispatcher struct {
	jVM         *C.JavaVM
	jDispatcher C.jobject
}

func (d *dispatcher) Lookup(suffix, method string) (interface{}, security.Authorizer, error) {
	// Get Java environment.
	jEnv, freeFunc := util.GetEnv(d.jVM)
	env := (*C.JNIEnv)(jEnv)
	defer freeFunc()

	// Call Java dispatcher's lookup() method.
	serviceObjectWithAuthorizerSign := util.ClassSign("io.veyron.veyron.veyron2.ipc.ServiceObjectWithAuthorizer")
	tempJObj, err := util.CallObjectMethod(env, d.jDispatcher, "lookup", []util.Sign{util.StringSign}, serviceObjectWithAuthorizerSign, suffix)
	jObj := C.jobject(tempJObj)
	if err != nil {
		return nil, nil, fmt.Errorf("error invoking Java dispatcher's lookup() method: %v", err)
	}
	if jObj == nil {
		// Lookup returned null, which means that the dispatcher isn't handling the object -
		// this is not an error.
		return nil, nil, nil
	}

	// Extract the Java service object and Authorizer.
	jServiceObj, err := util.CallObjectMethod(env, jObj, "getServiceObject", nil, util.ObjectSign)
	if err != nil {
		return nil, nil, err
	}
	if jServiceObj == nil {
		return nil, nil, fmt.Errorf("null service object returned by Java's ServiceObjectWithAuthorizer")
	}
	authSign := util.ClassSign("io.veyron.veyron.veyron2.security.Authorizer")
	jAuth, err := util.CallObjectMethod(env, jObj, "getAuthorizer", nil, authSign)
	if err != nil {
		return nil, nil, err
	}

	// Create Go Invoker and Authorizer.
	i, err := goInvoker(env, C.jobject(jServiceObj))
	if err != nil {
		return nil, nil, err
	}
	a, err := isecurity.GoAuthorizer(env, jAuth)
	if err != nil {
		return nil, nil, err
	}
	return i, a, nil
}
