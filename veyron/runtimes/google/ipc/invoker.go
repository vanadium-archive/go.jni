// +build android

package ipc

import (
	"encoding/json"
	"fmt"
	"runtime"

	jutil "veyron.io/jni/util"
	jsecurity "veyron.io/jni/veyron2/security"
	"veyron.io/veyron/veyron2/ipc"
	"veyron.io/veyron/veyron2/verror"
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

	// Fetch the argGetter for the object.
	jPathArray, err := jutil.CallObjectMethod(env, jInvoker, "getImplementedServers", nil, jutil.ArraySign(jutil.StringSign))
	if err != nil {
		return nil, err
	}
	paths := jutil.GoStringArray(env, jPathArray)
	getter, err := newArgGetter(paths)
	if err != nil {
		return nil, err
	}

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
		jVM:       jVM,
		jInvoker:  jInvoker,
		argGetter: getter,
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
	jVM       *C.JavaVM
	jInvoker  C.jobject
	argGetter *argGetter
}

func (i *invoker) Prepare(method string, numArgs int) (argptrs, tags []interface{}, err error) {
	// NOTE(spetrovic): In the long-term, this method will return an array of
	// []vom.Value.  This will in turn result in VOM decoding all input
	// arguments into vom.Value objects, which we shall then de-serialize into
	// Java objects (see Invoke comments below).  This approach is blocked on
	// pending VOM encoder/decoder changes as well as Java (de)serializer.
	jEnv, freeFunc := jutil.GetEnv(i.jVM)
	env := (*C.JNIEnv)(jEnv)
	defer freeFunc()

	mArgs := i.argGetter.FindMethod(method, numArgs)
	if mArgs == nil {
		err = fmt.Errorf("couldn't find VDL method %q with %d args", method, numArgs)
		return
	}
	argptrs = mArgs.InPtrs()

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
	// NOTE(spetrovic): In the long-term, all input arguments will be of
	// vom.Value type (see comments for Prepare() method above).  Through JNI,
	// we will call Java functions that transform a serialized vom.Value into
	// Java objects. We will then pass those Java objects to Java's Invoke
	// method.  The returned Java objects will be converted into serialized
	// vom.Values, which will then be returned.  This approach is blocked on VOM
	// encoder/decoder changes as well as Java's (de)serializer.
	jEnv, freeFunc := jutil.GetEnv(i.jVM)
	env := (*C.JNIEnv)(jEnv)
	defer freeFunc()

	// Create a new Java server call instance.
	mArgs := i.argGetter.FindMethod(method, len(argptrs))
	if mArgs == nil {
		err = fmt.Errorf("couldn't find VDL method %q with %d args", method, len(argptrs))
	}
	sCall := newServerCall(call, mArgs)
	jServerCall, err := javaServerCall(env, sCall)
	if err != nil {
		return nil, err
	}

	// Translate input args to JSON.
	jArgs, err := i.encodeArgs(env, argptrs)
	if err != nil {
		return
	}
	// Invoke the method.
	callSign := jutil.ClassSign("io.veyron.veyron.veyron2.ipc.ServerCall")
	replySign := jutil.ClassSign("io.veyron.veyron.veyron.runtimes.google.ipc.VDLInvoker$InvokeReply")
	jReply, err := jutil.CallObjectMethod(env, i.jInvoker, "invoke", []jutil.Sign{jutil.StringSign, callSign, jutil.ArraySign(jutil.StringSign)}, replySign, jutil.CamelCase(method), jServerCall, jArgs)
	if err != nil {
		return nil, fmt.Errorf("error invoking Java method %q: %v", method, err)
	}
	// Decode and return results.
	return i.decodeResults(env, method, len(argptrs), C.jobject(jReply))
}

func (i *invoker) VGlob() *ipc.GlobState {
	// TODO(spetrovic): implement this method.
	return &ipc.GlobState{}
}

func (i *invoker) Signature(ctx ipc.ServerContext) ([]ipc.InterfaceSig, error) {
	// TODO(spetrovic): implement this method.
	return nil, fmt.Errorf("Java runtime doesn't yet support signatures.")
}

func (i *invoker) MethodSignature(ctx ipc.ServerContext, method string) (ipc.MethodSig, error) {
	// TODO(spetrovic): implement this method.
	return ipc.MethodSig{}, fmt.Errorf("Java runtime doesn't yet support signatures.")
}

// encodeArgs JSON-encodes the provided argument pointers, converts them into
// Java strings, and returns a Java string array response.
func (*invoker) encodeArgs(env *C.JNIEnv, argptrs []interface{}) (C.jobjectArray, error) {
	// JSON encode.
	jsonArgs := make([][]byte, len(argptrs))
	for i, argptr := range argptrs {
		// Remove the pointer from the argument.  Simply *argptr doesn't work
		// as argptr is of type interface{}.
		arg := jutil.DerefOrDie(argptr)
		var err error
		jsonArgs[i], err = json.Marshal(arg)
		if err != nil {
			return nil, fmt.Errorf("error marshalling %q into JSON", arg)
		}
	}

	// Convert to Java array of C.jstring.
	ret := C.NewObjectArray(env, C.jsize(len(argptrs)), jStringClass, nil)
	for i, arg := range jsonArgs {
		C.SetObjectArrayElement(env, ret, C.jsize(i), C.jobject(jutil.JString(env, string(arg))))
	}
	return ret, nil
}

// decodeResults JSON-decodes replies stored in the Java reply object and
// returns an array of Go reply objects.
func (i *invoker) decodeResults(env *C.JNIEnv, method string, numArgs int, jReply C.jobject) ([]interface{}, error) {
	// Unpack the replies.
	results, err := jutil.JStringArrayField(env, jReply, "results")
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

	// Get result instances.
	mArgs := i.argGetter.FindMethod(method, numArgs)
	if mArgs == nil {
		return nil, fmt.Errorf("couldn't find method %q with %d input args: %v", method, numArgs)
	}
	argptrs := mArgs.OutPtrs()

	// Check for app error.
	if hasAppErr {
		return resultsWithError(argptrs, verror.Make(verror.ID(errorID), errorMsg)), nil
	}
	// JSON-decode.
	if len(results) != len(argptrs) {
		return nil, fmt.Errorf("mismatch in number of output arguments, have: %d want: %d", len(results), len(argptrs))
	}
	for i, result := range results {
		if err := json.Unmarshal([]byte(result), argptrs[i]); err != nil {
			return nil, err
		}
	}
	return resultsWithError(argptrs, nil), nil
}

// resultsWithError dereferences the provided result pointers and appends the
// given error to the returned array.
func resultsWithError(resultptrs []interface{}, err error) []interface{} {
	ret := make([]interface{}, len(resultptrs)+1)
	for i, resultptr := range resultptrs {
		ret[i] = jutil.DerefOrDie(resultptr)
	}
	ret[len(resultptrs)] = err
	return ret
}
