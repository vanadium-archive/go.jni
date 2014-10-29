// +build android

package security

import (
	"fmt"
	"log"
	"runtime"
	"unsafe"

	"veyron.io/jni/util"
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
	env := (*C.JNIEnv)(unsafe.Pointer(util.PtrValue(jEnv)))
	util.GoRef(&roots) // Un-refed when the Java BlessingRootsImpl is finalized.
	jObj, err := util.NewObject(env, jBlessingRootsImplClass, []util.Sign{util.LongSign}, &roots)
	return C.jobject(jObj), err
}

// GoBlessingRoots creates an instance of security.BlessingRoots that uses the
// provided Java BlessingRoots as its underlying implementation.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoBlessingRoots(jEnv, jBlessingRootsObj interface{}) (security.BlessingRoots, error) {
	env := (*C.JNIEnv)(unsafe.Pointer(util.PtrValue(jEnv)))
	jBlessingRoots := C.jobject(unsafe.Pointer(util.PtrValue(jBlessingRootsObj)))

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
		envPtr, freeFunc := util.GetEnv(r.jVM)
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
	envPtr, freeFunc := util.GetEnv(r.jVM)
	env := (*C.JNIEnv)(envPtr)
	defer freeFunc()
	jRoot, err := JavaPublicKey(env, root)
	if err != nil {
		return err
	}
	jPattern, err := JavaBlessingPattern(env, pattern)
	if err != nil {
		return err
	}
	return util.CallVoidMethod(env, r.jBlessingRoots, "add", []util.Sign{publicKeySign, blessingPatternSign}, jRoot, jPattern)
}

func (r *blessingRoots) Recognized(root security.PublicKey, blessing string) error {
	envPtr, freeFunc := util.GetEnv(r.jVM)
	env := (*C.JNIEnv)(envPtr)
	defer freeFunc()
	jRoot, err := JavaPublicKey(env, root)
	if err != nil {
		return err
	}
	return util.CallVoidMethod(env, r.jBlessingRoots, "recognized", []util.Sign{publicKeySign, util.StringSign}, jRoot, blessing)
}

func (r *blessingRoots) DebugString() string {
	envPtr, freeFunc := util.GetEnv(r.jVM)
	env := (*C.JNIEnv)(envPtr)
	defer freeFunc()
	jString, err := util.CallStringMethod(env, r.jBlessingRoots, "debugString", nil)
	if err != nil {
		log.Printf("Coudln't get Java DebugString: %v", err)
		return ""
	}
	return util.GoString(env, jString)
}
