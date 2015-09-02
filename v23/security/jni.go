// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package security

import (
	"fmt"
	"unsafe"

	"v.io/v23/security"
	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
	vsecurity "v.io/x/ref/lib/security"
)

// #include "jni.h"
import "C"

var (
	dischargeSign        = jutil.ClassSign("io.v.v23.security.Discharge")
	dischargeImpetusSign = jutil.ClassSign("io.v.v23.security.DischargeImpetus")
	principalSign        = jutil.ClassSign("io.v.v23.security.VPrincipal")
	blessingsSign        = jutil.ClassSign("io.v.v23.security.Blessings")
	wireBlessingsSign    = jutil.ClassSign("io.v.v23.security.WireBlessings")
	wireDischargeSign    = jutil.ClassSign("io.v.v23.security.WireDischarge")
	blessingStoreSign    = jutil.ClassSign("io.v.v23.security.BlessingStore")
	blessingRootsSign    = jutil.ClassSign("io.v.v23.security.BlessingRoots")
	blessingPatternSign  = jutil.ClassSign("io.v.v23.security.BlessingPattern")
	signerSign           = jutil.ClassSign("io.v.v23.security.VSigner")
	caveatSign           = jutil.ClassSign("io.v.v23.security.Caveat")
	contextSign          = jutil.ClassSign("io.v.v23.context.VContext")
	callSign             = jutil.ClassSign("io.v.v23.security.Call")
	signatureSign        = jutil.ClassSign("io.v.v23.security.VSignature")
	publicKeySign        = jutil.ClassSign("java.security.interfaces.ECPublicKey")

	// Global reference for io.v.v23.security.Blessings class.
	jBlessingsClass jutil.Class
	// Global reference for io.v.v23.security.Caveat class.
	jCaveatClass jutil.Class
	// Global reference for io.v.v23.security.WireBlessings class.
	jWireBlessingsClass jutil.Class
	// Global reference for io.v.v23.security.VPrincipalImpl class.
	jVPrincipalImplClass jutil.Class
	// Global reference for io.v.v23.security.BlessingStoreImpl class.
	jBlessingStoreImplClass jutil.Class
	// Global reference for io.v.v23.security.BlessingRootsImpl class.
	jBlessingRootsImplClass jutil.Class
	// Global reference for io.v.v23.security.CallImpl class.
	jCallImplClass jutil.Class
	// Global reference for io.v.v23.security.BlessingPattern class.
	jBlessingPatternClass jutil.Class
	// Global reference for io.v.v23.security.CaveatRegistry class.
	jCaveatRegistryClass jutil.Class
	// Global reference for io.v.v23.security.Util class.
	jUtilClass jutil.Class
	// Global reference for java.lang.Object class.
	jObjectClass jutil.Class
	// Global reference for io.v.v23.security.VSecurity class.
	jVSecurityClass jutil.Class
	// Global reference for io.v.v23.security.Discharge class.
	jDischargeClass jutil.Class
	// Global reference for io.v.v23.security.DischargeImpetus class.
	jDischargeImpetusClass jutil.Class
)

// Init initializes the JNI code with the given Java evironment. This method
// must be called from the main Java thread.
// interface and then cast into the package-local environment type.
func Init(env jutil.Env) error {
	security.OverrideCaveatValidation(caveatValidator)

	// Cache global references to all Java classes used by the package.  This is
	// necessary because JNI gets access to the class loader only in the system
	// thread, so we aren't able to invoke FindClass in other threads.
	var err error
	jBlessingsClass, err = jutil.JFindClass(env, "io/v/v23/security/Blessings")
	if err != nil {
		return err
	}
	jCaveatClass, err = jutil.JFindClass(env, "io/v/v23/security/Caveat")
	if err != nil {
		return err
	}
	jWireBlessingsClass, err = jutil.JFindClass(env, "io/v/v23/security/WireBlessings")
	if err != nil {
		return err
	}
	jVPrincipalImplClass, err = jutil.JFindClass(env, "io/v/v23/security/VPrincipalImpl")
	if err != nil {
		return err
	}
	jBlessingStoreImplClass, err = jutil.JFindClass(env, "io/v/v23/security/BlessingStoreImpl")
	if err != nil {
		return err
	}
	jBlessingRootsImplClass, err = jutil.JFindClass(env, "io/v/v23/security/BlessingRootsImpl")
	if err != nil {
		return err
	}
	jCallImplClass, err = jutil.JFindClass(env, "io/v/v23/security/CallImpl")
	if err != nil {
		return err
	}
	jBlessingPatternClass, err = jutil.JFindClass(env, "io/v/v23/security/BlessingPattern")
	if err != nil {
		return err
	}
	jCaveatRegistryClass, err = jutil.JFindClass(env, "io/v/v23/security/CaveatRegistry")
	if err != nil {
		return err
	}
	jUtilClass, err = jutil.JFindClass(env, "io/v/v23/security/Util")
	if err != nil {
		return err
	}
	jObjectClass, err = jutil.JFindClass(env, "java/lang/Object")
	if err != nil {
		return err
	}
	jVSecurityClass, err = jutil.JFindClass(env, "io/v/v23/security/VSecurity")
	if err != nil {
		return err
	}
	jDischargeClass, err = jutil.JFindClass(env, "io/v/v23/security/Discharge")
	if err != nil {
		return err
	}
	jDischargeImpetusClass, err = jutil.JFindClass(env, "io/v/v23/security/DischargeImpetus")
	if err != nil {
		return err
	}
	return nil
}

