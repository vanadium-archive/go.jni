// +build android

package security

import (
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"unsafe"

	jutil "veyron.io/jni/util"
	"veyron.io/veyron/veyron/security/acl"
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
		return nil, nil
	}
	wire, err := extractWire(blessings)
	if err != nil {
		return nil, err
	}
	encoded, err := encodeWireBlessings(wire)
	if err != nil {
		return nil, err
	}
	jBlessings, err := jutil.CallStaticObjectMethod(jEnv, jUtilClass, "decodeBlessings", []jutil.Sign{jutil.StringSign}, blessingsSign, encoded)
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
	jBlessings := getObject(jBless)

	if jBlessings == nil {
		return nil, nil
	}
	encoded, err := jutil.CallStaticStringMethod(jEnv, jUtilClass, "encodeBlessings", []jutil.Sign{blessingsSign}, jBlessings)
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
	encoded, err := encodeWireBlessings(wire)
	if err != nil {
		return nil, err
	}
	jWireBlessings, err := jutil.CallStaticObjectMethod(jEnv, jUtilClass, "decodeWireBlessings", []jutil.Sign{jutil.StringSign}, wireBlessingsSign, encoded)
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

	encoded, err := jutil.CallStaticStringMethod(jEnv, jUtilClass, "encodeWireBlessings", []jutil.Sign{wireBlessingsSign}, jWireBlessings)
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
	encoded, err := json.Marshal(sig)
	if err != nil {
		return nil, err
	}
	jSignature, err := jutil.CallStaticObjectMethod(jEnv, jUtilClass, "decodeSignature", []jutil.Sign{jutil.StringSign}, signatureSign, string(encoded))
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
	encoded, err := jutil.CallStaticStringMethod(jEnv, jUtilClass, "encodeSignature", []jutil.Sign{signatureSign}, jSignature)
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
	encoded, err := json.Marshal(caveat)
	if err != nil {
		return nil, err
	}
	jCaveat, err := jutil.CallStaticObjectMethod(jEnv, jUtilClass, "decodeCaveat", []jutil.Sign{jutil.StringSign}, caveatSign, string(encoded))
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
	encoded, err := jutil.CallStaticStringMethod(jEnv, jUtilClass, "encodeCaveat", []jutil.Sign{caveatSign}, jCaveat)
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
	encoded, err := json.Marshal(caveats)
	if err != nil {
		return nil, err
	}
	jCaveats, err := jutil.CallStaticObjectMethod(jEnv, jUtilClass, "decodeCaveats", []jutil.Sign{jutil.StringSign}, jutil.ArraySign(caveatSign), string(encoded))
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
	jCaveats := getObject(jCavs)
	encoded, err := jutil.CallStaticStringMethod(jEnv, jUtilClass, "encodeCaveats", []jutil.Sign{jutil.ArraySign(caveatSign)}, jCaveats)
	if err != nil {
		return nil, err
	}
	var caveats []security.Caveat
	if err := json.Unmarshal([]byte(encoded), &caveats); err != nil {
		return nil, fmt.Errorf("couldn't JSON-decode caveats %q: %v", encoded, err)
	}
	return caveats, nil
}

// JavaACL converts the provided Go ACL into a Java ACL.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaACL(jEnv interface{}, acl acl.ACL) (C.jobject, error) {
	encoded, err := json.Marshal(acl)
	if err != nil {
		return nil, err
	}
	jACL, err := jutil.CallStaticObjectMethod(jEnv, jUtilClass, "decodeACL", []jutil.Sign{jutil.StringSign}, aclSign, string(encoded))
	if err != nil {
		return nil, err
	}
	return C.jobject(jACL), nil
}

// GoACL converts the provided Java ACL into a Go ACL.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoACL(jEnv, jACL interface{}) (acl.ACL, error) {
	encoded, err := jutil.CallStaticStringMethod(jEnv, jUtilClass, "encodeACL", []jutil.Sign{aclSign}, jACL)
	if err != nil {
		return acl.ACL{}, err
	}
	var a acl.ACL
	if err := json.Unmarshal([]byte(encoded), &a); err != nil {
		return acl.ACL{}, fmt.Errorf("couldn't JSON-decode ACL %q: %v", encoded, err)
	}
	return a, nil
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

// JavaACLWrapper converts the provided Go ACL into a Java ACLWrapper object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaACLWrapper(jEnv interface{}, acl acl.ACL) (C.jobject, error) {
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

// javaTag is a placeholder for a tag that was obtained from Java.
type javaTag struct {
	jTag C.jobject
	jVM  *C.JavaVM
}

// GoTags converts the provided Java tags into Go tags.  These tags will be mostly
// useless except for later conversion back to Java tags.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoTags(jEnv, jTags interface{}) ([]interface{}, error) {
	env := (*C.JNIEnv)(unsafe.Pointer(jutil.PtrValue(jEnv)))

	// We cannot cache Java environments as they are only valid in the current
	// thread.  We can, however, cache the Java VM and obtain an environment
	// from it in whatever thread happens to be running at the time.
	var jVM *C.JavaVM
	if status := C.GetJavaVM(env, &jVM); status != 0 {
		return nil, fmt.Errorf("couldn't get Java VM from the (Java) environment")
	}

	tagsJava := jutil.GoObjectArray(env, jTags)
	if tagsJava == nil {
		return nil, nil
	}
	tags := make([]interface{}, len(tagsJava))
	for i, tag := range tagsJava {
		jniTag := &javaTag{
			// Reference the Java tag; it will be de-referenced when this Go tag
			// is garbage-collected (through the finalizer callback we setup
			// just below).
			jTag: C.NewGlobalRef(env, C.jobject(tag)),
			jVM:  jVM,
		}
		runtime.SetFinalizer(jniTag, func(t *javaTag) {
			jEnv, freeFunc := jutil.GetEnv(t.jVM)
			env := (*C.JNIEnv)(jEnv)
			defer freeFunc()
			C.DeleteGlobalRef(env, t.jTag)
		})
		tags[i] = jniTag
	}
	return tags, nil
}

// JavaTags converts the provided Go tags into Java tags.  It assumes that the
// tags were produced by a call to GoTags above.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaTags(jEnv interface{}, tags []interface{}) (C.jobjectArray, error) {
	if tags == nil {
		return nil, nil
	}
	tagsJava := make([]interface{}, len(tags))
	for i, tag := range tags {
		// Make sure that that the tag is a Java tag.  (That must be the case
		// because the tags are injected into the Veyron runtime by the
		// invoker's Prepare() method, which in out runtime implementation
		// obtains the tags from Java.)
		jniTag, ok := tag.(*javaTag)
		if !ok {
			return nil, fmt.Errorf("Encountered method tag of unsupported type %T: %v", tag, tag)
		}
		tagsJava[i] = jniTag.jTag
	}
	jTags, err := jutil.JObjectArray(jEnv, tagsJava)
	if err != nil {
		return nil, err
	}
	return C.jobjectArray(jTags), nil
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

func getObject(jObj interface{}) C.jobject {
	return C.jobject(unsafe.Pointer(jutil.PtrValue(jObj)))
}
