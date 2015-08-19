// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package namespace

import (
	"log"
	"unsafe"

	"v.io/v23/namespace"
	"v.io/v23/security/access"
	jchannel "v.io/x/jni/impl/google/channel"
	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
)

// #include "jni.h"
import "C"

var (
	// Global reference for io.v.impl.google.namespace.NamespaceImpl class.
	jNamespaceImplClass jutil.Class
	// Global reference for io.v.v23.naming.GlobReply class.
	jGlobReplyClass jutil.Class
	// Global reference for io.v.v23.naming.MountEntry class.
	jMountEntryClass jutil.Class
	// Global reference for io.v.v23.security.access.Permissions
	jPermissionsClass jutil.Class
)

// Init initializes the JNI code with the given Java environment. This method
// must be called from the main Java thread.
// interface and then cast into the package-local environment type.
func Init(env jutil.Env) error {
	var err error
	jNamespaceImplClass, err = jutil.JFindClass(env, "io/v/impl/google/namespace/NamespaceImpl")
	if err != nil {
		return err
	}
	jGlobReplyClass, err = jutil.JFindClass(env, "io/v/v23/naming/GlobReply")
	if err != nil {
		return err
	}
	jMountEntryClass, err = jutil.JFindClass(env, "io/v/v23/naming/MountEntry")
	if err != nil {
		return err
	}
	jPermissionsClass, err = jutil.JFindClass(env, "io/v/v23/security/access/Permissions")
	if err != nil {
		return err
	}
	return nil
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeGlob
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeGlob(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jPattern C.jstring, jOptions C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	context, err := jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	opts, err := namespaceOptions(env, jutil.WrapObject(jOptions))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	entryChan, err := n.Glob(context, jutil.GoString(env, jutil.WrapObject(jPattern)), opts...)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}

	retChan := make(chan jutil.Object, 100)
	go func() {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()

		for globReply := range entryChan {
			jGlobReply, err := jutil.JVomCopy(env, globReply, jGlobReplyClass)
			if err != nil {
				log.Printf("Couldn't convert Go glob result %v to Java\n", globReply)
				continue
			}
			// The other side of the channel is responsible for deleting this
			// global reference.
			retChan <- jutil.NewGlobalRef(env, jGlobReply)
		}
		close(retChan)
	}()
	jIterable, err := jchannel.JavaIterable(env, &retChan, &entryChan)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jIterable))
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeMount
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeMount(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jServer C.jstring, jDuration C.jobject, jOptions C.jobject) {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	context, err := jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	duration, err := jutil.GoDuration(env, jutil.WrapObject(jDuration))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	options, err := namespaceOptions(env, jutil.WrapObject(jOptions))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := n.Mount(context, jutil.GoString(env, jutil.WrapObject(jName)), jutil.GoString(env, jutil.WrapObject(jServer)), duration, options...); err != nil {
		jutil.JThrowV(env, err)
	}
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeUnmount
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeUnmount(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jServer C.jstring, jOptions C.jobject) {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	context, err := jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	options, err := namespaceOptions(env, jutil.WrapObject(jOptions))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := n.Unmount(context, jutil.GoString(env, jutil.WrapObject(jName)), jutil.GoString(env, jutil.WrapObject(jServer)), options...); err != nil {
		jutil.JThrowV(env, err)
	}
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeDelete
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeDelete(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jDeleteSubtree C.jboolean, jOptions C.jobject) {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	context, err := jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	options, err := namespaceOptions(env, jutil.WrapObject(jOptions))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := n.Delete(context, jutil.GoString(env, jutil.WrapObject(jName)), jDeleteSubtree == C.JNI_TRUE, options...); err != nil {
		jutil.JThrowV(env, err)
	}
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeResolve
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeResolve(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jOptions C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	context, err := jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	options, err := namespaceOptions(env, jutil.WrapObject(jOptions))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	entry, err := n.Resolve(context, jutil.GoString(env, jutil.WrapObject(jName)), options...)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jEntry, err := jutil.JVomCopy(env, entry, jMountEntryClass)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jEntry))
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeResolveToMountTable
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeResolveToMountTable(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jOptions C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	context, err := jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	options, err := namespaceOptions(env, jutil.WrapObject(jOptions))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	entry, err := n.ResolveToMountTable(context, jutil.GoString(env, jutil.WrapObject(jName)), options...)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jEntry, err := jutil.JVomCopy(env, entry, jMountEntryClass)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jEntry))
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeFlushCacheEntry
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeFlushCacheEntry(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName string) C.jboolean {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	context, err := jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
		return C.JNI_FALSE
	}
	result := n.FlushCacheEntry(context, jutil.GoString(env, jutil.WrapObject(jName)))
	if result {
		return C.JNI_TRUE
	} else {
		return C.JNI_FALSE
	}
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeSetRoots
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeSetRoots(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jNames C.jobject) {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	names, err := jutil.GoStringList(env, jutil.WrapObject(jNames))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	err = n.SetRoots(names...)
	if err != nil {
		jutil.JThrowV(env, err)
	}
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeSetPermissions
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeSetPermissions(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jPermissions C.jobject, jVersion C.jstring, jOptions C.jobject) {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	context, err := jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	var permissions access.Permissions
	err = jutil.GoVomCopy(env, jutil.WrapObject(jPermissions), jPermissionsClass, &permissions)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	options, err := namespaceOptions(env, jutil.WrapObject(jOptions))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := n.SetPermissions(context, jutil.GoString(env, jutil.WrapObject(jName)), permissions, jutil.GoString(env, jutil.WrapObject(jVersion)), options...); err != nil {
		jutil.JThrowV(env, err)
	}
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeGetPermissions
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeGetPermissions(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jOptions C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	context, err := jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	options, err := namespaceOptions(env, jutil.WrapObject(jOptions))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	permissions, version, err := n.GetPermissions(context, jutil.GoString(env, jutil.WrapObject(jName)), options...)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jPermissions, err := jutil.JVomCopy(env, permissions, jPermissionsClass)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	result := make(map[jutil.Object]jutil.Object)
	result[jutil.JString(env, version)] = jPermissions
	jResult, err := jutil.JObjectMap(env, result)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jResult))
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeFinalize
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeFinalize(jenv *C.JNIEnv, jNamespace C.jobject, goNamespacePtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goNamespacePtr))
}