//export Java_io_v_v23_security_CallImpl_nativeTimestamp
func Java_io_v_v23_security_CallImpl_nativeTimestamp(jenv *C.JNIEnv, jCall C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	t := (*(*security.Call)(jutil.NativePtr(goPtr))).Timestamp()
	jTime, err := jutil.JTime(env, t)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jTime))
}

//export Java_io_v_v23_security_CallImpl_nativeMethod
func Java_io_v_v23_security_CallImpl_nativeMethod(jenv *C.JNIEnv, jCall C.jobject, goPtr C.jlong) C.jstring {
	env := jutil.WrapEnv(jenv)
	method := (*(*security.Call)(jutil.NativePtr(goPtr))).Method()
	jMethod := jutil.JString(env, jutil.CamelCase(method))
	return C.jstring(unsafe.Pointer(jMethod))
}

//export Java_io_v_v23_security_CallImpl_nativeMethodTags
func Java_io_v_v23_security_CallImpl_nativeMethodTags(jenv *C.JNIEnv, jCall C.jobject, goPtr C.jlong) C.jobjectArray {
	env := jutil.WrapEnv(jenv)
	tags := (*(*security.Call)(jutil.NativePtr(goPtr))).MethodTags()
	jTags, err := jutil.JVDLValueArray(env, tags)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobjectArray(unsafe.Pointer(jTags))
}

//export Java_io_v_v23_security_CallImpl_nativeSuffix
func Java_io_v_v23_security_CallImpl_nativeSuffix(jenv *C.JNIEnv, jCall C.jobject, goPtr C.jlong) C.jstring {
	env := jutil.WrapEnv(jenv)
	jSuffix := jutil.JString(env, (*(*security.Call)(jutil.NativePtr(goPtr))).Suffix())
	return C.jstring(unsafe.Pointer(jSuffix))
}

//export Java_io_v_v23_security_CallImpl_nativeRemoteDischarges
func Java_io_v_v23_security_CallImpl_nativeRemoteDischarges(jenv *C.JNIEnv, jCall C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	remoteDischarges := (*(*security.Call)(jutil.NativePtr(goPtr))).RemoteDischarges()
	jObjectMap, err := javaDischargeMap(env, remoteDischarges)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jObjectMap))
}

//export Java_io_v_v23_security_CallImpl_nativeLocalDischarges
func Java_io_v_v23_security_CallImpl_nativeLocalDischarges(jenv *C.JNIEnv, jCall C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	localDischarges := (*(*security.Call)(jutil.NativePtr(goPtr))).LocalDischarges()
	jObjectMap, err := javaDischargeMap(env, localDischarges)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jObjectMap))
}

//export Java_io_v_v23_security_CallImpl_nativeLocalEndpoint
func Java_io_v_v23_security_CallImpl_nativeLocalEndpoint(jenv *C.JNIEnv, jCall C.jobject, goPtr C.jlong) C.jstring {
	env := jutil.WrapEnv(jenv)
	jEndpoint := jutil.JString(env, (*(*security.Call)(jutil.NativePtr(goPtr))).LocalEndpoint().String())
	return C.jstring(unsafe.Pointer(jEndpoint))
}

//export Java_io_v_v23_security_CallImpl_nativeRemoteEndpoint
func Java_io_v_v23_security_CallImpl_nativeRemoteEndpoint(jenv *C.JNIEnv, jCall C.jobject, goPtr C.jlong) C.jstring {
	env := jutil.WrapEnv(jenv)
	jEndpoint := jutil.JString(env, (*(*security.Call)(jutil.NativePtr(goPtr))).RemoteEndpoint().String())
	return C.jstring(unsafe.Pointer(jEndpoint))

}

//export Java_io_v_v23_security_CallImpl_nativeLocalPrincipal
func Java_io_v_v23_security_CallImpl_nativeLocalPrincipal(jenv *C.JNIEnv, jCall C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	principal := (*(*security.Call)(jutil.NativePtr(goPtr))).LocalPrincipal()
	jPrincipal, err := JavaPrincipal(env, principal)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jPrincipal))
}

