// +build android

package security

import (
	"fmt"
	"log"
	"runtime"
	"unsafe"

	jutil "veyron.io/jni/util"
	"veyron.io/veyron/veyron2/security"
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
	// We cannot cache Java environments as they are only valid in the current
	// thread.  We can, however, cache the Java VM and obtain an environment
	// from it in whatever thread happens to be running at the time.
	var jVM *C.JavaVM
	if status := C.GetJavaVM(env, &jVM); status != 0 {
		return nil, fmt.Errorf("couldn't get Java VM from the (Java) environment")
	}
	// Reference Java Signer; it will be de-referenced when the Go Signer
	// created below is garbage-collected (through the finalizer callback we
	// setup just below).
	jSigner = C.NewGlobalRef(env, jSigner)
	s := &signer{
		jVM:     jVM,
		jSigner: jSigner,
	}
	runtime.SetFinalizer(s, func(s *signer) {
		jEnv, freeFunc := jutil.GetEnv(s.jVM)
		env := (*C.JNIEnv)(jEnv)
		defer freeFunc()
		C.DeleteGlobalRef(env, s.jSigner)
	})
	return s, nil
}

type signer struct {
	jVM     *C.JavaVM
	jSigner C.jobject
}

func (s *signer) Sign(purpose, message []byte) (security.Signature, error) {
	env, freeFunc := jutil.GetEnv(s.jVM)
	defer freeFunc()
	signatureSign := jutil.ClassSign("io.veyron.veyron.veyron2.security.Signature")
	jSig, err := jutil.CallObjectMethod(env, s.jSigner, "sign", []jutil.Sign{jutil.ArraySign(jutil.ByteSign), jutil.ArraySign(jutil.ByteSign)}, signatureSign, purpose, message)
	if err != nil {
		return security.Signature{}, err
	}
	return GoSignature(env, jSig)
}

func (s *signer) PublicKey() security.PublicKey {
	env, freeFunc := jutil.GetEnv(s.jVM)
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
