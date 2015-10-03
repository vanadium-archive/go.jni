// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package namespace

import (
	"log"
	"time"
	"unsafe"

	"v.io/v23/context"
	"v.io/v23/namespace"
	"v.io/v23/naming"
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

func globArgs(env jutil.Env, jContext C.jobject, jPattern C.jstring, jOptions C.jobject) (context *context.T, pattern string, opts []naming.NamespaceOpt, err error) {
	context, err = jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		return
	}
	opts, err = namespaceOptions(env, jutil.WrapObject(jOptions))
	if err != nil {
		return
	}
	pattern = jutil.GoString(env, jutil.WrapObject(jPattern))
	return
}

func doGlob(env jutil.Env, n namespace.T, context *context.T, pattern string, opts []naming.NamespaceOpt) (jutil.Object, error) {
	entryChan, err := n.Glob(context, pattern, opts...)
	if err != nil {
		return jutil.NullObject, err
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
		return jutil.NullObject, err
	}
	return jIterable, nil
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeGlob
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeGlob(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jPattern C.jstring, jOptions C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	context, pattern, opts, err := globArgs(env, jContext, jPattern, jOptions)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jIterable, err := doGlob(env, n, context, pattern, opts)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jIterable))
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeGlobAsync
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeGlobAsync(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jPattern C.jstring, jOptions C.jobject, jCallback C.jobject) {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	callback := jutil.WrapObject(jCallback)
	context, pattern, opts, err := globArgs(env, jContext, jPattern, jOptions)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	jutil.DoAsyncCall(env, callback, func(env jutil.Env) (jutil.Object, error) {
		return doGlob(env, n, context, pattern, opts)
	})
}