//export Java_io_v_v23_security_CallImpl_nativeLocalBlessings
func Java_io_v_v23_security_CallImpl_nativeLocalBlessings(jenv *C.JNIEnv, jCall C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	blessings := (*(*security.Call)(jutil.NativePtr(goPtr))).LocalBlessings()
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jBlessings))
}

//export Java_io_v_v23_security_CallImpl_nativeRemoteBlessings
func Java_io_v_v23_security_CallImpl_nativeRemoteBlessings(jenv *C.JNIEnv, jCall C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	blessings := (*(*security.Call)(jutil.NativePtr(goPtr))).RemoteBlessings()
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jBlessings))
}

//export Java_io_v_v23_security_CallImpl_nativeFinalize
func Java_io_v_v23_security_CallImpl_nativeFinalize(jenv *C.JNIEnv, jCall C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goPtr))
}

//export Java_io_v_v23_security_VPrincipalImpl_nativeCreate
func Java_io_v_v23_security_VPrincipalImpl_nativeCreate(jenv *C.JNIEnv, jclass C.jclass) C.jobject {
	env := jutil.WrapEnv(jenv)
	principal, err := vsecurity.NewPrincipal()
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jPrincipal, err := JavaPrincipal(env, principal)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jPrincipal))
}

//export Java_io_v_v23_security_VPrincipalImpl_nativeCreateForSigner
func Java_io_v_v23_security_VPrincipalImpl_nativeCreateForSigner(jenv *C.JNIEnv, jclass C.jclass, jSigner C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	signerObj := jutil.WrapObject(jSigner)
	signer, err := GoSigner(env, signerObj)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	principal, err := vsecurity.NewPrincipalFromSigner(signer, nil)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jPrincipal, err := jutil.NewObject(env, jVPrincipalImplClass, []jutil.Sign{jutil.LongSign, signerSign, blessingStoreSign, blessingRootsSign}, int64(jutil.PtrValue(&principal)), signerObj, jutil.NullObject, jutil.NullObject)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jutil.GoRef(&principal) // Un-refed when the Java VPrincipalImpl is finalized.
	return C.jobject(unsafe.Pointer(jPrincipal))
}

//export Java_io_v_v23_security_VPrincipalImpl_nativeCreateForAll
func Java_io_v_v23_security_VPrincipalImpl_nativeCreateForAll(jenv *C.JNIEnv, jclass C.jclass, jSigner C.jobject, jStore C.jobject, jRoots C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	signerObj := jutil.WrapObject(jSigner)
	storeObj := jutil.WrapObject(jStore)
	rootsObj := jutil.WrapObject(jRoots)
	signer, err := GoSigner(env, signerObj)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	store, err := GoBlessingStore(env, storeObj)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	roots, err := GoBlessingRoots(env, rootsObj)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	principal, err := security.CreatePrincipal(signer, store, roots)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jPrincipal, err := jutil.NewObject(env, jVPrincipalImplClass, []jutil.Sign{jutil.LongSign, signerSign, blessingStoreSign, blessingRootsSign}, int64(jutil.PtrValue(&principal)), signerObj, storeObj, rootsObj)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jutil.GoRef(&principal) // Un-refed when the Java VPrincipalImpl is finalized.
	return C.jobject(unsafe.Pointer(jPrincipal))
}

//export Java_io_v_v23_security_VPrincipalImpl_nativeCreatePersistent
func Java_io_v_v23_security_VPrincipalImpl_nativeCreatePersistent(jenv *C.JNIEnv, jclass C.jclass, jPassphrase C.jstring, jDir C.jstring) C.jobject {
	env := jutil.WrapEnv(jenv)
	passphrase := jutil.GoString(env, jutil.WrapObject(jPassphrase))
	dir := jutil.GoString(env, jutil.WrapObject(jDir))
	principal, err := vsecurity.LoadPersistentPrincipal(dir, []byte(passphrase))
	if err != nil {
		if principal, err = vsecurity.CreatePersistentPrincipal(dir, []byte(passphrase)); err != nil {
			jutil.JThrowV(env, err)
			return nil
		}
	}
	jPrincipal, err := JavaPrincipal(env, principal)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jPrincipal))
}

//export Java_io_v_v23_security_VPrincipalImpl_nativeCreatePersistentForSigner
func Java_io_v_v23_security_VPrincipalImpl_nativeCreatePersistentForSigner(jenv *C.JNIEnv, jclass C.jclass, jSigner C.jobject, jDir C.jstring) C.jobject {
	env := jutil.WrapEnv(jenv)
	signerObj := jutil.WrapObject(jSigner)
	signer, err := GoSigner(env, signerObj)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	dir := jutil.GoString(env, jutil.WrapObject(jDir))
	stateSerializer, err := vsecurity.NewPrincipalStateSerializer(dir)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	principal, err := vsecurity.NewPrincipalFromSigner(signer, stateSerializer)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jPrincipal, err := jutil.NewObject(env, jVPrincipalImplClass, []jutil.Sign{jutil.LongSign, signerSign, blessingStoreSign, blessingRootsSign}, int64(jutil.PtrValue(&principal)), signerObj, jutil.NullObject, jutil.NullObject)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jutil.GoRef(&principal) // Un-refed when the Java VPrincipalImpl is finalized.
	return C.jobject(unsafe.Pointer(jPrincipal))
}

