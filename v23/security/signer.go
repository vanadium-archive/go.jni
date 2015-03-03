// +build android

package security

import (
	"log"
	"runtime"
	"unsafe"

	"v.io/v23/security"
	jutil "v.io/x/jni/util"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// GoSigner creates an instance of security.Signer that uses the provided
// Java Signer as its underlying implementation.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoSigner(jEnv interface{}, jSigner C.jobject) (security.Signer, error) {
	env := (*C.JNIEnv)(unsafe.Pointer(jutil.PtrValue(jEnv)))
	// Reference Java Signer; it will be de-referenced when the Go Signer
	// created below is garbage-collected (through the finalizer callback we
	// setup just below).
	jSigner = C.NewGlobalRef(env, jSigner)
	s := &signer{
		jSigner: jSigner,
	}
	runtime.SetFinalizer(s, func(s *signer) {
		jEnv, freeFunc := jutil.GetEnv()
		env := (*C.JNIEnv)(jEnv)
		defer freeFunc()
		C.DeleteGlobalRef(env, s.jSigner)
	})
	return s, nil
}

type signer struct {
	jSigner C.jobject
}

func (s *signer) Sign(purpose, message []byte) (security.Signature, error) {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	signatureSign := jutil.ClassSign("io.v.v23.security.Signature")
	jSig, err := jutil.CallObjectMethod(env, s.jSigner, "sign", []jutil.Sign{jutil.ArraySign(jutil.ByteSign), jutil.ArraySign(jutil.ByteSign)}, signatureSign, purpose, message)
	if err != nil {
		return security.Signature{}, err
	}
	return GoSignature(env, jSig)
}

func (s *signer) PublicKey() security.PublicKey {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	publicKeySign := jutil.ClassSign("java.security.interfaces.ECPublicKey")
	jPublicKey, err := jutil.CallObjectMethod(env, s.jSigner, "publicKey", nil, publicKeySign)
	if err != nil {
		log.Printf("Couldn't get Java public key: %v", err)
		return nil
	}
	// Get the encoded version of the public key.
	encoded, err := jutil.CallByteArrayMethod(env, jPublicKey, "getEncoded", nil)
	if err != nil {
		log.Printf("Couldn't get encoded data for Java public key: %v", err)
		return nil
	}
	key, err := security.UnmarshalPublicKey(encoded)
	if err != nil {
		log.Printf("Couldn't parse Java ECDSA public key: " + err.Error())
		return nil
	}
	return key
}
