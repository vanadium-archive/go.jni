// +build android

package security

import (
	"log"
	"runtime"
	"time"
	"unsafe"

	inaming "v.io/core/veyron/runtimes/google/naming"
	jutil "v.io/jni/util"
	jcontext "v.io/jni/v23/context"

	"v.io/v23/context"
	"v.io/v23/naming"
	"v.io/v23/security"
	"v.io/v23/vdl"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// JavaContext converts the provided Go (security) Context into a Java Context object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaContext(jEnv interface{}, ctx security.Context) (unsafe.Pointer, error) {
	jContext, err := jutil.NewObject(jEnv, jContextImplClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&ctx)))
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&ctx) // Un-refed when the Java ContextImpl object is finalized.
	return jContext, nil
}

// GoContext creates instance of security.Context that uses the provided Java
// Context as its underlying implementation.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoContext(jEnv, jContextObj interface{}) (security.Context, error) {
	env := (*C.JNIEnv)(unsafe.Pointer(jutil.PtrValue(jEnv)))
	jContext := C.jobject(unsafe.Pointer(jutil.PtrValue(jContextObj)))

	// Reference Java context; it will be de-referenced when the go context
	// created below is garbage-collected (through the finalizer callback we
	// setup just below).
	jContext = C.NewGlobalRef(env, jContext)
	ctx := &contextImpl{
		jContext: jContext,
	}
	runtime.SetFinalizer(ctx, func(c *contextImpl) {
		javaEnv, freeFunc := jutil.GetEnv()
		jenv := (*C.JNIEnv)(javaEnv)
		defer freeFunc()
		C.DeleteGlobalRef(jenv, c.jContext)
	})
	return ctx, nil
}

// contextImpl is the go interface to the java implementation of security.Context
type contextImpl struct {
	jContext C.jobject
}

func (c *contextImpl) Timestamp() time.Time {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jTime, err := jutil.CallObjectMethod(env, c.jContext, "timestamp", nil, jutil.DateTimeSign)
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

func (c *contextImpl) Method() string {
	return c.callStringMethod("method")
}

func (c *contextImpl) MethodTags() []*vdl.Value {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jTags, err := jutil.CallObjectArrayMethod(env, c.jContext, "methodTags", nil, jutil.VdlValueSign)
	if err != nil {
		log.Println("Couldn't call Java methodTags method: ", err)
		return nil
	}
	log.Println("JNI goContext: converting method tags")
	tags, err := jutil.GoVDLValueArray(env, jTags)
	if err != nil {
		log.Println("Couldn't convert Java tags to Go: ", err)
		return nil
	}
	log.Println("JNI goContext: success converting method tags")
	return tags
}

func (c *contextImpl) Suffix() string {
	return jutil.UpperCamelCase(c.callStringMethod("suffix"))
}

func (c *contextImpl) RemoteDischarges() map[string]security.Discharge {
	// TODO(spetrovic): implement this method.
	return nil
}

func (c *contextImpl) LocalEndpoint() naming.Endpoint {
	epStr := c.callStringMethod("localEndpoint")
	ep, err := inaming.NewEndpoint(epStr)
	if err != nil {
		log.Printf("Couldn't parse endpoint string %q: %v", epStr, err)
		return nil
	}
	return ep
}

func (c *contextImpl) LocalPrincipal() security.Principal {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jPrincipal, err := jutil.CallObjectMethod(env, c.jContext, "localPrincipal", nil, principalSign)
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

func (c *contextImpl) LocalBlessings() security.Blessings {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jBlessings, err := jutil.CallObjectMethod(env, c.jContext, "localBlessings", nil, blessingsSign)
	if err != nil {
		log.Printf("Couldn't call Java localBlessings method: %v", err)
		return nil
	}
	blessings, err := GoBlessings(env, jBlessings)
	if err != nil {
		log.Printf("Couldn't convert Java Blessings into Go: %v", err)
		return nil
	}
	return blessings
}

func (c *contextImpl) RemoteBlessings() security.Blessings {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jBlessings, err := jutil.CallObjectMethod(env, c.jContext, "remoteBlessings", nil, blessingsSign)
	if err != nil {
		log.Printf("Couldn't call Java remoteBlessings method: %v", err)
		return nil
	}
	blessings, err := GoBlessings(env, jBlessings)
	if err != nil {
		log.Printf("Couldn't convert Java Blessings into Go: %v", err)
		return nil
	}
	return blessings
}

func (c *contextImpl) RemoteEndpoint() naming.Endpoint {
	epStr := c.callStringMethod("remoteEndpoint")
	ep, err := inaming.NewEndpoint(epStr)
	if err != nil {
		log.Printf("Couldn't parse endpoint string %q: %v", epStr, err)
		return nil
	}
	return ep
}

func (c *contextImpl) Context() *context.T {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	contextSign := jutil.Sign("io.v.v23.context.VContext")
	jCtx, err := jutil.CallObjectMethod(env, c.jContext, "context", nil, contextSign)
	if err != nil {
		log.Printf("Couldn't get Java Vanadium context: %v", err)
	}
	ctx, err := jcontext.GoContext(env, jCtx)
	if err != nil {
		log.Printf("Couldn't convert Java Vanadium context to Go: %v", err)
	}
	return ctx
}

func (c *contextImpl) callStringMethod(methodName string) string {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	ret, err := jutil.CallStringMethod(env, c.jContext, methodName, nil)
	if err != nil {
		log.Printf("Couldn't call Java %q method: %v", methodName, err)
		return ""
	}
	return ret
}