//export Java_io_v_v23_security_VPrincipalImpl_nativeBless
func Java_io_v_v23_security_VPrincipalImpl_nativeBless(jenv *C.JNIEnv, jVPrincipalImpl C.jobject, goPtr C.jlong, jKey C.jobject, jWith C.jobject, jExtension C.jstring, jCaveat C.jobject, jAdditionalCaveats C.jobjectArray) C.jobject {
	env := jutil.WrapEnv(jenv)
	key, err := GoPublicKey(env, jutil.WrapObject(jKey))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	with, err := GoBlessings(env, jutil.WrapObject(jWith))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	extension := jutil.GoString(env, jutil.WrapObject(jExtension))
	caveat, err := GoCaveat(env, jutil.WrapObject(jCaveat))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	additionalCaveats, err := GoCaveats(env, jutil.WrapObject(jAdditionalCaveats))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	blessings, err := (*(*security.Principal)(jutil.NativePtr(goPtr))).Bless(key, with, extension, caveat, additionalCaveats...)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jBlessings))
}

//export Java_io_v_v23_security_VPrincipalImpl_nativeBlessSelf
func Java_io_v_v23_security_VPrincipalImpl_nativeBlessSelf(jenv *C.JNIEnv, jVPrincipalImpl C.jobject, goPtr C.jlong, jName C.jstring, jCaveats C.jobjectArray) C.jobject {
	env := jutil.WrapEnv(jenv)
	name := jutil.GoString(env, jutil.WrapObject(jName))
	caveats, err := GoCaveats(env, jutil.WrapObject(jCaveats))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	blessings, err := (*(*security.Principal)(jutil.NativePtr(goPtr))).BlessSelf(name, caveats...)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jBlessings))
}

//export Java_io_v_v23_security_VPrincipalImpl_nativeSign
func Java_io_v_v23_security_VPrincipalImpl_nativeSign(jenv *C.JNIEnv, jVPrincipalImpl C.jobject, goPtr C.jlong, jMessage C.jbyteArray) C.jobject {
	env := jutil.WrapEnv(jenv)
	message := jutil.GoByteArray(env, jutil.WrapObject(jMessage))
	sig, err := (*(*security.Principal)(jutil.NativePtr(goPtr))).Sign(message)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jSig, err := JavaSignature(env, sig)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jSig))
}

//export Java_io_v_v23_security_VPrincipalImpl_nativePublicKey
func Java_io_v_v23_security_VPrincipalImpl_nativePublicKey(jenv *C.JNIEnv, jVPrincipalImpl C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	key := (*(*security.Principal)(jutil.NativePtr(goPtr))).PublicKey()
	jKey, err := JavaPublicKey(env, key)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jKey))
}

//export Java_io_v_v23_security_VPrincipalImpl_nativeBlessingsByName
func Java_io_v_v23_security_VPrincipalImpl_nativeBlessingsByName(jenv *C.JNIEnv, jVPrincipalImpl C.jobject, goPtr C.jlong, jPattern C.jobject) C.jobjectArray {
	env := jutil.WrapEnv(jenv)
	pattern, err := GoBlessingPattern(env, jutil.WrapObject(jPattern))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	blessings := (*(*security.Principal)(jutil.NativePtr(goPtr))).BlessingsByName(pattern)
	barr := make([]jutil.Object, len(blessings))
	for i, b := range blessings {
		var err error
		if barr[i], err = JavaBlessings(env, b); err != nil {
			jutil.JThrowV(env, err)
			return nil
		}
	}
	jArr, err := jutil.JObjectArray(env, barr, jBlessingsClass)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobjectArray(unsafe.Pointer(jArr))
}

//export Java_io_v_v23_security_VPrincipalImpl_nativeBlessingsInfo
func Java_io_v_v23_security_VPrincipalImpl_nativeBlessingsInfo(jenv *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong, jBlessings C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	blessings, err := GoBlessings(env, jutil.WrapObject(jBlessings))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	info := (*(*security.Principal)(jutil.NativePtr(goPtr))).BlessingsInfo(blessings)
	infomap := make(map[jutil.Object]jutil.Object)
	for name, caveats := range info {
		jName := jutil.JString(env, name)
		jCaveatArray, err := JavaCaveats(env, caveats)
		if err != nil {
			jutil.JThrowV(env, err)
			return nil
		}
		infomap[jName] = jCaveatArray
	}
	jInfo, err := jutil.JObjectMap(env, infomap)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jInfo))
}

