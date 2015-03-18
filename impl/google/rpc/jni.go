// +build android

package rpc

import (
	"io"
	"log"
	"unsafe"

	"v.io/v23/rpc"
	"v.io/v23/vdl"
	"v.io/v23/vom"

	jchannel "v.io/x/jni/impl/google/channel"
	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
	jsecurity "v.io/x/jni/v23/security"
)

// #include "jni.h"
// #include <stdlib.h>
import "C"

var (
	optionsSign     = jutil.ClassSign("io.v.v23.Options")
	streamSign      = jutil.ClassSign("io.v.impl.google.rpc.Stream")
	listenAddrSign  = jutil.ClassSign("io.v.v23.rpc.ListenSpec$Address")
	serverStateSign = jutil.ClassSign("io.v.v23.rpc.ServerState")
	// Global reference for io.v.impl.google.rpc.Server class.
	jServerClass C.jclass
	// Global reference for io.v.impl.google.rpc.Client class.
	jClientClass C.jclass
	// Global reference for io.v.impl.google.rpc.Call class.
	jCallClass C.jclass
	// Global reference for io.v.impl.google.rpc.StreamServerCall class.
	jStreamServerCallClass C.jclass
	// Global reference for io.v.impl.google.rpc.Stream class.
	jStreamClass C.jclass
	// Global reference for io.v.impl.google.rpc.VDLInvoker class.
	jVDLInvokerClass C.jclass
	// Global reference for io.v.v23.rpc.ServerStatus class.
	jServerStatusClass C.jclass
	// Global reference for io.v.v23.rpc.ServerState class.
	jServerStateClass C.jclass
	// Global reference for io.v.v23.rpc.MountStatus class.
	jMountStatusClass C.jclass
	// Global reference for io.v.v23.rpc.ProxyStatus class.
	jProxyStatusClass C.jclass
	// Global reference for io.v.v23.rpc.ListenSpec class.
	jListenSpecClass C.jclass
	// Global reference for io.v.v23.rpc.ListenSpec$Address class.
	jListenSpecAddressClass C.jclass
	// Global reference for io.v.v23.rpc.NetworkChange class.
	jNetworkChangeClass C.jclass
	// Global reference for io.v.v23.OptionDefs class.
	jOptionDefsClass C.jclass
	// Global reference for java.io.EOFException class.
	jEOFExceptionClass C.jclass
	// Global reference for java.lang.String class.
	jStringClass C.jclass
	// Global reference for io.v.v23.vdlroot.signature.Interface class.
	jInterfaceClass C.jclass
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
	class, err := jutil.JFindClass(jEnv, "io/v/impl/google/rpc/Server")
	if err != nil {
		return err
	}
	jServerClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/impl/google/rpc/Client")
	if err != nil {
		return err
	}
	jClientClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/impl/google/rpc/Call")
	if err != nil {
		return err
	}
	jCallClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/impl/google/rpc/StreamServerCall")
	if err != nil {
		return err
	}
	jStreamServerCallClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/impl/google/rpc/Stream")
	if err != nil {
		return err
	}
	jStreamClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/impl/google/rpc/VDLInvoker")
	if err != nil {
		return err
	}
	jVDLInvokerClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/rpc/ServerStatus")
	if err != nil {
		return err
	}
	jServerStatusClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/rpc/ServerState")
	if err != nil {
		return err
	}
	jServerStateClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/rpc/MountStatus")
	if err != nil {
		return err
	}
	jMountStatusClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/rpc/ProxyStatus")
	if err != nil {
		return err
	}
	jProxyStatusClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/rpc/ListenSpec")
	if err != nil {
		return err
	}
	jListenSpecClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/rpc/ListenSpec$Address")
	if err != nil {
		return err
	}
	jListenSpecAddressClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/rpc/NetworkChange")
	if err != nil {
		return err
	}
	jNetworkChangeClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/OptionDefs")
	if err != nil {
		return err
	}
	jOptionDefsClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "java/io/EOFException")
	if err != nil {
		return err
	}
	jEOFExceptionClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "java/lang/String")
	if err != nil {
		return err
	}
	jStringClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/vdlroot/signature/Interface")
	if err != nil {
		return err
	}
	jInterfaceClass = C.jclass(class)
	return nil
}

