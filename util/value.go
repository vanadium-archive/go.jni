// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package util

import (
	"reflect"
	"time"
)

// #include "jni_wrapper.h"
//
// jvalue jBoolValue(jboolean val) {
//   jvalue ret = { .z = val };
//   return ret;
// }
// jvalue jByteValue(jbyte val) {
//   jvalue ret = { .b = val };
//   return ret;
// }
// jvalue jCharValue(jchar val) {
//   jvalue ret = { .c = val };
//   return ret;
// }
// jvalue jShortValue(jshort val) {
//   jvalue ret = { .s = val };
//   return ret;
// }
// jvalue jIntValue(jint val) {
//   jvalue ret = { .i = val };
//   return ret;
// }
// jvalue jLongValue(jlong val) {
//   jvalue ret = { .j = val };
//   return ret;
// }
// jvalue jFloatValue(jfloat val) {
//   jvalue ret = { .f = val };
//   return ret;
// }
// jvalue jDoubleValue(jdouble val) {
//   jvalue ret = { .d = val };
//   return ret;
// }
// jvalue jObjectValue(jobject val) {
//   jvalue ret = { .l = val };
//   return ret;
// }
import "C"

var errJValue = C.jObjectValue(nil)

// jValue converts a Go value into a Java value with the given sign.
func jValue(env Env, v interface{}, sign Sign) (C.jvalue, bool) {
	switch sign {
	case BoolSign:
		return jBoolValue(v)
	case ByteSign:
		return jByteValue(v)
	case CharSign:
		return jCharValue(v)
	case ShortSign:
		return jShortValue(v)
	case IntSign:
		return jIntValue(v)
	case LongSign:
		return jLongValue(v)
	case StringSign:
		return jStringValue(env, v)
	case DateTimeSign:
		return jDateTimeValue(env, v)
	case DurationSign:
		return jDurationValue(env, v)
	case VExceptionSign:
		return jVExceptionValue(env, v)
	case ArraySign(ByteSign):
		return jByteArrayValue(env, v)
	case ArraySign(StringSign):
		return jStringArrayValue(env, v)
	default:
		return jObjectValue(v)
	}
}

func jBoolValue(v interface{}) (C.jvalue, bool) {
	val, ok := v.(bool)
	if !ok {
		return errJValue, false
	}
	jBool := C.jboolean(C.JNI_FALSE)
	if val {
		jBool = C.jboolean(C.JNI_TRUE)
	}
	return C.jBoolValue(jBool), true
}

func jByteValue(v interface{}) (C.jvalue, bool) {
	val, ok := intValue(v)
	if !ok {
		return errJValue, false
	}
	return C.jByteValue(C.jbyte(val)), true
}

func jCharValue(v interface{}) (C.jvalue, bool) {
	val, ok := intValue(v)
	if !ok {
		return errJValue, false
	}
	return C.jCharValue(C.jchar(val)), true
}

func jShortValue(v interface{}) (C.jvalue, bool) {
	val, ok := intValue(v)
	if !ok {
		return errJValue, false
	}
	return C.jShortValue(C.jshort(val)), true
}

func jIntValue(v interface{}) (C.jvalue, bool) {
	val, ok := intValue(v)
	if !ok {
		return errJValue, false
	}
	return C.jIntValue(C.jint(val)), true
}

func jLongValue(v interface{}) (C.jvalue, bool) {
	val, ok := intValue(v)
	if !ok {
		return errJValue, false
	}
	return C.jLongValue(C.jlong(val)), true
}

func jStringValue(env Env, v interface{}) (C.jvalue, bool) {
	str, ok := v.(string)
	if !ok {
		return errJValue, false
	}
	return jObjectValue(JString(env, str))
}

func jDateTimeValue(env Env, v interface{}) (C.jvalue, bool) {
	t, ok := v.(time.Time)
	if !ok {
		return errJValue, false
	}
	jTime, err := JTime(env, t)
	if err != nil {
		return errJValue, false
	}
	return jObjectValue(jTime)
}

func jDurationValue(env Env, v interface{}) (C.jvalue, bool) {
	d, ok := v.(time.Duration)
	if !ok {
		return errJValue, false
	}
	jDuration, err := JDuration(env, d)
	if err != nil {
		return errJValue, false
	}
	return jObjectValue(jDuration)
}

func jVExceptionValue(env Env, v interface{}) (C.jvalue, bool) {
	err, ok := v.(error)
	if !ok {
		return errJValue, false
	}
	jVException, err := JVException(env, err)
	if err != nil {
		return errJValue, false
	}
	return jObjectValue(jVException)
}

func jByteArrayValue(env Env, v interface{}) (C.jvalue, bool) {
	arr, ok := v.([]byte)
	if !ok {
		return errJValue, false
	}
	jArr, err := JByteArray(env, arr)
	if err != nil {
		return errJValue, false
	}
	return jObjectValue(jArr)
}

func jStringArrayValue(env Env, v interface{}) (C.jvalue, bool) {
	arr, ok := v.([]string)
	if !ok {
		return errJValue, false
	}
	jArr, err := JStringArray(env, arr)
	if err != nil {
		return errJValue, false
	}
	return jObjectValue(jArr)
}

func jObjectValue(v interface{}) (C.jvalue, bool) {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() { // nil value
		return C.jObjectValue(nil), true
	}
	switch val := v.(type) {
	case Object:
		return C.jObjectValue(val.value()), true
	case Class:
		return C.jObjectValue(C.jobject(val.value())), true
	default:
		return errJValue, false
	}
}

func intValue(v interface{}) (int64, bool) {
	switch val := v.(type) {
	case int64:
		return val, true
	case int:
		return int64(val), true
	case int32:
		return int64(val), true
	case int16:
		return int64(val), true
	case int8:
		return int64(val), true
	case uint64:
		return int64(val), true
	case uint:
		return int64(val), true
	case uint32:
		return int64(val), true
	case uint16:
		return int64(val), true
	case uint8:
		return int64(val), true
	default:
		return 0, false
	}
}