//export Java_io_v_v23_security_VPrincipalImpl_nativeBlessingStore
func Java_io_v_v23_security_VPrincipalImpl_nativeBlessingStore(jenv *C.JNIEnv, jVPrincipalImpl C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	store := (*(*security.Principal)(jutil.NativePtr(goPtr))).BlessingStore()
	jStore, err := JavaBlessingStore(env, store)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jStore))
}

//export Java_io_v_v23_security_VPrincipalImpl_nativeRoots
func Java_io_v_v23_security_VPrincipalImpl_nativeRoots(jenv *C.JNIEnv, jVPrincipalImpl C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	roots := (*(*security.Principal)(jutil.NativePtr(goPtr))).Roots()
	jRoots, err := JavaBlessingRoots(env, roots)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jRoots))
}

//export Java_io_v_v23_security_VPrincipalImpl_nativeAddToRoots
func Java_io_v_v23_security_VPrincipalImpl_nativeAddToRoots(jenv *C.JNIEnv, jVPrincipalImpl C.jobject, goPtr C.jlong, jBlessings C.jobject) {
	env := jutil.WrapEnv(jenv)
	blessings, err := GoBlessings(env, jutil.WrapObject(jBlessings))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := (*(*security.Principal)(jutil.NativePtr(goPtr))).AddToRoots(blessings); err != nil {
		jutil.JThrowV(env, err)
	}
}

//export Java_io_v_v23_security_VPrincipalImpl_nativeFinalize
func Java_io_v_v23_security_VPrincipalImpl_nativeFinalize(jenv *C.JNIEnv, jVPrincipalImpl C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goPtr))
}

//export Java_io_v_v23_security_Blessings_nativeCreate
func Java_io_v_v23_security_Blessings_nativeCreate(jenv *C.JNIEnv, jBlessingsClass C.jclass, jWire C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	var blessings security.Blessings
	if err := jutil.GoVomCopy(env, jutil.WrapObject(jWire), jWireBlessingsClass, &blessings); err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jBlessings))
}

//export Java_io_v_v23_security_Blessings_nativeCreateUnion
func Java_io_v_v23_security_Blessings_nativeCreateUnion(jenv *C.JNIEnv, jBlessingsClass C.jclass, jBlessingsArr C.jobjectArray) C.jobject {
	env := jutil.WrapEnv(jenv)
	blessingsArr, err := GoBlessingsArray(env, jutil.WrapObject(jBlessingsArr))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	blessings, err := security.UnionOfBlessings(blessingsArr...)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jBlessings))
}

//export Java_io_v_v23_security_Blessings_nativePublicKey
func Java_io_v_v23_security_Blessings_nativePublicKey(jenv *C.JNIEnv, jBlessings C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	key := (*(*security.Blessings)(jutil.NativePtr(goPtr))).PublicKey()
	jPublicKey, err := JavaPublicKey(env, key)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jPublicKey))
}

//export Java_io_v_v23_security_Blessings_nativeSigningBlessings
func Java_io_v_v23_security_Blessings_nativeSigningBlessings(jenv *C.JNIEnv, jBlessings C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	blessings := security.SigningBlessings(*((*security.Blessings)(jutil.NativePtr(goPtr))))
	jSigningBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jSigningBlessings))
}

//export Java_io_v_v23_security_Blessings_nativeFinalize
func Java_io_v_v23_security_Blessings_nativeFinalize(jenv *C.JNIEnv, jBlessings C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goPtr))
}