//export Java_io_v_impl_google_rpc_Server_nativeListen
func Java_io_v_impl_google_rpc_Server_nativeListen(env *C.JNIEnv, jServer C.jobject, goPtr C.jlong, jSpec C.jobject) C.jobjectArray {
	spec, err := GoListenSpec(env, jSpec)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	eps, err := (*(*rpc.Server)(jutil.Ptr(goPtr))).Listen(spec)
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

//export Java_io_v_impl_google_rpc_Server_nativeServe
func Java_io_v_impl_google_rpc_Server_nativeServe(env *C.JNIEnv, jServer C.jobject, goPtr C.jlong, jName C.jstring, jDispatcher C.jobject) {
	name := jutil.GoString(env, jName)
	d, err := goDispatcher(env, jDispatcher)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := (*(*rpc.Server)(jutil.Ptr(goPtr))).ServeDispatcher(name, d); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_v_impl_google_rpc_Server_nativeAddName
func Java_io_v_impl_google_rpc_Server_nativeAddName(env *C.JNIEnv, jServer C.jobject, goPtr C.jlong, jName C.jstring) {
	name := jutil.GoString(env, jName)
	if err := (*(*rpc.Server)(jutil.Ptr(goPtr))).AddName(name); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_v_impl_google_rpc_Server_nativeRemoveName
func Java_io_v_impl_google_rpc_Server_nativeRemoveName(env *C.JNIEnv, jServer C.jobject, goPtr C.jlong, jName C.jstring) {
	name := jutil.GoString(env, jName)
	(*(*rpc.Server)(jutil.Ptr(goPtr))).RemoveName(name)
}

//export Java_io_v_impl_google_rpc_Server_nativeGetStatus
func Java_io_v_impl_google_rpc_Server_nativeGetStatus(env *C.JNIEnv, jServer C.jobject, goPtr C.jlong) C.jobject {
	status := (*(*rpc.Server)(jutil.Ptr(goPtr))).Status()
	jStatus, err := JavaServerStatus(env, status)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jStatus)
}

//export Java_io_v_impl_google_rpc_Server_nativeWatchNetwork
func Java_io_v_impl_google_rpc_Server_nativeWatchNetwork(env *C.JNIEnv, jServer C.jobject, goPtr C.jlong) C.jobject {
	networkChan := make(chan rpc.NetworkChange, 100)
	(*(*rpc.Server)(jutil.Ptr(goPtr))).WatchNetwork(networkChan)
	retChan := make(chan C.jobject, 100)
	go func() {
		for change := range networkChan {
			jEnv, freeFunc := jutil.GetEnv()
			env := (*C.JNIEnv)(jEnv)
			defer freeFunc()
			jChangeObj, err := JavaNetworkChange(env, change)
			if err != nil {
				log.Println("Couldn't convert Go NetworkChange %v to Java", change)
				continue
			}
			jChange := C.jobject(jutil.NewGlobalRef(env, jChangeObj))
			retChan <- jChange
		}
		close(retChan)
	}()
	jInputChannel, err := jchannel.JavaInputChannel(env, &retChan, &networkChan)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jInputChannel)
}

//export Java_io_v_impl_google_rpc_Server_nativeUnwatchNetwork
func Java_io_v_impl_google_rpc_Server_nativeUnwatchNetwork(env *C.JNIEnv, jServer C.jobject, goPtr C.jlong, jInputChannel C.jobject) {
	goNetworkChanPtr, err := jutil.CallLongMethod(env, jInputChannel, "getSourceNativePtr", nil)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	networkChan := *(*chan rpc.NetworkChange)(unsafe.Pointer(uintptr(goNetworkChanPtr)))
	(*(*rpc.Server)(jutil.Ptr(goPtr))).UnwatchNetwork(networkChan)
}

//export Java_io_v_impl_google_rpc_Server_nativeStop
func Java_io_v_impl_google_rpc_Server_nativeStop(env *C.JNIEnv, server C.jobject, goPtr C.jlong) {
	s := (*rpc.Server)(jutil.Ptr(goPtr))
	if err := (*s).Stop(); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_v_impl_google_rpc_Server_nativeFinalize
func Java_io_v_impl_google_rpc_Server_nativeFinalize(env *C.JNIEnv, server C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*rpc.Server)(jutil.Ptr(goPtr)))
}

