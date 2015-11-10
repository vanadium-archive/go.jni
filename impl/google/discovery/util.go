// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package discovery

import (
	jutil "v.io/x/jni/util"

	"v.io/v23/discovery"
	idiscovery "v.io/x/ref/lib/discovery"
)
// #include "jni.h"
import "C"

// JavaDiscovery converts a Go discovery instance into a Java discovery instance.
func JavaDiscovery(env jutil.Env, d discovery.T) (jutil.Object, error) {
	trigger := idiscovery.NewTrigger()
	// This reference will get unrefed when the jDiscovery object below is finalized.
	jutil.GoRef(trigger)

	jDiscovery, err := jutil.NewObject(env, jVDiscoveryImplClass, []jutil.Sign{jutil.LongSign, jutil.LongSign}, int64(jutil.PtrValue(&d)), int64(jutil.PtrValue(trigger)))
	if err != nil {
		return jutil.NullObject, err
	}
	return jDiscovery, nil
}
