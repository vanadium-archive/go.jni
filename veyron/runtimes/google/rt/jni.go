// +build android

package rt

import (
	"unsafe"

	jutil "veyron.io/jni/util"
	jipc "veyron.io/jni/veyron/runtimes/google/ipc"
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

//export Java_io_veyron_veyron_veyron_runtimes_google_VRuntime_nativeInit
func Java_io_veyron_veyron_veyron_runtimes_google_VRuntime_nativeInit(env *C.JNIEnv, jRuntime C.jclass, jOptions C.jobject) C.jlong {
	opts, err := getRuntimeOpts(env, jOptions)
	if err != nil {
		jutil.JThrowV(env, err)
		return C.jlong(0)
	}
	r := rt.Init(opts...)
	jutil.GoRef(&r) // Un-refed when the Java Runtime object is finalized.
	return C.jlong(jutil.PtrValue(&r))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_VRuntime_nativeNewRuntime
func Java_io_veyron_veyron_veyron_runtimes_google_VRuntime_nativeNewRuntime(env *C.JNIEnv, jRuntime C.jclass, jOptions C.jobject) C.jlong {
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
	jutil.GoRef(&r)
	return C.jlong(jutil.PtrValue(&r))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_VRuntime_nativeNewClient
func Java_io_veyron_veyron_veyron_runtimes_google_VRuntime_nativeNewClient(env *C.JNIEnv, jRuntime C.jobject, goPtr C.jlong, jOptions C.jobject) C.jobject {
	r := (*veyron2.Runtime)(jutil.Ptr(goPtr))
	// No options supported yet.
	rc, err := (*r).NewClient()
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jClient, err := jipc.JavaClient(env, rc)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jClient)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_VRuntime_nativeNewServer
func Java_io_veyron_veyron_veyron_runtimes_google_VRuntime_nativeNewServer(env *C.JNIEnv, jRuntime C.jobject, goPtr C.jlong, jOptions C.jobject) C.jobject {
	r := (*veyron2.Runtime)(jutil.Ptr(goPtr))
	// No options supported yet.
	s, err := (*r).NewServer()
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jServer, err := jipc.JavaServer(env, s)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jServer)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_VRuntime_nativeGetClient
func Java_io_veyron_veyron_veyron_runtimes_google_VRuntime_nativeGetClient(env *C.JNIEnv, jRuntime C.jobject, goPtr C.jlong) C.jobject {
	r := (*veyron2.Runtime)(jutil.Ptr(goPtr))
	rc := (*r).Client()
	jClient, err := jipc.JavaClient(env, rc)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jClient)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_VRuntime_nativeNewContext
func Java_io_veyron_veyron_veyron_runtimes_google_VRuntime_nativeNewContext(env *C.JNIEnv, jRuntime C.jobject, goPtr C.jlong) C.jobject {
	r := (*veyron2.Runtime)(jutil.Ptr(goPtr))
	c := (*r).NewContext()
	jContext, err := jcontext.JavaContext(env, c, nil)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jContext)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_VRuntime_nativeGetNamespace
func Java_io_veyron_veyron_veyron_runtimes_google_VRuntime_nativeGetNamespace(env *C.JNIEnv, jRuntime C.jobject, goPtr C.jlong) C.jlong {
	r := (*veyron2.Runtime)(jutil.Ptr(goPtr))
	n := (*r).Namespace()
	jutil.GoRef(&n) // Un-refed when the Java Namespace object is finalized.
	return C.jlong(jutil.PtrValue(&n))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_VRuntime_nativeFinalize
func Java_io_veyron_veyron_veyron_runtimes_google_VRuntime_nativeFinalize(env *C.JNIEnv, jRuntime C.jobject, goPtr C.jlong) {
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
