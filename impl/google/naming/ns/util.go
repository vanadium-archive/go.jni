// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package ns

import (
	"unsafe"

	"v.io/v23/naming"
	"v.io/v23/naming/ns"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

// JavaNamespace converts the provided Go Namespace into a Java Namespace
// object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaNamespace(jEnv interface{}, namespace ns.Namespace) (unsafe.Pointer, error) {
	jNamespace, err := jutil.NewObject(jEnv, jNamespaceImplClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&namespace)))
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&namespace) // Un-refed when the Java PrincipalImpl is finalized.
	return jNamespace, nil
}

// JavaMountEntry converts the Go MountEntry into a Java MountEntry.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaMountEntry(jEnv interface{}, entry *naming.MountEntry) (unsafe.Pointer, error) {
	var vdlEntry naming.MountEntry
	vdlEntry.Name = entry.Name
	for _, server := range entry.Servers {
		var vdlServer naming.MountedServer
		vdlServer.Server = server.Server
		vdlServer.Deadline = server.Deadline
		vdlEntry.Servers = append(vdlEntry.Servers, vdlServer)
	}
	jEntry, err := jutil.JVomCopy(jEnv, vdlEntry, nil)
	if err != nil {
		return nil, err
	}
	return jEntry, nil
}
