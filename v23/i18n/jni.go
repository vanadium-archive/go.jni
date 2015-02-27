// +build android

package i18n

import (
	jutil "v.io/jni/util"
	"v.io/v23/i18n"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// Init initializes the JNI code with the given Java environment.  This method
// must be invoked before any other method in this package and must be called
// from the main Java thread (e.g., On_Load()).
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java environment is passed in an empty
// interface and then cast into the package-local environment type.
func Init(jEnv interface{}) error {
	return nil
}

//export Java_io_v_v23_i18n_Catalog_nativeFormatParams
func Java_io_v_v23_i18n_Catalog_nativeFormatParams(env *C.JNIEnv, jCatalog C.jclass, jFormat C.jstring, jParams C.jobjectArray) C.jstring {
	format := jutil.GoString(env, jFormat)
	strParams := jutil.GoStringArray(env, jParams)
	params := make([]interface{}, len(strParams))
	for i, strParam := range strParams {
		params[i] = strParam
	}
	result := i18n.FormatParams(format, params...)
	ret := jutil.JString(env, result)
	return C.jstring(ret)
}