//export Java_io_v_v23_security_BlessingRootsImpl_nativeAdd
func Java_io_v_v23_security_BlessingRootsImpl_nativeAdd(jenv *C.JNIEnv, jBlessingRootsImpl C.jobject, goPtr C.jlong, jRoot C.jobject, jPattern C.jobject) {
	env := jutil.WrapEnv(jenv)
	root, err := JavaPublicKeyToDER(env, jutil.WrapObject(jRoot))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	pattern, err := GoBlessingPattern(env, jutil.WrapObject(jPattern))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := (*(*security.BlessingRoots)(jutil.NativePtr(goPtr))).Add(root, pattern); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_v_v23_security_BlessingRootsImpl_nativeRecognized
func Java_io_v_v23_security_BlessingRootsImpl_nativeRecognized(jenv *C.JNIEnv, jBlessingRootsImpl C.jobject, goPtr C.jlong, jRoot C.jobject, jBlessing C.jstring) {
	env := jutil.WrapEnv(jenv)
	root, err := JavaPublicKeyToDER(env, jutil.WrapObject(jRoot))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	blessing := jutil.GoString(env, jutil.WrapObject(jBlessing))
	if err := (*(*security.BlessingRoots)(jutil.NativePtr(goPtr))).Recognized(root, blessing); err != nil {
		jutil.JThrowV(env, err)
	}
}

//export Java_io_v_v23_security_BlessingRootsImpl_nativeDebugString
func Java_io_v_v23_security_BlessingRootsImpl_nativeDebugString(jenv *C.JNIEnv, jBlessingRootsImpl C.jobject, goPtr C.jlong) C.jstring {
	env := jutil.WrapEnv(jenv)
	debug := (*(*security.BlessingRoots)(jutil.NativePtr(goPtr))).DebugString()
	jDebug := jutil.JString(env, debug)
	return C.jstring(unsafe.Pointer(jDebug))
}

//export Java_io_v_v23_security_BlessingRootsImpl_nativeToString
func Java_io_v_v23_security_BlessingRootsImpl_nativeToString(jenv *C.JNIEnv, jBlessingRootsImpl C.jobject, goPtr C.jlong) C.jstring {
	env := jutil.WrapEnv(jenv)
	str := fmt.Sprintf("%v", (*(*security.BlessingRoots)(jutil.NativePtr(goPtr))))
	jStr := jutil.JString(env, str)
	return C.jstring(unsafe.Pointer(jStr))
}

//export Java_io_v_v23_security_BlessingRootsImpl_nativeDump
func Java_io_v_v23_security_BlessingRootsImpl_nativeDump(jenv *C.JNIEnv, jBlessingRootsImpl C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	dump := (*(*security.BlessingRoots)(jutil.NativePtr(goPtr))).Dump()
	result := make(map[jutil.Object][]jutil.Object)
	for pattern, keys := range dump {
		jBlessingPattern, err := JavaBlessingPattern(env, pattern)
		if err != nil {
			jutil.JThrowV(env, err)
			return nil
		}
		jPublicKeys := make([]jutil.Object, len(keys))
		for i, key := range keys {
			var err error
			if jPublicKeys[i], err = JavaPublicKey(env, key); err != nil {
				jutil.JThrowV(env, err)
				return nil
			}
		}
		result[jBlessingPattern] = jPublicKeys
	}
	jMap, err := jutil.JObjectMultimap(env, result)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jMap))
}

//export Java_io_v_v23_security_BlessingRootsImpl_nativeFinalize
func Java_io_v_v23_security_BlessingRootsImpl_nativeFinalize(jenv *C.JNIEnv, jBlessingRootsImpl C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goPtr))
}

//export Java_io_v_v23_security_BlessingStoreImpl_nativeSet
func Java_io_v_v23_security_BlessingStoreImpl_nativeSet(jenv *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong, jBlessings C.jobject, jForPeers C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	blessings, err := GoBlessings(env, jutil.WrapObject(jBlessings))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	forPeers, err := GoBlessingPattern(env, jutil.WrapObject(jForPeers))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	oldBlessings, err := (*(*security.BlessingStore)(jutil.NativePtr(goPtr))).Set(blessings, forPeers)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jOldBlessings, err := JavaBlessings(env, oldBlessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jOldBlessings))
}

//export Java_io_v_v23_security_BlessingStoreImpl_nativeForPeer
func Java_io_v_v23_security_BlessingStoreImpl_nativeForPeer(jenv *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong, jPeerBlessings C.jobjectArray) C.jobject {
	env := jutil.WrapEnv(jenv)
	peerBlessings, err := jutil.GoStringArray(env, jutil.WrapObject(jPeerBlessings))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	blessings := (*(*security.BlessingStore)(jutil.NativePtr(goPtr))).ForPeer(peerBlessings...)
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jBlessings))
}

//export Java_io_v_v23_security_BlessingStoreImpl_nativeSetDefaultBlessings
func Java_io_v_v23_security_BlessingStoreImpl_nativeSetDefaultBlessings(jenv *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong, jBlessings C.jobject) {
	env := jutil.WrapEnv(jenv)
	blessings, err := GoBlessings(env, jutil.WrapObject(jBlessings))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := (*(*security.BlessingStore)(jutil.NativePtr(goPtr))).SetDefault(blessings); err != nil {
		jutil.JThrowV(env, err)
	}
}

//export Java_io_v_v23_security_BlessingStoreImpl_nativeDefaultBlessings
func Java_io_v_v23_security_BlessingStoreImpl_nativeDefaultBlessings(jenv *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	blessings := (*(*security.BlessingStore)(jutil.NativePtr(goPtr))).Default()
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jBlessings))
}

//export Java_io_v_v23_security_BlessingStoreImpl_nativePublicKey
func Java_io_v_v23_security_BlessingStoreImpl_nativePublicKey(jenv *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	key := (*(*security.BlessingStore)(jutil.NativePtr(goPtr))).PublicKey()
	jKey, err := JavaPublicKey(env, key)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jKey))
}