func mountArgs(env jutil.Env, jContext C.jobject, jName, jServer C.jstring, jDuration, jOptions C.jobject) (context *context.T, name, server string, duration time.Duration, options []naming.NamespaceOpt, err error) {
	context, err = jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		return
	}
	name = jutil.GoString(env, jutil.WrapObject(jName))
	server = jutil.GoString(env, jutil.WrapObject(jServer))
	duration, err = jutil.GoDuration(env, jutil.WrapObject(jDuration))
	if err != nil {
		return
	}
	options, err = namespaceOptions(env, jutil.WrapObject(jOptions))
	return
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeMount
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeMount(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName, jServer C.jstring, jDuration C.jobject, jOptions C.jobject) {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	context, name, server, duration, options, err := mountArgs(env, jContext, jName, jServer, jDuration, jOptions)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := n.Mount(context, name, server, duration, options...); err != nil {
		jutil.JThrowV(env, err)
	}
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeMountAsync
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeMountAsync(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jServer C.jstring, jDuration C.jobject, jOptions C.jobject, jCallback C.jobject) {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	callback := jutil.WrapObject(jCallback)
	context, name, server, duration, options, err := mountArgs(env, jContext, jName, jServer, jDuration, jOptions)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	jutil.DoAsyncCall(env, callback, func(env jutil.Env) (jutil.Object, error) {
		return jutil.NullObject, n.Mount(context, name, server, duration, options...)
	})
}

func unmountArgs(env jutil.Env, jName, jServer C.jstring, jContext, jOptions C.jobject) (name, server string, context *context.T, options []naming.NamespaceOpt, err error) {
	name = jutil.GoString(env, jutil.WrapObject(jName))
	server = jutil.GoString(env, jutil.WrapObject(jServer))
	context, err = jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		return
	}
	options, err = namespaceOptions(env, jutil.WrapObject(jOptions))
	return
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeUnmount
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeUnmount(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jServer C.jstring, jOptions C.jobject) {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	name, server, context, options, err := unmountArgs(env, jName, jServer, jContext, jOptions)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := n.Unmount(context, name, server, options...); err != nil {
		jutil.JThrowV(env, err)
	}
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeUnmountAsync
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeUnmountAsync(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jServer C.jstring, jOptions C.jobject, jCallback C.jobject) {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	callback := jutil.WrapObject(jCallback)
	name, server, context, options, err := unmountArgs(env, jName, jServer, jContext, jOptions)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	jutil.DoAsyncCall(env, callback, func(env jutil.Env) (jutil.Object, error) {
		return jutil.NullObject, n.Unmount(context, name, server, options...)
	})
}

func deleteArgs(env jutil.Env, jContext, jOptions C.jobject, jName C.jstring, jDeleteSubtree C.jboolean) (context *context.T, options []naming.NamespaceOpt, name string, deleteSubtree bool, err error) {
	context, err = jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		return
	}
	options, err = namespaceOptions(env, jutil.WrapObject(jOptions))
	if err != nil {
		return
	}
	name = jutil.GoString(env, jutil.WrapObject(jName))
	deleteSubtree = jDeleteSubtree == C.JNI_TRUE
	return
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeDelete
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeDelete(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jDeleteSubtree C.jboolean, jOptions C.jobject) {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	context, options, name, deleteSubtree, err := deleteArgs(env, jContext, jOptions, jName, jDeleteSubtree)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := n.Delete(context, name, deleteSubtree, options...); err != nil {
		jutil.JThrowV(env, err)
	}
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeDeleteAsync
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeDeleteAsync(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jDeleteSubtree C.jboolean, jOptions C.jobject, jCallback C.jobject) {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	callback := jutil.WrapObject(jCallback)
	context, options, name, deleteSubtree, err := deleteArgs(env, jContext, jOptions, jName, jDeleteSubtree)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	jutil.DoAsyncCall(env, callback, func(env jutil.Env) (jutil.Object, error) {
		return jutil.NullObject, n.Delete(context, name, deleteSubtree, options...)
	})
}

func resolveArgs(env jutil.Env, jName C.jstring, jContext, jOptions C.jobject) (context *context.T, name string, options []naming.NamespaceOpt, err error) {
	context, err = jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	options, err = namespaceOptions(env, jutil.WrapObject(jOptions))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	name = jutil.GoString(env, jutil.WrapObject(jName))
	return
}

func doResolve(env jutil.Env, n namespace.T, context *context.T, name string, options []naming.NamespaceOpt) (jutil.Object, error) {
	entry, err := n.Resolve(context, name, options...)
	if err != nil {
		return jutil.NullObject, err
	}
	return jutil.JVomCopy(env, entry, jMountEntryClass)
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeResolve
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeResolve(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jOptions C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	context, name, options, err := resolveArgs(env, jName, jContext, jOptions)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jEntry, err := doResolve(env, n, context, name, options)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jEntry))
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeResolveAsync
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeResolveAsync(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jOptions C.jobject, jCallback C.jobject) {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	callback := jutil.WrapObject(jCallback)
	context, name, options, err := resolveArgs(env, jName, jContext, jOptions)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	jutil.DoAsyncCall(env, callback, func(env jutil.Env) (jutil.Object, error) {
		return doResolve(env, n, context, name, options)
	})
}

func resolveToMountTableArgs(env jutil.Env, jContext, jOptions C.jobject, jName C.jstring) (context *context.T, options []naming.NamespaceOpt, name string, err error) {
	context, err = jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	options, err = namespaceOptions(env, jutil.WrapObject(jOptions))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	name = jutil.GoString(env, jutil.WrapObject(jName))
	return
}

func doResolveToMountTable(env jutil.Env, n namespace.T, context *context.T, name string, options []naming.NamespaceOpt) (jutil.Object, error) {
	entry, err := n.ResolveToMountTable(context, name, options...)
	if err != nil {
		return jutil.NullObject, err
	}
	return jutil.JVomCopy(env, entry, jMountEntryClass)
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeResolveToMountTable
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeResolveToMountTable(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jOptions C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	context, options, name, err := resolveToMountTableArgs(env, jContext, jOptions, jName)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jEntry, err := doResolveToMountTable(env, n, context, name, options)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jEntry))
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeResolveToMountTableAsync
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeResolveToMountTableAsync(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jOptions C.jobject, jCallback C.jobject) {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	callback := jutil.WrapObject(jCallback)
	context, options, name, err := resolveToMountTableArgs(env, jContext, jOptions, jName)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	jutil.DoAsyncCall(env, callback, func(env jutil.Env) (jutil.Object, error) {
		return doResolveToMountTable(env, n, context, name, options)
	})
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

func setPermissionsArgs(env jutil.Env, jContext, jPermissions C.jobject, jName, jVersion C.jstring, jOptions C.jobject) (context *context.T, permissions access.Permissions, name, version string, options []naming.NamespaceOpt, err error) {
	context, err = jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		return
	}
	err = jutil.GoVomCopy(env, jutil.WrapObject(jPermissions), jPermissionsClass, &permissions)
	if err != nil {
		return
	}
	options, err = namespaceOptions(env, jutil.WrapObject(jOptions))
	name = jutil.GoString(env, jutil.WrapObject(jName))
	version = jutil.GoString(env, jutil.WrapObject(jVersion))
	return
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeSetPermissions
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeSetPermissions(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jPermissions C.jobject, jVersion C.jstring, jOptions C.jobject) {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	context, permissions, name, version, options, err := setPermissionsArgs(env, jContext, jPermissions, jName, jVersion, jOptions)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := n.SetPermissions(context, name, permissions, version, options...); err != nil {
		jutil.JThrowV(env, err)
	}
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeSetPermissionsAsync
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeSetPermissionsAsync(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jPermissions C.jobject, jVersion C.jstring, jOptions C.jobject, jCallback C.jobject) {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	callback := jutil.WrapObject(jCallback)
	context, permissions, name, version, options, err := setPermissionsArgs(env, jContext, jPermissions, jName, jVersion, jOptions)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	jutil.DoAsyncCall(env, callback, func(env jutil.Env) (jutil.Object, error) {
		return jutil.NullObject, n.SetPermissions(context, name, permissions, version, options...)
	})
}

func getPermissionsArgs(env jutil.Env, jContext C.jobject, jName C.jstring, jOptions C.jobject) (context *context.T, name string, options []naming.NamespaceOpt, err error) {
	context, err = jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		return
	}
	options, err = namespaceOptions(env, jutil.WrapObject(jOptions))
	if err != nil {
		return
	}
	name = jutil.GoString(env, jutil.WrapObject(jName))
	return
}

func doGetPermissions(env jutil.Env, n namespace.T, context *context.T, name string, options []naming.NamespaceOpt) (jutil.Object, error) {
	permissions, version, err := n.GetPermissions(context, name, options...)
	if err != nil {
		return jutil.NullObject, err
	}
	jPermissions, err := jutil.JVomCopy(env, permissions, jPermissionsClass)
	if err != nil {
		return jutil.NullObject, err
	}
	result := make(map[jutil.Object]jutil.Object)
	result[jutil.JString(env, version)] = jPermissions
	return jutil.JObjectMap(env, result)
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeGetPermissions
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeGetPermissions(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jOptions C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	context, name, options, err := getPermissionsArgs(env, jContext, jName, jOptions)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jResult, err := doGetPermissions(env, n, context, name, options)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jResult))
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeGetPermissionsAsync
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeGetPermissionsAsync(jenv *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jOptions C.jobject, jCallback C.jobject) {
	env := jutil.WrapEnv(jenv)
	n := *(*namespace.T)(jutil.NativePtr(goNamespacePtr))
	callback := jutil.WrapObject(jCallback)
	context, name, options, err := getPermissionsArgs(env, jContext, jName, jOptions)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	jutil.DoAsyncCall(env, callback, func(env jutil.Env) (jutil.Object, error) {
		return doGetPermissions(env, n, context, name, options)
	})
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeFinalize
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeFinalize(jenv *C.JNIEnv, jNamespace C.jobject, goNamespacePtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goNamespacePtr))
}
