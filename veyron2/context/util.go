// +build android

package context

import (
	"fmt"
	"log"
	"runtime"
	"unsafe"

	"veyron.io/jni/util"
	"veyron.io/veyron/veyron2/context"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

type goContextKey string

type goContextValue struct {
	jObj C.jobject
}

// JavaContext converts the provided Go Context into a Java Context.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaContext(jEnv interface{}, ctx context.T, cancel context.CancelFunc) (C.jobject, error) {
	cancelPtr := int64(0)
	if cancel != nil {
		cancelPtr = int64(util.PtrValue(&cancel))
	}
	jCtx, err := util.NewObject(jEnv, jContextImplClass, []util.Sign{util.LongSign, util.LongSign}, int64(util.PtrValue(&ctx)), cancelPtr)
	if err != nil {
		return nil, err
	}
	util.GoRef(&ctx) // Un-refed when the Java context object is finalized.
	if cancel != nil {
		util.GoRef(&cancel) // Un-refed when the Java context object is finalized.
	}
	return C.jobject(jCtx), err
}

// GoContext converts the provided Java Context into a Go context.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoContext(jEnv, jContextObj interface{}) (context.T, error) {
	jContext := getObject(jContextObj)
	if jContext == nil {
		return nil, nil
	}
	goCtxPtr, err := util.CallLongMethod(jEnv, jContext, "nativePtr", nil)
	if err != nil {
		return nil, err
	}
	return *(*context.T)(util.Ptr(goCtxPtr)), nil
}

// JavaCountDownLatch creates a Java CountDownLatch object with an initial count
// of one that counts down (to zero) the moment the value is sent on the
// provided Go channel or if the channel gets closed.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaCountDownLatch(jEnv interface{}, c <-chan struct{}) (C.jobject, error) {
	env := getEnv(jEnv)
	jLatchObj, err := util.NewObject(env, jCountDownLatchClass, []util.Sign{util.IntSign}, int(1))
	if err != nil {
		return nil, err
	}
	jLatch := C.jobject(jLatchObj)
	// We cannot cache Java environments as they are only valid in the current
	// thread.  We can, however, cache the Java VM and obtain an environment
	// from it in whatever thread happens to be running at the time.
	var jVM *C.JavaVM
	if status := C.GetJavaVM(env, &jVM); status != 0 {
		return nil, fmt.Errorf("couldn't get Java VM from the (Java) environment")
	}
	// Reference Java CountDownLatch; it will be de-referenced when the goroutine below exits.
	jLatch = C.NewGlobalRef(env, jLatch)
	go func() {
		<-c
		javaEnv, freeFunc := util.GetEnv(jVM)
		jenv := (*C.JNIEnv)(javaEnv)
		defer freeFunc()
		if err := util.CallVoidMethod(jenv, jLatch, "countDown", nil); err != nil {
			log.Printf("Error decrementing CountDownLatch: %v", err)
		}
		C.DeleteGlobalRef(jenv, jLatch)
	}()
	return jLatch, nil
}

// GoContextKey creates a Go Context key given the Java Context key.  The
// returned key guarantees that the two Java keys will be equal iff (1) they
// belong to the same class, and (2) they have the same hashCode().
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoContextKey(jEnv, jKeyObj interface{}) (interface{}, error) {
	env := getEnv(jEnv)
	jKey := getObject(jKeyObj)

	// Create a lookup key we use to map Java context keys to Go context keys.
	hashCode, err := util.CallIntMethod(env, jKey, "hashCode", nil)
	if err != nil {
		return nil, err
	}
	jClass, err := util.CallObjectMethod(env, jKey, "getClass", nil, classSign)
	if err != nil {
		return nil, err
	}
	className, err := util.CallStringMethod(env, jClass, "getName", nil)
	if err != nil {
		return nil, err
	}
	return goContextKey(fmt.Sprintf("%s:%d", className, hashCode)), nil
}

// GoContextValue returns the Go Context value given the Java Context value.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoContextValue(jEnv, jValueObj interface{}) (interface{}, error) {
	env := getEnv(jEnv)
	jValue := getObject(jValueObj)
	// We cannot cache Java environments as they are only valid in the current
	// thread.  We can, however, cache the Java VM and obtain an environment
	// from it in whatever thread happens to be running at the time.
	var jVM *C.JavaVM
	if status := C.GetJavaVM(env, &jVM); status != 0 {
		return nil, fmt.Errorf("couldn't get Java VM from the (Java) environment")
	}

	// Reference Java object; it will be de-referenced when the Go wrapper
	// object created below is garbage-collected (via the finalizer we setup
	// just below.)
	jValue = C.NewGlobalRef(env, jValue)
	value := &goContextValue{
		jObj: jValue,
	}
	runtime.SetFinalizer(value, func(value *goContextValue) {
		javaEnv, freeFunc := util.GetEnv(jVM)
		jenv := (*C.JNIEnv)(javaEnv)
		defer freeFunc()
		C.DeleteGlobalRef(jenv, value.jObj)
	})
	return value, nil
}

// JavaContextValue returns the Java Context value given the Go Context value.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaContextValue(jEnv interface{}, value interface{}) (C.jobject, error) {
	val, ok := value.(*goContextValue)
	if !ok {
		return nil, fmt.Errorf("Invalid type %T for value %v, wanted goContextValue", value, value)
	}
	return val.jObj, nil
}

func getEnv(jEnv interface{}) *C.JNIEnv {
	return (*C.JNIEnv)(unsafe.Pointer(util.PtrValue(jEnv)))
}
func getObject(jObj interface{}) C.jobject {
	return C.jobject(unsafe.Pointer(util.PtrValue(jObj)))
}
