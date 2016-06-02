// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package syncbase

import (
	"os"
	"path/filepath"
	"unsafe"

	"v.io/v23"
	"v.io/v23/options"
	"v.io/x/ref/services/syncbase/server"
	"v.io/x/ref/services/syncbase/vsync"

	jrpc "v.io/x/jni/impl/google/rpc"
	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
	jaccess "v.io/x/jni/v23/security/access"
)

// #include "jni.h"
import "C"

var (
	permissionsSign   = jutil.ClassSign("io.v.v23.security.access.Permissions")
	contextSign       = jutil.ClassSign("io.v.v23.context.VContext")
	storageEngineSign = jutil.ClassSign("io.v.impl.google.services.syncbase.SyncbaseServer$StorageEngine")
	serverSign        = jutil.ClassSign("io.v.v23.rpc.Server")

	jVRuntimeImplClass jutil.Class
)

// Init initializes the JNI code with the given Java environment.  This method
// must be invoked before any other method in this package and must be called
// from the main Java thread (e.g., On_Load()).
func Init(env jutil.Env) error {
	var err error
	jVRuntimeImplClass, err = jutil.JFindClass(env, "io/v/impl/google/rt/VRuntimeImpl")
	if err != nil {
		return err
	}
	return nil
}

//export Java_io_v_impl_google_services_syncbase_SyncbaseServer_nativeWithNewServer
func Java_io_v_impl_google_services_syncbase_SyncbaseServer_nativeWithNewServer(jenv *C.JNIEnv, jSyncbaseServerClass C.jclass, jContext C.jobject, jSyncbaseServerParams C.jobject) C.jobject {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	jCtx := jutil.Object(uintptr(unsafe.Pointer(jContext)))
	jParams := jutil.Object(uintptr(unsafe.Pointer(jSyncbaseServerParams)))

	// Read and translate all of the server params.
	jPerms, err := jutil.CallObjectMethod(env, jParams, "getPermissions", nil, permissionsSign)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	perms, err := jaccess.GoPermissions(env, jPerms)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	name, err := jutil.CallStringMethod(env, jParams, "getName", nil)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	rootDir, err := jutil.CallStringMethod(env, jParams, "getStorageRootDir", nil)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	if rootDir == "" {
		rootDir = filepath.Join(os.TempDir(), "syncbaseserver")
		if err := os.Mkdir(rootDir, 0755); err != nil && !os.IsExist(err) {
			jutil.JThrowV(env, err)
			return nil
		}
	}
	jEngine, err := jutil.CallObjectMethod(env, jParams, "getStorageEngine", nil, storageEngineSign)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	engine, err := GoStorageEngine(env, jEngine)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	ctx, cancel, err := jcontext.GoContext(env, jCtx)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}

	// Start the server.
	service, err := server.NewService(ctx, server.ServiceOptions{
		Perms:   perms,
		RootDir: rootDir,
		Engine:  engine,
	})
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	d := server.NewDispatcher(service)
	newCtx, s, err := v23.WithNewDispatchingServer(ctx, name, d, options.ChannelTimeout(vsync.NeighborConnectionTimeout))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	if err := service.AddNames(ctx, s); err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jNewCtx, err := jcontext.JavaContext(env, newCtx, cancel)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jServer, err := jrpc.JavaServer(env, s)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	// Attach a server to the new context.
	jServerAttCtx, err := jutil.CallStaticObjectMethod(env, jVRuntimeImplClass, "withServer", []jutil.Sign{contextSign, serverSign}, contextSign, jNewCtx, jServer)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jServerAttCtx))
}
