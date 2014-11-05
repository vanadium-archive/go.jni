// +build android

package ipc

import (
	"io"
	"unsafe"

	jutil "veyron.io/jni/util"
	jsecurity "veyron.io/jni/veyron2/security"
	_ "veyron.io/veyron/veyron/profiles"
	"veyron.io/veyron/veyron2/ipc"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
// #include <stdlib.h>
import "C"

var (
	streamSign = jutil.ClassSign("io.veyron.veyron.veyron.runtimes.google.ipc.Stream")

	// Global reference for io.veyron.veyron.veyron.runtimes.google.ipc.Server class.
	jServerClass C.jclass
	// Global reference for io.veyron.veyron.veyron.runtimes.google.ipc.Client class.
	jClientClass C.jclass
	// Global reference for io.veyron.veyron.veyron.runtimes.google.ipc.ClientCall class.
	jClientCallClass C.jclass
	// Global reference for io.veyron.veyron.veyron.runtimes.google.ipc.ServerCall class.
	jServerCallClass C.jclass
	// Global reference for io.veyron.veyron.veyron.runtimes.google.ipc.Stream class.
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
	env := (*C.JNIEnv)(unsafe.Pointer(jutil.PtrValue(jEnv)))
	// Cache global references to all Java classes used by the package.  This is
	// necessary because JNI gets access to the class loader only in the system
	// thread, so we aren't able to invoke FindClass in other threads.
	jServerClass = C.jclass(jutil.JFindClassOrPrint(env, "io/veyron/veyron/veyron/runtimes/google/ipc/Server"))
	jClientClass = C.jclass(jutil.JFindClassOrPrint(env, "io/veyron/veyron/veyron/runtimes/google/ipc/Client"))
	jClientCallClass = C.jclass(jutil.JFindClassOrPrint(env, "io/veyron/veyron/veyron/runtimes/google/ipc/ClientCall"))
	jServerCallClass = C.jclass(jutil.JFindClassOrPrint(env, "io/veyron/veyron/veyron/runtimes/google/ipc/ServerCall"))
	jStreamClass = C.jclass(jutil.JFindClassOrPrint(env, "io/veyron/veyron/veyron/runtimes/google/ipc/Stream"))
	jVDLInvokerClass = C.jclass(jutil.JFindClassOrPrint(env, "io/veyron/veyron/veyron/runtimes/google/ipc/VDLInvoker"))
	jOptionDefsClass = C.jclass(jutil.JFindClassOrPrint(env, "io/veyron/veyron/veyron2/OptionDefs"))
	jEOFExceptionClass = C.jclass(jutil.JFindClassOrPrint(env, "java/io/EOFException"))
	jStringClass = C.jclass(jutil.JFindClassOrPrint(env, "java/lang/String"))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_ipc_Server_nativeListen
func Java_io_veyron_veyron_veyron_runtimes_google_ipc_Server_nativeListen(env *C.JNIEnv, server C.jobject, goServerPtr C.jlong, jSpec C.jobject) C.jstring {
	spec, err := GoListenSpec(env, jSpec)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	ep, err := (*(*ipc.Server)(jutil.Ptr(goServerPtr))).Listen(spec)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jstring(jutil.JString(env, ep.String()))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_ipc_Server_nativeServe
func Java_io_veyron_veyron_veyron_runtimes_google_ipc_Server_nativeServe(env *C.JNIEnv, jServer C.jobject, goServerPtr C.jlong, name C.jstring, jDispatcher C.jobject) {
	s := (*ipc.Server)(jutil.Ptr(goServerPtr))
	d, err := goDispatcher(env, jDispatcher)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := (*s).ServeDispatcher(jutil.GoString(env, name), d); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_veyron_veyron_veyron_runtimes_google_ipc_Server_nativeGetPublishedNames
func Java_io_veyron_veyron_veyron_runtimes_google_ipc_Server_nativeGetPublishedNames(env *C.JNIEnv, jServer C.jobject, goServerPtr C.jlong) C.jobjectArray {
	names, err := (*(*ipc.Server)(jutil.Ptr(goServerPtr))).Published()
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	ret := C.NewObjectArray(env, C.jsize(len(names)), jStringClass, nil)
	for i, name := range names {
		C.SetObjectArrayElement(env, ret, C.jsize(i), C.jobject(jutil.JString(env, string(name))))
	}
	return ret
}

//export Java_io_veyron_veyron_veyron_runtimes_google_ipc_Server_nativeStop
func Java_io_veyron_veyron_veyron_runtimes_google_ipc_Server_nativeStop(env *C.JNIEnv, server C.jobject, goServerPtr C.jlong) {
	s := (*ipc.Server)(jutil.Ptr(goServerPtr))
	if err := (*s).Stop(); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_veyron_veyron_veyron_runtimes_google_ipc_Server_nativeFinalize
func Java_io_veyron_veyron_veyron_runtimes_google_ipc_Server_nativeFinalize(env *C.JNIEnv, server C.jobject, goServerPtr C.jlong) {
	jutil.GoUnref((*ipc.Server)(jutil.Ptr(goServerPtr)))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_ipc_Client_nativeStartCall
func Java_io_veyron_veyron_veyron_runtimes_google_ipc_Client_nativeStartCall(env *C.JNIEnv, jClient C.jobject, goClientPtr C.jlong, jContext C.jobject, name C.jstring, method C.jstring, jsonArgs C.jobjectArray, jOptions C.jobject) C.jobject {
	c := (*client)(jutil.Ptr(goClientPtr))
	call, err := c.StartCall(env, jContext, jutil.GoString(env, name), jutil.GoString(env, method), jsonArgs, jOptions)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jCall, err := javaClientCall(env, call)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return jCall
}

//export Java_io_veyron_veyron_veyron_runtimes_google_ipc_Client_nativeClose
func Java_io_veyron_veyron_veyron_runtimes_google_ipc_Client_nativeClose(env *C.JNIEnv, jClient C.jobject, goClientPtr C.jlong) {
	(*client)(jutil.Ptr(goClientPtr)).Close()
}

//export Java_io_veyron_veyron_veyron_runtimes_google_ipc_Client_nativeFinalize
func Java_io_veyron_veyron_veyron_runtimes_google_ipc_Client_nativeFinalize(env *C.JNIEnv, jClient C.jobject, goClientPtr C.jlong) {
	jutil.GoUnref((*client)(jutil.Ptr(goClientPtr)))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_ipc_Stream_nativeSend
func Java_io_veyron_veyron_veyron_runtimes_google_ipc_Stream_nativeSend(env *C.JNIEnv, jStream C.jobject, goPtr C.jlong, jItem C.jstring) {
	(*stream)(jutil.Ptr(goPtr)).Send(env, jItem)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_ipc_Stream_nativeRecv
func Java_io_veyron_veyron_veyron_runtimes_google_ipc_Stream_nativeRecv(env *C.JNIEnv, jStream C.jobject, goPtr C.jlong) C.jstring {
	ret, err := (*stream)(jutil.Ptr(goPtr)).Recv(env)
	if err != nil {
		if err == io.EOF {
			jutil.JThrow(env, jEOFExceptionClass, err.Error())
			return nil
		}
		jutil.JThrowV(env, err)
		return nil
	}
	return ret
}

//export Java_io_veyron_veyron_veyron_runtimes_google_ipc_Stream_nativeFinalize
func Java_io_veyron_veyron_veyron_runtimes_google_ipc_Stream_nativeFinalize(env *C.JNIEnv, jStream C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*stream)(jutil.Ptr(goPtr)))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_ipc_ClientCall_nativeFinish
func Java_io_veyron_veyron_veyron_runtimes_google_ipc_ClientCall_nativeFinish(env *C.JNIEnv, jClientCall C.jobject, goClientCallPtr C.jlong) C.jobjectArray {
	ret, err := (*clientCall)(jutil.Ptr(goClientCallPtr)).Finish(env)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return ret
}

//export Java_io_veyron_veyron_veyron_runtimes_google_ipc_ClientCall_nativeCancel
func Java_io_veyron_veyron_veyron_runtimes_google_ipc_ClientCall_nativeCancel(env *C.JNIEnv, jClientCall C.jobject, goClientCallPtr C.jlong) {
	(*clientCall)(jutil.Ptr(goClientCallPtr)).Cancel()
}

//export Java_io_veyron_veyron_veyron_runtimes_google_ipc_ClientCall_nativeFinalize
func Java_io_veyron_veyron_veyron_runtimes_google_ipc_ClientCall_nativeFinalize(env *C.JNIEnv, jClientCall C.jobject, goClientCallPtr C.jlong) {
	jutil.GoUnref((*clientCall)(jutil.Ptr(goClientCallPtr)))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_ipc_ServerCall_nativeBlessings
func Java_io_veyron_veyron_veyron_runtimes_google_ipc_ServerCall_nativeBlessings(env *C.JNIEnv, jServerCall C.jobject, goPtr C.jlong) C.jobject {
	blessings := (*serverCall)(jutil.Ptr(goPtr)).Blessings()
	jBlessings, err := jsecurity.JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jBlessings)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_ipc_ServerCall_nativeFinalize
func Java_io_veyron_veyron_veyron_runtimes_google_ipc_ServerCall_nativeFinalize(env *C.JNIEnv, jServerCall C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*serverCall)(jutil.Ptr(goPtr)))
}
