// +build android

package security

import (
	"encoding/json"
	"fmt"
	"reflect"
	"unsafe"

	jutil "veyron.io/jni/util"
	"veyron.io/veyron/veyron2/security"
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
		return C.jobject(nil), nil
	}
	env := getEnv(jEnv)
	wire, err := extractWire(blessings)
	if err != nil {
		return nil, err
	}
	encoded, err := encodeWireBlessings(wire)
	if err != nil {
		return nil, err
	}
	jBlessings, err := jutil.CallStaticObjectMethod(env, jUtilClass, "decodeBlessings", []jutil.Sign{jutil.StringSign}, blessingsSign, encoded)
	if err != nil {
		return nil, err
	}
	return C.jobject(jBlessings), nil
}

// GoBlessings converts the provided Java Blessings into Go Blessings.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoBlessings(jEnv, jBless interface{}) (security.Blessings, error) {
	env := getEnv(jEnv)
	jBlessings := getObject(jBless)

	if jBlessings == C.jobject(nil) {
		return nil, nil
	}
	encoded, err := jutil.CallStaticStringMethod(env, jUtilClass, "encodeBlessings", []jutil.Sign{blessingsSign}, jBlessings)
	if err != nil {
		return nil, err
	}
	wire, err := decodeWireBlessings(encoded)
	if err != nil {
		return nil, err
	}
	return security.NewBlessings(wire)
}

// JavaWireBlessings converts the provided Go WireBlessings into Java WireBlessings.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaWireBlessings(jEnv interface{}, wire security.WireBlessings) (C.jobject, error) {
	env := getEnv(jEnv)
	encoded, err := encodeWireBlessings(wire)
	if err != nil {
		return nil, err
	}
	jWireBlessings, err := jutil.CallStaticObjectMethod(env, jUtilClass, "decodeWireBlessings", []jutil.Sign{jutil.StringSign}, wireBlessingsSign, encoded)
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
	env := getEnv(jEnv)
	jWireBlessings := getObject(jWireBless)

	encoded, err := jutil.CallStaticStringMethod(env, jUtilClass, "encodeWireBlessings", []jutil.Sign{wireBlessingsSign}, jWireBlessings)
	if err != nil {
		return security.WireBlessings{}, err
	}

	return decodeWireBlessings(encoded)
}

// JavaBlessingPattern converts the provided Go BlessingPattern into Java BlessingPattern.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaBlessingPattern(jEnv interface{}, pattern security.BlessingPattern) (C.jobject, error) {
	env := getEnv(jEnv)
	jBlessingPattern, err := jutil.CallStaticObjectMethod(env, jUtilClass, "decodeBlessingPattern", []jutil.Sign{jutil.StringSign}, blessingPatternSign, string(pattern))
	return C.jobject(jBlessingPattern), err
}

// GoBlessingPattern converts the provided Java BlessingPattern into Go BlessingPattern.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoBlessingPattern(jEnv, jPatt interface{}) (security.BlessingPattern, error) {
	env := getEnv(jEnv)
	jPattern := getObject(jPatt)
	encoded, err := jutil.CallStaticStringMethod(env, jUtilClass, "encodeBlessingPattern", []jutil.Sign{blessingPatternSign}, jPattern)
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
		return C.jobject(nil), nil
	}
	env := getEnv(jEnv)
	encoded, err := key.MarshalBinary()
	if err != nil {
		return nil, err
	}
	jPublicKey, err := jutil.CallStaticObjectMethod(env, jUtilClass, "decodePublicKey", []jutil.Sign{jutil.ArraySign(jutil.ByteSign)}, publicKeySign, encoded)
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
	env := getEnv(jEnv)
	jPublicKey := getObject(jKey)
	encoded, err := jutil.CallStaticByteArrayMethod(env, jUtilClass, "encodePublicKey", []jutil.Sign{publicKeySign}, jPublicKey)
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
	env := getEnv(jEnv)
	encoded, err := json.Marshal(sig)
	if err != nil {
		return nil, err
	}
	jSignature, err := jutil.CallStaticObjectMethod(env, jUtilClass, "decodeSignature", []jutil.Sign{jutil.StringSign}, signatureSign, encoded)
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
	env := getEnv(jEnv)
	jSignature := getObject(jSig)
	encoded, err := jutil.CallStaticStringMethod(env, jUtilClass, "encodeSignature", []jutil.Sign{signatureSign}, jSignature)
	if err != nil {
		return security.Signature{}, err
	}
	var sig security.Signature
	if err := json.Unmarshal([]byte(encoded), &sig); err != nil {
		return security.Signature{}, fmt.Errorf("couldn't JSON-decode Signature %q: %v", encoded, err)
	}
	return sig, nil
}

