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
	wire "v.io/v23/services/syncbase"
	"v.io/x/ref/lib/dispatcher"
	"v.io/x/ref/services/syncbase/discovery"
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
	idSign            = jutil.ClassSign("io.v.v23.services.syncbase.Id")
	inviteSign        = jutil.ClassSign("io.v.v23.syncbase.Invite")

	jVRuntimeImplClass jutil.Class
	jInviteClass       jutil.Class
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
	jInviteClass, err = jutil.JFindClass(env, "io/v/v23/syncbase/Invite")
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

	// Create the rpc server before the service so that connections are shared between
	// clients in the service and the rpc server. (i.e. connections are shared if the
	// context returned from WithNewDispatchingServer is used for client calls).
	d := dispatcher.NewDispatcherWrapper()
	ctx, s, err := v23.WithNewDispatchingServer(ctx, name, d, options.ChannelTimeout(vsync.NeighborConnectionTimeout))
	if err != nil {
		cancel()
		jutil.JThrowV(env, err)
		return nil
	}

	service, err := server.NewService(ctx, server.ServiceOptions{
		Perms:   perms,
		RootDir: rootDir,
		Engine:  engine,
	})
	if err != nil {
		cancel()
		jutil.JThrowV(env, err)
		return nil
	}
	// Set the dispatcher in the dispatcher wrapper for the server to start responding
	// to incoming rpcs.
	d.SetDispatcher(server.NewDispatcher(service))

	if err := service.AddNames(ctx, s); err != nil {
		cancel()
		jutil.JThrowV(env, err)
		return nil
	}
	jNewCtx, err := jcontext.JavaContext(env, ctx, cancel)
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

//export Java_io_v_v23_syncbase_DatabaseImpl_nativeListenForInvites
func Java_io_v_v23_syncbase_DatabaseImpl_nativeListenForInvites(jenv *C.JNIEnv, jDatabase C.jobject, jContext C.jobject, jInviteHandler C.jobject) {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	database := jutil.Object(uintptr(unsafe.Pointer(jDatabase)))
	handler := jutil.Object(uintptr(unsafe.Pointer(jInviteHandler)))

	ctx, _, err := jcontext.GoContext(env, jutil.Object(uintptr(unsafe.Pointer(jContext))))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}

	jDbId, err := jutil.CallObjectMethod(env, database, "id", nil, idSign)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	dbName, err := jutil.CallStringMethod(env, jDbId, "getName", nil)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	dbBlessing, err := jutil.CallStringMethod(env, jDbId, "getBlessing", nil)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}

	dbId := wire.Id{Name: dbName, Blessing: dbBlessing}
	// Note: There is no need to use a buffered channel here.  ListenForInvites
	// spawns a goroutine for this listener that acts as an infinite buffer so we can
	// process invites at our own pace.
	ch := make(chan discovery.Invite)
	if err := discovery.ListenForInvites(ctx, dbId, ch); err != nil {
		jutil.JThrowV(env, err)
		return
	}

	handler = jutil.NewGlobalRef(env, handler)
	go func() {
		for invite := range ch {
			if err := handleInvite(invite, handler); err != nil {
				// TODO(mattr): We should cancel the stream and return an error to
				// the user here.
				ctx.Errorf("couldn't call invite handler: %v", err)
			}
		}
		env, free := jutil.GetEnv()
		jutil.DeleteGlobalRef(env, handler)
		free()
	}()
}

func handleInvite(invite discovery.Invite, handler jutil.Object) error {
	env, free := jutil.GetEnv()
	defer free()
	jInvite, err := jutil.NewObject(env, jInviteClass,
		[]jutil.Sign{jutil.StringSign, jutil.StringSign},
		invite.Syncgroup.Blessing, invite.Syncgroup.Name)
	if err != nil {
		return err
	}
	return jutil.CallVoidMethod(env, handler, "handleInvite",
		[]jutil.Sign{inviteSign}, jInvite)
}
