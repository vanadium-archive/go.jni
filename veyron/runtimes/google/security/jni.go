// +build android

package security

import (
	"unsafe"

	"veyron.io/jni/util"
	jsecurity "veyron.io/jni/veyron2/security"
	"veyron.io/veyron/veyron2/security"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

var (
	// Global reference for io.veyron.veyron.veyron.runtimes.google.security.Context class.
	jContextImplClass C.jclass
)

// Init initializes the JNI code with the given Java evironment. This method
// must be called from the main Java thread.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java environment is passed in an empty
// interface and then cast into the package-local environment type.
func Init(jEnv interface{}) {
	env := (*C.JNIEnv)(unsafe.Pointer(util.PtrValue(jEnv)))
	// Cache global references to all Java classes used by the package.  This is
	// necessary because JNI gets access to the class loader only in the system
	// thread, so we aren't able to invoke FindClass in other threads.
	jContextImplClass = C.jclass(util.JFindClassOrPrint(env, "io/veyron/veyron/veyron/runtimes/google/security/Context"))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeTimestamp
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeTimestamp(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jobject {
	t := (*(*security.Context)(util.Ptr(goContextPtr))).Timestamp()
	jTime, err := util.JTime(env, t)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return C.jobject(jTime)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeMethod
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeMethod(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jstring {
	return C.jstring(util.JString(env, (*(*security.Context)(util.Ptr(goContextPtr))).Method()))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeMethodTags
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeMethodTags(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jobjectArray {
	tags := (*(*security.Context)(util.Ptr(goContextPtr))).MethodTags()
	if tags == nil {
		return nil
	}
	tagsJava := make([]interface{}, len(tags))
	for i, tag := range tags {
		tagsJava[i] = C.jobject(unsafe.Pointer(util.PtrValue(tag)))
	}
	jTags, err := util.JObjectArray(env, tagsJava)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return C.jobjectArray(jTags)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeName
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeName(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jstring {
	return C.jstring(util.JString(env, (*(*security.Context)(util.Ptr(goContextPtr))).Name()))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeSuffix
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeSuffix(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jstring {
	return C.jstring(util.JString(env, (*(*security.Context)(util.Ptr(goContextPtr))).Suffix()))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLabel
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLabel(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jint {
	return C.jint((*(*security.Context)(util.Ptr(goContextPtr))).Label())
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLocalEndpoint
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLocalEndpoint(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jstring {
	return C.jstring(util.JString(env, (*(*security.Context)(util.Ptr(goContextPtr))).LocalEndpoint().String()))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeRemoteEndpoint
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeRemoteEndpoint(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jstring {
	return C.jstring(util.JString(env, (*(*security.Context)(util.Ptr(goContextPtr))).RemoteEndpoint().String()))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLocalPrincipal
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLocalPrincipal(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jobject {
	principal := (*(*security.Context)(util.Ptr(goContextPtr))).LocalPrincipal()
	jPrincipal, err := jsecurity.JavaPrincipal(env, principal)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return C.jobject(jPrincipal)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLocalBlessings
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLocalBlessings(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jobject {
	blessings := (*(*security.Context)(util.Ptr(goContextPtr))).LocalBlessings()
	jBlessings, err := jsecurity.JavaBlessings(env, blessings)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return C.jobject(jBlessings)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeRemoteBlessings
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeRemoteBlessings(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jobject {
	blessings := (*(*security.Context)(util.Ptr(goContextPtr))).RemoteBlessings()
	jBlessings, err := jsecurity.JavaBlessings(env, blessings)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return C.jobject(jBlessings)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeFinalize
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeFinalize(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) {
	util.GoUnref((*security.Context)(util.Ptr(goContextPtr)))
}
