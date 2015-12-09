// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package nosql

import (
	"unsafe"

	wire "v.io/v23/services/syncbase/nosql"
	"v.io/v23/syncbase/nosql"

	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
)

// #include "jni.h"
import "C"

var (
	// Global reference for io.v.v23.syncbase.nosql.DatabaseImpl class.
	jDatabaseImplClass jutil.Class
	// Global reference for io.v.v23.syncbase.nosql.Conflict class.
	jConflictClass jutil.Class
	// Global reference for io.v.v23.syncbase.nosql.Resolution class.
	jResolutionClass jutil.Class
	// Global reference for io.v.v23.services.syncbase.nosql.BatchOptions class.
	jBatchOptionsClass jutil.Class
	// Global reference for io.v.v23.services.syncbase.nosql.SchemaMetadata class.
	jSchemaMetadataClass jutil.Class
)

// Init initializes the JNI code with the given Java environment.  This method
// must be invoked before any other method in this package and must be called
// from the main Java thread (e.g., On_Load()).
func Init(env jutil.Env) error {
	var err error
	jDatabaseImplClass, err = jutil.JFindClass(env, "io/v/v23/syncbase/nosql/DatabaseImpl")
	if err != nil {
		return err
	}
	jConflictClass, err = jutil.JFindClass(env, "io/v/v23/syncbase/nosql/Conflict")
	if err != nil {
		return err
	}
	jResolutionClass, err = jutil.JFindClass(env, "io/v/v23/syncbase/nosql/Resolution")
	if err != nil {
		return err
	}
	jBatchOptionsClass, err = jutil.JFindClass(env, "io/v/v23/services/syncbase/nosql/BatchOptions")
	if err != nil {
		return err
	}
	jSchemaMetadataClass, err = jutil.JFindClass(env, "io/v/v23/services/syncbase/nosql/SchemaMetadata")
	if err != nil {
		return err
	}
	return nil
}

//export Java_io_v_v23_syncbase_nosql_DatabaseImpl_nativeCreate
func Java_io_v_v23_syncbase_nosql_DatabaseImpl_nativeCreate(jenv *C.JNIEnv, jDatabaseImplClass C.jclass, jParentFullName C.jstring, jRelativeName C.jstring, jSchemaObj C.jobject) C.jobject {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	parentFullName := jutil.GoString(env, jutil.Object(uintptr(unsafe.Pointer(jParentFullName))))
	relativeName := jutil.GoString(env, jutil.Object(uintptr(unsafe.Pointer(jRelativeName))))
	jSchema := jutil.Object(uintptr(unsafe.Pointer(jSchemaObj)))
	schema, err := GoSchema(env, jSchema)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	db := nosql.NewDatabase(parentFullName, relativeName, schema)
	jdb := newJNIDatabase(env, db, parentFullName, schema, jSchema)
	jDatabase, err := javaDatabase(env, jdb)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jDatabase))
}

//export Java_io_v_v23_syncbase_nosql_DatabaseImpl_nativeBeginBatch
func Java_io_v_v23_syncbase_nosql_DatabaseImpl_nativeBeginBatch(jenv *C.JNIEnv, jDatabaseImpl C.jobject, goPtr C.jlong, jContext C.jobject, jBatchOptsObj C.jobject, jCallbackObj C.jobject) {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	jCallback := jutil.Object(uintptr(unsafe.Pointer(jCallbackObj)))
	jBatchOpts := jutil.Object(uintptr(unsafe.Pointer(jBatchOptsObj)))
	jdb := (*jniDatabase)(jutil.NativePtr(goPtr))
	var batchOpts wire.BatchOptions
	if err := jutil.GoVomCopy(env, jBatchOpts, jBatchOptionsClass, &batchOpts); err != nil {
		jutil.CallbackOnFailure(env, jCallback, err)
		return
	}
	ctx, err := jcontext.GoContext(env, jutil.Object(uintptr(unsafe.Pointer(jContext))))
	if err != nil {
		jutil.CallbackOnFailure(env, jCallback, err)
		return
	}
	jutil.DoAsyncCall(env, jCallback, func() (jutil.Object, error) {
		batchDB, err := jdb.BeginBatch(ctx, batchOpts)
		if err != nil {
			return jutil.NullObject, err
		}
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()
		jBatchDB, err := javaBatchDatabase(env, batchDB, jdb.parentFullName, jdb.jSchema)
		if err != nil {
			return jutil.NullObject, err
		}
		// Must grab a global reference as we free up the env and all local references that come
		// along with it.
		return jutil.NewGlobalRef(env, jBatchDB), nil
	})
}

//export Java_io_v_v23_syncbase_nosql_DatabaseImpl_nativeEnforceSchema
func Java_io_v_v23_syncbase_nosql_DatabaseImpl_nativeEnforceSchema(jenv *C.JNIEnv, jDatabaseImpl C.jobject, goPtr C.jlong, jContext C.jobject, jCallbackObj C.jobject) {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	jCallback := jutil.Object(uintptr(unsafe.Pointer(jCallbackObj)))
	jdb := (*jniDatabase)(jutil.NativePtr(goPtr))
	ctx, err := jcontext.GoContext(env, jutil.Object(uintptr(unsafe.Pointer(jContext))))
	if err != nil {
		jutil.CallbackOnFailure(env, jCallback, err)
		return
	}
	jutil.DoAsyncCall(env, jCallback, func() (jutil.Object, error) {
		return jutil.NullObject, jdb.EnforceSchema(ctx)
	})
}

//export Java_io_v_v23_syncbase_nosql_DatabaseImpl_nativeFinalize
func Java_io_v_v23_syncbase_nosql_DatabaseImpl_nativeFinalize(jenv *C.JNIEnv, jDatabaseImpl C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goPtr))
}
