// +build android

package security

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"veyron.io/jni/util"
	isecurity "veyron.io/veyron/veyron/runtimes/google/security"
	"veyron.io/veyron/veyron2/security"
	"veyron.io/veyron/veyron2/security/wire"
	"veyron.io/veyron/veyron2/vom"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// JavaContext converts the provided Go (security) Context into a Java Context object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaContext(jEnv interface{}, context security.Context) (C.jobject, error) {
	jContext, err := util.NewObject(jEnv, jContextImplClass, []util.Sign{util.LongSign}, &context)
	if err != nil {
		return nil, err
	}
	util.GoRef(&context) // Un-refed when the Java Context object is finalized.
	return C.jobject(jContext), nil
}

// EncodeChains JSON-encodes the chains stored in the provided PublicID.
func EncodeChains(id security.PublicID) ([]string, error) {
	chains, err := extractChains(id)
	if err != nil {
		return nil, err
	}
	ret := make([]string, len(chains))
	for i, chain := range chains {
		enc, err := json.Marshal(chain)
		if err != nil {
			return nil, err
		}
		ret[i] = string(enc)
	}
	return ret, nil
}

func extractChains(id security.PublicID) ([]wire.ChainPublicID, error) {
	m := reflect.ValueOf(id).MethodByName("VomEncode")
	if !m.IsValid() {
		return nil, fmt.Errorf("type %T doesn't implement VomEncode()", id)
	}
	results := m.Call(nil)
	if len(results) != 2 {
		return nil, fmt.Errorf("type %T has wrong number of return arguments for VomEncode()", id)
	}
	if !results[1].IsNil() {
		err, ok := results[1].Interface().(error)
		if !ok {
			return nil, fmt.Errorf("second return argument must be an error, got %T", results[1].Interface())
		}
		return nil, fmt.Errorf("error invoking VomEncode(): %v", err)
	}
	if results[0].IsNil() {
		return nil, fmt.Errorf("VomEncode() returned nil encoding value")
	}
	switch result := results[0].Interface().(type) {
	case *wire.ChainPublicID:
		return []wire.ChainPublicID{*result}, nil
	case []security.PublicID:
		var ret []wire.ChainPublicID
		for _, childID := range result {
			chains, err := extractChains(childID)
			if err != nil {
				return nil, err
			}
			ret = append(ret, chains...)
		}
		return ret, nil
	default:
		return nil, fmt.Errorf("unexpected return value of type %T for VomEncode", result)
	}
}

// DecodeChains JSON-decodes the provided chains and creates a new PublicID
// from them.
func DecodeChains(encoded []string) (security.PublicID, error) {
	// JSON-decode chains.
	chains := make([]wire.ChainPublicID, len(encoded))
	for i, str := range encoded {
		if err := json.Unmarshal([]byte(str), &chains[i]); err != nil {
			return nil, fmt.Errorf("couldn't JSON-decode chain %q: %v", str, err)
		}
	}
	// Create PublicIDs.
	ids := make([]security.PublicID, len(chains))
	for i, chain := range chains {
		// Total hack to obtain a PublicID from wire.ChainPublicID.
		// TODO(spetrovic): make sure this goes away when we switch to Principal/Blessing API.
		var buf bytes.Buffer
		if err := vom.NewEncoder(&buf).Encode(&chain); err != nil {
			return nil, fmt.Errorf("couldn't VOM-encode chain: %v", err)
		}
		privID, err := isecurity.NewPrivateID("dummy", nil)
		if err != nil {
			return nil, fmt.Errorf("couldn't mint new private id")
		}
		ids[i] = privID.PublicID()
		if err := vom.NewDecoder(&buf).Decode(&ids[i]); err != nil {
			return nil, fmt.Errorf("couldn't VOM-decode chain: %v", err)
		}
	}
	return isecurity.NewSetPublicID(ids...)
}
