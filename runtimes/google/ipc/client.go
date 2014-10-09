// +build android

package ipc

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"veyron.io/jni/runtimes/google/util"
	"veyron.io/veyron/veyron2/ipc"
	"veyron.io/veyron/veyron2/rt"
)

// #cgo LDFLAGS: -ljniwrapper
// #include <stdlib.h>
// #include "jni_wrapper.h"
import "C"

type client struct {
	client ipc.Client
}

func newClient(c ipc.Client) *client {
	return &client{
		client: c,
	}
}

func (c *client) StartCall(env *C.JNIEnv, jContext C.jobject, name, method string, jArgs C.jobjectArray, jOptions C.jobject) (*clientCall, error) {
	// NOTE(spetrovic): In the long-term, we will decode JSON arguments into an
	// array of vom.Value instances and send this array across the wire.

	// Convert Java argument array into []string.
	argStrs := make([]string, int(C.GetArrayLength(env, C.jarray(jArgs))))
	for i := 0; i < len(argStrs); i++ {
		argStrs[i] = util.GoString(env, C.GetObjectArrayElement(env, jArgs, C.jsize(i)))
	}
	// Get argument instances that correspond to the provided method.
	vdlPackagePathPtr, err := getVDLPathOpt(env, jOptions)
	if err != nil {
		return nil, err
	}
	if vdlPackagePathPtr == nil {
		return nil, fmt.Errorf("couldn't find VDL_INTERFACE_PATH option")
	}
	getter, err := newArgGetter([]string{*vdlPackagePathPtr})
	if err != nil {
		return nil, err
	}
	mArgs := getter.FindMethod(method, len(argStrs))
	if mArgs == nil {
		return nil, fmt.Errorf("couldn't find method %s with %d args in VDL interface at path %q", method, len(argStrs), util.GoString(env, *vdlPackagePathPtr))
	}
	argptrs := mArgs.InPtrs()
	if len(argptrs) != len(argStrs) {
		return nil, fmt.Errorf("invalid number of arguments for method %s, want %d, have %d", method, len(argStrs), len(argptrs))
	}
	// JSON decode.
	args := make([]interface{}, len(argptrs))
	for i, argStr := range argStrs {
		if err := json.Unmarshal([]byte(argStr), argptrs[i]); err != nil {
			return nil, err
		}
		// Remove the pointer from the argument.  Simply *argptr[i] doesn't work
		// as argptr[i] is of type interface{}.
		args[i] = util.DerefOrDie(argptrs[i])
	}

	context, _ := rt.R().NewContext().WithTimeout(10 * time.Second)

	// Invoke StartCall
	call, err := c.client.StartCall(context, name, method, args)
	if err != nil {
		return nil, err
	}
	return &clientCall{
		stream: newStream(call, mArgs),
		call:   call,
	}, nil
}

func (c *client) Close() {
	c.client.Close()
}

type clientCall struct {
	stream
	call ipc.Call
}

func (c *clientCall) Finish(env *C.JNIEnv) (C.jobjectArray, error) {
	var resultptrs []interface{}
	if c.mArgs.IsStreaming() {
		resultptrs = c.mArgs.StreamFinishPtrs()
	} else {
		resultptrs = c.mArgs.OutPtrs()
	}
	// argGetter doesn't store the (mandatory) error result, so we add it here.
	var appErr error
	if err := c.call.Finish(append(resultptrs, &appErr)...); err != nil {
		// invocation error
		return nil, fmt.Errorf("Invocation error: %v", err)
	}
	if appErr != nil { // application error
		return nil, appErr
	}
	// JSON encode the results.
	jsonResults := make([][]byte, len(resultptrs))
	for i, resultptr := range resultptrs {
		// Remove the pointer from the result.  Simply *resultptr doesn't work
		// as resultptr is of type interface{}.
		result := util.DerefOrDie(resultptr)

		// See if the result has a VomEncode method, which converts the result into a VDL type.
		// If it does, invoke the method as the resulting VDL type is guaranteed to be
		// JSON-encodeable (while the original result will likely not).
		value := interface{}(result)
		if v := reflect.ValueOf(result); v.IsValid() {
			if m := v.MethodByName("VomEncode"); m.IsValid() && m.Kind() == reflect.Func {
				if data := m.Call(nil); len(data) == 2 && data[0].CanInterface() {
					value = data[0].Interface()
				}
			}
		}
		var err error
		jsonResults[i], err = json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("error marshalling %q into JSON", resultptr)
		}
	}

	// Convert to Java array of C.jstring.
	ret := C.NewObjectArray(env, C.jsize(len(jsonResults)), jStringClass, nil)
	for i, result := range jsonResults {
		C.SetObjectArrayElement(env, ret, C.jsize(i), C.jobject(util.JStringPtr(env, string(result))))
	}
	return ret, nil
}

func (c *clientCall) Cancel() {
	c.call.Cancel()
}
