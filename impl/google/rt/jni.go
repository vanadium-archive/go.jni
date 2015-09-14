// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package rt

import (
	"runtime"
	"unsafe"

	"v.io/v23"
	"v.io/v23/context"
	_ "v.io/x/ref/runtime/factories/roaming"

	jns "v.io/x/jni/impl/google/namespace"
	jrpc "v.io/x/jni/impl/google/rpc"
	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
	jsecurity "v.io/x/jni/v23/security"
)

// #include "jni.h"
import "C"

var (
	contextSign = jutil.ClassSign("io.v.v23.context.VContext")
	serverSign = jutil.ClassSign("io.v.v23.rpc.Server")

	jVRuntimeImplClass jutil.Class
)

// Init initializes the JNI code with the given Java environment.  This method
// must be invoked before any other method in this package and must be called
// from the main Java thread (e.g., On_Load()).
// interface and then cast into the package-local environment type.
func Init(env jutil.Env) error {
	var err error
	jVRuntimeImplClass, err = jutil.JFindClass(env, "io/v/impl/google/rt/VRuntimeImpl")
	if err != nil {
		return err
	}
	return nil
}

type shutdownKey struct{}

//export Java_io_v_impl_google_rt_VRuntimeImpl_nativeInit
func Java_io_v_impl_google_rt_VRuntimeImpl_nativeInit(jenv *C.JNIEnv, jRuntime C.jclass, jNumCpus C.jint) C.jobject {
	env := jutil.WrapEnv(jenv)
	runtime.GOMAXPROCS(int(jNumCpus))
	ctx, shutdownFunc := v23.Init()
	ctx = context.WithValue(ctx, shutdownKey{}, shutdownFunc)
	jCtx, err := jcontext.JavaContext(env, ctx, nil)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jCtx))
}

//export Java_io_v_impl_google_rt_VRuntimeImpl_nativeShutdown
func Java_io_v_impl_google_rt_VRuntimeImpl_nativeShutdown(jenv *C.JNIEnv, jRuntime C.jclass, jContext C.jobject) {
	env := jutil.WrapEnv(jenv)
	ctx, err := jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
	}
	value := ctx.Value(shutdownKey{})
	if shutdownFunc, ok := value.(v23.Shutdown); ok {
		shutdownFunc()
	}
}

//export Java_io_v_impl_google_rt_VRuntimeImpl_nativeWithNewClient
func Java_io_v_impl_google_rt_VRuntimeImpl_nativeWithNewClient(jenv *C.JNIEnv, jRuntime C.jclass, jContext C.jobject, jOptions C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	// TODO(spetrovic): Have Java context support nativePtr()?
	ctx, err := jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	// No options supported yet.
	newCtx, _, err := v23.WithNewClient(ctx)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jNewCtx, err := jcontext.JavaContext(env, newCtx, nil)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jNewCtx))
}

//export Java_io_v_impl_google_rt_VRuntimeImpl_nativeGetClient
func Java_io_v_impl_google_rt_VRuntimeImpl_nativeGetClient(jenv *C.JNIEnv, jRuntime C.jclass, jContext C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	// TODO(spetrovic): Have Java context support nativePtr()?
	ctx, err := jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	client := v23.GetClient(ctx)
	jClient, err := jrpc.JavaClient(env, client)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jClient))
}

//export Java_io_v_impl_google_rt_VRuntimeImpl_nativeWithNewServer
func Java_io_v_impl_google_rt_VRuntimeImpl_nativeWithNewServer(jenv *C.JNIEnv, jRuntime C.jclass, jContext C.jobject, jName C.jstring, jDispatcher C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	// TODO(spetrovic): Have Java context support nativePtr()?
	ctx, err := jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	name := jutil.GoString(env, jutil.WrapObject(jName))
	d, err := jrpc.GoDispatcher(env, jutil.WrapObject(jDispatcher))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	newCtx, server, err := v23.WithNewDispatchingServer(ctx, name, d)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jServer, err := jrpc.JavaServer(env, server)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jNewCtx, err := jcontext.JavaContext(env, newCtx, nil)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	// Explicitly attach a server to the new context.
	jServerAttCtx, err := jutil.CallStaticObjectMethod(env, jVRuntimeImplClass, "withServer", []jutil.Sign{contextSign, serverSign}, contextSign, jNewCtx, jServer)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jServerAttCtx))
}

//export Java_io_v_impl_google_rt_VRuntimeImpl_nativeWithPrincipal
func Java_io_v_impl_google_rt_VRuntimeImpl_nativeWithPrincipal(jenv *C.JNIEnv, jRuntime C.jclass, jContext C.jobject, jPrincipal C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	// TODO(spetrovic): Have Java context support nativePtr()?
	ctx, err := jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	principal, err := jsecurity.GoPrincipal(env, jutil.WrapObject(jPrincipal))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	newCtx, err := v23.WithPrincipal(ctx, principal)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jNewCtx, err := jcontext.JavaContext(env, newCtx, nil)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jNewCtx))
}

//export Java_io_v_impl_google_rt_VRuntimeImpl_nativeGetPrincipal
func Java_io_v_impl_google_rt_VRuntimeImpl_nativeGetPrincipal(jenv *C.JNIEnv, jRuntime C.jclass, jContext C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	// TODO(spetrovic): Have Java context support nativePtr()?
	ctx, err := jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	principal := v23.GetPrincipal(ctx)
	jPrincipal, err := jsecurity.JavaPrincipal(env, principal)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jPrincipal))
}

//export Java_io_v_impl_google_rt_VRuntimeImpl_nativeWithNamespace
func Java_io_v_impl_google_rt_VRuntimeImpl_nativeWithNamespace(jenv *C.JNIEnv, jRuntime C.jclass, jContext C.jobject, jRoots C.jobjectArray) C.jobject {
	env := jutil.WrapEnv(jenv)
	// TODO(spetrovic): Have Java context support nativePtr()?
	ctx, err := jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	roots, err := jutil.GoStringArray(env, jutil.WrapObject(jRoots))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}

	newCtx, _, err := v23.WithNewNamespace(ctx, roots...)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jNewCtx, err := jcontext.JavaContext(env, newCtx, nil)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jNewCtx))
}

//export Java_io_v_impl_google_rt_VRuntimeImpl_nativeGetNamespace
func Java_io_v_impl_google_rt_VRuntimeImpl_nativeGetNamespace(jenv *C.JNIEnv, jRuntime C.jclass, jContext C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	// TODO(spetrovic): Have Java context support nativePtr()?
	ctx, err := jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	namespace := v23.GetNamespace(ctx)
	jNamespace, err := jns.JavaNamespace(env, namespace)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jNamespace))
}

//export Java_io_v_impl_google_rt_VRuntimeImpl_nativeWithListenSpec
func Java_io_v_impl_google_rt_VRuntimeImpl_nativeWithListenSpec(jenv *C.JNIEnv, jRuntime C.jclass, jContext C.jobject, jSpec C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	ctx, err := jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	spec, err := jrpc.GoListenSpec(env, jutil.WrapObject(jSpec))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	newCtx := v23.WithListenSpec(ctx, spec)
	jNewCtx, err := jcontext.JavaContext(env, newCtx, nil)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jNewCtx))
}

//export Java_io_v_impl_google_rt_VRuntimeImpl_nativeGetListenSpec
func Java_io_v_impl_google_rt_VRuntimeImpl_nativeGetListenSpec(jenv *C.JNIEnv, jRuntime C.jclass, jContext C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	ctx, err := jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	spec := v23.GetListenSpec(ctx)
	jSpec, err := jrpc.JavaListenSpec(env, spec)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jSpec))
}