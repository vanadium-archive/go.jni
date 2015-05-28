// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package rt

import (
	"runtime"

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

// Init initializes the JNI code with the given Java environment.  This method
// must be invoked before any other method in this package and must be called
// from the main Java thread (e.g., On_Load()).
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java environment is passed in an empty
// interface and then cast into the package-local environment type.
func Init(jEnv interface{}) error {
	return nil
}

type shutdownKey struct{}

//export Java_io_v_impl_google_rt_VRuntimeImpl_nativeInit
func Java_io_v_impl_google_rt_VRuntimeImpl_nativeInit(env *C.JNIEnv, jRuntime C.jclass, jNumCpus C.jint) C.jobject {
	runtime.GOMAXPROCS(int(jNumCpus))
	ctx, shutdownFunc := v23.Init()
	ctx = context.WithValue(ctx, shutdownKey{}, shutdownFunc)
	jCtx, err := jcontext.JavaContext(env, ctx, nil)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jCtx)
}

//export Java_io_v_impl_google_rt_VRuntimeImpl_nativeShutdown
func Java_io_v_impl_google_rt_VRuntimeImpl_nativeShutdown(env *C.JNIEnv, jRuntime C.jclass, jContext C.jobject) {
	ctx, err := jcontext.GoContext(env, jContext)
	if err != nil {
		jutil.JThrowV(env, err)
	}
	value := ctx.Value(shutdownKey{})
	if shutdownFunc, ok := value.(v23.Shutdown); ok {
		shutdownFunc()
	}
}

//export Java_io_v_impl_google_rt_VRuntimeImpl_nativeSetNewClient
func Java_io_v_impl_google_rt_VRuntimeImpl_nativeSetNewClient(env *C.JNIEnv, jRuntime C.jclass, jContext C.jobject, jOptions C.jobject) C.jobject {
	// TODO(spetrovic): Have Java context support nativePtr()?
	ctx, err := jcontext.GoContext(env, jContext)
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
	return C.jobject(jNewCtx)
}

//export Java_io_v_impl_google_rt_VRuntimeImpl_nativeGetClient
func Java_io_v_impl_google_rt_VRuntimeImpl_nativeGetClient(env *C.JNIEnv, jRuntime C.jclass, jContext C.jobject) C.jobject {
	// TODO(spetrovic): Have Java context support nativePtr()?
	ctx, err := jcontext.GoContext(env, jContext)
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
	return C.jobject(jClient)
}

//export Java_io_v_impl_google_rt_VRuntimeImpl_nativeNewServer
func Java_io_v_impl_google_rt_VRuntimeImpl_nativeNewServer(env *C.JNIEnv, jRuntime C.jclass, jContext C.jobject) C.jobject {
	// TODO(spetrovic): Have Java context support nativePtr()?
	ctx, err := jcontext.GoContext(env, jContext)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	server, err := v23.NewServer(ctx)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jServer, err := jrpc.JavaServer(env, server)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jServer)
}

//export Java_io_v_impl_google_rt_VRuntimeImpl_nativeSetPrincipal
func Java_io_v_impl_google_rt_VRuntimeImpl_nativeSetPrincipal(env *C.JNIEnv, jRuntime C.jclass, jContext C.jobject, jPrincipal C.jobject) C.jobject {
	// TODO(spetrovic): Have Java context support nativePtr()?
	ctx, err := jcontext.GoContext(env, jContext)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	principal, err := jsecurity.GoPrincipal(env, jPrincipal)
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
	return C.jobject(jNewCtx)
}

//export Java_io_v_impl_google_rt_VRuntimeImpl_nativeGetPrincipal
func Java_io_v_impl_google_rt_VRuntimeImpl_nativeGetPrincipal(env *C.JNIEnv, jRuntime C.jclass, jContext C.jobject) C.jobject {
	// TODO(spetrovic): Have Java context support nativePtr()?
	ctx, err := jcontext.GoContext(env, jContext)
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
	return C.jobject(jPrincipal)
}

//export Java_io_v_impl_google_rt_VRuntimeImpl_nativeSetNamespace
func Java_io_v_impl_google_rt_VRuntimeImpl_nativeSetNamespace(env *C.JNIEnv, jRuntime C.jclass, jContext C.jobject, jRoots C.jobjectArray) C.jobject {
	// TODO(spetrovic): Have Java context support nativePtr()?
	ctx, err := jcontext.GoContext(env, jContext)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	roots := jutil.GoStringArray(env, jRoots)
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
	return C.jobject(jNewCtx)
}

//export Java_io_v_impl_google_rt_VRuntimeImpl_nativeGetNamespace
func Java_io_v_impl_google_rt_VRuntimeImpl_nativeGetNamespace(env *C.JNIEnv, jRuntime C.jclass, jContext C.jobject) C.jobject {
	// TODO(spetrovic): Have Java context support nativePtr()?
	ctx, err := jcontext.GoContext(env, jContext)
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
	return C.jobject(jNamespace)
}

//export Java_io_v_impl_google_rt_VRuntimeImpl_nativeGetListenSpec
func Java_io_v_impl_google_rt_VRuntimeImpl_nativeGetListenSpec(env *C.JNIEnv, jRuntime C.jclass, jContext C.jobject) C.jobject {
	ctx, err := jcontext.GoContext(env, jContext)
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
	return C.jobject(jSpec)
}
