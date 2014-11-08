// +build android

package security

import (
	"fmt"
	"log"
	"runtime"
	"time"
	"unsafe"

	jutil "veyron.io/jni/util"
	inaming "veyron.io/veyron/veyron/runtimes/google/naming"
	"veyron.io/veyron/veyron2/naming"
	"veyron.io/veyron/veyron2/security"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// JavaContext converts the provided Go (security) Context into a Java Context object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaContext(jEnv interface{}, context security.Context) (C.jobject, error) {
	jContext, err := jutil.NewObject(jEnv, jContextImplClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&context)))
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&context) // Un-refed when the Java ContextImpl object is finalized.
	return C.jobject(jContext), nil
}

// GoContext creates instance of security.Context that uses the provided Java
// Context as its underlying implementation.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoContext(jEnv, jContextObj interface{}) (security.Context, error) {
	env := (*C.JNIEnv)(unsafe.Pointer(jutil.PtrValue(jEnv)))
	jContext := C.jobject(unsafe.Pointer(jutil.PtrValue(jContextObj)))

	// We cannot cache Java environments as they are only valid in the current
	// thread.  We can, however, cache the Java VM and obtain an environment
	// from it in whatever thread happens to be running at the time.
	var jVM *C.JavaVM
	if status := C.GetJavaVM(env, &jVM); status != 0 {
		return nil, fmt.Errorf("couldn't get Java VM from the (Java) environment")
	}
	// Reference Java context; it will be de-referenced when the go context
	// created below is garbage-collected (through the finalizer callback we
	// setup just below).
	jContext = C.NewGlobalRef(env, jContext)
	c := &context{
		jVM:      jVM,
		jContext: jContext,
	}
	runtime.SetFinalizer(c, func(c *context) {
		javaEnv, freeFunc := jutil.GetEnv(c.jVM)
		jenv := (*C.JNIEnv)(javaEnv)
		defer freeFunc()
		C.DeleteGlobalRef(jenv, c.jContext)
	})
	return c, nil
}

// context is the go interface to the java implementation of security.Context
type context struct {
	jVM      *C.JavaVM
	jContext C.jobject
}

func (c *context) Timestamp() time.Time {
	env, freeFunc := jutil.GetEnv(c.jVM)
	defer freeFunc()
	dateTimeSign := jutil.ClassSign("org.joda.time.DateTime")
	jTime, err := jutil.CallObjectMethod(env, c.jContext, "timestamp", nil, dateTimeSign)
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

func (c *context) Method() string {
	return c.callStringMethod("method")
}

func (c *context) MethodTags() []interface{} {
	env, freeFunc := jutil.GetEnv(c.jVM)
	defer freeFunc()
	jTags, err := jutil.CallObjectArrayMethod(env, c.jContext, "methodTags", nil, jutil.ObjectSign)
	if err != nil {
		log.Println("Couldn't call Java methodTags method: ", err)
		return nil
	}
	tags, err := GoTags(env, jTags)
	if err != nil {
		log.Println("Couldn't convert Java tags to Go: ", err)
		return nil
	}
	return tags
}

func (c *context) Name() string {
	return c.callStringMethod("name")
}

func (c *context) Suffix() string {
	return c.callStringMethod("suffix")
}

// TODO(spetrovic): remove when the method is removed from the Context interface.
func (c *context) Label() security.Label {
	return security.Label(0)
}

func (c *context) Discharges() map[string]security.Discharge {
	// TODO(spetrovic): implement this method.
	return nil
}

func (c *context) LocalEndpoint() naming.Endpoint {
	epStr := c.callStringMethod("localEndpoint")
	ep, err := inaming.NewEndpoint(epStr)
	if err != nil {
		log.Printf("Couldn't parse endpoint string %q: %v", epStr, err)
		return nil
	}
	return ep
}

func (c *context) LocalPrincipal() security.Principal {
	env, freeFunc := jutil.GetEnv(c.jVM)
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

func (c *context) LocalBlessings() security.Blessings {
	env, freeFunc := jutil.GetEnv(c.jVM)
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

func (c *context) RemoteBlessings() security.Blessings {
	env, freeFunc := jutil.GetEnv(c.jVM)
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

func (c *context) RemoteEndpoint() naming.Endpoint {
	epStr := c.callStringMethod("remoteEndpoint")
	ep, err := inaming.NewEndpoint(epStr)
	if err != nil {
		log.Printf("Couldn't parse endpoint string %q: %v", epStr, err)
		return nil
	}
	return ep
}

func (c *context) callStringMethod(methodName string) string {
	env, freeFunc := jutil.GetEnv(c.jVM)
	defer freeFunc()
	ret, err := jutil.CallStringMethod(env, c.jContext, methodName, nil)
	if err != nil {
		log.Printf("Couldn't call Java %q method: %v", methodName, err)
		return ""
	}
	return ret
}
