// +build android

package util

import (
	"bytes"

	"v.io/v23/vdl"
	"v.io/v23/vom"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// VomDecodeToValue VOM-decodes the provided value into *vdl.Value using a new
// instance of a VOM decoder.
func VomDecodeToValue(data []byte) (*vdl.Value, error) {
	decoder, err := vom.NewDecoder(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	var value *vdl.Value
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

// VomCopy copies the provided Go value by encoding/decoding it from VOM.
func VomCopy(src interface{}, dstptr interface{}) error {
	data, err := vom.Encode(src)
	if err != nil {
		return err
	}
	return vom.Decode(data, dstptr)
}

// JVomDecode VOM-decodes the provided value into a Java object.
func JVomDecode(jEnv interface{}, data []byte, jClass interface{}) (C.jobject, error) {
	class := getClass(jClass)
	if class == nil {
		class = jObjectClass
	}
	return CallStaticObjectMethod(jEnv, jVomUtilClass, "decode", []Sign{ByteArraySign, TypeSign}, ObjectSign, data, class)
}

// JVomCopy copies the provided Go value into a corresponding Java object by
// encoding/decoding it from VOM.
func JVomCopy(jEnv interface{}, src interface{}, jClass interface{}) (C.jobject, error) {
	data, err := vom.Encode(src)
	if err != nil {
		return nil, err
	}
	return JVomDecode(jEnv, data, jClass)
}
