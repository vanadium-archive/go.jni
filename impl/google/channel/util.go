// +build android

package channel

import (
	"unsafe"

	jutil "v.io/x/jni/util"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// JavaInputChannel converts the provided Go channel of C.jobject values into a Java
// InputChannel object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaInputChannel(jEnv, chPtr, sourceChanPtr interface{}) (unsafe.Pointer, error) {
	jInputChannel, err := jutil.NewObject(jEnv, jInputChannelImplClass, []jutil.Sign{jutil.LongSign, jutil.LongSign}, int64(jutil.PtrValue(chPtr)), int64(jutil.PtrValue(sourceChanPtr)))
	if err != nil {
		return nil, err
	}
	jutil.GoRef(chPtr)         // Un-refed when the InputChannel is finalized.
	jutil.GoRef(sourceChanPtr) // Un-refed when the InputChannel is finalized.
	return jInputChannel, nil
}
