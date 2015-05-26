// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package v23

import (
	jcontext "v.io/x/jni/v23/context"
	ji18n "v.io/x/jni/v23/i18n"
	jnaming "v.io/x/jni/v23/naming"
	jsecurity "v.io/x/jni/v23/security"
	jaccess "v.io/x/jni/v23/security/access"
)

// #include "jni.h"
import "C"

// Init initializes the JNI code with the given Java environment.  This method
// must be invoked before any other method in this package and must be called
// from the main Java thread (e.g., On_Load()).
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java environment is passed in an empty
// interface and then cast into the package-local environment type.
func Init(jEnv interface{}) error {
	if err := jcontext.Init(jEnv); err != nil {
		return err
	}
	if err := ji18n.Init(jEnv); err != nil {
		return err
	}
	if err := jnaming.Init(jEnv); err != nil {
		return err
	}
	if err := jsecurity.Init(jEnv); err != nil {
		return err
	}
	if err := jaccess.Init(jEnv); err != nil {
		return err
	}
	return nil
}
