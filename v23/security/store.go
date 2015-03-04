// +build android

package security

import (
	"log"
	"runtime"
	"unsafe"

	"v.io/v23/security"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

// JavaBlessingStore creates an instance of Java BlessingStore that uses the provided Go
// BlessingStore as its underlying implementation.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaBlessingStore(jEnv interface{}, store security.BlessingStore) (unsafe.Pointer, error) {
	jObj, err := jutil.NewObject(jEnv, jBlessingStoreImplClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&store)))
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&store) // Un-refed when the Java BlessingStoreImpl is finalized.
	return jObj, nil
}

// GoBlessingStore creates an instance of security.BlessingStore that uses the
// provided Java BlessingStore as its underlying implementation.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoBlessingStore(jEnv, jBlessingStoreObj interface{}) (security.BlessingStore, error) {
	if jBlessingStoreObj == nil {
		return nil, nil
	}
	// Reference Java BlessingStore; it will be de-referenced when the Go
	// BlessingStore created below is garbage-collected (through the finalizer
	// callback we setup just below).
	jBlessingStore := C.jobject(jutil.NewGlobalRef(jEnv, jBlessingStoreObj))
	s := &blessingStore{
		jBlessingStore: jBlessingStore,
	}
	runtime.SetFinalizer(s, func(s *blessingStore) {
		envPtr, freeFunc := jutil.GetEnv()
		env := (*C.JNIEnv)(envPtr)
		defer freeFunc()
		jutil.DeleteGlobalRef(env, s.jBlessingStore)
	})
	return s, nil
}

type blessingStore struct {
	jBlessingStore C.jobject
}

func (s *blessingStore) Set(blessings security.Blessings, forPeers security.BlessingPattern) (security.Blessings, error) {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		return security.Blessings{}, err
	}
	jForPeers, err := JavaBlessingPattern(env, forPeers)
	if err != nil {
		return security.Blessings{}, err
	}
	jOldBlessings, err := jutil.CallObjectMethod(env, s.jBlessingStore, "set", []jutil.Sign{blessingsSign, blessingPatternSign}, blessingsSign, jBlessings, jForPeers)
	if err != nil {
		return security.Blessings{}, err
	}
	return GoBlessings(env, jOldBlessings)
}

func (s *blessingStore) ForPeer(peerBlessings ...string) security.Blessings {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jBlessings, err := jutil.CallObjectMethod(env, s.jBlessingStore, "forPeer", []jutil.Sign{jutil.ArraySign(jutil.StringSign)}, blessingsSign, peerBlessings)
	if err != nil {
		log.Printf("Couldn't call Java forPeer method: %v", err)
		return security.Blessings{}
	}
	blessings, err := GoBlessings(env, jBlessings)
	if err != nil {
		log.Printf("Couldn't convert Java Blessings into Go: %v", err)
		return security.Blessings{}
	}
	return blessings
}

func (s *blessingStore) SetDefault(blessings security.Blessings) error {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		return err
	}
	return jutil.CallVoidMethod(env, s.jBlessingStore, "setDefaultBlessings", []jutil.Sign{blessingsSign}, jBlessings)
}

func (s *blessingStore) Default() security.Blessings {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jBlessings, err := jutil.CallObjectMethod(env, s.jBlessingStore, "defaultBlessings", nil, blessingsSign)
	if err != nil {
		log.Printf("Couldn't call Java defaultBlessings method: %v", err)
		return security.Blessings{}
	}
	blessings, err := GoBlessings(env, jBlessings)
	if err != nil {
		log.Printf("Couldn't convert Java Blessings to Go Blessings: %v", err)
		return security.Blessings{}
	}
	return blessings
}

func (s *blessingStore) PublicKey() security.PublicKey {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jPublicKey, err := jutil.CallObjectMethod(env, s.jBlessingStore, "publicKey", nil, publicKeySign)
	if err != nil {
		log.Printf("Couldn't get Java public key: %v", err)
		return nil
	}
	publicKey, err := GoPublicKey(env, jPublicKey)
	if err != nil {
		log.Printf("Couldn't convert Java ECPublicKey to Go PublicKey: %v", err)
		return nil
	}
	return publicKey
}

func (s *blessingStore) PeerBlessings() map[security.BlessingPattern]security.Blessings {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jBlessingsMap, err := jutil.CallObjectMethod(env, s.jBlessingStore, "peerBlessings", nil, jutil.MapSign)
	if err != nil {
		log.Printf("Couldn't get Java peer blessings: %v", err)
		return nil
	}
	bmap, err := jutil.GoObjectMap(env, jBlessingsMap)
	if err != nil {
		log.Printf("Couldn't convert Java object map into a Go object map: %v", err)
		return nil
	}
	ret := make(map[security.BlessingPattern]security.Blessings)
	for jPattern, jBlessings := range bmap {
		pattern, err := GoBlessingPattern(env, C.jobject(jPattern))
		if err != nil {
			log.Printf("Couldn't convert Java pattern into Go: %v", err)
			return nil
		}
		blessings, err := GoBlessings(env, C.jobject(jBlessings))
		if err != nil {
			log.Printf("Couldn't convert Java blessings into Go: %v", err)
			return nil
		}
		ret[pattern] = blessings
	}
	return ret
}

func (r *blessingStore) DebugString() string {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jString, err := jutil.CallStringMethod(env, r.jBlessingStore, "debugString", nil)
	if err != nil {
		log.Printf("Couldn't call Java debugString: %v", err)
		return ""
	}
	return jutil.GoString(env, jString)
}
