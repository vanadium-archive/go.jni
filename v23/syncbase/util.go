// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package syncbase

import (
	"runtime"

	wire "v.io/v23/services/syncbase"
	"v.io/v23/syncbase"

	jutil "v.io/x/jni/util"
)

type jniDatabase struct {
	syncbase.Database
	parentFullName string
	schema         *syncbase.Schema
	jSchema        jutil.Object
}

func newJNIDatabase(env jutil.Env, db syncbase.Database, parentFullName string, schema *syncbase.Schema, jSchema jutil.Object) *jniDatabase {
	// Reference Java schema; it will be de-referenced when the Go database
	// created below is garbage-collected (through the finalizer callback we
	// setup just below).
	jSchema = jutil.NewGlobalRef(env, jSchema)
	jdb := &jniDatabase{
		Database:       db,
		parentFullName: parentFullName,
		schema:         schema,
		jSchema:        jSchema,
	}
	runtime.SetFinalizer(jdb, func(jdb *jniDatabase) {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()
		jutil.DeleteGlobalRef(env, jdb.jSchema)
	})
	return jdb
}

func javaDatabase(env jutil.Env, jdb *jniDatabase) (jutil.Object, error) {
	schemaSign := jutil.ClassSign("io.v.v23.syncbase.Schema")
	ref := jutil.GoNewRef(jdb) // Un-refed when jDatabase is finalized
	jDatabase, err := jutil.NewObject(env, jDatabaseImplClass, []jutil.Sign{jutil.LongSign, jutil.StringSign, jutil.StringSign, jutil.StringSign, schemaSign}, int64(ref), jdb.parentFullName, jdb.FullName(), jdb.Name(), jdb.jSchema)
	if err != nil {
		jutil.GoDecRef(ref)
		return jutil.NullObject, err
	}
	return jDatabase, nil
}

func javaBatchDatabase(env jutil.Env, batchDB syncbase.BatchDatabase, parentFullName string, jSchema jutil.Object) (jutil.Object, error) {
	schemaSign := jutil.ClassSign("io.v.v23.syncbase.Schema")
	return jutil.NewObject(env, jDatabaseImplClass, []jutil.Sign{jutil.LongSign, jutil.StringSign, jutil.StringSign, jutil.StringSign, schemaSign}, 0, parentFullName, batchDB.FullName(), batchDB.Name(), jSchema)
}

// GoSchema converts the provided Java Schema object into a Go syncbase.Schema
// type.
func GoSchema(env jutil.Env, jSchema jutil.Object) (*syncbase.Schema, error) {
	if jSchema.IsNull() {
		return nil, nil
	}
	metadataSign := jutil.ClassSign("io.v.v23.services.syncbase.SchemaMetadata")
	jMetadata, err := jutil.CallObjectMethod(env, jSchema, "getMetadata", nil, metadataSign)
	if err != nil {
		return nil, err
	}
	var metadata wire.SchemaMetadata
	if err := jutil.GoVomCopy(env, jMetadata, jSchemaMetadataClass, &metadata); err != nil {
		return nil, err
	}
	resolverSign := jutil.ClassSign("io.v.v23.syncbase.ConflictResolver")
	jResolver, err := jutil.CallObjectMethod(env, jSchema, "getResolver", nil, resolverSign)
	if err != nil {
		return nil, err
	}
	resolver := GoResolver(env, jResolver)
	return &syncbase.Schema{
		Metadata: metadata,
		Resolver: resolver,
	}, nil
}
