// +build android

package util

import (
	"v.io/v23/vdl"
	"v.io/v23/vom"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// VomDecodeToValue VOM-decodes the provided value into *vdl.Value using a new
// instance of a VOM decoder.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func VomDecodeToValue(data []byte) (*vdl.Value, error) {
	var value *vdl.Value
	if err := vom.Decode(data, &value); err != nil {
		return nil, err
	}
	return value, nil
}

// VomCopy copies the provided Go value by encoding/decoding it from VOM.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func VomCopy(src interface{}, dstptr interface{}) error {
	data, err := vom.Encode(src)
	if err != nil {
		return err
	}
	return vom.Decode(data, dstptr)
}

// JVomEncode VOM-encodes the provided Java object of the given class.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JVomEncode(jEnv, jObj, jClass interface{}) ([]byte, error) {
	return CallStaticByteArrayMethod(jEnv, jVomUtilClass, "encode", []Sign{ObjectSign, TypeSign}, jObj, jClass)
}

// JVomEncode VOM-encodes the provided Java VdlValue object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JVomEncodeValue(jEnv, jVdlValue interface{}) ([]byte, error) {
	return CallStaticByteArrayMethod(jEnv, jVomUtilClass, "encode", []Sign{VdlValueSign}, jVdlValue)
}

// JVomDecode VOM-decodes the provided data into a Java object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JVomDecode(jEnv interface{}, data []byte, jClass interface{}) (C.jobject, error) {
	class := getClass(jClass)
	if class == nil {
		class = jObjectClass
	}
	return CallStaticObjectMethod(jEnv, jVomUtilClass, "decode", []Sign{ByteArraySign, TypeSign}, ObjectSign, data, class)
}

// JVomCopy copies the provided Go value into a corresponding Java object by
// encoding/decoding it from VOM.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JVomCopy(jEnv interface{}, src interface{}, jClass interface{}) (C.jobject, error) {
	data, err := vom.Encode(src)
	if err != nil {
		return nil, err
	}
	return JVomDecode(jEnv, data, jClass)
}

// GoVomCopy copies the provided Java object into a provided Go value pointer by
// encoding/decoding it from VOM.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoVomCopy(jEnv, jObj, jClass, dstptr interface{}) error {
	data, err := JVomEncode(jEnv, jObj, jClass)
	if err != nil {
		return err
	}
	return vom.Decode(data, dstptr)
}

// GoVomCopyValue copies the provided Java VDLValue object into a Go *vdl.Value
// by encoding/decoding it from VOM.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoVomCopyValue(jEnv, jVdlValue interface{}) (*vdl.Value, error) {
	data, err := JVomEncodeValue(jEnv, jVdlValue)
	if err != nil {
		return nil, err
	}
	return VomDecodeToValue(data)
}
