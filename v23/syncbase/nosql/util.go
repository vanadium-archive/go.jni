// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package nosql

import (
	"runtime"

	wire "v.io/v23/services/syncbase/nosql"
	"v.io/v23/syncbase/nosql"

	jutil "v.io/x/jni/util"
)

type jniDatabase struct {
	nosql.Database
	parentFullName string
	schema         *nosql.Schema
	jSchema        jutil.Object
}

func newJNIDatabase(env jutil.Env, db nosql.Database, parentFullName string, schema *nosql.Schema, jSchema jutil.Object) *jniDatabase {
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
	schemaSign := jutil.ClassSign("io.v.v23.syncbase.nosql.Schema")
	jDatabase, err := jutil.NewObject(env, jDatabaseImplClass, []jutil.Sign{jutil.LongSign, jutil.StringSign, jutil.StringSign, jutil.StringSign, schemaSign}, int64(jutil.PtrValue(jdb)), jdb.parentFullName, jdb.FullName(), jdb.Name(), jdb.jSchema)
	if err != nil {
		return jutil.NullObject, err
	}
	jutil.GoRef(jdb) // Un-refed when jDatabase is finalized
	return jDatabase, nil
}

func javaBatchDatabase(env jutil.Env, batchDB nosql.BatchDatabase, parentFullName string, jSchema jutil.Object) (jutil.Object, error) {
	schemaSign := jutil.ClassSign("io.v.v23.syncbase.nosql.Schema")
	return jutil.NewObject(env, jDatabaseImplClass, []jutil.Sign{jutil.LongSign, jutil.StringSign, jutil.StringSign, jutil.StringSign, schemaSign}, 0, parentFullName, batchDB.FullName(), batchDB.Name(), jSchema)
}

// GoSchema converts the provided Java Schema object into a Go nosql.Schema type.
func GoSchema(env jutil.Env, jSchema jutil.Object) (*nosql.Schema, error) {
	if jSchema.IsNull() {
		return nil, nil
	}
	metadataSign := jutil.ClassSign("io.v.v23.services.syncbase.nosql.SchemaMetadata")
	jMetadata, err := jutil.CallObjectMethod(env, jSchema, "getMetadata", nil, metadataSign)
	if err != nil {
		return nil, err
	}
	var metadata wire.SchemaMetadata
	if err := jutil.GoVomCopy(env, jMetadata, jSchemaMetadataClass, &metadata); err != nil {
		return nil, err
	}
	upgraderSign := jutil.ClassSign("io.v.v23.syncbase.nosql.SchemaUpgrader")
	jUpgrader, err := jutil.CallObjectMethod(env, jSchema, "getUpgrader", nil, upgraderSign)
	if err != nil {
		return nil, err
	}
	upgrader := GoUpgrader(env, jUpgrader)
	resolverSign := jutil.ClassSign("io.v.v23.syncbase.nosql.ConflictResolver")
	jResolver, err := jutil.CallObjectMethod(env, jSchema, "getResolver", nil, resolverSign)
	if err != nil {
		return nil, err
	}
	resolver := GoResolver(env, jResolver)
	return &nosql.Schema{
		Metadata: metadata,
		Upgrader: upgrader,
		Resolver: resolver,
	}, nil
}
