// +build android

package naming

import (
	"time"
	"unsafe"

	"v.io/core/veyron2/naming"
	jutil "v.io/jni/util"
	jcontext "v.io/jni/veyron2/context"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

var (
	// Global reference for io.v.core.veyron.runtimes.google.naming.Namespace class.
	jNamespaceImplClass C.jclass
)

// Init initializes the JNI code with the given Java environment. This method
// must be called from the main Java thread.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java environment is passed in an empty
// interface and then cast into the package-local environment type.
func Init(jEnv interface{}) {
	env := (*C.JNIEnv)(unsafe.Pointer(jutil.PtrValue(jEnv)))
	jNamespaceImplClass = C.jclass(jutil.JFindClassOrPrint(env, "io/v/core/veyron/veyron/runtimes/google/naming/Namespace"))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_naming_Namespace_nativeGlob
func Java_io_veyron_veyron_veyron_runtimes_google_naming_Namespace_nativeGlob(env *C.JNIEnv, jNamespace C.jobject, goNamespacePtr C.jlong, jContext C.jobject, pattern C.jstring) C.jlong {
	n := *(*naming.Namespace)(jutil.Ptr(goNamespacePtr))
	context, err := jcontext.GoContext(env, jContext)
	if err != nil {
		jutil.JThrowV(env, err)
		return C.jlong(0)
	}
	entryChan, err := n.Glob(context, jutil.GoString(env, pattern))
	if err != nil {
		jutil.JThrowV(env, err)
		return C.jlong(0)
	}
	// We want to return chan interface{}, not chan naming.MountEntry, so we
	// convert here.  We also convert naming.MountEntry into naming.VDLMountEntry
	// which can be VOM-encoded.
	retChan := make(chan interface{}, 100)
	go func() {
		for entry := range entryChan {
			var vdlEntry naming.VDLMountEntry
			vdlEntry.Name = entry.Name
			for _, server := range entry.Servers {
				var vdlServer naming.VDLMountedServer
				vdlServer.Server = server.Server
				vdlServer.TTL = uint32(server.Expires.Sub(time.Now()).Seconds())
				vdlEntry.Servers = append(vdlEntry.Servers, vdlServer)
			}
			retChan <- vdlEntry
		}
		close(retChan)
	}()
	jutil.GoRef(&retChan) // Un-refed when the InputChannel is finalized.
	return C.jlong(jutil.PtrValue(&retChan))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_naming_Namespace_nativeFinalize
func Java_io_veyron_veyron_veyron_runtimes_google_naming_Namespace_nativeFinalize(env *C.JNIEnv, jNamespace C.jobject, goNamespacePtr C.jlong) {
	jutil.GoUnref((*naming.Namespace)(jutil.Ptr(goNamespacePtr)))
}
