// +build android

package rt

import (
	"unsafe"

	jutil "veyron.io/jni/util"
	jipc "veyron.io/jni/veyron/runtimes/google/ipc"
	jnaming "veyron.io/jni/veyron/runtimes/google/naming"
	jcontext "veyron.io/jni/veyron2/context"
	jsecurity "veyron.io/jni/veyron2/security"

	"veyron.io/veyron/veyron2"
	"veyron.io/veyron/veyron2/options"
	"veyron.io/veyron/veyron2/rt"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

var (
	// Global reference for io.veyron.veyron.veyron2.OptionDefs class.
	jOptionDefsClass C.jclass
)

// Init initializes the JNI code with the given Java environment.  This method
// must be invoked before any other method in this package and must be called
// from the main Java thread (e.g., On_Load()).
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java environment is passed in an empty
// interface and then cast into the package-local environment type.
func Init(jEnv interface{}) {
	env := (*C.JNIEnv)(unsafe.Pointer(jutil.PtrValue(jEnv)))
	jOptionDefsClass = C.jclass(jutil.JFindClassOrPrint(env, "io/veyron/veyron/veyron2/OptionDefs"))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_VRuntimeImpl_nativeInit
func Java_io_veyron_veyron_veyron_runtimes_google_VRuntimeImpl_nativeInit(env *C.JNIEnv, jRuntime C.jclass, jOptions C.jobject) C.jlong {
	opts, err := getRuntimeOpts(env, jOptions)
	if err != nil {
		jutil.JThrowV(env, err)
		return C.jlong(0)
	}
	r, err := rt.New(opts...)
	if err != nil {
		jutil.JThrowV(env, err)
		return C.jlong(0)
	}
	jutil.GoRef(&r) // Un-refed when the Java Runtime object is finalized.
	return C.jlong(jutil.PtrValue(&r))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_VRuntimeImpl_nativeNewClient
func Java_io_veyron_veyron_veyron_runtimes_google_VRuntimeImpl_nativeNewClient(env *C.JNIEnv, jRuntime C.jobject, goPtr C.jlong, jOptions C.jobject) C.jobject {
	// No options supported yet.
	client, err := (*(*veyron2.Runtime)(jutil.Ptr(goPtr))).NewClient()
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jClient, err := jipc.JavaClient(env, client)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jClient)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_VRuntimeImpl_nativeNewServer
func Java_io_veyron_veyron_veyron_runtimes_google_VRuntimeImpl_nativeNewServer(env *C.JNIEnv, jRuntime C.jobject, goPtr C.jlong, jOptions C.jobject) C.jobject {
	// No options supported yet.
	server, err := (*(*veyron2.Runtime)(jutil.Ptr(goPtr))).NewServer()
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jServer, err := jipc.JavaServer(env, server)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jServer)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_VRuntimeImpl_nativeGetClient
func Java_io_veyron_veyron_veyron_runtimes_google_VRuntimeImpl_nativeGetClient(env *C.JNIEnv, jRuntime C.jobject, goPtr C.jlong) C.jobject {
	client := (*(*veyron2.Runtime)(jutil.Ptr(goPtr))).Client()
	jClient, err := jipc.JavaClient(env, client)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jClient)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_VRuntimeImpl_nativeNewContext
func Java_io_veyron_veyron_veyron_runtimes_google_VRuntimeImpl_nativeNewContext(env *C.JNIEnv, jRuntime C.jobject, goPtr C.jlong) C.jobject {
	context := (*(*veyron2.Runtime)(jutil.Ptr(goPtr))).NewContext()
	jContext, err := jcontext.JavaContext(env, context, nil)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jContext)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_VRuntimeImpl_nativeGetPrincipal
func Java_io_veyron_veyron_veyron_runtimes_google_VRuntimeImpl_nativeGetPrincipal(env *C.JNIEnv, jRuntime C.jobject, goPtr C.jlong) C.jobject {
	principal := (*(*veyron2.Runtime)(jutil.Ptr(goPtr))).Principal()
	jPrincipal, err := jsecurity.JavaPrincipal(env, principal)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jPrincipal)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_VRuntimeImpl_nativeGetNamespace
func Java_io_veyron_veyron_veyron_runtimes_google_VRuntimeImpl_nativeGetNamespace(env *C.JNIEnv, jRuntime C.jobject, goPtr C.jlong) C.jobject {
	namespace := (*(*veyron2.Runtime)(jutil.Ptr(goPtr))).Namespace()
	jNamespace, err := jnaming.JavaNamespace(env, namespace)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jNamespace)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_VRuntimeImpl_nativeFinalize
func Java_io_veyron_veyron_veyron_runtimes_google_VRuntimeImpl_nativeFinalize(env *C.JNIEnv, jRuntime C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*veyron2.Runtime)(jutil.Ptr(goPtr)))
}

// getRuntimeOpts converts Java runtime options into Go runtime options.
func getRuntimeOpts(env *C.JNIEnv, jOptions C.jobject) (ret []veyron2.ROpt, err error) {
	if jOptions == nil {
		return
	}
	runtimePrincipalKey, err := jutil.JStaticStringField(env, jOptionDefsClass, "RUNTIME_PRINCIPAL")
	if err != nil {
		return nil, err
	}
	if has, err := jutil.CallBooleanMethod(env, jOptions, "has", []jutil.Sign{jutil.StringSign}, runtimePrincipalKey); err != nil {
		return nil, err
	} else if has {
		jPrincipal, err := jutil.CallObjectMethod(env, jOptions, "get", []jutil.Sign{jutil.StringSign}, jutil.ObjectSign, runtimePrincipalKey)
		if err != nil {
			return nil, err
		}
		principal, err := jsecurity.GoPrincipal(env, jPrincipal)
		if err != nil {
			return nil, err
		}
		ret = append(ret, options.RuntimePrincipal{principal})
	}
	return
}
