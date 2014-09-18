// +build android

package naming

import (
	"time"

	"veyron.io/jni/runtimes/google/util"
	"veyron.io/veyron/veyron2/naming"
	"veyron.io/veyron/veyron2/rt"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// Init initializes the JNI code with the given Java environment. This method
// must be called from the main Java thread.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java environment is passed in an empty
// interface and then cast into the package-local environment type.
func Init(jEnv interface{}) {}

//export Java_com_veyron_runtimes_google_naming_Namespace_nativeGlob
func Java_com_veyron_runtimes_google_naming_Namespace_nativeGlob(env *C.JNIEnv, jNamespace C.jobject, goNamespacePtr C.jlong, jContext C.jobject, pattern C.jstring) C.jlong {
	n := *(*naming.Namespace)(util.Ptr(goNamespacePtr))
	// TODO(spetrovic): Implement context correctly in Java and use it here.
	ctx, _ := rt.R().NewContext().WithTimeout(20 * time.Second)
	entryChan, err := n.Glob(ctx, util.GoString(env, pattern))
	if err != nil {
		util.JThrowV(env, err)
		return C.jlong(0)
	}
	// We want to return chan interface{}, not chan naming.MountEntry, so we
	// convert here.  We also convert naming.MountEntry into a similar struct
	// which can be JSON-serialized.  (MounEntry has a field of type "error"
	// which doesn't get JSON-serialized correctly.)
	retChan := make(chan interface{}, 100)
	go func() {
		for entry := range entryChan {
			s := struct {
				Name    string
				Servers []naming.MountedServer
				Error   string
			}{}
			s.Name = entry.Name
			s.Servers = entry.Servers
			if entry.Error != nil {
				s.Error = entry.Error.Error()
			}
			retChan <- s
		}
		close(retChan)
	}()
	util.GoRef(&retChan) // Un-refed when the InputChannel is finalized.
	return C.jlong(util.PtrValue(&retChan))
}

//export Java_com_veyron_runtimes_google_naming_Namespace_nativeFinalize
func Java_com_veyron_runtimes_google_naming_Namespace_nativeFinalize(env *C.JNIEnv, jNamespace C.jobject, goNamespacePtr C.jlong) {
	util.GoUnref((*naming.Namespace)(util.Ptr(goNamespacePtr)))
}
