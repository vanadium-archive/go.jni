// +build android

package security

import (
	"unsafe"

	jutil "v.io/jni/util"
	"v.io/v23/security"
	"v.io/v23/vom"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// JavaBlessings converts the provided Go Blessings into Java Blessings.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaBlessings(jEnv interface{}, blessings security.Blessings) (C.jobject, error) {
	if blessings == nil {
		return nil, nil
	}
	wire := security.MarshalBlessings(blessings)
	jWire, err := JavaWireBlessings(jEnv, wire)
	if err != nil {
		return nil, err
	}
	jBlessings, err := jutil.NewObject(jEnv, jBlessingsImplClass, []jutil.Sign{jutil.LongSign, wireBlessingsSign}, int64(jutil.PtrValue(&blessings)), jWire)
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&blessings) // Un-refed when the Java BlessingsImpl object is finalized.
	return C.jobject(jBlessings), nil
}

// GoBlessings converts the provided Java Blessings into Go Blessings.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoBlessings(jEnv, jBlessingsObj interface{}) (security.Blessings, error) {
	jBlessings := C.jobject(unsafe.Pointer(jutil.PtrValue(jBlessingsObj)))
	if jBlessings == nil {
		return nil, nil
	}
	jWire, err := jutil.CallObjectMethod(jEnv, jBlessings, "wireFormat", nil, wireBlessingsSign)
	if err != nil {
		return nil, err
	}
	wire, err := GoWireBlessings(jEnv, jWire)
	if err != nil {
		return nil, err
	}
	return security.NewBlessings(wire)
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

// JavaWireBlessings converts the provided Go WireBlessings into Java WireBlessings.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaWireBlessings(jEnv interface{}, wire security.WireBlessings) (C.jobject, error) {
	var err error
	encoded, err := vom.Encode(wire)
	if err != nil {
		return nil, err
	}
	jWireBlessings, err := jutil.CallStaticObjectMethod(jEnv, jUtilClass, "decodeWireBlessings", []jutil.Sign{jutil.ByteArraySign}, wireBlessingsSign, encoded)
	if err != nil {
		return nil, err
	}
	return C.jobject(jWireBlessings), nil
}

// GoWireBlessings converts the provided Java WireBlessings into Go WireBlessings.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoWireBlessings(jEnv, jWireBless interface{}) (security.WireBlessings, error) {
	jWireBlessings := getObject(jWireBless)
	encoded, err := jutil.CallStaticByteArrayMethod(jEnv, jUtilClass, "encodeWireBlessings", []jutil.Sign{wireBlessingsSign}, jWireBlessings)
	if err != nil {
		return security.WireBlessings{}, err
	}
	var wire security.WireBlessings
	if err := vom.Decode(encoded, &wire); err != nil {
		return security.WireBlessings{}, err
	}
	return wire, nil
}

// JavaCaveat converts the provided Go Caveat into a Java Caveat.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaCaveat(jEnv interface{}, caveat security.Caveat) (C.jobject, error) {
	encoded, err := vom.Encode(caveat)
	if err != nil {
		return nil, err
	}
	jCaveat, err := jutil.CallStaticObjectMethod(jEnv, jUtilClass, "decodeCaveat", []jutil.Sign{jutil.ByteArraySign}, caveatSign, encoded)
	if err != nil {
		return nil, err
	}
	return C.jobject(jCaveat), nil
}

// GoCaveat converts the provided Java Caveat into a Go Caveat.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoCaveat(jEnv, jCav interface{}) (security.Caveat, error) {
	jCaveat := getObject(jCav)
	encoded, err := jutil.CallStaticByteArrayMethod(jEnv, jUtilClass, "encodeCaveat", []jutil.Sign{caveatSign}, jCaveat)
	if err != nil {
		return security.Caveat{}, err
	}
	var caveat security.Caveat
	if err := vom.Decode(encoded, &caveat); err != nil {
		return security.Caveat{}, err
	}
	return caveat, nil
}

// JavaCaveats converts the provided Go Caveat slice into a Java Caveat array.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaCaveats(jEnv interface{}, caveats []security.Caveat) (C.jobjectArray, error) {
	cavarr := make([]interface{}, len(caveats))
	for i, caveat := range caveats {
		var err error
		if cavarr[i], err = JavaCaveat(jEnv, caveat); err != nil {
			return nil, err
		}
	}
	jCaveats := jutil.JObjectArray(jEnv, cavarr, jCaveatClass)
	return C.jobjectArray(jCaveats), nil
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
func JavaBlessingPattern(jEnv interface{}, pattern security.BlessingPattern) (C.jobject, error) {
	jBlessingPattern, err := jutil.CallStaticObjectMethod(jEnv, jUtilClass, "decodeBlessingPattern", []jutil.Sign{jutil.StringSign}, blessingPatternSign, string(pattern))
	return C.jobject(jBlessingPattern), err
}

// GoBlessingPattern converts the provided Java BlessingPattern into Go BlessingPattern.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoBlessingPattern(jEnv, jPatt interface{}) (security.BlessingPattern, error) {
	jPattern := getObject(jPatt)
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
func JavaPublicKey(jEnv interface{}, key security.PublicKey) (C.jobject, error) {
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
	return C.jobject(jPublicKey), nil
}

// GoPublicKey converts the provided Java PublicKey into Go PublicKey.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoPublicKey(jEnv, jKey interface{}) (security.PublicKey, error) {
	jPublicKey := getObject(jKey)
	encoded, err := jutil.CallStaticByteArrayMethod(jEnv, jUtilClass, "encodePublicKey", []jutil.Sign{publicKeySign}, jPublicKey)
	if err != nil {
		return nil, err
	}
	return security.UnmarshalPublicKey(encoded)
}

// JavaSignature converts the provided Go Signature into a Java Signature.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaSignature(jEnv interface{}, sig security.Signature) (C.jobject, error) {
	encoded, err := vom.Encode(sig)
	if err != nil {
		return nil, err
	}
	jSignature, err := jutil.CallStaticObjectMethod(jEnv, jUtilClass, "decodeSignature", []jutil.Sign{jutil.ByteArraySign}, signatureSign, encoded)
	if err != nil {
		return nil, err
	}
	return C.jobject(jSignature), nil
}

// GoSignature converts the provided Java Signature into a Go Signature.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoSignature(jEnv, jSig interface{}) (security.Signature, error) {
	jSignature := getObject(jSig)
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
func JavaBlessingPatternWrapper(jEnv interface{}, pattern security.BlessingPattern) (C.jobject, error) {
	jPattern, err := JavaBlessingPattern(jEnv, pattern)
	if err != nil {
		return nil, err
	}
	jWrapper, err := jutil.NewObject(jEnv, jBlessingPatternWrapperClass, []jutil.Sign{jutil.LongSign, blessingPatternSign}, int64(jutil.PtrValue(&pattern)), jPattern)
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&pattern) // Un-refed when the Java BlessingPatternWrapper object is finalized.
	return C.jobject(jWrapper), nil
}

func getObject(jObj interface{}) C.jobject {
	return C.jobject(unsafe.Pointer(jutil.PtrValue(jObj)))
}
