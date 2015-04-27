// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package security

import (
	"unsafe"

	"v.io/v23/security"
	"v.io/v23/vom"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

// JavaBlessings converts the provided Go Blessings into Java Blessings.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaBlessings(jEnv interface{}, blessings security.Blessings) (unsafe.Pointer, error) {
	jWire, err := jutil.JVomCopy(jEnv, blessings, jWireBlessingsClass)
	if err != nil {
		return nil, err
	}
	jBlessings, err := jutil.NewObject(jEnv, jBlessingsClass, []jutil.Sign{jutil.LongSign, wireBlessingsSign}, int64(jutil.PtrValue(&blessings)), jWire)
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&blessings) // Un-refed when the Java Blessings object is finalized.
	return jBlessings, nil
}

// GoBlessings converts the provided Java Blessings into Go Blessings.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoBlessings(jEnv, jBlessings interface{}) (security.Blessings, error) {
	if jutil.IsNull(jBlessings) {
		return security.Blessings{}, nil
	}
	jWire, err := jutil.CallObjectMethod(jEnv, jBlessings, "wireFormat", nil, wireBlessingsSign)
	if err != nil {
		return security.Blessings{}, err
	}
	var blessings security.Blessings
	if err := jutil.GoVomCopy(jEnv, jWire, jWireBlessingsClass, &blessings); err != nil {
		return security.Blessings{}, err
	}
	return blessings, nil
}

// GoBlessingsArray converts the provided Java Blessings array into a Go
// Blessings slice.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoBlessingsArray(jEnv, jBlessingsArr interface{}) ([]security.Blessings, error) {
	barr := jutil.GoObjectArray(jEnv, jBlessingsArr)
	ret := make([]security.Blessings, len(barr))
	for i, jBlessings := range barr {
		var err error
		if ret[i], err = GoBlessings(jEnv, jBlessings); err != nil {
			return nil, err
		}
	}
	return ret, nil
}

// JavaCaveat converts the provided Go Caveat into a Java Caveat.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaCaveat(jEnv interface{}, caveat security.Caveat) (unsafe.Pointer, error) {
	return jutil.JVomCopy(jEnv, caveat, jCaveatClass)
}

// GoCaveat converts the provided Java Caveat into a Go Caveat.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoCaveat(jEnv, jCav interface{}) (security.Caveat, error) {
	var caveat security.Caveat
	if err := jutil.GoVomCopy(jEnv, jCav, jCaveatClass, &caveat); err != nil {
		return security.Caveat{}, err
	}
	return caveat, nil
}

// JavaCaveats converts the provided Go Caveat slice into a Java Caveat array.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaCaveats(jEnv interface{}, caveats []security.Caveat) (unsafe.Pointer, error) {
	cavarr := make([]interface{}, len(caveats))
	for i, caveat := range caveats {
		var err error
		if cavarr[i], err = JavaCaveat(jEnv, caveat); err != nil {
			return nil, err
		}
	}
	jCaveats := jutil.JObjectArray(jEnv, cavarr, jCaveatClass)
	return jCaveats, nil
}

// GoCaveats converts the provided Java Caveat array into a Go Caveat slice.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoCaveats(jEnv, jCaveats interface{}) ([]security.Caveat, error) {
	cavarr := jutil.GoObjectArray(jEnv, jCaveats)
	caveats := make([]security.Caveat, len(cavarr))
	for i, jCaveat := range cavarr {
		var err error
		if caveats[i], err = GoCaveat(jEnv, jCaveat); err != nil {
			return nil, err
		}
	}
	return caveats, nil
}

// JavaBlessingPattern converts the provided Go BlessingPattern into Java BlessingPattern.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaBlessingPattern(jEnv interface{}, pattern security.BlessingPattern) (unsafe.Pointer, error) {
	jBlessingPattern, err := jutil.CallStaticObjectMethod(jEnv, jUtilClass, "decodeBlessingPattern", []jutil.Sign{jutil.StringSign}, blessingPatternSign, string(pattern))
	return jBlessingPattern, err
}

// GoBlessingPattern converts the provided Java BlessingPattern into Go BlessingPattern.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoBlessingPattern(jEnv, jPattern interface{}) (security.BlessingPattern, error) {
	encoded, err := jutil.CallStaticStringMethod(jEnv, jUtilClass, "encodeBlessingPattern", []jutil.Sign{blessingPatternSign}, jPattern)
	if err != nil {
		return security.BlessingPattern(""), err
	}
	return security.BlessingPattern(encoded), nil
}

// JavaPublicKey converts the provided Go PublicKey into Java PublicKey.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaPublicKey(jEnv interface{}, key security.PublicKey) (unsafe.Pointer, error) {
	if key == nil {
		return nil, nil
	}
	encoded, err := key.MarshalBinary()
	if err != nil {
		return nil, err
	}
	jPublicKey, err := jutil.CallStaticObjectMethod(jEnv, jUtilClass, "decodePublicKey", []jutil.Sign{jutil.ArraySign(jutil.ByteSign)}, publicKeySign, encoded)
	if err != nil {
		return nil, err
	}
	return jPublicKey, nil
}

// GoPublicKey converts the provided Java PublicKey into Go PublicKey.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoPublicKey(jEnv, jKey interface{}) (security.PublicKey, error) {
	encoded, err := jutil.CallStaticByteArrayMethod(jEnv, jUtilClass, "encodePublicKey", []jutil.Sign{publicKeySign}, jKey)
	if err != nil {
		return nil, err
	}
	return security.UnmarshalPublicKey(encoded)
}

// JavaSignature converts the provided Go Signature into a Java Signature.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaSignature(jEnv interface{}, sig security.Signature) (unsafe.Pointer, error) {
	encoded, err := vom.Encode(sig)
	if err != nil {
		return nil, err
	}
	jSignature, err := jutil.CallStaticObjectMethod(jEnv, jUtilClass, "decodeSignature", []jutil.Sign{jutil.ByteArraySign}, signatureSign, encoded)
	if err != nil {
		return nil, err
	}
	return jSignature, nil
}

// GoSignature converts the provided Java Signature into a Go Signature.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoSignature(jEnv, jSignature interface{}) (security.Signature, error) {
	encoded, err := jutil.CallStaticByteArrayMethod(jEnv, jUtilClass, "encodeSignature", []jutil.Sign{signatureSign}, jSignature)
	if err != nil {
		return security.Signature{}, err
	}
	var sig security.Signature
	if err := vom.Decode(encoded, &sig); err != nil {
		return security.Signature{}, err
	}
	return sig, nil
}

// JavaBlessingPatternWrapper converts the provided Go BlessingPattern into a Java
// BlessingPatternWrapper object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaBlessingPatternWrapper(jEnv interface{}, pattern security.BlessingPattern) (unsafe.Pointer, error) {
	jPattern, err := JavaBlessingPattern(jEnv, pattern)
	if err != nil {
		return nil, err
	}
	jWrapper, err := jutil.NewObject(jEnv, jBlessingPatternWrapperClass, []jutil.Sign{jutil.LongSign, blessingPatternSign}, int64(jutil.PtrValue(&pattern)), jPattern)
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&pattern) // Un-refed when the Java BlessingPatternWrapper object is finalized.
	return jWrapper, nil
}
