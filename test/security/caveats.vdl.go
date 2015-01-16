// This file was auto-generated by the veyron vdl tool.
// Source: caveats.vdl

package security

import (
	// The non-user imports are prefixed with "__" to prevent collisions.
	__vdl "v.io/core/veyron2/vdl"
)

// TestCaveat is a caveat that's used in various security tests.
type TestCaveat string

func (TestCaveat) __VDLReflect(struct {
	Name string "v.io/jni/test/security.TestCaveat"
}) {
}

func init() {
	__vdl.Register(TestCaveat(""))
}