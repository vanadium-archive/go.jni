// +build android

package security

import (
	"fmt"
	"log"
	"runtime"
	"unsafe"

	"v.io/v23/security"
	jutil "v.io/x/jni/util"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// JavaPrincipal converts the provided Go Principal into a Java Principal
// object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaPrincipal(jEnv interface{}, principal security.Principal) (unsafe.Pointer, error) {
	if principal == nil {
		return nil, nil
	}
	jPrincipal, err := jutil.NewObject(jEnv, jPrincipalImplClass, []jutil.Sign{jutil.LongSign, signerSign, blessingStoreSign, blessingRootsSign}, int64(jutil.PtrValue(&principal)), C.jobject(nil), C.jobject(nil), C.jobject(nil))
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&principal) // Un-refed when the Java PrincipalImpl is finalized.
	return jPrincipal, nil
}

// GoPrincipal converts the provided Java Principal object into a Go Principal.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoPrincipal(jEnv, jPrincipalObj interface{}) (security.Principal, error) {
	env := (*C.JNIEnv)(unsafe.Pointer(jutil.PtrValue(jEnv)))
	jPrincipal := C.jobject(unsafe.Pointer(jutil.PtrValue(jPrincipalObj)))

	if jPrincipal == nil {
		return nil, nil
	}
	// Reference Java Principal; it will be de-referenced when the Go Principal
	// created below is garbage-collected (through the finalizer callback we
	// setup just below).
	jPrincipal = C.NewGlobalRef(env, jPrincipal)
	// Create Go Principal.
	p := &principal{
		jPrincipal: jPrincipal,
	}
	runtime.SetFinalizer(p, func(p *principal) {
		jEnv, freeFunc := jutil.GetEnv()
		env := (*C.JNIEnv)(jEnv)
		defer freeFunc()
		C.DeleteGlobalRef(env, p.jPrincipal)
	})
	return p, nil
}

type principal struct {
	jPrincipal C.jobject
}

func (p *principal) Bless(key security.PublicKey, with security.Blessings, extension string, caveat security.Caveat, additionalCaveats ...security.Caveat) (security.Blessings, error) {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()

	jKey, err := JavaPublicKey(env, key)
	if err != nil {
		return security.Blessings{}, err
	}
	jWith, err := JavaBlessings(env, with)
	if err != nil {
		return security.Blessings{}, err
	}
	jCaveat, err := JavaCaveat(env, caveat)
	if err != nil {
		return security.Blessings{}, err
	}
	jAdditionalCaveats, err := JavaCaveats(env, additionalCaveats)
	if err != nil {
		return security.Blessings{}, err
	}
	jBlessings, err := jutil.CallObjectMethod(env, p.jPrincipal, "bless", []jutil.Sign{publicKeySign, blessingsSign, jutil.StringSign, caveatSign, jutil.ArraySign(caveatSign)}, blessingsSign, jKey, jWith, extension, jCaveat, jAdditionalCaveats)
	if err != nil {
		return security.Blessings{}, err
	}
	return GoBlessings(env, jBlessings)
}

func (p *principal) BlessSelf(name string, caveats ...security.Caveat) (security.Blessings, error) {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jCaveats, err := JavaCaveats(env, caveats)
	if err != nil {
		return security.Blessings{}, err
	}
	jBlessings, err := jutil.CallObjectMethod(env, p.jPrincipal, "blessSelf", []jutil.Sign{jutil.StringSign, jutil.ArraySign(caveatSign)}, blessingsSign, name, jCaveats)
	if err != nil {
		return security.Blessings{}, err
	}
	return GoBlessings(env, jBlessings)
}

func (p *principal) Sign(message []byte) (security.Signature, error) {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jSig, err := jutil.CallObjectMethod(env, p.jPrincipal, "sign", []jutil.Sign{jutil.ArraySign(jutil.ByteSign)}, signatureSign, message)
	if err != nil {
		return security.Signature{}, err
	}
	return GoSignature(env, jSig)
}

func (p *principal) MintDischarge(forThirdPartyCaveat, caveatOnDischarge security.Caveat, additionalCaveatsOnDischarge ...security.Caveat) (security.Discharge, error) {
	return security.Discharge{}, fmt.Errorf("MintDischarge not yet implemented")
}

func (p *principal) PublicKey() security.PublicKey {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jPublicKey, err := jutil.CallObjectMethod(env, p.jPrincipal, "publicKey", nil, publicKeySign)
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

func (p *principal) BlessingsByName(name security.BlessingPattern) []security.Blessings {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jName, err := JavaBlessingPattern(env, name)
	if err != nil {
		log.Printf("Couldn't convert Go blessing pattern: %v", err)
		return nil
	}
	barr, err := jutil.CallObjectArrayMethod(env, p.jPrincipal, "blessingsByName", []jutil.Sign{blessingPatternSign}, blessingsSign, jName)
	if err != nil {
		log.Printf("Couldn't get Java blessings for name: %v", err)
		return nil
	}
	ret := make([]security.Blessings, len(barr))
	for i, jBlessings := range barr {
		var err error
		if ret[i], err = GoBlessings(env, jBlessings); err != nil {
			log.Printf("Couldn't convert Java blessings to Go: %v", err)
			return nil
		}
	}
	return ret
}

func (p *principal) BlessingsInfo(blessings security.Blessings) map[string][]security.Caveat {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		log.Printf("Couldn't convert Go blessings to Java: %v", err)
		return nil
	}
	infomap, err := jutil.CallMapMethod(env, p.jPrincipal, "blessingsInfo", []jutil.Sign{blessingsSign}, jBlessings)
	if err != nil {
		log.Printf("Couldn't get Java blessings info: %v", err)
		return nil
	}
	ret := make(map[string][]security.Caveat)
	for jName, jCaveats := range infomap {
		name := jutil.GoString(env, jName)
		caveats, err := GoCaveats(env, jCaveats)
		if err != nil {
			log.Printf("Couldn't convert Java Caveats to Go: %v", err)
			return nil
		}
		ret[name] = caveats
	}
	return ret
}

func (p *principal) BlessingStore() security.BlessingStore {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jBlessingStore, err := jutil.CallObjectMethod(env, p.jPrincipal, "blessingStore", nil, blessingStoreSign)
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
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jBlessingRoots, err := jutil.CallObjectMethod(env, p.jPrincipal, "roots", nil, blessingRootsSign)
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
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		return err
	}
	return jutil.CallVoidMethod(env, p.jPrincipal, "addToRoots", []jutil.Sign{blessingsSign}, jBlessings)
}
