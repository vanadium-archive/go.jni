// +build android

package ipc

import (
	"io"
	"unsafe"

	"v.io/core/veyron2/ipc"
	"v.io/core/veyron2/vdl"
	"v.io/core/veyron2/vom"

	jutil "v.io/jni/util"
	jcontext "v.io/jni/veyron2/context"
	jsecurity "v.io/jni/veyron2/security"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
// #include <stdlib.h>
import "C"

var (
	optionsSign    = jutil.ClassSign("io.v.core.veyron2.Options")
	streamSign     = jutil.ClassSign("io.v.core.veyron.runtimes.google.ipc.Stream")
	listenAddrSign = jutil.ClassSign("io.v.core.veyron2.ipc.ListenSpec$Address")

	// Global reference for io.v.core.veyron.runtimes.google.ipc.Server class.
	jServerClass C.jclass
	// Global reference for io.v.core.veyron.runtimes.google.ipc.Client class.
	jClientClass C.jclass
	// Global reference for io.v.core.veyron.runtimes.google.ipc.Call class.
	jCallClass C.jclass
	// Global reference for io.v.core.veyron.runtimes.google.ipc.ServerCall class.
	jServerCallClass C.jclass
	// Global reference for io.v.core.veyron.runtimes.google.ipc.Stream class.
	jStreamClass C.jclass
	// Global reference for io.v.core.veyron.runtimes.google.ipc.VDLInvoker class.
	jVDLInvokerClass C.jclass
	// Global reference for io.v.core.veyron2.OptionDefs class.
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
	jServerClass = C.jclass(jutil.JFindClassOrPrint(env, "io/v/core/veyron/runtimes/google/ipc/Server"))
	jClientClass = C.jclass(jutil.JFindClassOrPrint(env, "io/v/core/veyron/runtimes/google/ipc/Client"))
	jCallClass = C.jclass(jutil.JFindClassOrPrint(env, "io/v/core/veyron/runtimes/google/ipc/Call"))
	jServerCallClass = C.jclass(jutil.JFindClassOrPrint(env, "io/v/core/veyron/runtimes/google/ipc/ServerCall"))
	jStreamClass = C.jclass(jutil.JFindClassOrPrint(env, "io/v/core/veyron/runtimes/google/ipc/Stream"))
	jVDLInvokerClass = C.jclass(jutil.JFindClassOrPrint(env, "io/v/core/veyron/runtimes/google/ipc/VDLInvoker"))
	jOptionDefsClass = C.jclass(jutil.JFindClassOrPrint(env, "io/v/core/veyron2/OptionDefs"))
	jEOFExceptionClass = C.jclass(jutil.JFindClassOrPrint(env, "java/io/EOFException"))
	jStringClass = C.jclass(jutil.JFindClassOrPrint(env, "java/lang/String"))
}

//export Java_io_v_core_veyron_runtimes_google_ipc_Server_nativeListen
func Java_io_v_core_veyron_runtimes_google_ipc_Server_nativeListen(env *C.JNIEnv, server C.jobject, goPtr C.jlong, jSpec C.jobject) C.jobjectArray {
	spec, err := GoListenSpec(env, jSpec)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	eps, err := (*(*ipc.Server)(jutil.Ptr(goPtr))).Listen(spec)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	epStrs := make([]string, len(eps))
	for i, ep := range eps {
		epStrs[i] = ep.String()
	}
	return C.jobjectArray(jutil.JStringArray(env, epStrs))
}

//export Java_io_v_core_veyron_runtimes_google_ipc_Server_nativeServe
func Java_io_v_core_veyron_runtimes_google_ipc_Server_nativeServe(env *C.JNIEnv, jServer C.jobject, goPtr C.jlong, name C.jstring, jDispatcher C.jobject) {
	s := (*ipc.Server)(jutil.Ptr(goPtr))
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

//export Java_io_v_core_veyron_runtimes_google_ipc_Server_nativeGetPublishedNames
func Java_io_v_core_veyron_runtimes_google_ipc_Server_nativeGetPublishedNames(env *C.JNIEnv, jServer C.jobject, goPtr C.jlong) C.jobjectArray {
	names, err := (*(*ipc.Server)(jutil.Ptr(goPtr))).Published()
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

//export Java_io_v_core_veyron_runtimes_google_ipc_Server_nativeStop
func Java_io_v_core_veyron_runtimes_google_ipc_Server_nativeStop(env *C.JNIEnv, server C.jobject, goPtr C.jlong) {
	s := (*ipc.Server)(jutil.Ptr(goPtr))
	if err := (*s).Stop(); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_v_core_veyron_runtimes_google_ipc_Server_nativeFinalize
func Java_io_v_core_veyron_runtimes_google_ipc_Server_nativeFinalize(env *C.JNIEnv, server C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*ipc.Server)(jutil.Ptr(goPtr)))
}

//export Java_io_v_core_veyron_runtimes_google_ipc_Client_nativeStartCall
func Java_io_v_core_veyron_runtimes_google_ipc_Client_nativeStartCall(env *C.JNIEnv, jClient C.jobject, goPtr C.jlong, jContext C.jobject, jName C.jstring, jMethod C.jstring, jVomArgs C.jobjectArray, jOptions C.jobject) C.jobject {
	name := jutil.GoString(env, jName)
	method := jutil.GoString(env, jMethod)
	context, err := jcontext.GoContext(env, jContext)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	vomArgs := jutil.GoByteArrayArray(env, jVomArgs)
	// VOM-decode each arguments into a *vdl.Value.
	args := make([]interface{}, len(vomArgs))
	for i := 0; i < len(vomArgs); i++ {
		var err error
		if args[i], err = jutil.VomDecodeToValue(vomArgs[i]); err != nil {
			jutil.JThrowV(env, err)
			return nil
		}
	}

	// Invoke StartCall
	call, err := (*(*ipc.Client)(jutil.Ptr(goPtr))).StartCall(context, name, method, args)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jCall, err := javaCall(env, call)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return jCall
}

//export Java_io_v_core_veyron_runtimes_google_ipc_Client_nativeClose
func Java_io_v_core_veyron_runtimes_google_ipc_Client_nativeClose(env *C.JNIEnv, jClient C.jobject, goPtr C.jlong) {
	(*(*ipc.Client)(jutil.Ptr(goPtr))).Close()
}

//export Java_io_v_core_veyron_runtimes_google_ipc_Client_nativeFinalize
func Java_io_v_core_veyron_runtimes_google_ipc_Client_nativeFinalize(env *C.JNIEnv, jClient C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*ipc.Client)(jutil.Ptr(goPtr)))
}

