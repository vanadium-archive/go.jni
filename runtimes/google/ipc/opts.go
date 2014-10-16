// +build android

package ipc

import (
	jnisecurity "veyron.io/jni/runtimes/google/security"
	"veyron.io/jni/runtimes/google/util"
	"veyron.io/veyron/veyron2"
	"veyron.io/veyron/veyron2/options"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// getRuntimeOpts converts Java runtime options into Go runtime options.
func getRuntimeOpts(env *C.JNIEnv, jOptions C.jobject) (ret []veyron2.ROpt) {
	if jOptions == nil {
		return
	}
	// Process RuntimeIDOpt.
	runtimeIDKey := util.JStaticStringField(env, jOptionDefsClass, "RUNTIME_ID")
	if util.CallBooleanMethodOrCatch(env, jOptions, "has", []util.Sign{util.StringSign}, runtimeIDKey) {
		jPrivateID := C.jobject(util.CallObjectMethodOrCatch(env, jOptions, "get", []util.Sign{util.StringSign}, util.ObjectSign, runtimeIDKey))
		id := jnisecurity.NewPrivateID(env, jPrivateID)
		ret = append(ret, options.RuntimeID(id))
	}
	// TODO(ashankar,ataly,spetrovic): Replace witht he new security API
	/*
		// Process RuntimePublicIDStoreOpt
		runtimePublicIDStoreKey := util.JStaticStringField(env, jOptionDefsClass, "RUNTIME_PUBLIC_ID_STORE")
		if util.CallBooleanMethodOrCatch(env, jOptions, "has", []util.Sign{util.StringSign}, runtimePublicIDStoreKey) {
			jStore := C.jobject(util.CallObjectMethodOrCatch(env, jOptions, "get", []util.Sign{util.StringSign}, util.ObjectSign, runtimePublicIDStoreKey))
			store := jnisecurity.NewPublicIDStore(env, jStore)
			ret = append(ret, options.RuntimePublicIDStore(store))
		}
	*/
	return
}

// getLocalIDOpt converts the Java LocalID option (encoded) into Go LocalId option.
func getLocalIDOpt(env *C.JNIEnv, jOptions C.jobject) (*veyron2.LocalIDOpt, error) {
	if jOptions == nil {
		return nil, nil
	}
	localIDKey := util.JStaticStringField(env, jOptionDefsClass, "LOCAL_ID")
	if !util.CallBooleanMethodOrCatch(env, jOptions, "has", []util.Sign{util.StringSign}, localIDKey) {
		return nil, nil
	}
	jEncodedChains := C.jobject(util.CallObjectMethodOrCatch(env, jOptions, "get", []util.Sign{util.StringSign}, util.ObjectSign, localIDKey))
	if jEncodedChains == nil {
		return nil, nil
	}
	encodedChains := util.GoStringArray(env, jEncodedChains)
	id, err := jnisecurity.DecodeChains(encodedChains)
	if err != nil {
		return nil, err
	}
	opt := options.LocalID(id)
	return &opt, nil
}

// getVDLPathOpt retrieves the Java VDL_PATH option.
func getVDLPathOpt(env *C.JNIEnv, jOptions C.jobject) (*string, error) {
	if jOptions == nil {
		return nil, nil
	}
	vdlPathKey := util.JStaticStringField(env, jOptionDefsClass, "VDL_INTERFACE_PATH")
	if !util.CallBooleanMethodOrCatch(env, jOptions, "has", []util.Sign{util.StringSign}, vdlPathKey) {
		return nil, nil
	}
	jPath := C.jobject(util.CallObjectMethodOrCatch(env, jOptions, "get", []util.Sign{util.StringSign}, util.ObjectSign, vdlPathKey))
	if jPath == nil {
		return nil, nil
	}
	path := util.GoString(env, jPath)
	return &path, nil
}
