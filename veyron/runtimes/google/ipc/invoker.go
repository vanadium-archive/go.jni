// +build android

package ipc

import (
	"fmt"
	"runtime"
	"v.io/core/veyron2/ipc"
	"v.io/core/veyron2/vdl"
	"v.io/core/veyron2/vdl/vdlroot/src/signature"
	"v.io/core/veyron2/verror"
	"v.io/core/veyron2/vom2"
	jutil "v.io/jni/util"
	jsecurity "v.io/jni/veyron2/security"
)

// #cgo LDFLAGS: -llog -ljniwrapper
// #include "jni_wrapper.h"
import "C"

func goInvoker(env *C.JNIEnv, jObj C.jobject) (ipc.Invoker, error) {
	// Create a new Java VDLInvoker object.
	jInvokerObj, err := jutil.NewObject(env, jVDLInvokerClass, []jutil.Sign{jutil.ObjectSign}, jObj)
	if err != nil {
		return nil, fmt.Errorf("error creating Java VDLInvoker object: %v", err)
	}
	jInvoker := C.jobject(jInvokerObj)

	// We cannot cache Java environments as they are only valid in the current
	// thread.  We can, however, cache the Java VM and obtain an environment
	// from it in whatever thread happens to be running at the time.
	var jVM *C.JavaVM
	if status := C.GetJavaVM(env, &jVM); status != 0 {
		return nil, fmt.Errorf("couldn't get Java VM from the (Java) environment")
	}

	// Reference Java invoker; it will be de-referenced when the go invoker
	// created below is garbage-collected (through the finalizer callback we
	// setup just below).
	jInvoker = C.NewGlobalRef(env, jInvoker)
	i := &invoker{
		jVM:      jVM,
		jInvoker: jInvoker,
	}
	runtime.SetFinalizer(i, func(i *invoker) {
		jEnv, freeFunc := jutil.GetEnv(i.jVM)
		env := (*C.JNIEnv)(jEnv)
		defer freeFunc()
		C.DeleteGlobalRef(env, i.jInvoker)
	})
	return i, nil
}

type invoker struct {
	jVM      *C.JavaVM
	jInvoker C.jobject
}

func (i *invoker) Prepare(method string, numArgs int) (argptrs, tags []interface{}, err error) {
	env, freeFunc := jutil.GetEnv(i.jVM)
	defer freeFunc()

	// Have all input arguments be decoded into *vdl.Value.
	argptrs = make([]interface{}, numArgs)
	for i := 0; i < numArgs; i++ {
		value := new(vdl.Value)
		argptrs[i] = &value
	}
	// Get the method tags.
	jTags, err := jutil.CallObjectMethod(env, i.jInvoker, "getMethodTags", []jutil.Sign{jutil.StringSign}, jutil.ArraySign(jutil.ObjectSign), jutil.CamelCase(method))
	if err != nil {
		return nil, nil, err
	}
	tags, err = jsecurity.GoTags(env, jTags)
	if err != nil {
		return nil, nil, err
	}
	return
}

func (i *invoker) Invoke(method string, call ipc.ServerCall, argptrs []interface{}) (results []interface{}, err error) {
	jEnv, freeFunc := jutil.GetEnv(i.jVM)
	env := (*C.JNIEnv)(jEnv)
	defer freeFunc()

	jServerCall, err := javaServerCall(env, call)
	if err != nil {
		return nil, err
	}

	// VOM-encode the input arguments.
	jVomArgs, err := encodeArgs(env, argptrs)
	if err != nil {
		return nil, err
	}
	// Invoke the method.
	callSign := jutil.ClassSign("io.v.core.veyron2.ipc.ServerCall")
	replySign := jutil.ClassSign("io.v.core.veyron.runtimes.google.ipc.VDLInvoker$InvokeReply")
	jReply, err := jutil.CallObjectMethod(env, i.jInvoker, "invoke", []jutil.Sign{jutil.StringSign, callSign, jutil.ArraySign(jutil.ArraySign(jutil.ByteSign))}, replySign, jutil.CamelCase(method), jServerCall, jVomArgs)
	if err != nil {
		return nil, fmt.Errorf("error invoking Java method %q: %v", method, err)
	}
	// Decode and return results.
	return decodeResults(env, C.jobject(jReply))
}

func (i *invoker) Globber() *ipc.GlobState {
	// TODO(spetrovic): implement this method.
	return &ipc.GlobState{}
}

func (i *invoker) Signature(ctx ipc.ServerContext) ([]signature.Interface, error) {
	// TODO(spetrovic): implement this method.
	return nil, fmt.Errorf("Java runtime doesn't yet support signatures.")
}

func (i *invoker) MethodSignature(ctx ipc.ServerContext, method string) (signature.Method, error) {
	// TODO(spetrovic): implement this method.
	return signature.Method{}, fmt.Errorf("Java runtime doesn't yet support signatures.")
}

// encodeArgs VOM-encodes the provided arguments pointers and returns them as a
// Java array of byte arrays.
func encodeArgs(env *C.JNIEnv, argptrs []interface{}) (C.jobjectArray, error) {
	vomArgs := make([][]byte, len(argptrs))
	for i, argptr := range argptrs {
		arg := interface{}(jutil.DerefOrDie(argptr))
		var err error
		if vomArgs[i], err = vom2.Encode(arg); err != nil {
			return nil, err
		}
	}
	return C.jobjectArray(jutil.JByteArrayArray(env, vomArgs)), nil
}

// decodeResults VOM-decodes replies stored in the Java reply object into
// an array *vdl.Value.
func decodeResults(env *C.JNIEnv, jReply C.jobject) ([]interface{}, error) {
	// Unpack the replies.
	results, err := jutil.JByteArrayArrayField(env, jReply, "results")
	if err != nil {
		return nil, err
	}
	hasAppErr, err := jutil.JBoolField(env, jReply, "hasApplicationError")
	if err != nil {
		return nil, err
	}
	errorID, err := jutil.JStringField(env, jReply, "errorID")
	if err != nil {
		return nil, err
	}
	errorMsg, err := jutil.JStringField(env, jReply, "errorMsg")
	if err != nil {
		return nil, err
	}
	// Check for app error.
	var appErr error
	if hasAppErr {
		appErr = verror.Make(verror.ID(errorID), errorMsg)
	}
	// VOM-decode results into *vdl.Value instances and append the error (if any).
	ret := make([]interface{}, len(results)+1)
	for i, result := range results {
		var err error
		if ret[i], err = jutil.VomDecodeToValue(result); err != nil {
			return nil, err
		}
	}
	ret[len(results)] = appErr
	return ret, nil
}
