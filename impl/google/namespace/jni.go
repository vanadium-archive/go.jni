// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package namespace

import (
	"log"

	"v.io/v23/namespace"
	"v.io/v23/naming"
	"v.io/v23/options"
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
	return nil
}

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeGlob
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeGlob(env *C.JNIEnv, jNamespace C.jobject, goNamespacePtr C.jlong, jContext C.jobject, pattern C.jstring, jSkipServerAuth C.jboolean) C.jobject {
	n := *(*namespace.T)(jutil.Ptr(goNamespacePtr))
	context, err := jcontext.GoContext(env, jContext)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	var opts []naming.NamespaceOpt
	if jSkipServerAuth == C.JNI_TRUE {
		opts = append(opts, options.SkipServerEndpointAuthorization{})
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

//export Java_io_v_impl_google_namespace_NamespaceImpl_nativeFinalize
func Java_io_v_impl_google_namespace_NamespaceImpl_nativeFinalize(env *C.JNIEnv, jNamespace C.jobject, goNamespacePtr C.jlong) {
	jutil.GoUnref(jutil.Ptr(goNamespacePtr))
}