//export Java_io_v_v23_security_BlessingStoreImpl_nativePeerBlessings
func Java_io_v_v23_security_BlessingStoreImpl_nativePeerBlessings(jenv *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	blessingsMap := (*(*security.BlessingStore)(jutil.NativePtr(goPtr))).PeerBlessings()
	bmap := make(map[jutil.Object]jutil.Object)
	for pattern, blessings := range blessingsMap {
		jPattern, err := JavaBlessingPattern(env, pattern)
		if err != nil {
			jutil.JThrowV(env, err)
			return nil
		}
		jBlessings, err := JavaBlessings(env, blessings)
		if err != nil {
			jutil.JThrowV(env, err)
			return nil
		}
		bmap[jPattern] = jBlessings
	}
	jBlessingsMap, err := jutil.JObjectMap(env, bmap)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jBlessingsMap))
}

//export Java_io_v_v23_security_BlessingStoreImpl_nativeCacheDischarge
func Java_io_v_v23_security_BlessingStoreImpl_nativeCacheDischarge(jenv *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong, jDischarge C.jobject, jCaveat C.jobject, jImpetus C.jobject) {
	env := jutil.WrapEnv(jenv)
	blessingStore := *(*security.BlessingStore)(jutil.NativePtr(goPtr))
	discharge, err := GoDischarge(env, jutil.WrapObject(jDischarge))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	caveat, err := GoCaveat(env, jutil.WrapObject(jCaveat))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	var impetus security.DischargeImpetus
	err = jutil.GoVomCopy(env, jutil.WrapObject(jImpetus), jDischargeImpetusClass, &impetus)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	blessingStore.CacheDischarge(discharge, caveat, impetus)
}

//export Java_io_v_v23_security_BlessingStoreImpl_nativeClearDischarges
func Java_io_v_v23_security_BlessingStoreImpl_nativeClearDischarges(jenv *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong, jDischarges C.jobject) {
	env := jutil.WrapEnv(jenv)
	blessingStore := *(*security.BlessingStore)(jutil.NativePtr(goPtr))
	arr, err := jutil.GoObjectArray(env, jutil.WrapObject(jDischarges))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	var discharges []security.Discharge
	for _, jDischarge := range arr {
		discharge, err := GoDischarge(env, jDischarge)
		if err != nil {
			jutil.JThrowV(env, err)
			return
		}
		discharges = append(discharges, discharge)
	}
	blessingStore.ClearDischarges(discharges...)
}

//export Java_io_v_v23_security_BlessingStoreImpl_nativeDischarge
func Java_io_v_v23_security_BlessingStoreImpl_nativeDischarge(jenv *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong, jCaveat C.jobject, jImpetus C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	blessingStore := *(*security.BlessingStore)(jutil.NativePtr(goPtr))
	caveat, err := GoCaveat(env, jutil.WrapObject(jCaveat))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	var impetus security.DischargeImpetus
	err = jutil.GoVomCopy(env, jutil.WrapObject(jImpetus), jDischargeImpetusClass, &impetus)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	discharge := blessingStore.Discharge(caveat, impetus)
	jDischarge, err := JavaDischarge(env, discharge)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jDischarge))
}

//export Java_io_v_v23_security_BlessingStoreImpl_nativeDebugString
func Java_io_v_v23_security_BlessingStoreImpl_nativeDebugString(jenv *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong) C.jstring {
	env := jutil.WrapEnv(jenv)
	debug := (*(*security.BlessingStore)(jutil.NativePtr(goPtr))).DebugString()
	jDebug := jutil.JString(env, debug)
	return C.jstring(unsafe.Pointer(jDebug))
}

//export Java_io_v_v23_security_BlessingStoreImpl_nativeToString
func Java_io_v_v23_security_BlessingStoreImpl_nativeToString(jenv *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong) C.jstring {
	env := jutil.WrapEnv(jenv)
	str := fmt.Sprintf("%s", (*(*security.BlessingStore)(jutil.NativePtr(goPtr))))
	jStr := jutil.JString(env, str)
	return C.jstring(unsafe.Pointer(jStr))
}

//export Java_io_v_v23_security_BlessingStoreImpl_nativeFinalize
func Java_io_v_v23_security_BlessingStoreImpl_nativeFinalize(jenv *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goPtr))
}

//export Java_io_v_v23_security_BlessingPattern_nativeCreate
func Java_io_v_v23_security_BlessingPattern_nativeCreate(jenv *C.JNIEnv, jBlessingPattern C.jobject) C.jlong {
	env := jutil.WrapEnv(jenv)
	pattern, err := GoBlessingPattern(env, jutil.WrapObject(jBlessingPattern))
	if err != nil {
		jutil.JThrowV(env, err)
		return C.jlong(0)
	}
	jutil.GoRef(&pattern) // Un-refed when the BlessingPattern object is finalized.
	return C.jlong(jutil.PtrValue(&pattern))
}

