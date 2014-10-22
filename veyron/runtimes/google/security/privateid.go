// +build android

package security

import (
	"fmt"
	"runtime"
	"time"
	"unsafe"

	"veyron.io/jni/util"
	"veyron.io/veyron/veyron2/security"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// NewPrivateID creates an instance of security.PrivateID that uses the provided
// Java PrivateID as its underlying implementation.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func NewPrivateID(jEnv, jPrivID interface{}) security.PrivateID {
	env := (*C.JNIEnv)(unsafe.Pointer(util.PtrValue(jEnv)))
	jPrivateID := C.jobject(unsafe.Pointer(util.PtrValue(jPrivID)))

	// We cannot cache Java environments as they are only valid in the current
	// thread.  We can, however, cache the Java VM and obtain an environment
	// from it in whatever thread happens to be running at the time.
	var jVM *C.JavaVM
	if status := C.GetJavaVM(env, &jVM); status != 0 {
		panic("couldn't get Java VM from the (Java) environment")
	}
	// Reference Java PrivateID; it will be de-referenced when the Go PrivateID
	// created below is garbage-collected (through the finalizer callback we
	// setup just below).
	jPrivateID = C.NewGlobalRef(env, jPrivateID)
	// Create Go PrivateID.
	id := &privateID{
		jVM:        jVM,
		jPrivateID: jPrivateID,
	}
	runtime.SetFinalizer(id, func(id *privateID) {
		jEnv, freeFunc := util.GetEnv(id.jVM)
		env := (*C.JNIEnv)(jEnv)
		defer freeFunc()
		C.DeleteGlobalRef(env, id.jPrivateID)
	})
	return id
}

type privateID struct {
	jVM        *C.JavaVM
	jPrivateID C.jobject
}

func (id *privateID) PublicID() security.PublicID {
	jEnv, freeFunc := util.GetEnv(id.jVM)
	env := (*C.JNIEnv)(jEnv)
	defer freeFunc()
	jPublicID := C.jobject(util.CallObjectMethodOrCatch(env, id.jPrivateID, "publicID", nil, publicIDSign))
	publicIDPtr := util.CallLongMethodOrCatch(env, jPublicID, "getNativePtr", nil)
	return (*(*security.PublicID)(util.Ptr(publicIDPtr)))
}

func (id *privateID) Bless(blessee security.PublicID, blessingName string, duration time.Duration, caveats []security.Caveat) (security.PublicID, error) {
	if len(caveats) > 0 {
		return nil, fmt.Errorf("Caveats currently not supported in Java.")
	}
	jEnv, freeFunc := util.GetEnv(id.jVM)
	env := (*C.JNIEnv)(jEnv)
	defer freeFunc()
	util.GoRef(&blessee) // Un-refed when the Java blessee object created below is finalized.
	jBlessee := C.jobject(util.NewObjectOrCatch(env, jPublicIDImplClass, []util.Sign{util.LongSign}, &blessee))
	jDuration := C.jobject(util.NewObjectOrCatch(env, jDurationClass, []util.Sign{util.LongSign}, int64(duration.Seconds()*1000)))
	durationSign := util.ClassSign("org.joda.time.Duration")
	jPublicID, err := util.CallObjectMethod(env, id.jPrivateID, "bless", []util.Sign{publicIDSign, util.StringSign, durationSign}, publicIDSign, jBlessee, blessingName, jDuration)
	if err != nil {
		return nil, err
	}
	publicIDPtr := util.CallLongMethodOrCatch(env, jPublicID, "getNativePtr", nil)
	return (*(*security.PublicID)(util.Ptr(publicIDPtr))), nil
}

func (id *privateID) Derive(publicID security.PublicID) (security.PrivateID, error) {
	jEnv, freeFunc := util.GetEnv(id.jVM)
	env := (*C.JNIEnv)(jEnv)
	defer freeFunc()
	util.GoRef(&publicID) // Un-refed when the Java PublicID object created below is finalized.
	jPublicID := C.jobject(util.NewObjectOrCatch(env, jPublicIDImplClass, []util.Sign{util.LongSign}, &publicID))
	privateIDSign := util.ClassSign("io.veyron.veyron.veyron2.security.PrivateID")
	jPrivateID, err := util.CallObjectMethod(env, id.jPrivateID, "derive", []util.Sign{publicIDSign}, privateIDSign, jPublicID)
	if err != nil {
		return nil, err
	}
	return NewPrivateID(env, C.jobject(jPrivateID)), nil
}

func (id *privateID) MintDischarge(caveat security.ThirdPartyCaveat, context security.Context, duration time.Duration, caveats []security.Caveat) (security.Discharge, error) {
	return nil, fmt.Errorf("MintDischarge currently not implemented.")
}

func (id *privateID) Sign(message []byte) (security.Signature, error) {
	jEnv, freeFunc := util.GetEnv(id.jVM)
	env := (*C.JNIEnv)(jEnv)
	defer freeFunc()
	signatureSign := util.ClassSign("io.veyron.veyron.veyron2.security.Signature")
	jSig, err := util.CallObjectMethod(env, id.jPrivateID, "sign", []util.Sign{util.ArraySign(util.ByteSign)}, signatureSign, message)
	if err != nil {
		return security.Signature{}, err
	}
	jHash := util.CallObjectMethodOrCatch(env, jSig, "getHash", nil, util.ClassSign("io.veyron.veyron.veyron2.security.Hash"))
	sig := security.Signature{
		Purpose: nil,
		Hash:    security.Hash(util.CallStringMethodOrCatch(env, jHash, "getValue", nil)),
		R:       util.CallByteArrayMethodOrCatch(env, jSig, "getR", nil),
		S:       util.CallByteArrayMethodOrCatch(env, jSig, "getS", nil),
	}
	return sig, nil
}

func (id *privateID) PublicKey() security.PublicKey {
	jEnv, freeFunc := util.GetEnv(id.jVM)
	env := (*C.JNIEnv)(jEnv)
	defer freeFunc()
	publicKeySign := util.ClassSign("java.security.interfaces.ECPublicKey")
	jPublicKey := C.jobject(util.CallObjectMethodOrCatch(env, id.jPrivateID, "publicKey", nil, publicKeySign))
	// Get the encoded version of the public key.
	encoded := util.CallByteArrayMethodOrCatch(env, jPublicKey, "getEncoded", nil)
	key, err := security.UnmarshalPublicKey(encoded)
	if err != nil {
		panic("couldn't parse Java ECDSA public key: " + err.Error())
	}
	return key
}
