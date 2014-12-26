// +build android

package access

import (
	"unsafe"

	jutil "v.io/jni/util"
	jsecurity "v.io/jni/veyron2/security"
	"v.io/core/veyron2/services/security/access"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

var (
	aclSign = jutil.ClassSign("io.veyron.veyron.veyron2.services.security.access.ACL")

	// Global reference for io.veyron.veyron.veyron2.services.security.access.ACLWrapper class.
	jACLWrapperClass C.jclass
	// Global reference for io.veyron.veyron.veyron2.services.security.access.Util class.
	jUtilClass C.jclass
)

func Init(jEnv interface{}) {
	env := (*C.JNIEnv)(unsafe.Pointer(jutil.PtrValue(jEnv)))
	jACLWrapperClass = C.jclass(jutil.JFindClassOrPrint(env, "io/core/veyron/veyron2/services/security/access/ACLWrapper"))
	jUtilClass = C.jclass(jutil.JFindClassOrPrint(env, "io/core/veyron/veyron2/services/security/access/Util"))
}

//export Java_io_veyron_veyron_veyron2_services_security_access_ACLWrapper_nativeWrap
func Java_io_veyron_veyron_veyron2_services_security_access_ACLWrapper_nativeWrap(env *C.JNIEnv, jACLWrapperClass C.jclass, jACL C.jobject) C.jobject {
	acl, err := GoACL(env, jACL)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jWrapper, err := JavaACLWrapper(env, acl)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return jWrapper
}

//export Java_io_veyron_veyron_veyron2_services_security_access_ACLWrapper_nativeIncludes
func Java_io_veyron_veyron_veyron2_services_security_access_ACLWrapper_nativeIncludes(env *C.JNIEnv, jACLWrapper C.jobject, goPtr C.jlong, jBlessings C.jobjectArray) C.jboolean {
	blessings := jutil.GoStringArray(env, jBlessings)
	ok := (*(*access.ACL)(jutil.Ptr(goPtr))).Includes(blessings...)
	if ok {
		return C.JNI_TRUE
	}
	return C.JNI_FALSE
}

//export Java_io_veyron_veyron_veyron2_services_security_access_ACLWrapper_nativeAuthorize
func Java_io_veyron_veyron_veyron2_services_security_access_ACLWrapper_nativeAuthorize(env *C.JNIEnv, jACLWrapper C.jobject, goPtr C.jlong, jContext C.jobject) {
	ctx, err := jsecurity.GoContext(env, jContext)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := (*(*access.ACL)(jutil.Ptr(goPtr))).Authorize(ctx); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_veyron_veyron_veyron2_services_security_access_ACLWrapper_nativeFinalize
func Java_io_veyron_veyron_veyron2_services_security_access_ACLWrapper_nativeFinalize(env *C.JNIEnv, jACLWrapper C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*access.ACL)(jutil.Ptr(goPtr)))
}