//export Java_io_v_v23_security_BlessingPattern_nativeIsMatchedBy
func Java_io_v_v23_security_BlessingPattern_nativeIsMatchedBy(jenv *C.JNIEnv, jBlessingPattern C.jobject, goPtr C.jlong, jBlessings C.jobjectArray) C.jboolean {
	env := jutil.WrapEnv(jenv)
	blessings, err := jutil.GoStringArray(env, jutil.WrapObject(jBlessings))
	if err != nil {
		jutil.JThrowV(env, err)
		return C.JNI_FALSE
	}

	matched := (*(*security.BlessingPattern)(jutil.NativePtr(goPtr))).MatchedBy(blessings...)
	if matched {
		return C.JNI_TRUE
	}
	return C.JNI_FALSE
}

//export Java_io_v_v23_security_BlessingPattern_nativeIsValid
func Java_io_v_v23_security_BlessingPattern_nativeIsValid(jenv *C.JNIEnv, jBlessingPattern C.jobject, goPtr C.jlong) C.jboolean {
	valid := (*(*security.BlessingPattern)(jutil.NativePtr(goPtr))).IsValid()
	if valid {
		return C.JNI_TRUE
	}
	return C.JNI_FALSE
}

//export Java_io_v_v23_security_BlessingPattern_nativeMakeNonExtendable
func Java_io_v_v23_security_BlessingPattern_nativeMakeNonExtendable(jenv *C.JNIEnv, jBlessingPattern C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	pattern := (*(*security.BlessingPattern)(jutil.NativePtr(goPtr))).MakeNonExtendable()
	jPattern, err := JavaBlessingPattern(env, pattern)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jPattern))
}

//export Java_io_v_v23_security_BlessingPattern_nativeFinalize
func Java_io_v_v23_security_BlessingPattern_nativeFinalize(jenv *C.JNIEnv, jBlessingPattern C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goPtr))
}

//export Java_io_v_v23_security_PublicKeyThirdPartyCaveatValidator_nativeValidate
func Java_io_v_v23_security_PublicKeyThirdPartyCaveatValidator_nativeValidate(jenv *C.JNIEnv, jThirdPartyValidatorClass C.jclass, jContext C.jobject, jCall C.jobject, jCaveatParam C.jobject) {
	env := jutil.WrapEnv(jenv)
	param, err := jutil.GoVomCopyValue(env, jutil.WrapObject(jCaveatParam))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	ctx, err := jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	call, err := GoCall(env, jutil.WrapObject(jCall))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	caveat, err := security.NewCaveat(security.PublicKeyThirdPartyCaveat, param)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := caveat.Validate(ctx, call); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_v_v23_security_VSecurity_nativeGetRemoteBlessingNames
func Java_io_v_v23_security_VSecurity_nativeGetRemoteBlessingNames(jenv *C.JNIEnv, jVSecurityClass C.jclass, jCtx C.jobject, jCall C.jobject) C.jobjectArray {
	env := jutil.WrapEnv(jenv)
	ctx, err := jcontext.GoContext(env, jutil.WrapObject(jCtx))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	call, err := GoCall(env, jutil.WrapObject(jCall))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	blessingStrs, _ := security.RemoteBlessingNames(ctx, call)
	jArr, err := jutil.JStringArray(env, blessingStrs)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobjectArray(unsafe.Pointer(jArr))
}

//export Java_io_v_v23_security_VSecurity_nativeGetSigningBlessingNames
func Java_io_v_v23_security_VSecurity_nativeGetSigningBlessingNames(jenv *C.JNIEnv, jVSecurityClass C.jclass, jCtx C.jobject, jPrincipal C.jobject, jBlessings C.jobject) C.jobjectArray {
	env := jutil.WrapEnv(jenv)
	ctx, err := jcontext.GoContext(env, jutil.WrapObject(jCtx))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	principal, err := GoPrincipal(env, jutil.WrapObject(jPrincipal))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	blessings, err := GoBlessings(env, jutil.WrapObject(jBlessings))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	blessingStrs, _ := security.SigningBlessingNames(ctx, principal, blessings)
	jArr, err := jutil.JStringArray(env, blessingStrs)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobjectArray(unsafe.Pointer(jArr))
}

//export Java_io_v_v23_security_VSecurity_nativeGetLocalBlessingNames
func Java_io_v_v23_security_VSecurity_nativeGetLocalBlessingNames(jenv *C.JNIEnv, jVSecurityClass C.jclass, jCtx C.jobject, jCall C.jobject) C.jobjectArray {
	env := jutil.WrapEnv(jenv)
	ctx, err := jcontext.GoContext(env, jutil.WrapObject(jCtx))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	call, err := GoCall(env, jutil.WrapObject(jCall))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	blessingStrs := security.LocalBlessingNames(ctx, call)
	jArr, err := jutil.JStringArray(env, blessingStrs)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobjectArray(unsafe.Pointer(jArr))
}
