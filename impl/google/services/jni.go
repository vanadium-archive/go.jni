// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package services

import (
	jgroups "v.io/x/jni/impl/google/services/groups"
	jmounttable "v.io/x/jni/impl/google/services/mounttable"
	// See TODO below.
	//jsyncbase "v.io/x/jni/impl/google/services/syncbase"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

// Init initializes the JNI code with the given Java environment.  This method
// must be invoked before any other method in this package and must be called
// from the main Java thread (e.g., On_Load()).
func Init(env jutil.Env) error {
	if err := jgroups.Init(env); err != nil {
		return err
	}
	if err := jmounttable.Init(env); err != nil {
		return err
	}
	// TODO(sadovsky): Temporarily disable Syncbase Java code (as of 2016-03-28),
	// while we update the code to the simplified API.
	/*
		if err := jsyncbase.Init(env); err != nil {
			return err
		}
	*/
	return nil
}
