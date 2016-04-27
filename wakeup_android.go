// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package jni

import (
	"fmt"
	"time"

	"v.io/v23/context"
	"v.io/v23/namespace"
	"v.io/v23/naming"
	"v.io/v23/security/access"
)

// wakeupNamespace is a namespace used by the persistent Android services
// to mount themselves in a way that they can be woken up whenever a
// client resolves their names.
//
// This namespace consists of two sub-parts:
//    1) The original namespace (OrigNS),
//    2) The mount root used for waking up the service (WakeMTRoot)
//
// Let's say the service wishes to mount itself under a relative name
// 'a/b'. (Absolute names are disallowed.)  The service first mounts its
// endpoints under 'WakeMTRoot/server/a/b' and then mounts name
// 'WakeMTRoot/client/a/b' under 'OrigNS/a/b'.  When the client resolves
// 'OrigNS/a/b', the name resolution will trigger resolution of
// 'WakeMTRoot/client/a/b', which in turn triggers the process of waking
// up the server.
//
// Note that we prefix mounted names by 'client' and 'server' above.
// This is needed in order to distinguish the cases of a server mounting itself
// into the mounttable (i.e., no wakeup needed) from the case of a client
// resolving a name (i.e., wakeup needed).
type wakeupNamespace struct {
	wakeupMountRoot string
	ns              namespace.T
}

func (w *wakeupNamespace) Mount(ctx *context.T, name, server string, ttl time.Duration, opts ...naming.NamespaceOpt) error {
	if naming.Rooted(name) {
		return fmt.Errorf("Mount(%q): rooted names not allowed in wakeup namespace", name)
	}
	mountName := naming.Join(w.wakeupMountRoot, "server", name)
	if err := w.ns.Mount(ctx, mountName, server, ttl, opts...); err != nil {
		return err
	}
	// We set an infinite TTL at the original mount location as it allows us to wake up a
	// server even after its mount entries have expired.  (It's fine if the wakeup MT entries
	// do expire.)  Since entries are mounted forever, we explicitly have to clean up previous
	// entries.
	w.ns.Delete(ctx, name, false)
	wakeupName := naming.Join(w.wakeupMountRoot, "client", name)
	var zeroTTL time.Duration
	return w.ns.Mount(ctx, name, wakeupName, zeroTTL, append(opts, naming.ServesMountTable(true))...)
}

func (w *wakeupNamespace) Unmount(ctx *context.T, name, server string, opts ...naming.NamespaceOpt) error {
	if naming.Rooted(name) {
		return fmt.Errorf("Unmount(%q): rooted names not allowed in wakeup namespace", name)
	}
	mountName := naming.Join(w.wakeupMountRoot, "server", name)
	return w.ns.Unmount(ctx, mountName, server, opts...)
	// We never unmount at the original location as it allows us to wake up a sleeping server.
}

func (w *wakeupNamespace) Delete(ctx *context.T, name string, deleteSubtree bool, opts ...naming.NamespaceOpt) error {
	if naming.Rooted(name) {
		return fmt.Errorf("Delete(%q): rooted names not allowed in wakeup namespace", name)
	}
	mountName := naming.Join(w.wakeupMountRoot, "server", name)
	if err := w.ns.Delete(ctx, mountName, deleteSubtree, opts...); err != nil {
		return err
	}
	return w.ns.Delete(ctx, name, deleteSubtree, opts...)
}

func (w *wakeupNamespace) Resolve(ctx *context.T, name string, opts ...naming.NamespaceOpt) (entry *naming.MountEntry, err error) {
	return w.ns.Resolve(ctx, name, opts...)
}

func (w *wakeupNamespace) ShallowResolve(ctx *context.T, name string, opts ...naming.NamespaceOpt) (entry *naming.MountEntry, err error) {
	return w.ns.ShallowResolve(ctx, name, opts...)
}

func (w *wakeupNamespace) ResolveToMountTable(ctx *context.T, name string, opts ...naming.NamespaceOpt) (entry *naming.MountEntry, err error) {
	return w.ns.ResolveToMountTable(ctx, name, opts...)
}

func (w *wakeupNamespace) FlushCacheEntry(ctx *context.T, name string) bool {
	if naming.Rooted(name) {
		return false
	}
	mountName := naming.Join(w.wakeupMountRoot, "server", name)
	return w.ns.FlushCacheEntry(ctx, mountName) || w.ns.FlushCacheEntry(ctx, name)
}

func (w *wakeupNamespace) CacheCtl(ctls ...naming.CacheCtl) []naming.CacheCtl {
	return w.ns.CacheCtl(ctls...)
}

func (w *wakeupNamespace) Glob(ctx *context.T, pattern string, opts ...naming.NamespaceOpt) (<-chan naming.GlobReply, error) {
	return w.ns.Glob(ctx, pattern, opts...)
}

func (w *wakeupNamespace) SetRoots(roots ...string) error {
	return w.ns.SetRoots(roots...)
}

func (w *wakeupNamespace) Roots() []string {
	return w.ns.Roots()
}

func (w *wakeupNamespace) SetPermissions(ctx *context.T, name string, perms access.Permissions, version string, opts ...naming.NamespaceOpt) error {
	if naming.Rooted(name) {
		return fmt.Errorf("SetPermissions(%q): rooted names not allowed in wakeup namespace", name)
	}
	mountName := naming.Join(w.wakeupMountRoot, "server", name)
	if err := w.ns.SetPermissions(ctx, mountName, perms, version, opts...); err != nil {
		return err
	}
	return w.ns.SetPermissions(ctx, name, perms, version, opts...)
}

func (w *wakeupNamespace) GetPermissions(ctx *context.T, name string, opts ...naming.NamespaceOpt) (perms access.Permissions, version string, err error) {
	return w.ns.GetPermissions(ctx, name, opts...)
}
