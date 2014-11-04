// +build android

package ipc

import (
	"io"
	"unsafe"

	"veyron.io/jni/util"
	jcontext "veyron.io/jni/veyron2/context"
	jsecurity "veyron.io/jni/veyron2/security"
	_ "veyron.io/veyron/veyron/profiles"
	"veyron.io/veyron/veyron2"
	"veyron.io/veyron/veyron2/ipc"
	"veyron.io/veyron/veyron2/rt"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
// #include <stdlib.h>
import "C"

var (
	streamSign = util.ClassSign("io.veyron.veyron.veyron.runtimes.google.Runtime$Stream")

	// Global reference for io.veyron.veyron.veyron.runtimes.google.ipc.Runtime$Server class.
	jServerClass C.jclass
	// Global reference for io.veyron.veyron.veyron.runtimes.google.ipc.Runtime$Client class.
	jClientClass C.jclass
	// Global reference for io.veyron.veyron.veyron.runtimes.google.ipc.Runtime$ClientCall class.
	jClientCallClass C.jclass
	// Global reference for io.veyron.veyron.veyron.runtimes.google.ipc.Runtime$ServerCall class.
	jServerCallClass C.jclass
	// Global reference for io.veyron.veyron.veyron.runtimes.google.ipc.Runtime$Stream class.
	jStreamClass C.jclass
	// Global reference for io.veyron.veyron.veyron.runtimes.google.ipc.VDLInvoker class.
	jVDLInvokerClass C.jclass
	// Global reference for io.veyron.veyron.veyron2.OptionDefs class.
	jOptionDefsClass C.jclass
	// Global reference for java.io.EOFException class.
	jEOFExceptionClass C.jclass
	// Global reference for java.lang.String class.
	jStringClass C.jclass
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
	jServerClass = C.jclass(util.JFindClassOrPrint(env, "io/veyron/veyron/veyron/runtimes/google/Runtime$Server"))
	jClientClass = C.jclass(util.JFindClassOrPrint(env, "io/veyron/veyron/veyron/runtimes/google/Runtime$Client"))
	jClientCallClass = C.jclass(util.JFindClassOrPrint(env, "io/veyron/veyron/veyron/runtimes/google/Runtime$ClientCall"))
	jServerCallClass = C.jclass(util.JFindClassOrPrint(env, "io/veyron/veyron/veyron/runtimes/google/Runtime$ServerCall"))
	jStreamClass = C.jclass(util.JFindClassOrPrint(env, "io/veyron/veyron/veyron/runtimes/google/Runtime$Stream"))
	jVDLInvokerClass = C.jclass(util.JFindClassOrPrint(env, "io/veyron/veyron/veyron/runtimes/google/VDLInvoker"))
	jOptionDefsClass = C.jclass(util.JFindClassOrPrint(env, "io/veyron/veyron/veyron2/OptionDefs"))
	jEOFExceptionClass = C.jclass(util.JFindClassOrPrint(env, "java/io/EOFException"))
	jStringClass = C.jclass(util.JFindClassOrPrint(env, "java/lang/String"))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_nativeInit
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_nativeInit(env *C.JNIEnv, jRuntime C.jclass, jOptions C.jobject) C.jlong {
	opts, err := getRuntimeOpts(env, jOptions)
	if err != nil {
		util.JThrowV(env, err)
		return C.jlong(0)
	}
	r := rt.Init(opts...)
	util.GoRef(&r) // Un-refed when the Java Runtime object is finalized.
	return C.jlong(util.PtrValue(&r))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_nativeNewRuntime
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_nativeNewRuntime(env *C.JNIEnv, jRuntime C.jclass, jOptions C.jobject) C.jlong {
	opts, err := getRuntimeOpts(env, jOptions)
	if err != nil {
		util.JThrowV(env, err)
		return C.jlong(0)
	}
	r, err := rt.New(opts...)
	if err != nil {
		util.JThrowV(env, err)
		return C.jlong(0)
	}
	util.GoRef(&r)
	return C.jlong(util.PtrValue(&r))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_nativeNewClient
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_nativeNewClient(env *C.JNIEnv, jRuntime C.jobject, goPtr C.jlong, jOptions C.jobject) C.jobject {
	r := (*veyron2.Runtime)(util.Ptr(goPtr))
	// No options supported yet.
	rc, err := (*r).NewClient()
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	c := newClient(rc)
	jClient, err := javaClient(env, c)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return jClient
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_nativeNewServer
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_nativeNewServer(env *C.JNIEnv, jRuntime C.jobject, goPtr C.jlong, jOptions C.jobject) C.jobject {
	r := (*veyron2.Runtime)(util.Ptr(goPtr))
	// No options supported yet.
	s, err := (*r).NewServer()
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	jServer, err := javaServer(env, s)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return jServer
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_nativeGetClient
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_nativeGetClient(env *C.JNIEnv, jRuntime C.jobject, goPtr C.jlong) C.jobject {
	r := (*veyron2.Runtime)(util.Ptr(goPtr))
	rc := (*r).Client()
	c := newClient(rc)
	jClient, err := javaClient(env, c)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return jClient
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_nativeNewContext
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_nativeNewContext(env *C.JNIEnv, jRuntime C.jobject, goPtr C.jlong) C.jobject {
	r := (*veyron2.Runtime)(util.Ptr(goPtr))
	c := (*r).NewContext()
	jContext, err := jcontext.JavaContext(env, c, nil)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return C.jobject(jContext)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_nativeGetNamespace
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_nativeGetNamespace(env *C.JNIEnv, jRuntime C.jobject, goPtr C.jlong) C.jlong {
	r := (*veyron2.Runtime)(util.Ptr(goPtr))
	n := (*r).Namespace()
	util.GoRef(&n) // Un-refed when the Java Namespace object is finalized.
	return C.jlong(util.PtrValue(&n))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_nativeFinalize
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_nativeFinalize(env *C.JNIEnv, jRuntime C.jobject, goPtr C.jlong) {
	util.GoUnref((*veyron2.Runtime)(util.Ptr(goPtr)))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024Server_nativeListen
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024Server_nativeListen(env *C.JNIEnv, server C.jobject, goServerPtr C.jlong, jSpec C.jobject) C.jstring {
	spec, err := GoListenSpec(env, jSpec)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	ep, err := (*(*ipc.Server)(util.Ptr(goServerPtr))).Listen(spec)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return C.jstring(util.JString(env, ep.String()))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024Server_nativeServe
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024Server_nativeServe(env *C.JNIEnv, jServer C.jobject, goServerPtr C.jlong, name C.jstring, jDispatcher C.jobject) {
	s := (*ipc.Server)(util.Ptr(goServerPtr))
	d, err := goDispatcher(env, jDispatcher)
	if err != nil {
		util.JThrowV(env, err)
		return
	}
	if err := (*s).ServeDispatcher(util.GoString(env, name), d); err != nil {
		util.JThrowV(env, err)
		return
	}
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024Server_nativeGetPublishedNames
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024Server_nativeGetPublishedNames(env *C.JNIEnv, jServer C.jobject, goServerPtr C.jlong) C.jobjectArray {
	names, err := (*(*ipc.Server)(util.Ptr(goServerPtr))).Published()
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	ret := C.NewObjectArray(env, C.jsize(len(names)), jStringClass, nil)
	for i, name := range names {
		C.SetObjectArrayElement(env, ret, C.jsize(i), C.jobject(util.JString(env, string(name))))
	}
	return ret
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024Server_nativeStop
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024Server_nativeStop(env *C.JNIEnv, server C.jobject, goServerPtr C.jlong) {
	s := (*ipc.Server)(util.Ptr(goServerPtr))
	if err := (*s).Stop(); err != nil {
		util.JThrowV(env, err)
		return
	}
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024Server_nativeFinalize
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024Server_nativeFinalize(env *C.JNIEnv, server C.jobject, goServerPtr C.jlong) {
	util.GoUnref((*ipc.Server)(util.Ptr(goServerPtr)))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024Client_nativeStartCall
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024Client_nativeStartCall(env *C.JNIEnv, jClient C.jobject, goClientPtr C.jlong, jContext C.jobject, name C.jstring, method C.jstring, jsonArgs C.jobjectArray, jOptions C.jobject) C.jobject {
	c := (*client)(util.Ptr(goClientPtr))
	call, err := c.StartCall(env, jContext, util.GoString(env, name), util.GoString(env, method), jsonArgs, jOptions)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	jCall, err := javaClientCall(env, call)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return jCall
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024Client_nativeClose
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024Client_nativeClose(env *C.JNIEnv, jClient C.jobject, goClientPtr C.jlong) {
	(*client)(util.Ptr(goClientPtr)).Close()
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024Client_nativeFinalize
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024Client_nativeFinalize(env *C.JNIEnv, jClient C.jobject, goClientPtr C.jlong) {
	util.GoUnref((*client)(util.Ptr(goClientPtr)))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024Stream_nativeSend
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024Stream_nativeSend(env *C.JNIEnv, jStream C.jobject, goPtr C.jlong, jItem C.jstring) {
	(*stream)(util.Ptr(goPtr)).Send(env, jItem)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024Stream_nativeRecv
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024Stream_nativeRecv(env *C.JNIEnv, jStream C.jobject, goPtr C.jlong) C.jstring {
	ret, err := (*stream)(util.Ptr(goPtr)).Recv(env)
	if err != nil {
		if err == io.EOF {
			util.JThrow(env, jEOFExceptionClass, err.Error())
			return nil
		}
		util.JThrowV(env, err)
		return nil
	}
	return ret
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024Stream_nativeFinalize
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024Stream_nativeFinalize(env *C.JNIEnv, jStream C.jobject, goPtr C.jlong) {
	util.GoUnref((*stream)(util.Ptr(goPtr)))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024ClientCall_nativeFinish
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024ClientCall_nativeFinish(env *C.JNIEnv, jClientCall C.jobject, goClientCallPtr C.jlong) C.jobjectArray {
	ret, err := (*clientCall)(util.Ptr(goClientCallPtr)).Finish(env)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return ret
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024ClientCall_nativeCancel
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024ClientCall_nativeCancel(env *C.JNIEnv, jClientCall C.jobject, goClientCallPtr C.jlong) {
	(*clientCall)(util.Ptr(goClientCallPtr)).Cancel()
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024ClientCall_nativeFinalize
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024ClientCall_nativeFinalize(env *C.JNIEnv, jClientCall C.jobject, goClientCallPtr C.jlong) {
	util.GoUnref((*clientCall)(util.Ptr(goClientCallPtr)))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024ServerCall_nativeBlessings
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024ServerCall_nativeBlessings(env *C.JNIEnv, jServerCall C.jobject, goPtr C.jlong) C.jobject {
	blessings := (*serverCall)(util.Ptr(goPtr)).Blessings()
	jBlessings, err := jsecurity.JavaBlessings(env, blessings)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return C.jobject(jBlessings)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024ServerCall_nativeFinalize
func Java_io_veyron_veyron_veyron_runtimes_google_Runtime_00024ServerCall_nativeFinalize(env *C.JNIEnv, jServerCall C.jobject, goPtr C.jlong) {
	util.GoUnref((*serverCall)(util.Ptr(goPtr)))
}
