// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package namespace

import (
	"log"

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
	jNamespaceImplClass C.jclass
	// Global reference for io.v.v23.naming.GlobReply class.
	jGlobReplyClass C.jclass
	// Global reference for io.v.v23.naming.MountEntry class.
	jMountEntryClass C.jclass
	// Global reference for io.v.v23.security.access.Permissions
	jPermissionsClass C.jclass
)

// Init initializes the JNI code with the given Java environment. This method
// must be called from the main Java thread.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java environment is passed in an empty
// interface and then cast into the package-local environment type.
func Init(jEnv interface{}) error {
	class, err := jutil.JFindClass(jEnv, "io/v/impl/google/namespace/NamespaceImpl")
	if err != nil {
		return err
	}
	jNamespaceImplClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/naming/GlobReply")
	if err != nil {
		return err
	}
	jGlobReplyClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/naming/MountEntry")
	if err != nil {
		return err
	}
	jMountEntryClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/security/access/Permissions")
	if err != nil {
		return err
	}
	jPermissionsClass = C.jclass(class)
	return nil
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeGlob
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeGlob(env *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, pattern C.jstring, jOptions C.jobject) C.jobject {
	n := *(*namespace.T)(jutil.Ptr(goNamespacePtr))
	context, err := jcontext.GoContext(env, jContext)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	opts, err := namespaceOptions(env, jOptions)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	entryChan, err := n.Glob(context, jutil.GoString(env, pattern), opts...)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}

	retChan := make(chan C.jobject, 100)
	go func() {
		jEnv, freeFunc := jutil.GetEnv()
		env := (*C.JNIEnv)(jEnv)
		defer freeFunc()

		for globReply := range entryChan {
			jGlobReply, err := jutil.JVomCopy(env, globReply, jGlobReplyClass)
			if err != nil {
				log.Printf("Couldn't convert Go glob result %v to Java\n", globReply)
				continue
			}
			// The other side of the channel is responsible
			// for deleting this global reference.
			jGlobalGlobReply := C.jobject(jutil.NewGlobalRef(env, jGlobReply))
			retChan <- jGlobalGlobReply
		}
		close(retChan)
	}()
	jInputChannel, err := jchannel.JavaInputChannel(env, &retChan, &entryChan)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jInputChannel)
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeMount
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeMount(env *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jServer C.jstring, jDuration C.jobject, jOptions C.jobject) {
	n := *(*namespace.T)(jutil.Ptr(goNamespacePtr))
	context, err := jcontext.GoContext(env, jContext)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	duration, err := jutil.GoDuration(env, jDuration)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	options, err := namespaceOptions(env, jOptions)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := n.Mount(context, jutil.GoString(env, jName), jutil.GoString(env, jServer), duration, options...); err != nil {
		jutil.JThrowV(env, err)
	}
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeUnmount
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeUnmount(env *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jServer C.jstring, jOptions C.jobject) {
	n := *(*namespace.T)(jutil.Ptr(goNamespacePtr))
	context, err := jcontext.GoContext(env, jContext)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	options, err := namespaceOptions(env, jOptions)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := n.Unmount(context, jutil.GoString(env, jName), jutil.GoString(env, jServer), options...); err != nil {
		jutil.JThrowV(env, err)
	}
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeDelete
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeDelete(env *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jDeleteSubtree C.jboolean, jOptions C.jobject) {
	n := *(*namespace.T)(jutil.Ptr(goNamespacePtr))
	context, err := jcontext.GoContext(env, jContext)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	options, err := namespaceOptions(env, jOptions)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := n.Delete(context, jutil.GoString(env, jName), jDeleteSubtree == C.JNI_TRUE, options...); err != nil {
		jutil.JThrowV(env, err)
	}
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeResolve
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeResolve(env *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jOptions C.jobject) C.jobject {
	n := *(*namespace.T)(jutil.Ptr(goNamespacePtr))
	context, err := jcontext.GoContext(env, jContext)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	options, err := namespaceOptions(env, jOptions)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	entry, err := n.Resolve(context, jutil.GoString(env, jName), options...)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jEntry, err := jutil.JVomCopy(env, entry, jMountEntryClass)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jEntry)
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeResolveToMountTable
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeResolveToMountTable(env *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jOptions C.jobject) C.jobject {
	n := *(*namespace.T)(jutil.Ptr(goNamespacePtr))
	context, err := jcontext.GoContext(env, jContext)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	options, err := namespaceOptions(env, jOptions)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	entry, err := n.ResolveToMountTable(context, jutil.GoString(env, jName), options...)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jEntry, err := jutil.JVomCopy(env, entry, jMountEntryClass)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jEntry)
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeFlushCacheEntry
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeFlushCacheEntry(env *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName string) C.jboolean {
	n := *(*namespace.T)(jutil.Ptr(goNamespacePtr))
	context, err := jcontext.GoContext(env, jContext)
	if err != nil {
		jutil.JThrowV(env, err)
		return C.JNI_FALSE
	}
	result := n.FlushCacheEntry(context, jutil.GoString(env, jName))
	if result {
		return C.JNI_TRUE
	} else {
		return C.JNI_FALSE
	}
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeSetRoots
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeSetRoots(env *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jNames C.jobject) {
	n := *(*namespace.T)(jutil.Ptr(goNamespacePtr))
	names, err := jutil.GoStringList(env, jNames)
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
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeSetPermissions(env *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jPermissions C.jobject, jVersion C.jstring, jOptions C.jobject) {
	n := *(*namespace.T)(jutil.Ptr(goNamespacePtr))
	context, err := jcontext.GoContext(env, jContext)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	var permissions access.Permissions
	err = jutil.GoVomCopy(env, jPermissions, jPermissionsClass, &permissions)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	options, err := namespaceOptions(env, jOptions)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := n.SetPermissions(context, jutil.GoString(env, jName), permissions, jutil.GoString(env, jVersion), options...); err != nil {
		jutil.JThrowV(env, err)
	}
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeGetPermissions
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeGetPermissions(env *C.JNIEnv, jNamespaceClass C.jclass, goNamespacePtr C.jlong, jContext C.jobject, jName C.jstring, jOptions C.jobject) C.jobject {
	n := *(*namespace.T)(jutil.Ptr(goNamespacePtr))
	context, err := jcontext.GoContext(env, jContext)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	options, err := namespaceOptions(env, jOptions)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	permissions, version, err := n.GetPermissions(context, jutil.GoString(env, jName), options...)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jPermissions, err := jutil.JVomCopy(env, permissions, jPermissionsClass)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	result := make(map[interface{}]interface{})
	result[C.jstring(jutil.JString(env, version))] = jPermissions
	jResult, err := jutil.JObjectMap(env, result)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jResult)
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeFinalize
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeFinalize(env *C.JNIEnv, jNamespace C.jobject, goNamespacePtr C.jlong) {
	jutil.GoUnref(jutil.Ptr(goNamespacePtr))
}
