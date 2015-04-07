// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package namespace

import (
	"log"

	"v.io/v23/namespace"
	"v.io/v23/naming"
	jchannel "v.io/x/jni/impl/google/channel"
	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
)

// #include "jni.h"
import "C"

var (
	// Global reference for io.v.impl.google.namespace.Namespace class.
	jNamespaceImplClass C.jclass
	// Global reference for io.v.v23.naming.MountEntry class.
	jMountEntryClass C.jclass
)

// Init initializes the JNI code with the given Java environment. This method
// must be called from the main Java thread.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java environment is passed in an empty
// interface and then cast into the package-local environment type.
func Init(jEnv interface{}) error {
	class, err := jutil.JFindClass(jEnv, "io/v/impl/google/namespace/Namespace")
	if err != nil {
		return err
	}
	jNamespaceImplClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/naming/MountEntry")
	if err != nil {
		return err
	}
	jMountEntryClass = C.jclass(class)
	return nil
}

//export Java_io_v_impl_google_namespace_Namespace_nativeGlob
func Java_io_v_impl_google_namespace_Namespace_nativeGlob(env *C.JNIEnv, jNamespace C.jobject, goNamespacePtr C.jlong, jContext C.jobject, pattern C.jstring) C.jobject {
	n := *(*namespace.T)(jutil.Ptr(goNamespacePtr))
	context, err := jcontext.GoContext(env, jContext)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	entryChan, err := n.Glob(context, jutil.GoString(env, pattern))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}

	retChan := make(chan C.jobject, 100)
	go func() {
		for entry := range entryChan {
			switch v := entry.(type) {
			case *naming.MountEntry:
				jEnv, freeFunc := jutil.GetEnv()
				env := (*C.JNIEnv)(jEnv)
				defer freeFunc()
				jMountEntryObj, err := JavaMountEntry(env, v)
				if err != nil {
					log.Println("Couldn't convert Go MountEntry %v to Java", v)
					continue
				}
				jMountEntry := C.jobject(jutil.NewGlobalRef(env, jMountEntryObj))
				retChan <- jMountEntry
			case *naming.GlobError:
				// Silently drop.
				// TODO(spetrovic): convert to Java counter-part.
			default:
				log.Printf("Encountered value %v of unexpected type %T", entry, entry)
			}
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

//export Java_io_v_impl_google_namespace_Namespace_nativeFinalize
func Java_io_v_impl_google_namespace_Namespace_nativeFinalize(env *C.JNIEnv, jNamespace C.jobject, goNamespacePtr C.jlong) {
	jutil.GoUnref((*namespace.T)(jutil.Ptr(goNamespacePtr)))
}
