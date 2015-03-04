// +build android

package context

import (
	"v.io/v23/context"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
// #include <stdlib.h>
import "C"

var (
	classSign = jutil.ClassSign("java.lang.Class")
	// Global reference for io.v.v23.context.VContextImpl class.
	jVContextImplClass C.jclass
	// Global reference for java.jutil.concurrent.CountDownLatch class.
	jCountDownLatchClass C.jclass
)

// Init initializes the JNI code with the given Java environment. This method
// must be called from the main Java thread.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java environment is passed in an empty
// interface and then cast into the package-local environment type.
func Init(jEnv interface{}) error {
	// Cache global references to all Java classes used by the package.  This is
	// necessary because JNI gets access to the class loader only in the system
	// thread, so we aren't able to invoke FindClass in other threads.
	class, err := jutil.JFindClass(jEnv, "io/v/v23/context/VContextImpl")
	if err != nil {
		return err
	}
	jVContextImplClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "java/util/concurrent/CountDownLatch")
	if err != nil {
		return err
	}
	jCountDownLatchClass = C.jclass(class)
	return nil
}

//export Java_io_v_v23_context_VContextImpl_nativeCreate
func Java_io_v_v23_context_VContextImpl_nativeCreate(env *C.JNIEnv, jVContextImpl C.jclass) C.jobject {
	ctx, cancel := context.RootContext()
	jContext, err := JavaContext(env, ctx, cancel)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jContext)
}

//export Java_io_v_v23_context_VContextImpl_nativeDeadline
func Java_io_v_v23_context_VContextImpl_nativeDeadline(env *C.JNIEnv, jVContextImpl C.jobject, goPtr C.jlong) C.jobject {
	d, ok := (*(*context.T)(jutil.Ptr(goPtr))).Deadline()
	if !ok {
		return nil
	}
	jDeadline, err := jutil.JTime(env, d)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jDeadline)
}

//export Java_io_v_v23_context_VContextImpl_nativeDone
func Java_io_v_v23_context_VContextImpl_nativeDone(env *C.JNIEnv, jVContextImpl C.jobject, goPtr C.jlong) C.jobject {
	c := (*(*context.T)(jutil.Ptr(goPtr))).Done()
	jCounter, err := JavaCountDownLatch(env, c)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jCounter)
}

//export Java_io_v_v23_context_VContextImpl_nativeValue
func Java_io_v_v23_context_VContextImpl_nativeValue(env *C.JNIEnv, jVContextImpl C.jobject, goPtr C.jlong, jKey C.jobject) C.jobject {
	key, err := GoContextKey(env, jKey)
	value := (*(*context.T)(jutil.Ptr(goPtr))).Value(key)
	jValue, err := JavaContextValue(env, value)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jValue)
}

//export Java_io_v_v23_context_VContextImpl_nativeWithCancel
func Java_io_v_v23_context_VContextImpl_nativeWithCancel(env *C.JNIEnv, jVContextImpl C.jobject, goPtr C.jlong) C.jobject {
	ctx, cancelFunc := context.WithCancel((*context.T)(jutil.Ptr(goPtr)))
	jCtx, err := JavaContext(env, ctx, cancelFunc)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jCtx)
}

//export Java_io_v_v23_context_VContextImpl_nativeWithDeadline
func Java_io_v_v23_context_VContextImpl_nativeWithDeadline(env *C.JNIEnv, jVContextImpl C.jobject, goPtr C.jlong, jDeadline C.jobject) C.jobject {
	deadline, err := jutil.GoTime(env, jDeadline)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	ctx, cancelFunc := context.WithDeadline((*context.T)(jutil.Ptr(goPtr)), deadline)
	jCtx, err := JavaContext(env, ctx, cancelFunc)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jCtx)
}

//export Java_io_v_v23_context_VContextImpl_nativeWithTimeout
func Java_io_v_v23_context_VContextImpl_nativeWithTimeout(env *C.JNIEnv, jVContextImpl C.jobject, goPtr C.jlong, jTimeout C.jobject) C.jobject {
	timeout, err := jutil.GoDuration(env, jTimeout)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	ctx, cancelFunc := context.WithTimeout((*context.T)(jutil.Ptr(goPtr)), timeout)
	jCtx, err := JavaContext(env, ctx, cancelFunc)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jCtx)
}

//export Java_io_v_v23_context_VContextImpl_nativeWithValue
func Java_io_v_v23_context_VContextImpl_nativeWithValue(env *C.JNIEnv, jVContextImpl C.jobject, goPtr C.jlong, jKey C.jobject, jValue C.jobject) C.jobject {
	key, err := GoContextKey(env, jKey)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	value, err := GoContextValue(env, jValue)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	ctx := context.WithValue((*context.T)(jutil.Ptr(goPtr)), key, value)
	jCtx, err := JavaContext(env, ctx, nil)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jCtx)
}

//export Java_io_v_v23_context_VContextImpl_nativeCancel
func Java_io_v_v23_context_VContextImpl_nativeCancel(env *C.JNIEnv, jVContextImpl C.jobject, goCancelPtr C.jlong) {
	if goCancelPtr != 0 {
		(*(*context.CancelFunc)(jutil.Ptr(goCancelPtr)))()
	}
}

//export Java_io_v_v23_context_VContextImpl_nativeFinalize
func Java_io_v_v23_context_VContextImpl_nativeFinalize(env *C.JNIEnv, jVContextImpl C.jobject, goPtr C.jlong, goCancelPtr C.jlong) {
	jutil.GoUnref((*context.T)(jutil.Ptr(goPtr)))
	if goCancelPtr != 0 {
		jutil.GoUnref((*context.CancelFunc)(jutil.Ptr(goCancelPtr)))
	}
}
