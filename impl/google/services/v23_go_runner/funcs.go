// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package v23_go_runner

import (
	"v.io/v23/context"
)

// v23GoRunnerFuncs is a map containing go functions keys by unique strings
// intended to be run by java/android applications using V23GoRunner.run(key).
// Users must add function entries to this map and rebuild lib/android-lib in
// the vanadium java repository.
var v23GoRunnerFuncs = map[string]func(*context.T) error{
	"bt-rpc": func(ctx *context.T) error {
		ctx.Errorf("bt-rpc test to be implemented")
		return nil
	},
}
