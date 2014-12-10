// +build android

package security

import (
	"fmt"
	"runtime"
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
		return nil, nil
	}
	wire := security.MarshalBlessings(blessings)
	encoded, err := jutil.VomEncode(wire)
	if err != nil {
		return nil, err
	}
	jBlessings, err := jutil.CallStaticObjectMethod(jEnv, jUtilClass, "decodeBlessings", []jutil.Sign{jutil.ByteArraySign}, blessingsSign, encoded)
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
	encoded, err := jutil.CallStaticByteArrayMethod(jEnv, jUtilClass, "encodeBlessings", []jutil.Sign{blessingsSign}, jBlessings)
	if err != nil {
		return nil, err
	}
	var wire security.WireBlessings
	if err := jutil.VomDecode(encoded, &wire); err != nil {
		return nil, err
	}
	return security.NewBlessings(wire)
}

// JavaWireBlessings converts the provided Go WireBlessings into Java WireBlessings.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaWireBlessings(jEnv interface{}, wire security.WireBlessings) (C.jobject, error) {
	encoded, err := jutil.VomEncode(wire)
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
	if err := jutil.VomDecode(encoded, &wire); err != nil {
		return security.WireBlessings{}, err
	}
	return wire, nil
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
	encoded, err := jutil.VomEncode(sig)
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
	if err := jutil.VomDecode(encoded, &sig); err != nil {
		return security.Signature{}, err
	}
	return sig, nil
}

// JavaCaveat converts the provided Go Caveat into a Java Caveat.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaCaveat(jEnv interface{}, caveat security.Caveat) (C.jobject, error) {
	encoded, err := jutil.VomEncode(caveat)
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
	if err := jutil.VomDecode(encoded, &caveat); err != nil {
		return security.Caveat{}, err
	}
	return caveat, nil
}

// JavaCaveats converts the provided Go Caveat slice into a Java Caveat array.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaCaveats(jEnv interface{}, caveats []security.Caveat) (C.jobjectArray, error) {
	encoded, err := jutil.VomEncode(caveats)
	if err != nil {
		return nil, err
	}
	jCaveats, err := jutil.CallStaticObjectMethod(jEnv, jUtilClass, "decodeCaveats", []jutil.Sign{jutil.ByteArraySign}, jutil.ArraySign(caveatSign), encoded)
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
	encoded, err := jutil.CallStaticByteArrayMethod(jEnv, jUtilClass, "encodeCaveats", []jutil.Sign{jutil.ArraySign(caveatSign)}, jCaveats)
	if err != nil {
		return nil, err
	}
	var caveats []security.Caveat
	if err := jutil.VomDecode(encoded, &caveats); err != nil {
		return nil, err
	}
	return caveats, nil
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
	jTags := jutil.JObjectArray(jEnv, tagsJava)
	return C.jobjectArray(jTags), nil
}

func getObject(jObj interface{}) C.jobject {
	return C.jobject(unsafe.Pointer(jutil.PtrValue(jObj)))
}