//export Java_io_v_impl_google_rpc_Client_nativeStartCall
func Java_io_v_impl_google_rpc_Client_nativeStartCall(env *C.JNIEnv, jClient C.jobject, goPtr C.jlong, jContext C.jobject, jName C.jstring, jMethod C.jstring, jVomArgs C.jobjectArray, jOptions C.jobject) C.jobject {
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
	call, err := (*(*rpc.Client)(jutil.Ptr(goPtr))).StartCall(context, name, method, args)
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

//export Java_io_v_impl_google_rpc_Client_nativeClose
func Java_io_v_impl_google_rpc_Client_nativeClose(env *C.JNIEnv, jClient C.jobject, goPtr C.jlong) {
	(*(*rpc.Client)(jutil.Ptr(goPtr))).Close()
}

//export Java_io_v_impl_google_rpc_Client_nativeFinalize
func Java_io_v_impl_google_rpc_Client_nativeFinalize(env *C.JNIEnv, jClient C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*rpc.Client)(jutil.Ptr(goPtr)))
}

//export Java_io_v_impl_google_rpc_Stream_nativeSend
func Java_io_v_impl_google_rpc_Stream_nativeSend(env *C.JNIEnv, jStream C.jobject, goPtr C.jlong, jVomItem C.jbyteArray) {
	vomItem := jutil.GoByteArray(env, jVomItem)
	item, err := jutil.VomDecodeToValue(vomItem)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := (*(*rpc.Stream)(jutil.Ptr(goPtr))).Send(item); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_v_impl_google_rpc_Stream_nativeRecv
func Java_io_v_impl_google_rpc_Stream_nativeRecv(env *C.JNIEnv, jStream C.jobject, goPtr C.jlong) C.jbyteArray {
	result := new(vdl.Value)
	if err := (*(*rpc.Stream)(jutil.Ptr(goPtr))).Recv(&result); err != nil {
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

//export Java_io_v_impl_google_rpc_Stream_nativeFinalize
func Java_io_v_impl_google_rpc_Stream_nativeFinalize(env *C.JNIEnv, jStream C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*rpc.Stream)(jutil.Ptr(goPtr)))
}

//export Java_io_v_impl_google_rpc_Call_nativeCloseSend
func Java_io_v_impl_google_rpc_Call_nativeCloseSend(env *C.JNIEnv, jCall C.jobject, goPtr C.jlong) {
	if err := (*(*rpc.ClientCall)(jutil.Ptr(goPtr))).CloseSend(); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_v_impl_google_rpc_Call_nativeFinish
func Java_io_v_impl_google_rpc_Call_nativeFinish(env *C.JNIEnv, jCall C.jobject, goPtr C.jlong, jNumResults C.jint) C.jobjectArray {
	// Have all the results be decoded into *vdl.Value.
	numResults := int(jNumResults)
	resultPtrs := make([]interface{}, numResults)
	for i := 0; i < numResults; i++ {
		value := new(vdl.Value)
		resultPtrs[i] = &value
	}
	if err := (*(*rpc.ClientCall)(jutil.Ptr(goPtr))).Finish(resultPtrs...); err != nil {
		// Invocation error.
		jutil.JThrowV(env, err)
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

//export Java_io_v_impl_google_rpc_Call_nativeFinalize
func Java_io_v_impl_google_rpc_Call_nativeFinalize(env *C.JNIEnv, jCall C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*rpc.ClientCall)(jutil.Ptr(goPtr)))
}

//export Java_io_v_impl_google_rpc_StreamServerCall_nativeBlessings
func Java_io_v_impl_google_rpc_StreamServerCall_nativeBlessings(env *C.JNIEnv, jStreamServerCall C.jobject, goPtr C.jlong) C.jobject {
	blessings := (*(*rpc.StreamServerCall)(jutil.Ptr(goPtr))).GrantedBlessings()
	jBlessings, err := jsecurity.JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jBlessings)
}

//export Java_io_v_impl_google_rpc_StreamServerCall_nativeFinalize
func Java_io_v_impl_google_rpc_StreamServerCall_nativeFinalize(env *C.JNIEnv, jStreamServerCall C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*rpc.StreamServerCall)(jutil.Ptr(goPtr)))
}