// JavaCaveat converts the provided Go Caveat into a Java Caveat.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaCaveat(jEnv interface{}, caveat security.Caveat) (C.jobject, error) {
	env := getEnv(jEnv)
	encoded, err := json.Marshal(caveat)
	if err != nil {
		return nil, err
	}
	jCaveat, err := jutil.CallStaticObjectMethod(env, jUtilClass, "decodeCaveat", []jutil.Sign{jutil.StringSign}, caveatSign, encoded)
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
	env := getEnv(jEnv)
	jCaveat := getObject(jCav)
	encoded, err := jutil.CallStaticStringMethod(env, jUtilClass, "encodeCaveat", []jutil.Sign{caveatSign}, jCaveat)
	if err != nil {
		return security.Caveat{}, err
	}
	var caveat security.Caveat
	if err := json.Unmarshal([]byte(encoded), &caveat); err != nil {
		return security.Caveat{}, fmt.Errorf("couldn't JSON-decode Caveat %q: %v", encoded, err)
	}
	return caveat, nil
}

// JavaCaveats converts the provided Go Caveat slice into a Java Caveat array.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaCaveats(jEnv interface{}, caveats []security.Caveat) (C.jobjectArray, error) {
	env := getEnv(jEnv)
	encoded, err := json.Marshal(caveats)
	if err != nil {
		return nil, err
	}
	jCaveats, err := jutil.CallStaticObjectMethod(env, jUtilClass, "decodeCaveats", []jutil.Sign{jutil.StringSign}, jutil.ArraySign(caveatSign), encoded)
	if err != nil {
		return nil, err
	}
	return C.jobjectArray(jCaveats), nil

}

// GoCaveats converts the provided Java Caveat array into a Go Caveat slice.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoCaveats(jEnv, jCavs interface{}) ([]security.Caveat, error) {
	env := getEnv(jEnv)
	jCaveats := getObject(jCavs)
	encoded, err := jutil.CallStaticStringMethod(env, jUtilClass, "encodeCaveats", []jutil.Sign{jutil.ArraySign(caveatSign)}, jCaveats)
	if err != nil {
		return nil, err
	}
	var caveats []security.Caveat
	if err := json.Unmarshal([]byte(encoded), &caveats); err != nil {
		return nil, fmt.Errorf("couldn't JSON-decode caveats %q: %v", encoded, err)
	}
	return caveats, nil
}

// extractWire extracts WireBlessings from the provided Blessings.
func extractWire(blessings security.Blessings) (security.WireBlessings, error) {
	m := reflect.ValueOf(blessings).MethodByName("VomEncode")
	if !m.IsValid() {
		return security.WireBlessings{}, fmt.Errorf("type %T doesn't implement VomEncode()", blessings)
	}
	results := m.Call(nil)
	if len(results) != 2 {
		return security.WireBlessings{}, fmt.Errorf("wrong number of return arguments for %T.VomEncode(), want 2, have %d", blessings, len(results))
	}
	if !results[1].IsNil() {
		err, ok := results[1].Interface().(error)
		if !ok {
			return security.WireBlessings{}, fmt.Errorf("second return argument must be an error, got %T", results[1].Interface())
		}
		return security.WireBlessings{}, fmt.Errorf("error invoking VomEncode(): %v", err)
	}
	result, ok := results[0].Interface().(security.WireBlessings)
	if !ok {
		return security.WireBlessings{}, fmt.Errorf("unexpected return value of type %T for VomEncode", result)
	}
	return result, nil
}

// encodeWireBlessings JSON-encodes the provided set of WireBlessings.
func encodeWireBlessings(wire security.WireBlessings) (string, error) {
	enc, err := json.Marshal(wire)
	if err != nil {
		return "", err
	}
	return string(enc), nil
}

// decodeWireBlessings decodes the provided JSON-encoded WireBlessings.
func decodeWireBlessings(encoded string) (security.WireBlessings, error) {
	// JSON-decode chains.
	var wire security.WireBlessings
	if err := json.Unmarshal([]byte(encoded), &wire); err != nil {
		return security.WireBlessings{}, fmt.Errorf("couldn't JSON-decode WireBlessings %q: %v", encoded, err)
	}
	return wire, nil
}

// Various functions that cast CGO types from other packages into this package's types.
func getEnv(jEnv interface{}) *C.JNIEnv {
	return (*C.JNIEnv)(unsafe.Pointer(jutil.PtrValue(jEnv)))
}
func getObject(jObj interface{}) C.jobject {
	return C.jobject(unsafe.Pointer(jutil.PtrValue(jObj)))
}
