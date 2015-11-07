// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package util

import (
	"fmt"
)

// #include "jni_wrapper.h"
import "C"

// HasOption returns true iff the option with the provided key exists.
func HasOption(env Env, jOpts Object, key string) (bool, error) {
	if jOpts.IsNull() {
		return false, nil
	}
	return CallBooleanMethod(env, jOpts, "has", []Sign{StringSign}, key)
}

// GetOption returns the option with the given key.  It returns NullObject if the option
// doesn't exist.
func GetOption(env Env, jOpts Object, key string) (Object, error) {
	if jOpts.IsNull() {
		return NullObject, nil
	}
	return CallObjectMethod(env, jOpts, "get", []Sign{StringSign}, ObjectSign, key)
}

// GetIntOption returns the integer option with the given key.  It returns 0 if the option
// doesn't exist.
func GetIntOption(env Env, jOpts Object, key string) (int, error) {
	jVal, err := GetOption(env, jOpts, key)
	if err != nil {
		return 0, err
	}
	if jVal.IsNull() {
		return 0, nil
	}
	if !IsInstanceOf(env, jVal, jIntegerClass) {
		return 0, fmt.Errorf("Expected option with key %s to be of Integer type", key)
	}
	return CallIntMethod(env, jVal, "intValue", nil)
}

// GetBooleanOption returns the boolean option with the given key.  It returns 'false' if the option
// doesn't exist.
func GetBooleanOption(env Env, jOpts Object, key string) (bool, error) {
	jVal, err := GetOption(env, jOpts, key)
	if err != nil {
		return false, err
	}
	if jVal.IsNull() {
		return false, nil
	}
	if !IsInstanceOf(env, jVal, jBooleanClass) {
		return false, fmt.Errorf("Expected option with key %s to be of Boolean type", key)
	}
	return CallBooleanMethod(env, jVal, "booleanValue", nil)
}

// StringOption returns the string option with the given key.  It returns an empty string if the
// option doesn't exist.
func GetStringOption(env Env, jOpts Object, key string) (string, error) {
	jVal, err := GetOption(env, jOpts, key)
	if err != nil {
		return "", err
	}
	if jVal.IsNull() {
		return "", nil
	}
	if !IsInstanceOf(env, jVal, jStringClass) {
		return "", fmt.Errorf("Expected option with key %s to be of String type", key)
	}
	return GoString(env, jVal), nil
}
