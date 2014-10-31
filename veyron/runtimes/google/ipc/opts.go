// +build android

package ipc

import (
	"veyron.io/jni/util"
	jsecurity "veyron.io/jni/veyron2/security"
	"veyron.io/veyron/veyron2"
	"veyron.io/veyron/veyron2/options"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// getRuntimeOpts converts Java runtime options into Go runtime options.
func getRuntimeOpts(env *C.JNIEnv, jOptions C.jobject) (ret []veyron2.ROpt, err error) {
	if jOptions == nil {
		return
	}
	runtimePrincipalKey := util.JStaticStringField(env, jOptionDefsClass, "RUNTIME_PRINCIPAL")
	if has, err := util.CallBooleanMethod(env, jOptions, "has", []util.Sign{util.StringSign}, runtimePrincipalKey); err != nil {
		return nil, err
	} else if has {
		jPrincipal, err := util.CallObjectMethod(env, jOptions, "get", []util.Sign{util.StringSign}, util.ObjectSign, runtimePrincipalKey)
		if err != nil {
			return nil, err
		}
		principal, err := jsecurity.GoPrincipal(env, jPrincipal)
		if err != nil {
			return nil, err
		}
		ret = append(ret, options.RuntimePrincipal{principal})
	}
	return
}

// getVDLPathOpt retrieves the Java VDL_PATH option.
func getVDLPathOpt(env *C.JNIEnv, jOptions C.jobject) (*string, error) {
	if jOptions == nil {
		return nil, nil
	}
	vdlPathKey := util.JStaticStringField(env, jOptionDefsClass, "VDL_INTERFACE_PATH")
	has, err := util.CallBooleanMethod(env, jOptions, "has", []util.Sign{util.StringSign}, vdlPathKey)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	jPath, err := util.CallObjectMethod(env, jOptions, "get", []util.Sign{util.StringSign}, util.ObjectSign, vdlPathKey)
	if err != nil {
		return nil, err
	}
	if jPath == nil {
		return nil, nil
	}
	path := util.GoString(env, jPath)
	return &path, nil
}
