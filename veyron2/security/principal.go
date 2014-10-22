// +build android

package security

import (
	"fmt"
	"log"
	"runtime"
	"unsafe"

	"veyron.io/jni/util"
	"veyron.io/veyron/veyron2/security"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// JavaPrincipal creates an instance of Java Principal that uses the provided Go
// Principal as its underlying implementation.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaPrincipal(jEnv interface{}, principal security.Principal) (C.jobject, error) {
	env := (*C.JNIEnv)(unsafe.Pointer(util.PtrValue(jEnv)))
	util.GoRef(&principal) // Un-refed when the Java PrincipalImpl is finalized.
	jObj, err := util.NewObject(env, jPrincipalImplClass, []util.Sign{util.LongSign, signerSign, blessingStoreSign, blessingRootsSign}, &principal, C.jobject(nil), C.jobject(nil), C.jobject(nil))
	return C.jobject(jObj), err
}

// GoPrincipal creates an instance of security.Principal that uses the provided
// Java Principal as its underlying implementation.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoPrincipal(jEnv, jPrinc interface{}) (security.Principal, error) {
	env := (*C.JNIEnv)(unsafe.Pointer(util.PtrValue(jEnv)))
	jPrincipal := C.jobject(unsafe.Pointer(util.PtrValue(jPrinc)))

	// We cannot cache Java environments as they are only valid in the current
	// thread.  We can, however, cache the Java VM and obtain an environment
	// from it in whatever thread happens to be running at the time.
	var jVM *C.JavaVM
	if status := C.GetJavaVM(env, &jVM); status != 0 {
		return nil, fmt.Errorf("couldn't get Java VM from the (Java) environment")
	}
	// Reference Java Principal; it will be de-referenced when the Go Principal
	// created below is garbage-collected (through the finalizer callback we
	// setup just below).
	jPrincipal = C.NewGlobalRef(env, jPrincipal)
	// Create Go Principal.
	p := &principal{
		jVM:        jVM,
		jPrincipal: jPrincipal,
	}
	runtime.SetFinalizer(p, func(p *principal) {
		jEnv, freeFunc := util.GetEnv(p.jVM)
		env := (*C.JNIEnv)(jEnv)
		defer freeFunc()
		C.DeleteGlobalRef(env, p.jPrincipal)
	})
	return p, nil
}

type principal struct {
	jVM        *C.JavaVM
	jPrincipal C.jobject
}

func (p *principal) Bless(key security.PublicKey, with security.Blessings, extension string, caveat security.Caveat, additionalCaveats ...security.Caveat) (security.Blessings, error) {
	jEnv, freeFunc := util.GetEnv(p.jVM)
	env := (*C.JNIEnv)(jEnv)
	defer freeFunc()

	jKey, err := JavaPublicKey(env, key)
	if err != nil {
		return nil, err
	}
	jWith, err := JavaBlessings(env, with)
	if err != nil {
		return nil, err
	}
	jCaveat, err := JavaCaveat(env, caveat)
	if err != nil {
		return nil, err
	}
	jAdditionalCaveats, err := JavaCaveats(env, additionalCaveats)
	if err != nil {
		return nil, err
	}
	jBlessings, err := util.CallObjectMethod(env, p.jPrincipal, "bless", []util.Sign{publicKeySign, blessingsSign, util.StringSign, caveatSign, util.ArraySign(caveatSign)}, blessingsSign, jKey, jWith, extension, jCaveat, jAdditionalCaveats)
	if err != nil {
		return nil, err
	}
	return GoBlessings(env, jBlessings)
}

func (p *principal) BlessSelf(name string, caveats ...security.Caveat) (security.Blessings, error) {
	jEnv, freeFunc := util.GetEnv(p.jVM)
	env := (*C.JNIEnv)(jEnv)
	defer freeFunc()
	jCaveats, err := JavaCaveats(env, caveats)
	if err != nil {
		return nil, err
	}
	jBlessings, err := util.CallObjectMethod(env, p.jPrincipal, "blessSelf", []util.Sign{util.StringSign, util.ArraySign(caveatSign)}, blessingsSign, name, jCaveats)
	if err != nil {
		return nil, err
	}
	return GoBlessings(env, jBlessings)
}

func (p *principal) Sign(message []byte) (security.Signature, error) {
	jEnv, freeFunc := util.GetEnv(p.jVM)
	env := (*C.JNIEnv)(jEnv)
	defer freeFunc()
	jSig, err := util.CallObjectMethod(env, p.jPrincipal, "sign", []util.Sign{util.ArraySign(util.ByteSign)}, signatureSign, message)
	if err != nil {
		return security.Signature{}, err
	}
	return GoSignature(env, jSig)
}

func (p *principal) MintDischarge(tp security.ThirdPartyCaveat, caveat security.Caveat, additionalCaveats ...security.Caveat) (security.Discharge, error) {
	return nil, fmt.Errorf("MintDischarge not yet implemented")
}

func (p *principal) PublicKey() security.PublicKey {
	jEnv, freeFunc := util.GetEnv(p.jVM)
	env := (*C.JNIEnv)(jEnv)
	defer freeFunc()
	jPublicKey, err := util.CallObjectMethod(env, p.jPrincipal, "publicKey", nil, publicKeySign)
	if err != nil {
		log.Printf("Couldn't get Java public key: %v", err)
		return nil
	}
	key, err := GoPublicKey(env, C.jobject(jPublicKey))
	if err != nil {
		log.Printf("Couldn't convert Java public key to Go: %v", err)
		return nil
	}
	return key
}

func (p *principal) BlessingStore() security.BlessingStore {
	jEnv, freeFunc := util.GetEnv(p.jVM)
	env := (*C.JNIEnv)(jEnv)
	defer freeFunc()
	jBlessingStore, err := util.CallObjectMethod(env, p.jPrincipal, "blessingStore", nil, blessingStoreSign)
	if err != nil {
		log.Printf("Couldn't get Java Blessing Store: %v", err)
		return nil
	}
	store, err := GoBlessingStore(env, C.jobject(jBlessingStore))
	if err != nil {
		log.Printf("Couldn't convert Java Blessing Store to Go: %v", err)
		return nil
	}
	return store
}

func (p *principal) Roots() security.BlessingRoots {
	jEnv, freeFunc := util.GetEnv(p.jVM)
	env := (*C.JNIEnv)(jEnv)
	defer freeFunc()
	jBlessingRoots, err := util.CallObjectMethod(env, p.jPrincipal, "roots", nil, blessingRootsSign)
	if err != nil {
		log.Printf("Couldn't get Java Blessing Roots: %v", err)
		return nil
	}
	roots, err := GoBlessingRoots(env, C.jobject(jBlessingRoots))
	if err != nil {
		log.Printf("Couldn't convert Java Blessing Roots to Go: %v", err)
		return nil
	}
	return roots
}

func (p *principal) AddToRoots(blessings security.Blessings) error {
	jEnv, freeFunc := util.GetEnv(p.jVM)
	env := (*C.JNIEnv)(jEnv)
	defer freeFunc()
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		return err
	}
	return util.CallVoidMethod(env, p.jPrincipal, "addToRoots", []util.Sign{blessingsSign}, jBlessings)
}
