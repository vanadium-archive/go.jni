// +build android

package security

import (
	"unsafe"

	"veyron.io/jni/util"
	vsecurity "veyron.io/veyron/veyron/security"
	"veyron.io/veyron/veyron2/security"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

var (
	blessingsSign       = util.ClassSign("io.veyron.veyron.veyron2.security.Blessings")
	wireBlessingsSign   = util.ClassSign("io.veyron.veyron.veyron2.security.WireBlessings")
	blessingStoreSign   = util.ClassSign("io.veyron.veyron.veyron2.security.BlessingStore")
	blessingRootsSign   = util.ClassSign("io.veyron.veyron.veyron2.security.BlessingRoots")
	blessingPatternSign = util.ClassSign("io.veyron.veyron.veyron2.security.BlessingPattern")
	signerSign          = util.ClassSign("io.veyron.veyron.veyron2.security.Signer")
	caveatSign          = util.ClassSign("io.veyron.veyron.veyron2.security.Caveat")
	signatureSign       = util.ClassSign("io.veyron.veyron.veyron2.security.Signature")
	publicKeySign       = util.ClassSign("java.security.interfaces.ECPublicKey")

	// Global reference for io.veyron.veyron.veyron2.security.PrincipalImpl
	jPrincipalImplClass C.jclass
	// Global reference for io.veyron.veyron.veyron2.security.BlessingStoreImpl
	jBlessingStoreImplClass C.jclass
	// Global reference for io.veyron.veyron.veyron2.security.BlessingRootsImpl
	jBlessingRootsImplClass C.jclass
	// Global reference for io.veyron.veyron.veyron2.security.Util class.
	jUtilClass C.jclass
)

// Init initializes the JNI code with the given Java evironment. This method
// must be called from the main Java thread.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java environment is passed in an empty
// interface and then cast into the package-local environment type.
func Init(jEnv interface{}) {
	env := (*C.JNIEnv)(unsafe.Pointer(util.PtrValue(jEnv)))
	// Cache global references to all Java classes used by the package.  This is
	// necessary because JNI gets access to the class loader only in the system
	// thread, so we aren't able to invoke FindClass in other threads.
	jPrincipalImplClass = C.jclass(util.JFindClassOrPrint(env, "io/veyron/veyron/veyron2/security/PrincipalImpl"))
	jBlessingStoreImplClass = C.jclass(util.JFindClassOrPrint(env, "io/veyron/veyron/veyron2/security/BlessingStoreImpl"))
	jBlessingRootsImplClass = C.jclass(util.JFindClassOrPrint(env, "io/veyron/veyron/veyron2/security/BlessingRootsImpl"))
	jUtilClass = C.jclass(util.JFindClassOrPrint(env, "io/veyron/veyron/veyron2/security/Util"))
}

//export Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeCreate
func Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeCreate(env *C.JNIEnv, jPrincipalImplClass C.jclass) C.jobject {
	principal, err := vsecurity.NewPrincipal()
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	jPrincipal, err := JavaPrincipal(env, principal)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	return jPrincipal
}

//export Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeCreateForSigner
func Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeCreateForSigner(env *C.JNIEnv, jPrincipalImplClass C.jclass, jSigner C.jobject) C.jobject {
	signer, err := GoSigner(env, jSigner)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	principal, err := vsecurity.NewPrincipalFromSigner(signer)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	util.GoRef(&principal) // Un-refed when the Java PrincipalImpl is finalized.
	jPrincipal, err := util.NewObject(env, jPrincipalImplClass, []util.Sign{util.LongSign, signerSign, blessingStoreSign, blessingRootsSign}, &principal, jSigner, C.jobject(nil), C.jobject(nil))
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	return C.jobject(jPrincipal)
}

//export Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeCreateForAll
func Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeCreateForAll(env *C.JNIEnv, jPrincipalImplClass C.jclass, jSigner C.jobject, jStore C.jobject, jRoots C.jobject) C.jobject {
	signer, err := GoSigner(env, jSigner)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	store, err := GoBlessingStore(env, jStore)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	roots, err := GoBlessingRoots(env, jRoots)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	principal, err := security.CreatePrincipal(signer, store, roots)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	util.GoRef(&principal) // Un-refed when the Java PrincipalImpl is finalized.
	jPrincipal, err := util.NewObject(env, jPrincipalImplClass, []util.Sign{util.LongSign, signerSign, blessingStoreSign, blessingRootsSign}, &principal, jSigner, jStore, jRoots)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	return C.jobject(jPrincipal)
}

//export Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeCreatePersistent
func Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeCreatePersistent(env *C.JNIEnv, jPrincipalImplClass C.jclass, jPassphrase C.jstring, jDir C.jstring) C.jobject {
	passphrase := util.GoString(env, jPassphrase)
	dir := util.GoString(env, jDir)
	principal, err := vsecurity.LoadPersistentPrincipal(dir, []byte(passphrase))
	if err != nil {
		if principal, err = vsecurity.CreatePersistentPrincipal(dir, []byte(passphrase)); err != nil {
			util.JThrowV(env, err)
			return C.jobject(nil)
		}
	}
	jPrincipal, err := JavaPrincipal(env, principal)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	return C.jobject(jPrincipal)
}

//export Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeCreatePersistentForSigner
func Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeCreatePersistentForSigner(env *C.JNIEnv, jPrincipalImplClass C.jclass, jSigner C.jobject, jDir C.jstring) C.jobject {
	signer, err := GoSigner(env, jSigner)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	dir := util.GoString(env, jDir)
	principal, err := vsecurity.NewPersistentPrincipalFromSigner(signer, dir)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	util.GoRef(&principal) // Un-refed when the Java PrincipalImpl is finalized.
	jPrincipal, err := util.NewObject(env, jPrincipalImplClass, []util.Sign{util.LongSign, signerSign, blessingStoreSign, blessingRootsSign}, &principal, jSigner, C.jobject(nil), C.jobject(nil))
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	return C.jobject(jPrincipal)
}

//export Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeBless
func Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeBless(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong, jKey C.jobject, jWith C.jobject, jExtension C.jstring, jCaveat C.jobject, jAdditionalCaveats C.jobjectArray) C.jobject {
	key, err := GoPublicKey(env, jKey)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	with, err := GoBlessings(env, jWith)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	extension := util.GoString(env, jExtension)
	caveat, err := GoCaveat(env, jCaveat)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	additionalCaveats, err := GoCaveats(env, jAdditionalCaveats)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	blessings, err := (*(*security.Principal)(util.Ptr(goPtr))).Bless(key, with, extension, caveat, additionalCaveats...)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	return jBlessings
}

//export Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeBlessSelf
func Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeBlessSelf(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong, jName C.jstring, jCaveats C.jobjectArray) C.jobject {
	name := util.GoString(env, jName)
	caveats, err := GoCaveats(env, jCaveats)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	blessings, err := (*(*security.Principal)(util.Ptr(goPtr))).BlessSelf(name, caveats...)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	return jBlessings
}

//export Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeSign
func Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeSign(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong, jMessage C.jbyteArray) C.jobject {
	message := util.GoByteArray(env, jMessage)
	sig, err := (*(*security.Principal)(util.Ptr(goPtr))).Sign(message)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	jSig, err := JavaSignature(env, sig)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	return jSig
}

//export Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativePublicKey
func Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativePublicKey(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong) C.jobject {
	key := (*(*security.Principal)(util.Ptr(goPtr))).PublicKey()
	jKey, err := JavaPublicKey(env, key)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	return jKey
}

//export Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeBlessingStore
func Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeBlessingStore(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong) C.jobject {
	store := (*(*security.Principal)(util.Ptr(goPtr))).BlessingStore()
	jStore, err := JavaBlessingStore(env, store)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	return jStore
}

//export Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeRoots
func Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeRoots(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong) C.jobject {
	roots := (*(*security.Principal)(util.Ptr(goPtr))).Roots()
	jRoots, err := JavaBlessingRoots(env, roots)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	return jRoots
}

//export Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeAddToRoots
func Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeAddToRoots(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong, jBlessings C.jobject) {
	blessings, err := GoBlessings(env, jBlessings)
	if err != nil {
		util.JThrowV(env, err)
		return
	}
	if err := (*(*security.Principal)(util.Ptr(goPtr))).AddToRoots(blessings); err != nil {
		util.JThrowV(env, err)
	}
}

//export Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeFinalize
func Java_io_veyron_veyron_veyron2_security_PrincipalImpl_nativeFinalize(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong) {
	util.GoUnref((*security.Principal)(util.Ptr(goPtr)))
}

//export Java_io_veyron_veyron_veyron2_security_BlessingsImpl_nativeCreate
func Java_io_veyron_veyron_veyron2_security_BlessingsImpl_nativeCreate(env *C.JNIEnv, jBlessingsImplClass C.jclass, jWire C.jobject) C.jlong {
	wire, err := GoWireBlessings(env, jWire)
	if err != nil {
		util.JThrowV(env, err)
		return C.jlong(0)
	}
	blessings, err := security.NewBlessings(wire)
	if err != nil {
		util.JThrowV(env, err)
		return C.jlong(0)
	}
	util.GoRef(&blessings) // Un-refed when the Java BlessingsImpl is finalized.
	return C.jlong(util.PtrValue(&blessings))
}

//export Java_io_veyron_veyron_veyron2_security_BlessingsImpl_nativeForContext
func Java_io_veyron_veyron_veyron2_security_BlessingsImpl_nativeForContext(env *C.JNIEnv, jBlessingsImpl C.jobject, goPtr C.jlong, jContext C.jobject) C.jobjectArray {
	context, err := GoContext(env, jContext)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	blessingStrs := (*(*security.Blessings)(util.Ptr(goPtr))).ForContext(context)
	return C.jobjectArray(util.JStringArray(env, blessingStrs))
}

//export Java_io_veyron_veyron_veyron2_security_BlessingsImpl_nativePublicKey
func Java_io_veyron_veyron_veyron2_security_BlessingsImpl_nativePublicKey(env *C.JNIEnv, jBlessingsImpl C.jobject, goPtr C.jlong) C.jobject {
	key := (*(*security.Blessings)(util.Ptr(goPtr))).PublicKey()
	jPublicKey, err := JavaPublicKey(env, key)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	return jPublicKey
}

//export Java_io_veyron_veyron_veyron2_security_BlessingsImpl_nativeFinalize
func Java_io_veyron_veyron_veyron2_security_BlessingsImpl_nativeFinalize(env *C.JNIEnv, jBlessingsImpl C.jobject, goPtr C.jlong) {
	util.GoUnref((*security.Blessings)(util.Ptr(goPtr)))
}

//export Java_io_veyron_veyron_veyron2_security_BlessingRootsImpl_nativeAdd
func Java_io_veyron_veyron_veyron2_security_BlessingRootsImpl_nativeAdd(env *C.JNIEnv, jBlessingRootsImpl C.jobject, goPtr C.jlong, jRoot C.jobject, jPattern C.jobject) {
	root, err := GoPublicKey(env, jRoot)
	if err != nil {
		util.JThrowV(env, err)
		return
	}
	pattern, err := GoBlessingPattern(env, jPattern)
	if err != nil {
		util.JThrowV(env, err)
		return
	}
	if err := (*(*security.BlessingRoots)(util.Ptr(goPtr))).Add(root, pattern); err != nil {
		util.JThrowV(env, err)
		return
	}
}

//export Java_io_veyron_veyron_veyron2_security_BlessingRootsImpl_nativeRecognized
func Java_io_veyron_veyron_veyron2_security_BlessingRootsImpl_nativeRecognized(env *C.JNIEnv, jBlessingRootsImpl C.jobject, goPtr C.jlong, jRoot C.jobject, jBlessing C.jstring) {
	root, err := GoPublicKey(env, jRoot)
	if err != nil {
		util.JThrowV(env, err)
		return
	}
	blessing := util.GoString(env, jBlessing)
	if err := (*(*security.BlessingRoots)(util.Ptr(goPtr))).Recognized(root, blessing); err != nil {
		util.JThrowV(env, err)
	}
}

//export Java_io_veyron_veyron_veyron2_security_BlessingRootsImpl_nativeDebugString
func Java_io_veyron_veyron_veyron2_security_BlessingRootsImpl_nativeDebugString(env *C.JNIEnv, jBlessingRootsImpl C.jobject, goPtr C.jlong) C.jstring {
	debug := (*(*security.BlessingRoots)(util.Ptr(goPtr))).DebugString()
	return C.jstring(util.JString(env, debug))
}

//export Java_io_veyron_veyron_veyron2_security_BlessingRootsImpl_nativeFinalize
func Java_io_veyron_veyron_veyron2_security_BlessingRootsImpl_nativeFinalize(env *C.JNIEnv, jBlessingRootsImpl C.jobject, goPtr C.jlong) {
	util.GoUnref((*security.BlessingRoots)(util.Ptr(goPtr)))
}

//export Java_io_veyron_veyron_veyron2_security_BlessingStoreImpl_nativeSet
func Java_io_veyron_veyron_veyron2_security_BlessingStoreImpl_nativeSet(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong, jBlessings C.jobject, jForPeers C.jobject) C.jobject {
	blessings, err := GoBlessings(env, jBlessings)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	forPeers, err := GoBlessingPattern(env, jForPeers)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	oldBlessings, err := (*(*security.BlessingStore)(util.Ptr(goPtr))).Set(blessings, forPeers)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	jOldBlessings, err := JavaBlessings(env, oldBlessings)
	return C.jobject(nil)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	return jOldBlessings
}

//export Java_io_veyron_veyron_veyron2_security_BlessingStoreImpl_nativeForPeer
func Java_io_veyron_veyron_veyron2_security_BlessingStoreImpl_nativeForPeer(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong, jPeerBlessings C.jobjectArray) C.jobject {
	peerBlessings := util.GoStringArray(env, jPeerBlessings)
	blessings := (*(*security.BlessingStore)(util.Ptr(goPtr))).ForPeer(peerBlessings...)
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	return jBlessings
}

//export Java_io_veyron_veyron_veyron2_security_BlessingStoreImpl_nativeSetDefaultBlessings
func Java_io_veyron_veyron_veyron2_security_BlessingStoreImpl_nativeSetDefaultBlessings(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong, jBlessings C.jobject) {
	blessings, err := GoBlessings(env, jBlessings)
	if err != nil {
		util.JThrowV(env, err)
		return
	}
	if err := (*(*security.BlessingStore)(util.Ptr(goPtr))).SetDefault(blessings); err != nil {
		util.JThrowV(env, err)
	}
}

//export Java_io_veyron_veyron_veyron2_security_BlessingStoreImpl_nativeDefaultBlessings
func Java_io_veyron_veyron_veyron2_security_BlessingStoreImpl_nativeDefaultBlessings(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong) C.jobject {
	blessings := (*(*security.BlessingStore)(util.Ptr(goPtr))).Default()
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	return jBlessings
}

//export Java_io_veyron_veyron_veyron2_security_BlessingStoreImpl_nativePublicKey
func Java_io_veyron_veyron_veyron2_security_BlessingStoreImpl_nativePublicKey(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong) C.jobject {
	key := (*(*security.BlessingStore)(util.Ptr(goPtr))).PublicKey()
	jKey, err := JavaPublicKey(env, key)
	if err != nil {
		util.JThrowV(env, err)
		return C.jobject(nil)
	}
	return jKey
}

//export Java_io_veyron_veyron_veyron2_security_BlessingStoreImpl_nativeDebugString
func Java_io_veyron_veyron_veyron2_security_BlessingStoreImpl_nativeDebugString(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong) C.jstring {
	debug := (*(*security.BlessingStore)(util.Ptr(goPtr))).DebugString()
	return C.jstring(util.JString(env, debug))
}

//export Java_io_veyron_veyron_veyron2_security_BlessingStoreImpl_nativeFinalize
func Java_io_veyron_veyron_veyron2_security_BlessingStoreImpl_nativeFinalize(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong) {
	util.GoUnref((*security.BlessingStore)(util.Ptr(goPtr)))
}
