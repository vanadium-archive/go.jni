// +build android

package context

import (
	"log"
	"unsafe"

	"veyron.io/jni/util"
	_ "veyron.io/veyron/veyron/profiles"
	"veyron.io/veyron/veyron2/context"
	"veyron.io/veyron/veyron2/rt"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
// #include <stdlib.h>
import "C"

var (
	classSign = util.ClassSign("java.lang.Class")
	// Global reference for io.veyron.veyron.veyron2.context.ContextImpl class.
	jContextImplClass C.jclass
	// Global reference for java.util.concurrent.CountDownLatch class.
	jCountDownLatchClass C.jclass
)

// Init initializes the JNI code with the given Java environment. This method
// must be called from the main Java thread.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java environment is passed in an empty
// interface and then cast into the package-local environment type.
func Init(jEnv interface{}) {
	env := (*C.JNIEnv)(unsafe.Pointer(util.PtrValue(jEnv)))
	// Cache global references to all Java classes used by the package.  This is
	// necessary because JNI gets access to the class loader only in the system
	// thread, so we aren't able to invoke FindClass in other threads.
	jContextImplClass = C.jclass(util.JFindClassOrPrint(env, "io/veyron/veyron/veyron2/context/ContextImpl"))
	jCountDownLatchClass = C.jclass(util.JFindClassOrPrint(env, "java/util/concurrent/CountDownLatch"))
}

//export Java_io_veyron_veyron_veyron2_context_ContextImpl_nativeCreate
func Java_io_veyron_veyron_veyron2_context_ContextImpl_nativeCreate(env *C.JNIEnv, jContextClass C.jclass) C.jobject {
	// NOTE(spetrovic): we create a context here using a default runtime.  Since Java doesn't really
	// expose the Runtime argument, this is probably OK.
	r := rt.R()
	ctx := r.NewContext()
	jCtx, err := JavaContext(env, ctx, nil)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return jCtx
}

//export Java_io_veyron_veyron_veyron2_context_ContextImpl_nativeDeadline
func Java_io_veyron_veyron_veyron2_context_ContextImpl_nativeDeadline(env *C.JNIEnv, jContextObj C.jobject, goPtr C.jlong) C.jobject {

	d, ok := (*(*context.T)(util.Ptr(goPtr))).Deadline()
	if !ok {
		return nil
	}
	log.Println("Go got deadline: ", d)
	jDeadline, err := util.JTime(env, d)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return C.jobject(jDeadline)
}

//export Java_io_veyron_veyron_veyron2_context_ContextImpl_nativeDone
func Java_io_veyron_veyron_veyron2_context_ContextImpl_nativeDone(env *C.JNIEnv, jContextObj C.jobject, goPtr C.jlong) C.jobject {
	c := (*(*context.T)(util.Ptr(goPtr))).Done()
	jCounter, err := JavaCountDownLatch(env, c)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return jCounter
}

//export Java_io_veyron_veyron_veyron2_context_ContextImpl_nativeValue
func Java_io_veyron_veyron_veyron2_context_ContextImpl_nativeValue(env *C.JNIEnv, jContextObj C.jobject, goPtr C.jlong, jKey C.jobject) C.jobject {
	key, err := GoContextKey(env, jKey)
	value := (*(*context.T)(util.Ptr(goPtr))).Value(key)
	jValue, err := JavaContextValue(env, value)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return jValue
}

//export Java_io_veyron_veyron_veyron2_context_ContextImpl_nativeWithCancel
func Java_io_veyron_veyron_veyron2_context_ContextImpl_nativeWithCancel(env *C.JNIEnv, jContextObj C.jobject, goPtr C.jlong) C.jobject {
	ctx, cancelFunc := (*(*context.T)(util.Ptr(goPtr))).WithCancel()
	jCtx, err := JavaContext(env, ctx, cancelFunc)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return jCtx
}

//export Java_io_veyron_veyron_veyron2_context_ContextImpl_nativeWithDeadline
func Java_io_veyron_veyron_veyron2_context_ContextImpl_nativeWithDeadline(env *C.JNIEnv, jContextObj C.jobject, goPtr C.jlong, jDeadline C.jobject) C.jobject {
	deadline, err := util.GoTime(env, jDeadline)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	ctx, cancelFunc := (*(*context.T)(util.Ptr(goPtr))).WithDeadline(deadline)
	jCtx, err := JavaContext(env, ctx, cancelFunc)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return jCtx
}

//export Java_io_veyron_veyron_veyron2_context_ContextImpl_nativeWithTimeout
func Java_io_veyron_veyron_veyron2_context_ContextImpl_nativeWithTimeout(env *C.JNIEnv, jContextObj C.jobject, goPtr C.jlong, jTimeout C.jobject) C.jobject {
	timeout, err := util.GoDuration(env, jTimeout)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	ctx, cancelFunc := (*(*context.T)(util.Ptr(goPtr))).WithTimeout(timeout)
	jCtx, err := JavaContext(env, ctx, cancelFunc)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return jCtx
}

//export Java_io_veyron_veyron_veyron2_context_ContextImpl_nativeWithValue
func Java_io_veyron_veyron_veyron2_context_ContextImpl_nativeWithValue(env *C.JNIEnv, jContextObj C.jobject, goPtr C.jlong, jKey C.jobject, jValue C.jobject) C.jobject {
	key, err := GoContextKey(env, jKey)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	value, err := GoContextValue(env, jValue)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	ctx := (*(*context.T)(util.Ptr(goPtr))).WithValue(key, value)
	jCtx, err := JavaContext(env, ctx, nil)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return jCtx
}

//export Java_io_veyron_veyron_veyron2_context_ContextImpl_nativeCancel
func Java_io_veyron_veyron_veyron2_context_ContextImpl_nativeCancel(env *C.JNIEnv, jContextObj C.jobject, goCancelPtr C.jlong) {
	if goCancelPtr != 0 {
		(*(*context.CancelFunc)(util.Ptr(goCancelPtr)))()
	}
}

//export Java_io_veyron_veyron_veyron2_context_ContextImpl_nativeFinalize
func Java_io_veyron_veyron_veyron2_context_ContextImpl_nativeFinalize(env *C.JNIEnv, jContextObj C.jobject, goPtr C.jlong, goCancelPtr C.jlong) {
	util.GoUnref((*context.T)(util.Ptr(goPtr)))
	if goCancelPtr != 0 {
		util.GoUnref((*context.CancelFunc)(util.Ptr(goPtr)))
	}
}
