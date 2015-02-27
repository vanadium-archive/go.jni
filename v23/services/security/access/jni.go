// +build android

package access

import (
	jutil "v.io/jni/util"
	jsecurity "v.io/jni/v23/security"
	"v.io/v23/services/security/access"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

var (
	aclSign = jutil.ClassSign("io.v.v23.services.security.access.ACL")

	// Global reference for io.v.v23.services.security.access.ACLWrapper class.
	jACLWrapperClass C.jclass
	// Global reference for io.v.v23.services.security.access.Util class.
	jUtilClass C.jclass
)

func Init(jEnv interface{}) error {
	class, err := jutil.JFindClass(jEnv, "io/v/v23/services/security/access/ACLWrapper")
	if err != nil {
		return err
	}
	jACLWrapperClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/services/security/access/Util")
	if err != nil {
		return err
	}
	jUtilClass = C.jclass(class)
	return nil
}

//export Java_io_v_v23_services_security_access_ACLWrapper_nativeWrap
func Java_io_v_v23_services_security_access_ACLWrapper_nativeWrap(env *C.JNIEnv, jACLWrapperClass C.jclass, jACL C.jobject) C.jobject {
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
	return C.jobject(jWrapper)
}

//export Java_io_v_v23_services_security_access_ACLWrapper_nativeIncludes
func Java_io_v_v23_services_security_access_ACLWrapper_nativeIncludes(env *C.JNIEnv, jACLWrapper C.jobject, goPtr C.jlong, jBlessings C.jobjectArray) C.jboolean {
	blessings := jutil.GoStringArray(env, jBlessings)
	ok := (*(*access.ACL)(jutil.Ptr(goPtr))).Includes(blessings...)
	if ok {
		return C.JNI_TRUE
	}
	return C.JNI_FALSE
}

//export Java_io_v_v23_services_security_access_ACLWrapper_nativeAuthorize
func Java_io_v_v23_services_security_access_ACLWrapper_nativeAuthorize(env *C.JNIEnv, jACLWrapper C.jobject, goPtr C.jlong, jContext C.jobject) {
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

//export Java_io_v_v23_services_security_access_ACLWrapper_nativeFinalize
func Java_io_v_v23_services_security_access_ACLWrapper_nativeFinalize(env *C.JNIEnv, jACLWrapper C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*access.ACL)(jutil.Ptr(goPtr)))
}
