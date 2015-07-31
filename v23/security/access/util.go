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
func GoAccessList(env jutil.Env, jAccessList jutil.Object) (acl access.AccessList, err error) {
	err = jutil.GoVomCopy(env, jAccessList, jAccessListClass, &acl)
	return
}
