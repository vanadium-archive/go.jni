// +build android

package util

import (
	"fmt"
	"log"
	"reflect"
	"runtime"
	"sync"
	"unsafe"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// GoRef creates a new reference to the value addressed by the provided pointer.
// The value will remain referenced until it is explicitly unreferenced using
// goUnref().
func GoRef(valptr interface{}) {
	if !IsPointer(valptr) {
		panic("must pass pointer value to goRef")
	}
	goRefs.ref(valptr)
}

// GoUnref removes a previously added reference to the value addressed by the
// provided pointer.  If the value hasn't been ref-ed (a bug?), this unref will
// be a no-op.
func GoUnref(valptr interface{}) {
	if !IsPointer(valptr) {
		panic("must pass pointer value to goUnref")
	}
	goRefs.unref(valptr)
}

// IsPointer returns true iff the provided value is a pointer.
func IsPointer(val interface{}) bool {
	return reflect.ValueOf(val).Kind() == reflect.Ptr
}

// DerefOrDie dereferences the provided (pointer) value, or panic-s if the value
// isn't of pointer type.
func DerefOrDie(i interface{}) interface{} {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("want reflect.Ptr value for %v, have %v", i, v.Kind()))
	}
	return v.Elem().Interface()
}

// Ptr returns the value of the provided Java pointer (of type C.jlong) as an
// unsafe.Pointer.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func Ptr(jPtr interface{}) unsafe.Pointer {
	v := reflect.ValueOf(jPtr)
	return unsafe.Pointer(uintptr(v.Int()))
}

// PtrValue returns the value of the pointer as a uintptr.
func PtrValue(ptr interface{}) uintptr {
	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr && v.Kind() != reflect.UnsafePointer {
		panic("must pass pointer value to PtrValue")
	}
	return v.Pointer()
}

// JConditionalRef creates a reference on the provided Java object that is valid
// during the lifetime of the provided Go pointer.  In other words, the Java
// object remains referenced from the moment this method returns until the Go
// pointer is garbage-collected by the Go runtime.
// In case the provided object is already being conditionally referenced (by
// some other Go pointer), this method returns the existing Go pointer and
// doesn't modify any state.  (If the provided object isn't being conditionally
// referenced, the returned Go pointer is the same as the provided Go pointer.
// NOTE: this method is heavyweight so use it only if you can't update
// references to Java objects yourselves.
func JConditionalRef(jEnv, jObj interface{}, valptr interface{}) (unsafe.Pointer, error) {
	env := getEnv(jEnv)
	obj := getObject(jObj)

	if !IsPointer(valptr) {
		return nil, fmt.Errorf("Provided Go value %v must be a pointer: ", valptr)
	}

	// Check if this Java object was seen before.
	identityHash, err := getIdentityHash(env, obj)
	if err != nil {
		return nil, err
	}
	goJavaRefsMutex.Lock()
	defer goJavaRefsMutex.Unlock()
	if goPtr, ok := javaToGoRefs[identityHash]; ok { // Go pointer already exists in the map - use it.
		return unsafe.Pointer(uintptr(goPtr)), nil
	}

	// We cannot cache Java environments as they are only valid in the current
	// thread.  We can, however, cache the Java VM and obtain an environment
	// from it in whatever thread happens to be running at the time.
	var jVM *C.JavaVM
	if status := C.GetJavaVM(env, &jVM); status != 0 {
		return nil, fmt.Errorf("couldn't get Java VM from the (Java) environment")
	}

	// Reference Java object: it will be de-referenced when the Go pointer is
	// garbage-collected (through the finalizer callback we setup just below.
	obj = C.NewGlobalRef(env, obj)

	// Map the Go pointer to the Java object and vice-versa.  We store the Go
	// pointer as an int64 to hide its reference from the Go runtime.  This
	// allows us to trigger the Go finalizer when the references disappear from
	// the rest of the Go code, which in turn allows us to remove the reference
	// to the Java object as soon as the Go pointer gets garbage-collected.
	goPtr := PtrValue(valptr)
	goToJavaRefs[int64(goPtr)] = obj
	javaToGoRefs[identityHash] = int64(goPtr)

	// Setup the finalizer.
	setupRefFinalizer(jVM, valptr)

	return unsafe.Pointer(goPtr), nil
}

func setupRefFinalizer(jVM *C.JavaVM, valptr interface{}) {
	runtime.SetFinalizer(valptr, func(valptr interface{}) {
		jEnv, freeFunc := GetEnv(jVM)
		env := (*C.JNIEnv)(jEnv)
		defer freeFunc()

		goJavaRefsMutex.Lock()
		defer goJavaRefsMutex.Unlock()
		goPtr := PtrValue(valptr)
		jObj, ok := goToJavaRefs[int64(goPtr)]
		if !ok {
			log.Printf("Couldn't find Java object associated with Go pointer: %v", valptr)
			return
		}
		identityHash, err := getIdentityHash(env, jObj)
		if err != nil {
			log.Printf("Couldn't get identity hash for Java object: %v", err)
			return
		}
		delete(goToJavaRefs, int64(goPtr))
		delete(javaToGoRefs, identityHash)
		C.DeleteGlobalRef(env, jObj)
	})
}

func getIdentityHash(env *C.JNIEnv, jObj C.jobject) (int, error) {
	return CallStaticIntMethod(env, jSystemClass, "identityHashCode", []Sign{ObjectSign}, jObj)
}

// goToJavaRefs maps from a Go pointer (represented by an int64) to a Java
// object.
var goToJavaRefs map[int64]C.jobject // GUARDED_BY(goJavaRefsMutex)

// javaToGoRefs maps from a Java object (represented by an int value of its
// identity hashCode) to a Go pointer (represented by an int64).
var javaToGoRefs map[int]int64 // GUARDED_BY(goJavaRefsMutex)

// goJavaRefsMutex guards the access to the above two maps.
var goJavaRefsMutex sync.Mutex

// goRefs stores references to instances of various Go types, namely instances
// that are referenced only by the Java code.  The only purpose of this store
// is to prevent Go runtime from garbage collecting those instances.
var goRefs = newSafeRefCounter()

// newSafeRefCounter returns a new instance of a thread-safe reference counter.
func newSafeRefCounter() *safeRefCounter {
	return &safeRefCounter{
		refs: make(map[interface{}]int),
	}
}

// safeRefCounter is a thread-safe reference counter.
type safeRefCounter struct {
	lock sync.Mutex
	refs map[interface{}]int
}

func (c *safeRefCounter) ref(value interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	count, ok := c.refs[value]
	if !ok {
		c.refs[value] = 1
	} else {
		c.refs[value] = count + 1
	}
}

func (c *safeRefCounter) unref(value interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	count, ok := c.refs[value]
	if !ok {
		log.Println("Unrefing value %v or type %T that hasn't been refed before", value, value)
		return
	}
	if count == 0 {
		log.Println("Ref count for value %v is zero: that shouldn't happen", value)
		return
	}
	if count > 1 {
		c.refs[value] = count - 1
		return
	}
	delete(c.refs, value)
}
