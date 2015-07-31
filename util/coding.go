// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package util

import (
	"v.io/v23/vdl"
	"v.io/v23/vom"
)

// #include "jni_wrapper.h"
import "C"

// VomDecodeToValue VOM-decodes the provided value into *vdl.Value using a new
// instance of a VOM decoder.
func VomDecodeToValue(data []byte) (*vdl.Value, error) {
	var value *vdl.Value
	if err := vom.Decode(data, &value); err != nil {
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

// JVomEncode VOM-encodes the provided Java object of the given type.
func JVomEncode(env Env, obj Object, typeObj Object) ([]byte, error) {
	return CallStaticByteArrayMethod(env, jVomUtilClass, "encode", []Sign{ObjectSign, TypeSign}, obj, typeObj)
}

// JVomEncode VOM-encodes the provided Java VdlValue object.
func JVomEncodeValue(env Env, vdlValue Object) ([]byte, error) {
	return CallStaticByteArrayMethod(env, jVomUtilClass, "encode", []Sign{VdlValueSign}, vdlValue)
}

// JVomDecode VOM-decodes the provided data into a Java object of the
// given class.
func JVomDecode(env Env, data []byte, class Class) (Object, error) {
	return JVomDecodeWithType(env, data, WrapObject(class.value()))
}

// JVomDecodeWithType VOM-decodes the provided data into a Java object
// of the given type.
func JVomDecodeWithType(env Env, data []byte, typeObj Object) (Object, error) {
	if typeObj.IsNull() {
		typeObj = WrapObject(jObjectClass.value())
	}
	return CallStaticObjectMethod(env, jVomUtilClass, "decode", []Sign{ByteArraySign, TypeSign}, ObjectSign, data, typeObj)
}

// JVomCopy copies the provided Go value into a Java object of the given class,
// by encoding/decoding it from VOM.
func JVomCopy(env Env, val interface{}, class Class) (Object, error) {
	return JVomCopyWithType(env, val, WrapObject(class.value()))
}

// JVomCopyWithType copies the provided Go value into a Java object of the
// given type, by encoding/decoding it from VOM.
func JVomCopyWithType(env Env, val interface{}, typeObj Object) (Object, error) {
	data, err := vom.Encode(val)
	if err != nil {
		return NullObject, err
	}
	return JVomDecodeWithType(env, data, typeObj)
}

// GoVomCopy copies the provided Java object into a provided Go value pointer by
// encoding/decoding it from VOM.
func GoVomCopy(env Env, obj Object, class Class, dstptr interface{}) error {
	data, err := JVomEncode(env, obj, WrapObject(class.value()))
	if err != nil {
		return err
	}
	return vom.Decode(data, dstptr)
}

// GoVomCopyValue copies the provided Java VDLValue object into a Go *vdl.Value
// by encoding/decoding it from VOM.
func GoVomCopyValue(env Env, vdlValue Object) (*vdl.Value, error) {
	data, err := JVomEncodeValue(env, vdlValue)
	if err != nil {
		return nil, err
	}
	return VomDecodeToValue(data)
}