//export Java_io_v_core_veyron_runtimes_google_ipc_Stream_nativeSend
func Java_io_v_core_veyron_runtimes_google_ipc_Stream_nativeSend(env *C.JNIEnv, jStream C.jobject, goPtr C.jlong, jVomItem C.jbyteArray) {
	vomItem := jutil.GoByteArray(env, jVomItem)
	item, err := jutil.VomDecodeToValue(vomItem)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := (*(*ipc.Stream)(jutil.Ptr(goPtr))).Send(item); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_v_core_veyron_runtimes_google_ipc_Stream_nativeRecv
func Java_io_v_core_veyron_runtimes_google_ipc_Stream_nativeRecv(env *C.JNIEnv, jStream C.jobject, goPtr C.jlong) C.jbyteArray {
	result := new(vdl.Value)
	if err := (*(*ipc.Stream)(jutil.Ptr(goPtr))).Recv(&result); err != nil {
		if err == io.EOF {
			jutil.JThrow(env, jEOFExceptionClass, err.Error())
			return nil
		}
		jutil.JThrowV(env, err)
		return nil
	}
	vomResult, err := vom.Encode(result)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jbyteArray(jutil.JByteArray(env, vomResult))
}

//export Java_io_v_core_veyron_runtimes_google_ipc_Stream_nativeFinalize
func Java_io_v_core_veyron_runtimes_google_ipc_Stream_nativeFinalize(env *C.JNIEnv, jStream C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*ipc.Stream)(jutil.Ptr(goPtr)))
}

//export Java_io_v_core_veyron_runtimes_google_ipc_Call_nativeCloseSend
func Java_io_v_core_veyron_runtimes_google_ipc_Call_nativeCloseSend(env *C.JNIEnv, jCall C.jobject, goPtr C.jlong) {
	if err := (*(*ipc.Call)(jutil.Ptr(goPtr))).CloseSend(); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_v_core_veyron_runtimes_google_ipc_Call_nativeFinish
func Java_io_v_core_veyron_runtimes_google_ipc_Call_nativeFinish(env *C.JNIEnv, jCall C.jobject, goPtr C.jlong, jNumResults C.jint) C.jobjectArray {
	// Have all the results be decoded into *vdl.Value.
	numResults := int(jNumResults)
	resultPtrs := make([]interface{}, numResults)
	for i := 0; i < numResults; i++ {
		value := new(vdl.Value)
		resultPtrs[i] = &value
	}
	// Append mandatory error result and call Finish.
	var appErr error
	if err := (*(*ipc.Call)(jutil.Ptr(goPtr))).Finish(append(resultPtrs, &appErr)...); err != nil {
		// Invocation error.
		jutil.JThrowV(env, err)
		return nil
	}
	if appErr != nil {
		// Application error
		jutil.JThrowV(env, appErr)
		return nil
	}

	// VOM-encode the results.
	vomResults := make([][]byte, numResults)
	for i, resultPtr := range resultPtrs {
		// Remove the pointer from the result.  Simply *resultPtr doesn't work
		// as resultPtr is of type interface{}.
		result := interface{}(jutil.DerefOrDie(resultPtr))
		var err error
		if vomResults[i], err = vom.Encode(result); err != nil {
			jutil.JThrowV(env, err)
			return nil
		}
	}
	return C.jobjectArray(jutil.JByteArrayArray(env, vomResults))
}

//export Java_io_v_core_veyron_runtimes_google_ipc_Call_nativeFinalize
func Java_io_v_core_veyron_runtimes_google_ipc_Call_nativeFinalize(env *C.JNIEnv, jCall C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*ipc.Call)(jutil.Ptr(goPtr)))
}

//export Java_io_v_core_veyron_runtimes_google_ipc_ServerCall_nativeBlessings
func Java_io_v_core_veyron_runtimes_google_ipc_ServerCall_nativeBlessings(env *C.JNIEnv, jServerCall C.jobject, goPtr C.jlong) C.jobject {
	blessings := (*(*ipc.ServerCall)(jutil.Ptr(goPtr))).Blessings()
	jBlessings, err := jsecurity.JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jBlessings)
}

//export Java_io_v_core_veyron_runtimes_google_ipc_ServerCall_nativeFinalize
func Java_io_v_core_veyron_runtimes_google_ipc_ServerCall_nativeFinalize(env *C.JNIEnv, jServerCall C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*ipc.ServerCall)(jutil.Ptr(goPtr)))
}
