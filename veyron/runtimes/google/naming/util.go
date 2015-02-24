// +build android

package naming

import (
	"time"

	jutil "v.io/jni/util"
	"v.io/v23/naming"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// JavaNamespace converts the provided Go Namespace into a Java Namespace
// object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaNamespace(jEnv interface{}, namespace naming.Namespace) (C.jobject, error) {
	jNamespace, err := jutil.NewObject(jEnv, jNamespaceImplClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&namespace)))
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&namespace) // Un-refed when the Java PrincipalImpl is finalized.
	return C.jobject(jNamespace), nil
}

// JavaMountEntry converts the Go MountEntry into a Java VDLMountEntry.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaMountEntry(jEnv interface{}, entry *naming.MountEntry) (C.jobject, error) {
	var vdlEntry naming.VDLMountEntry
	vdlEntry.Name = entry.Name
	for _, server := range entry.Servers {
		var vdlServer naming.VDLMountedServer
		vdlServer.Server = server.Server
		vdlServer.TTL = uint32(server.Expires.Sub(time.Now()).Seconds())
		vdlEntry.Servers = append(vdlEntry.Servers, vdlServer)
	}
	jEntry, err := jutil.JVomCopy(jEnv, vdlEntry, nil)
	if err != nil {
		return nil, err
	}
	return C.jobject(jEntry), nil
}
