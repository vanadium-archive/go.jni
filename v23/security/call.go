// +build android

package security

import (
	"log"
	"runtime"
	"time"
	"unsafe"

	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"

	"v.io/v23"
	"v.io/v23/context"
	"v.io/v23/naming"
	"v.io/v23/security"
	"v.io/v23/vdl"
)

// #include "jni.h"
import "C"

// JavaCall converts the provided Go (security) Call into a Java Call object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaCall(jEnv interface{}, call security.Call) (unsafe.Pointer, error) {
	jCall, err := jutil.NewObject(jEnv, jCallImplClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&call)))
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&call) // Un-refed when the Java CallImpl object is finalized.
	return jCall, nil
}

// GoCall creates instance of security.Call that uses the provided Java
// Call as its underlying implementation.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoCall(jEnv, jCallObj interface{}) (security.Call, error) {
	// Reference Java call; it will be de-referenced when the go call
	// created below is garbage-collected (through the finalizer callback we
	// setup just below).
	jCall := C.jobject(jutil.NewGlobalRef(jEnv, jCallObj))
	call := &callImpl{
		jCall: jCall,
	}
	runtime.SetFinalizer(call, func(c *callImpl) {
		javaEnv, freeFunc := jutil.GetEnv()
		jenv := (*C.JNIEnv)(javaEnv)
		defer freeFunc()
		jutil.DeleteGlobalRef(jenv, c.jCall)
	})
	return call, nil
}

// callImpl is the go interface to the java implementation of security.Call
type callImpl struct {
	jCall C.jobject
}

func (c *callImpl) Timestamp() time.Time {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jTime, err := jutil.CallObjectMethod(env, c.jCall, "timestamp", nil, jutil.DateTimeSign)
	if err != nil {
		log.Println("Couldn't call Java timestamp method: ", err)
		return time.Time{}
	}
	t, err := jutil.GoTime(env, jTime)
	if err != nil {
		log.Println("Couldn't convert Java time to Go: ", err)
		return time.Time{}
	}
	return t
}

func (c *callImpl) Method() string {
	return jutil.UpperCamelCase(c.callStringMethod("method"))
}

func (c *callImpl) MethodTags() []*vdl.Value {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jTags, err := jutil.CallObjectArrayMethod(env, c.jCall, "methodTags", nil, jutil.VdlValueSign)
	if err != nil {
		log.Println("Couldn't call Java methodTags method: ", err)
		return nil
	}
	tags, err := jutil.GoVDLValueArray(env, jTags)
	if err != nil {
		log.Println("Couldn't convert Java tags to Go: ", err)
		return nil
	}
	return tags
}

func (c *callImpl) Suffix() string {
	return c.callStringMethod("suffix")
}

func (c *callImpl) LocalDischarges() map[string]security.Discharge {
	// TODO(spetrovic): implement this method.
	return nil
}

func (c *callImpl) RemoteDischarges() map[string]security.Discharge {
	// TODO(spetrovic): implement this method.
	return nil
}

func (c *callImpl) LocalEndpoint() naming.Endpoint {
	epStr := c.callStringMethod("localEndpoint")
	ep, err := v23.NewEndpoint(epStr)
	if err != nil {
		log.Printf("Couldn't parse endpoint string %q: %v", epStr, err)
		return nil
	}
	return ep
}

func (c *callImpl) LocalPrincipal() security.Principal {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jPrincipal, err := jutil.CallObjectMethod(env, c.jCall, "localPrincipal", nil, principalSign)
	if err != nil {
		log.Printf("Couldn't call Java localPrincipal method: %v", err)
		return nil
	}
	principal, err := GoPrincipal(env, jPrincipal)
	if err != nil {
		log.Printf("Couldn't convert Java principal to Go: %v", err)
		return nil
	}
	return principal
}

func (c *callImpl) LocalBlessings() security.Blessings {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jBlessings, err := jutil.CallObjectMethod(env, c.jCall, "localBlessings", nil, blessingsSign)
	if err != nil {
		log.Printf("Couldn't call Java localBlessings method: %v", err)
		return security.Blessings{}
	}
	blessings, err := GoBlessings(env, jBlessings)
	if err != nil {
		log.Printf("Couldn't convert Java Blessings into Go: %v", err)
		return security.Blessings{}
	}
	return blessings
}

func (c *callImpl) RemoteBlessings() security.Blessings {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jBlessings, err := jutil.CallObjectMethod(env, c.jCall, "remoteBlessings", nil, blessingsSign)
	if err != nil {
		log.Printf("Couldn't call Java remoteBlessings method: %v", err)
		return security.Blessings{}
	}
	blessings, err := GoBlessings(env, jBlessings)
	if err != nil {
		log.Printf("Couldn't convert Java Blessings into Go: %v", err)
		return security.Blessings{}
	}
	return blessings
}

func (c *callImpl) RemoteEndpoint() naming.Endpoint {
	epStr := c.callStringMethod("remoteEndpoint")
	ep, err := v23.NewEndpoint(epStr)
	if err != nil {
		log.Printf("Couldn't parse endpoint string %q: %v", epStr, err)
		return nil
	}
	return ep
}

func (c *callImpl) Context() *context.T {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	contextSign := jutil.ClassSign("io.v.v23.context.VContext")
	jCtx, err := jutil.CallObjectMethod(env, c.jCall, "context", nil, contextSign)
	if err != nil {
		log.Printf("Couldn't get Java Vanadium context: %v", err)
	}
	ctx, err := jcontext.GoContext(env, jCtx)
	if err != nil {
		log.Printf("Couldn't convert Java Vanadium context to Go: %v", err)
	}
	return ctx
}

func (c *callImpl) callStringMethod(methodName string) string {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	ret, err := jutil.CallStringMethod(env, c.jCall, methodName, nil)
	if err != nil {
		log.Printf("Couldn't call Java %q method: %v", methodName, err)
		return ""
	}
	return ret
}
