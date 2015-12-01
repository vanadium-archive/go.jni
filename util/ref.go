// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package util

import (
	"fmt"
	"reflect"
	"sync"
	"unsafe"
)

// #include "jni_wrapper.h"
import "C"

// NewGlobalRef creates a new global reference to the object referred to by the
// obj argument.  The obj argument may be a global or local reference. Global
// references must be explicitly disposed of by calling DeleteGlobalRef().
func NewGlobalRef(env Env, obj Object) Object {
	return Object(uintptr(unsafe.Pointer(C.NewGlobalRef(env.value(), obj.value()))))
}

// DeleteGlobalRef deletes the global reference pointed to by obj.
func DeleteGlobalRef(env Env, obj Object) {
	C.DeleteGlobalRef(env.value(), obj.value())
}

// NewLocalRef creates a new local reference that refers to the same object
// as obj. The given obj may be a global or local reference. Returns null if
// ref refers to null.
func NewLocalRef(env Env, obj Object) Object {
	return Object(uintptr(unsafe.Pointer(C.NewLocalRef(env.value(), obj.value()))))
}

// DeleteLocalRef deletes the local reference pointed to by obj.
func DeleteLocalRef(env Env, obj Object) {
	C.DeleteLocalRef(env.value(), obj.value())
}

// IsGlobalRef returns true iff the reference pointed to by obj is a global reference.
func IsGlobalRef(env Env, obj Object) bool {
	return C.GetObjectRefType(env.value(), obj.value()) == C.JNIGlobalRefType
}

// IsLocalRef returns true iff the reference pointed to by obj is a local reference.
func IsLocalRef(env Env, obj Object) bool {
	return C.GetObjectRefType(env.value(), obj.value()) == C.JNILocalRefType
}

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
	v := reflect.ValueOf(val)
	return v.Kind() == reflect.Ptr || v.Kind() == reflect.UnsafePointer
}

// PtrValue returns the value of the pointer as a uintptr.
func PtrValue(ptr interface{}) uintptr {
	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr && v.Kind() != reflect.UnsafePointer {
		panic(fmt.Sprintf("must pass pointer value to PtrValue, was %v ", v.Type()))
	}
	return v.Pointer()
}

// DerefOrDie dereferences the provided (pointer) value, or panic-s if the value
// isn't of pointer type.
func DerefOrDie(i interface{}) interface{} {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("want reflect.Ptr value for %v, have %v", i, v.Type()))
	}
	return v.Elem().Interface()
}

// NativePtr returns the value of the provided Go pointer as an unsafe.Pointer.
// This function should only be used for converting Go pointers that have been
// passed in to Java and then back into Go, and are of (local-package) type
// C.jlong.
func NativePtr(goPtr interface{}) unsafe.Pointer {
	v := reflect.ValueOf(goPtr)
	return unsafe.Pointer(uintptr(v.Int()))
}

// goRefs stores references to instances of various Go types, namely instances
// that are referenced only by the Java code.  The only purpose of this store
// is to prevent Go runtime from garbage collecting those instances.
var goRefs = newSafeRefCounter()

type refData struct {
	instance interface{}
	count    int
}

// newSafeRefCounter returns a new instance of a thread-safe reference counter.
func newSafeRefCounter() *safeRefCounter {
	return &safeRefCounter{
		refs: make(map[uintptr]*refData),
	}
}

// safeRefCounter is a thread-safe reference counter.
type safeRefCounter struct {
	lock sync.Mutex
	refs map[uintptr]*refData
}

// ref increases the reference count to the given valptr by 1.
func (c *safeRefCounter) ref(valptr interface{}) {
	p := PtrValue(valptr)
	c.lock.Lock()
	defer c.lock.Unlock()
	ref, ok := c.refs[p]
	if !ok {
		c.refs[p] = &refData{
			instance: valptr,
			count:    1,
		}
	} else {
		ref.count++
	}
}

// unref decreases the reference count of the given valptr by 1, returning
// the new reference count value.
func (c *safeRefCounter) unref(valptr interface{}) int {
	c.lock.Lock()
	defer c.lock.Unlock()
	p := PtrValue(valptr)
	ref, ok := c.refs[p]
	if !ok {
		panic(fmt.Sprintf("Unrefing pointer %d of type %T that hasn't been refed before", int64(p), valptr))
	}
	count := ref.count
	if count == 0 {
		panic(fmt.Sprintf("Ref count for pointer %d of type %T is zero", int64(p), valptr))
	}
	if count > 1 {
		ref.count--
		return ref.count
	}
	delete(c.refs, p)
	return 0
}
