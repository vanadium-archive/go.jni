// +build android

package util

import (
	"fmt"
	"log"
	"reflect"
	"runtime/debug"
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
	if _, ok := val.(unsafe.Pointer); ok {
		return true
	}
	return reflect.ValueOf(val).Kind() == reflect.Ptr
}

// PtrValue returns the value of the pointer as a uintptr.
func PtrValue(ptr interface{}) uintptr {
	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr && v.Kind() != reflect.UnsafePointer {
		panic("must pass pointer value to PtrValue")
	}
	return v.Pointer()
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

func (c *safeRefCounter) ref(valptr interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	count, ok := c.refs[valptr]
	if !ok {
		c.refs[valptr] = 1
	} else {
		c.refs[valptr] = count + 1
	}
}

func (c *safeRefCounter) unref(valptr interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	count, ok := c.refs[valptr]
	if !ok {
		log.Printf("Unrefing pointer %d of type %T that hasn't been refed before, stack: %s", int64(PtrValue(valptr)), valptr, string(debug.Stack()))
		return
	}
	if count == 0 {
		log.Printf("Ref count for pointer %d of type %T is zero: that shouldn't happen, stack: %s", int64(PtrValue(valptr)), valptr, string(debug.Stack()))
		return
	}
	if count > 1 {
		c.refs[valptr] = count - 1
		return
	}
	delete(c.refs, valptr)
}
