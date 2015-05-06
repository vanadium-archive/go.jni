// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package access

import (
	"v.io/v23/security/access"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

// GoAccessList converts the provided Java AccessList into a Go AccessList.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoAccessList(jEnv, jAccessList interface{}) (acl access.AccessList, err error) {
	err = jutil.GoVomCopy(jEnv, jAccessList, jAccessListClass, &acl)
	return
}
