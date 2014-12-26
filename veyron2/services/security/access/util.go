// +build android

package access

import (
	jutil "v.io/jni/util"
	"v.io/core/veyron2/services/security/access"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// JavaACL converts the provided Go ACL into a Java ACL.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaACL(jEnv interface{}, acl access.ACL) (C.jobject, error) {
	encoded, err := jutil.VomEncode(acl)
	if err != nil {
		return nil, err
	}
	jACL, err := jutil.CallStaticObjectMethod(jEnv, jUtilClass, "decodeACL", []jutil.Sign{jutil.ByteArraySign}, aclSign, encoded)
	if err != nil {
		return nil, err
	}
	return C.jobject(jACL), nil
}

// GoACL converts the provided Java ACL into a Go ACL.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoACL(jEnv, jACL interface{}) (access.ACL, error) {
	encoded, err := jutil.CallStaticByteArrayMethod(jEnv, jUtilClass, "encodeACL", []jutil.Sign{aclSign}, jACL)
	if err != nil {
		return access.ACL{}, err
	}
	var a access.ACL
	if err := jutil.VomDecode(encoded, &a); err != nil {
		return access.ACL{}, err
	}
	return a, nil
}

// JavaACLWrapper converts the provided Go ACL into a Java ACLWrapper object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaACLWrapper(jEnv interface{}, acl access.ACL) (C.jobject, error) {
	jACL, err := JavaACL(jEnv, acl)
	if err != nil {
		return nil, err
	}
	jWrapper, err := jutil.NewObject(jEnv, jACLWrapperClass, []jutil.Sign{jutil.LongSign, aclSign}, int64(jutil.PtrValue(&acl)), jACL)
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&acl) // Un-refed when the Java ACLWrapper object is finalized.
	return C.jobject(jWrapper), nil
}
