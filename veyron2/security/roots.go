// +build android

package security

import (
	"fmt"
	"log"
	"runtime"
	"unsafe"

	jutil "veyron.io/jni/util"
	"veyron.io/veyron/veyron2/security"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// JavaBlessingRoots creates an instance of Java BlessingRoots that uses the provided Go
// BlessingRoots as its underlying implementation.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaBlessingRoots(jEnv interface{}, roots security.BlessingRoots) (C.jobject, error) {
	env := (*C.JNIEnv)(unsafe.Pointer(jutil.PtrValue(jEnv)))
	jObj, err := jutil.NewObject(env, jBlessingRootsImplClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&roots)))
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&roots) // Un-refed when the Java BlessingRootsImpl is finalized.
	return C.jobject(jObj), nil
}

// GoBlessingRoots creates an instance of security.BlessingRoots that uses the
// provided Java BlessingRoots as its underlying implementation.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoBlessingRoots(jEnv, jBlessingRootsObj interface{}) (security.BlessingRoots, error) {
	env := (*C.JNIEnv)(unsafe.Pointer(jutil.PtrValue(jEnv)))
	jBlessingRoots := C.jobject(unsafe.Pointer(jutil.PtrValue(jBlessingRootsObj)))

	// We cannot cache Java environments as they are only valid in the current
	// thread.  We can, however, cache the Java VM and obtain an environment
	// from it in whatever thread happens to be running at the time.
	var jVM *C.JavaVM
	if status := C.GetJavaVM(env, &jVM); status != 0 {
		return nil, fmt.Errorf("couldn't get Java VM from the (Java) environment")
	}
	// Reference Java BlessingRoots; it will be de-referenced when the Go
	// BlessingRoots created below is garbage-collected (through the finalizer
	// callback we setup just below).
	jBlessingRoots = C.NewGlobalRef(env, jBlessingRoots)
	r := &blessingRoots{
		jVM:            jVM,
		jBlessingRoots: jBlessingRoots,
	}
	runtime.SetFinalizer(r, func(r *blessingRoots) {
		envPtr, freeFunc := jutil.GetEnv(r.jVM)
		env := (*C.JNIEnv)(envPtr)
		defer freeFunc()
		C.DeleteGlobalRef(env, r.jBlessingRoots)
	})
	return r, nil
}

type blessingRoots struct {
	jVM            *C.JavaVM
	jBlessingRoots C.jobject
}

func (r *blessingRoots) Add(root security.PublicKey, pattern security.BlessingPattern) error {
	env, freeFunc := jutil.GetEnv(r.jVM)
	defer freeFunc()
	jRoot, err := JavaPublicKey(env, root)
	if err != nil {
		return err
	}
	jPattern, err := JavaBlessingPattern(env, pattern)
	if err != nil {
		return err
	}
	return jutil.CallVoidMethod(env, r.jBlessingRoots, "add", []jutil.Sign{publicKeySign, blessingPatternSign}, jRoot, jPattern)
}

func (r *blessingRoots) Recognized(root security.PublicKey, blessing string) error {
	env, freeFunc := jutil.GetEnv(r.jVM)
	defer freeFunc()
	jRoot, err := JavaPublicKey(env, root)
	if err != nil {
		return err
	}
	return jutil.CallVoidMethod(env, r.jBlessingRoots, "recognized", []jutil.Sign{publicKeySign, jutil.StringSign}, jRoot, blessing)
}

func (r *blessingRoots) DebugString() string {
	env, freeFunc := jutil.GetEnv(r.jVM)
	defer freeFunc()
	jString, err := jutil.CallStringMethod(env, r.jBlessingRoots, "debugString", nil)
	if err != nil {
		log.Printf("Coudln't get Java DebugString: %v", err)
		return ""
	}
	return jutil.GoString(env, jString)
}
