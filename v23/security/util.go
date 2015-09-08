// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package security

import (
	"v.io/v23/security"
	"v.io/v23/vom"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

// JavaBlessings converts the provided Go Blessings into Java Blessings.
func JavaBlessings(env jutil.Env, blessings security.Blessings) (jutil.Object, error) {
	jWire, err := jutil.JVomCopy(env, blessings, jWireBlessingsClass)
	if err != nil {
		return jutil.NullObject, err
	}
	jBlessings, err := jutil.NewObject(env, jBlessingsClass, []jutil.Sign{jutil.LongSign, wireBlessingsSign}, int64(jutil.PtrValue(&blessings)), jWire)
	if err != nil {
		return jutil.NullObject, err
	}
	jutil.GoRef(&blessings) // Un-refed when the Java Blessings object is finalized.
	return jBlessings, nil
}

// GoBlessings converts the provided Java Blessings into Go Blessings.
func GoBlessings(env jutil.Env, jBlessings jutil.Object) (security.Blessings, error) {
	if jBlessings.IsNull() {
		return security.Blessings{}, nil
	}
	jWire, err := jutil.CallObjectMethod(env, jBlessings, "wireFormat", nil, wireBlessingsSign)
	if err != nil {
		return security.Blessings{}, err
	}
	var blessings security.Blessings
	if err := jutil.GoVomCopy(env, jWire, jWireBlessingsClass, &blessings); err != nil {
		return security.Blessings{}, err
	}
	return blessings, nil
}

// GoBlessingsArray converts the provided Java Blessings array into a Go
// Blessings slice.
func GoBlessingsArray(env jutil.Env, jBlessingsArr jutil.Object) ([]security.Blessings, error) {
	barr, err := jutil.GoObjectArray(env, jBlessingsArr)
	if err != nil {
		return nil, err
	}
	ret := make([]security.Blessings, len(barr))
	for i, jBlessings := range barr {
		var err error
		if ret[i], err = GoBlessings(env, jBlessings); err != nil {
			return nil, err
		}
	}
	return ret, nil
}

// JavaCaveat converts the provided Go Caveat into a Java Caveat.
func JavaCaveat(env jutil.Env, caveat security.Caveat) (jutil.Object, error) {
	return jutil.JVomCopy(env, caveat, jCaveatClass)
}

// GoCaveat converts the provided Java Caveat into a Go Caveat.
func GoCaveat(env jutil.Env, jCav jutil.Object) (security.Caveat, error) {
	var caveat security.Caveat
	if err := jutil.GoVomCopy(env, jCav, jCaveatClass, &caveat); err != nil {
		return security.Caveat{}, err
	}
	return caveat, nil
}

// JavaCaveats converts the provided Go Caveat slice into a Java Caveat array.
func JavaCaveats(env jutil.Env, caveats []security.Caveat) (jutil.Object, error) {
	cavarr := make([]jutil.Object, len(caveats))
	for i, caveat := range caveats {
		var err error
		if cavarr[i], err = JavaCaveat(env, caveat); err != nil {
			return jutil.NullObject, err
		}
	}
	return jutil.JObjectArray(env, cavarr, jCaveatClass)
}

// GoCaveats converts the provided Java Caveat array into a Go Caveat slice.
func GoCaveats(env jutil.Env, jCaveats jutil.Object) ([]security.Caveat, error) {
	cavarr, err := jutil.GoObjectArray(env, jCaveats)
	if err != nil {
		return nil, err
	}
	caveats := make([]security.Caveat, len(cavarr))
	for i, jCaveat := range cavarr {
		var err error
		if caveats[i], err = GoCaveat(env, jCaveat); err != nil {
			return nil, err
		}
	}
	return caveats, nil
}

// JavaBlessingPattern converts the provided Go BlessingPattern into Java
// BlessingPattern.
func JavaBlessingPattern(env jutil.Env, pattern security.BlessingPattern) (jutil.Object, error) {
	return jutil.JVomCopy(env, pattern, jBlessingPatternClass)
}

// GoBlessingPattern converts the provided Java BlessingPattern into Go BlessingPattern.
func GoBlessingPattern(env jutil.Env, jPattern jutil.Object) (pattern security.BlessingPattern, err error) {
	err = jutil.GoVomCopy(env, jPattern, jBlessingPatternClass, &pattern)
	return
}

// JavaPublicKey converts the provided Go PublicKey into Java PublicKey.
func JavaPublicKey(env jutil.Env, key security.PublicKey) (jutil.Object, error) {
	if key == nil {
		return jutil.NullObject, nil
	}
	der, err := key.MarshalBinary()
	if err != nil {
		return jutil.NullObject, err
	}
	return JavaPublicKeyFromDER(env, der)
}

// JavaPublicKeyFromDER converts a DER-encoded public key into a Java PublicKey object.
func JavaPublicKeyFromDER(env jutil.Env, der []byte) (jutil.Object, error) {
	jPublicKey, err := jutil.CallStaticObjectMethod(env, jUtilClass, "decodePublicKey", []jutil.Sign{jutil.ArraySign(jutil.ByteSign)}, publicKeySign, der)
	if err != nil {
		return jutil.NullObject, err
	}
	return jPublicKey, nil
}

// JavaPublicKeyToDER returns the DER-encoded representations of a Java PublicKey object.
func JavaPublicKeyToDER(env jutil.Env, jKey jutil.Object) ([]byte, error) {
	return jutil.CallStaticByteArrayMethod(env, jUtilClass, "encodePublicKey", []jutil.Sign{publicKeySign}, jKey)
}

// GoPublicKey converts the provided Java PublicKey into Go PublicKey.
func GoPublicKey(env jutil.Env, jKey jutil.Object) (security.PublicKey, error) {
	der, err := JavaPublicKeyToDER(env, jKey)
	if err != nil {
		return nil, err
	}
	return security.UnmarshalPublicKey(der)
}

// JavaSignature converts the provided Go Signature into a Java VSignature.
func JavaSignature(env jutil.Env, sig security.Signature) (jutil.Object, error) {
	encoded, err := vom.Encode(sig)
	if err != nil {
		return jutil.NullObject, err
	}
	jSignature, err := jutil.CallStaticObjectMethod(env, jUtilClass, "decodeSignature", []jutil.Sign{jutil.ByteArraySign}, signatureSign, encoded)
	if err != nil {
		return jutil.NullObject, err
	}
	return jSignature, nil
}

// GoSignature converts the provided Java VSignature into a Go Signature.
func GoSignature(env jutil.Env, jSignature jutil.Object) (security.Signature, error) {
	encoded, err := jutil.CallStaticByteArrayMethod(env, jUtilClass, "encodeSignature", []jutil.Sign{signatureSign}, jSignature)
	if err != nil {
		return security.Signature{}, err
	}
	var sig security.Signature
	if err := vom.Decode(encoded, &sig); err != nil {
		return security.Signature{}, err
	}
	return sig, nil
}

// GoDischarge converts the provided Java Discharge into a Go Discharge.
func GoDischarge(env jutil.Env, jDischarge jutil.Object) (security.Discharge, error) {
	var discharge security.Discharge
	if err := jutil.GoVomCopy(env, jDischarge, jDischargeClass, &discharge); err != nil {
		return security.Discharge{}, err
	}
	return discharge, nil
}

// JavaDischarge converts the provided Go Discharge into a Java discharge.
func JavaDischarge(env jutil.Env, discharge security.Discharge) (jutil.Object, error) {
	return jutil.JVomCopy(env, discharge, jDischargeClass)
}

func javaDischargeMap(env jutil.Env, discharges map[string]security.Discharge) (jutil.Object, error) {
	objectMap := make(map[jutil.Object]jutil.Object)
	for key, discharge := range discharges {
		jKey := jutil.JString(env, key)
		jDischarge, err := JavaDischarge(env, discharge)
		if err != nil {
			return jutil.NullObject, err
		}
		objectMap[jKey] = jDischarge
	}
	return jutil.JObjectMap(env, objectMap)
}
