// +build android

package security

import (
	"fmt"
	"unsafe"

	vsecurity "v.io/core/veyron/security"
	"v.io/core/veyron2/security"
	jutil "v.io/jni/util"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

var (
	principalSign       = jutil.ClassSign("io.v.core.veyron2.security.Principal")
	blessingsSign       = jutil.ClassSign("io.v.core.veyron2.security.Blessings")
	wireBlessingsSign   = jutil.ClassSign("io.v.core.veyron2.security.WireBlessings")
	blessingStoreSign   = jutil.ClassSign("io.v.core.veyron2.security.BlessingStore")
	blessingRootsSign   = jutil.ClassSign("io.v.core.veyron2.security.BlessingRoots")
	blessingPatternSign = jutil.ClassSign("io.v.core.veyron2.security.BlessingPattern")
	signerSign          = jutil.ClassSign("io.v.core.veyron2.security.Signer")
	caveatSign          = jutil.ClassSign("io.v.core.veyron2.security.Caveat")
	caveatValidatorSign = jutil.ClassSign("io.v.core.veyron2.security.CaveatValidator")
	contextSign         = jutil.ClassSign("io.v.core.veyron2.security.VContext")
	signatureSign       = jutil.ClassSign("io.v.core.veyron2.security.Signature")
	publicKeySign       = jutil.ClassSign("java.security.interfaces.ECPublicKey")

	// Global reference for io.v.core.veyron2.security.Blessings class.
	jBlessingsClass C.jclass
	// Global reference for io.v.core.veyron2.security.Caveat class.
	jCaveatClass C.jclass
	// Global reference for io.v.core.veyron2.security.PrincipalImpl class.
	jPrincipalImplClass C.jclass
	// Global reference for io.v.core.veyron2.security.BlessingsImpl class.
	jBlessingsImplClass C.jclass
	// Global reference for io.v.core.veyron2.security.BlessingStoreImpl class.
	jBlessingStoreImplClass C.jclass
	// Global reference for io.v.core.veyron2.security.BlessingRootsImpl class.
	jBlessingRootsImplClass C.jclass
	// Global reference for io.v.core.veyron2.security.VContextImpl class.
	jContextImplClass C.jclass
	// Global reference for io.v.core.veyron2.security.BlessingPatternWrapper class.
	jBlessingPatternWrapperClass C.jclass
	// Global reference for io.v.core.veyron2.security.CaveatCoder class.
	jCaveatCoderClass C.jclass
	// Global reference for io.v.core.veyron2.security.Util class.
	jUtilClass C.jclass
	// Global reference for java.lang.Object class.
	jObjectClass C.jclass

	javaVM *C.JavaVM
)

// Init initializes the JNI code with the given Java evironment. This method
// must be called from the main Java thread.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java environment is passed in an empty
// interface and then cast into the package-local environment type.
func Init(jEnv interface{}) error {
	env := (*C.JNIEnv)(unsafe.Pointer(jutil.PtrValue(jEnv)))
	// We cannot cache Java environments as they are only valid in the current
	// thread.  We can, however, cache the Java VM and obtain an environment
	// from it in whatever thread happens to be running at the time.
	if status := C.GetJavaVM(env, &javaVM); status != 0 {
		return fmt.Errorf("couldn't get Java VM from the (Java) environment")
	}
	security.RegisterCaveatDecoder(caveatDecoder)

	// Cache global references to all Java classes used by the package.  This is
	// necessary because JNI gets access to the class loader only in the system
	// thread, so we aren't able to invoke FindClass in other threads.
	jBlessingsClass = C.jclass(jutil.JFindClassOrPrint(env, "io/v/core/veyron2/security/Blessings"))
	jCaveatClass = C.jclass(jutil.JFindClassOrPrint(env, "io/v/core/veyron2/security/Caveat"))
	jPrincipalImplClass = C.jclass(jutil.JFindClassOrPrint(env, "io/v/core/veyron2/security/PrincipalImpl"))
	jBlessingsImplClass = C.jclass(jutil.JFindClassOrPrint(env, "io/v/core/veyron2/security/BlessingsImpl"))
	jBlessingStoreImplClass = C.jclass(jutil.JFindClassOrPrint(env, "io/v/core/veyron2/security/BlessingStoreImpl"))
	jBlessingRootsImplClass = C.jclass(jutil.JFindClassOrPrint(env, "io/v/core/veyron2/security/BlessingRootsImpl"))
	jContextImplClass = C.jclass(jutil.JFindClassOrPrint(env, "io/v/core/veyron2/security/VContextImpl"))
	jBlessingPatternWrapperClass = C.jclass(jutil.JFindClassOrPrint(env, "io/v/core/veyron2/security/BlessingPatternWrapper"))
	jCaveatCoderClass = C.jclass(jutil.JFindClassOrPrint(env, "io/v/core/veyron2/security/CaveatCoder"))
	jUtilClass = C.jclass(jutil.JFindClassOrPrint(env, "io/v/core/veyron2/security/Util"))
	jObjectClass = C.jclass(jutil.JFindClassOrPrint(env, "java/lang/Object"))
	return nil
}

//export Java_io_v_core_veyron2_security_VContextImpl_nativeTimestamp
func Java_io_v_core_veyron2_security_VContextImpl_nativeTimestamp(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jobject {
	t := (*(*security.Context)(jutil.Ptr(goContextPtr))).Timestamp()
	jTime, err := jutil.JTime(env, t)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jTime)
}

//export Java_io_v_core_veyron2_security_VContextImpl_nativeMethod
func Java_io_v_core_veyron2_security_VContextImpl_nativeMethod(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jstring {
	return C.jstring(jutil.JString(env, (*(*security.Context)(jutil.Ptr(goContextPtr))).Method()))
}

//export Java_io_v_core_veyron2_security_VContextImpl_nativeMethodTags
func Java_io_v_core_veyron2_security_VContextImpl_nativeMethodTags(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jobjectArray {
	tags := (*(*security.Context)(jutil.Ptr(goContextPtr))).MethodTags()
	jTags, err := JavaTags(env, tags)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return jTags
}

//export Java_io_v_core_veyron2_security_VContextImpl_nativeName
func Java_io_v_core_veyron2_security_VContextImpl_nativeName(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jstring {
	return C.jstring(jutil.JString(env, (*(*security.Context)(jutil.Ptr(goContextPtr))).Name()))
}

//export Java_io_v_core_veyron2_security_VContextImpl_nativeSuffix
func Java_io_v_core_veyron2_security_VContextImpl_nativeSuffix(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jstring {
	return C.jstring(jutil.JString(env, (*(*security.Context)(jutil.Ptr(goContextPtr))).Suffix()))
}

//export Java_io_v_core_veyron2_security_VContextImpl_nativeLocalEndpoint
func Java_io_v_core_veyron2_security_VContextImpl_nativeLocalEndpoint(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jstring {
	return C.jstring(jutil.JString(env, (*(*security.Context)(jutil.Ptr(goContextPtr))).LocalEndpoint().String()))
}

//export Java_io_v_core_veyron2_security_VContextImpl_nativeRemoteEndpoint
func Java_io_v_core_veyron2_security_VContextImpl_nativeRemoteEndpoint(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jstring {
	return C.jstring(jutil.JString(env, (*(*security.Context)(jutil.Ptr(goContextPtr))).RemoteEndpoint().String()))
}

//export Java_io_v_core_veyron2_security_VContextImpl_nativeLocalPrincipal
func Java_io_v_core_veyron2_security_VContextImpl_nativeLocalPrincipal(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jobject {
	principal := (*(*security.Context)(jutil.Ptr(goContextPtr))).LocalPrincipal()
	jPrincipal, err := JavaPrincipal(env, principal)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jPrincipal)
}

//export Java_io_v_core_veyron2_security_VContextImpl_nativeLocalBlessings
func Java_io_v_core_veyron2_security_VContextImpl_nativeLocalBlessings(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jobject {
	blessings := (*(*security.Context)(jutil.Ptr(goContextPtr))).LocalBlessings()
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jBlessings)
}

//export Java_io_v_core_veyron2_security_VContextImpl_nativeRemoteBlessings
func Java_io_v_core_veyron2_security_VContextImpl_nativeRemoteBlessings(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jobject {
	blessings := (*(*security.Context)(jutil.Ptr(goContextPtr))).RemoteBlessings()
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jBlessings)
}

//export Java_io_v_core_veyron2_security_VContextImpl_nativeFinalize
func Java_io_v_core_veyron2_security_VContextImpl_nativeFinalize(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) {
	jutil.GoUnref((*security.Context)(jutil.Ptr(goContextPtr)))
}

//export Java_io_v_core_veyron2_security_PrincipalImpl_nativeCreate
func Java_io_v_core_veyron2_security_PrincipalImpl_nativeCreate(env *C.JNIEnv, jPrincipalImplClass C.jclass) C.jobject {
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
	return jPrincipal
}

//export Java_io_v_core_veyron2_security_PrincipalImpl_nativeCreateForSigner
func Java_io_v_core_veyron2_security_PrincipalImpl_nativeCreateForSigner(env *C.JNIEnv, jPrincipalImplClass C.jclass, jSigner C.jobject) C.jobject {
	signer, err := GoSigner(env, jSigner)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	principal, err := vsecurity.NewPrincipalFromSigner(signer, nil)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jPrincipal, err := jutil.NewObject(env, jPrincipalImplClass, []jutil.Sign{jutil.LongSign, signerSign, blessingStoreSign, blessingRootsSign}, &principal, jSigner, C.jobject(nil), C.jobject(nil))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jutil.GoRef(&principal) // Un-refed when the Java PrincipalImpl is finalized.
	return C.jobject(jPrincipal)
}

//export Java_io_v_core_veyron2_security_PrincipalImpl_nativeCreateForAll
func Java_io_v_core_veyron2_security_PrincipalImpl_nativeCreateForAll(env *C.JNIEnv, jPrincipalImplClass C.jclass, jSigner C.jobject, jStore C.jobject, jRoots C.jobject) C.jobject {
	signer, err := GoSigner(env, jSigner)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	store, err := GoBlessingStore(env, jStore)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	roots, err := GoBlessingRoots(env, jRoots)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	principal, err := security.CreatePrincipal(signer, store, roots)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jPrincipal, err := jutil.NewObject(env, jPrincipalImplClass, []jutil.Sign{jutil.LongSign, signerSign, blessingStoreSign, blessingRootsSign}, &principal, jSigner, jStore, jRoots)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jutil.GoRef(&principal) // Un-refed when the Java PrincipalImpl is finalized.
	return C.jobject(jPrincipal)
}

//export Java_io_v_core_veyron2_security_PrincipalImpl_nativeCreatePersistent
func Java_io_v_core_veyron2_security_PrincipalImpl_nativeCreatePersistent(env *C.JNIEnv, jPrincipalImplClass C.jclass, jPassphrase C.jstring, jDir C.jstring) C.jobject {
	passphrase := jutil.GoString(env, jPassphrase)
	dir := jutil.GoString(env, jDir)
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
	return C.jobject(jPrincipal)
}

//export Java_io_v_core_veyron2_security_PrincipalImpl_nativeCreatePersistentForSigner
func Java_io_v_core_veyron2_security_PrincipalImpl_nativeCreatePersistentForSigner(env *C.JNIEnv, jPrincipalImplClass C.jclass, jSigner C.jobject, jDir C.jstring) C.jobject {
	signer, err := GoSigner(env, jSigner)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	dir := jutil.GoString(env, jDir)
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
	jPrincipal, err := jutil.NewObject(env, jPrincipalImplClass, []jutil.Sign{jutil.LongSign, signerSign, blessingStoreSign, blessingRootsSign}, &principal, jSigner, C.jobject(nil), C.jobject(nil))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jutil.GoRef(&principal) // Un-refed when the Java PrincipalImpl is finalized.
	return C.jobject(jPrincipal)
}

//export Java_io_v_core_veyron2_security_PrincipalImpl_nativeBless
func Java_io_v_core_veyron2_security_PrincipalImpl_nativeBless(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong, jKey C.jobject, jWith C.jobject, jExtension C.jstring, jCaveat C.jobject, jAdditionalCaveats C.jobjectArray) C.jobject {
	key, err := GoPublicKey(env, jKey)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	with, err := GoBlessings(env, jWith)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	extension := jutil.GoString(env, jExtension)
	caveat, err := GoCaveat(env, jCaveat)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	additionalCaveats, err := GoCaveats(env, jAdditionalCaveats)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	blessings, err := (*(*security.Principal)(jutil.Ptr(goPtr))).Bless(key, with, extension, caveat, additionalCaveats...)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return jBlessings
}

//export Java_io_v_core_veyron2_security_PrincipalImpl_nativeBlessSelf
func Java_io_v_core_veyron2_security_PrincipalImpl_nativeBlessSelf(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong, jName C.jstring, jCaveats C.jobjectArray) C.jobject {
	name := jutil.GoString(env, jName)
	caveats, err := GoCaveats(env, jCaveats)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	blessings, err := (*(*security.Principal)(jutil.Ptr(goPtr))).BlessSelf(name, caveats...)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return jBlessings
}

//export Java_io_v_core_veyron2_security_PrincipalImpl_nativeSign
func Java_io_v_core_veyron2_security_PrincipalImpl_nativeSign(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong, jMessage C.jbyteArray) C.jobject {
	message := jutil.GoByteArray(env, jMessage)
	sig, err := (*(*security.Principal)(jutil.Ptr(goPtr))).Sign(message)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jSig, err := JavaSignature(env, sig)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return jSig
}

//export Java_io_v_core_veyron2_security_PrincipalImpl_nativePublicKey
func Java_io_v_core_veyron2_security_PrincipalImpl_nativePublicKey(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong) C.jobject {
	key := (*(*security.Principal)(jutil.Ptr(goPtr))).PublicKey()
	jKey, err := JavaPublicKey(env, key)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return jKey
}

//export Java_io_v_core_veyron2_security_PrincipalImpl_nativeBlessingsByName
func Java_io_v_core_veyron2_security_PrincipalImpl_nativeBlessingsByName(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong, jPattern C.jobject) C.jobjectArray {
	pattern, err := GoBlessingPattern(env, jPattern)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	blessings := (*(*security.Principal)(jutil.Ptr(goPtr))).BlessingsByName(pattern)
	barr := make([]interface{}, len(blessings))
	for i, b := range blessings {
		var err error
		if barr[i], err = JavaBlessings(env, b); err != nil {
			jutil.JThrowV(env, err)
			return nil
		}
	}
	return C.jobjectArray(jutil.JObjectArray(env, barr, jBlessingsClass))
}

//export Java_io_v_core_veyron2_security_PrincipalImpl_nativeBlessingsInfo
func Java_io_v_core_veyron2_security_PrincipalImpl_nativeBlessingsInfo(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong, jBlessings C.jobject) C.jobject {
	blessings, err := GoBlessings(env, jBlessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	info := (*(*security.Principal)(jutil.Ptr(goPtr))).BlessingsInfo(blessings)
	infomap := make(map[interface{}]interface{})
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
	return C.jobject(jInfo)
}

//export Java_io_v_core_veyron2_security_PrincipalImpl_nativeBlessingStore
func Java_io_v_core_veyron2_security_PrincipalImpl_nativeBlessingStore(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong) C.jobject {
	store := (*(*security.Principal)(jutil.Ptr(goPtr))).BlessingStore()
	jStore, err := JavaBlessingStore(env, store)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return jStore
}

//export Java_io_v_core_veyron2_security_PrincipalImpl_nativeRoots
func Java_io_v_core_veyron2_security_PrincipalImpl_nativeRoots(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong) C.jobject {
	roots := (*(*security.Principal)(jutil.Ptr(goPtr))).Roots()
	jRoots, err := JavaBlessingRoots(env, roots)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return jRoots
}

//export Java_io_v_core_veyron2_security_PrincipalImpl_nativeAddToRoots
func Java_io_v_core_veyron2_security_PrincipalImpl_nativeAddToRoots(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong, jBlessings C.jobject) {
	blessings, err := GoBlessings(env, jBlessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := (*(*security.Principal)(jutil.Ptr(goPtr))).AddToRoots(blessings); err != nil {
		jutil.JThrowV(env, err)
	}
}

//export Java_io_v_core_veyron2_security_PrincipalImpl_nativeFinalize
func Java_io_v_core_veyron2_security_PrincipalImpl_nativeFinalize(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*security.Principal)(jutil.Ptr(goPtr)))
}

//export Java_io_v_core_veyron2_security_BlessingsImpl_nativeCreate
func Java_io_v_core_veyron2_security_BlessingsImpl_nativeCreate(env *C.JNIEnv, jBlessingsImplClass C.jclass, jWire C.jobject) C.jobject {
	wire, err := GoWireBlessings(env, jWire)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	blessings, err := security.NewBlessings(wire)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return jBlessings
}

//export Java_io_v_core_veyron2_security_BlessingsImpl_nativeCreateUnion
func Java_io_v_core_veyron2_security_BlessingsImpl_nativeCreateUnion(env *C.JNIEnv, jBlessingsImplClass C.jclass, jBlessingsArr C.jobjectArray) C.jobject {
	blessingsArr, err := GoBlessingsArray(env, jBlessingsArr)
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
	return jBlessings
}

//export Java_io_v_core_veyron2_security_BlessingsImpl_nativeForContext
func Java_io_v_core_veyron2_security_BlessingsImpl_nativeForContext(env *C.JNIEnv, jBlessingsImpl C.jobject, goPtr C.jlong, jContext C.jobject) C.jobjectArray {
	context, err := GoContext(env, jContext)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	blessingStrs := (*(*security.Blessings)(jutil.Ptr(goPtr))).ForContext(context)
	return C.jobjectArray(jutil.JStringArray(env, blessingStrs))
}

//export Java_io_v_core_veyron2_security_BlessingsImpl_nativePublicKey
func Java_io_v_core_veyron2_security_BlessingsImpl_nativePublicKey(env *C.JNIEnv, jBlessingsImpl C.jobject, goPtr C.jlong) C.jobject {
	key := (*(*security.Blessings)(jutil.Ptr(goPtr))).PublicKey()
	jPublicKey, err := JavaPublicKey(env, key)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return jPublicKey
}

//export Java_io_v_core_veyron2_security_BlessingsImpl_nativeFinalize
func Java_io_v_core_veyron2_security_BlessingsImpl_nativeFinalize(env *C.JNIEnv, jBlessingsImpl C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*security.Blessings)(jutil.Ptr(goPtr)))
}

//export Java_io_v_core_veyron2_security_BlessingRootsImpl_nativeAdd
func Java_io_v_core_veyron2_security_BlessingRootsImpl_nativeAdd(env *C.JNIEnv, jBlessingRootsImpl C.jobject, goPtr C.jlong, jRoot C.jobject, jPattern C.jobject) {
	root, err := GoPublicKey(env, jRoot)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	pattern, err := GoBlessingPattern(env, jPattern)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := (*(*security.BlessingRoots)(jutil.Ptr(goPtr))).Add(root, pattern); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_v_core_veyron2_security_BlessingRootsImpl_nativeRecognized
func Java_io_v_core_veyron2_security_BlessingRootsImpl_nativeRecognized(env *C.JNIEnv, jBlessingRootsImpl C.jobject, goPtr C.jlong, jRoot C.jobject, jBlessing C.jstring) {
	root, err := GoPublicKey(env, jRoot)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	blessing := jutil.GoString(env, jBlessing)
	if err := (*(*security.BlessingRoots)(jutil.Ptr(goPtr))).Recognized(root, blessing); err != nil {
		jutil.JThrowV(env, err)
	}
}

//export Java_io_v_core_veyron2_security_BlessingRootsImpl_nativeDebugString
func Java_io_v_core_veyron2_security_BlessingRootsImpl_nativeDebugString(env *C.JNIEnv, jBlessingRootsImpl C.jobject, goPtr C.jlong) C.jstring {
	debug := (*(*security.BlessingRoots)(jutil.Ptr(goPtr))).DebugString()
	return C.jstring(jutil.JString(env, debug))
}

//export Java_io_v_core_veyron2_security_BlessingRootsImpl_nativeToString
func Java_io_v_core_veyron2_security_BlessingRootsImpl_nativeToString(env *C.JNIEnv, jBlessingRootsImpl C.jobject, goPtr C.jlong) C.jstring {
	str := fmt.Sprintf("%v", (*(*security.BlessingRoots)(jutil.Ptr(goPtr))))
	return C.jstring(jutil.JString(env, str))
}

//export Java_io_v_core_veyron2_security_BlessingRootsImpl_nativeFinalize
func Java_io_v_core_veyron2_security_BlessingRootsImpl_nativeFinalize(env *C.JNIEnv, jBlessingRootsImpl C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*security.BlessingRoots)(jutil.Ptr(goPtr)))
}

//export Java_io_v_core_veyron2_security_BlessingStoreImpl_nativeSet
func Java_io_v_core_veyron2_security_BlessingStoreImpl_nativeSet(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong, jBlessings C.jobject, jForPeers C.jobject) C.jobject {
	blessings, err := GoBlessings(env, jBlessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	forPeers, err := GoBlessingPattern(env, jForPeers)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	oldBlessings, err := (*(*security.BlessingStore)(jutil.Ptr(goPtr))).Set(blessings, forPeers)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jOldBlessings, err := JavaBlessings(env, oldBlessings)
	return nil
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return jOldBlessings
}

//export Java_io_v_core_veyron2_security_BlessingStoreImpl_nativeForPeer
func Java_io_v_core_veyron2_security_BlessingStoreImpl_nativeForPeer(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong, jPeerBlessings C.jobjectArray) C.jobject {
	peerBlessings := jutil.GoStringArray(env, jPeerBlessings)
	blessings := (*(*security.BlessingStore)(jutil.Ptr(goPtr))).ForPeer(peerBlessings...)
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return jBlessings
}

//export Java_io_v_core_veyron2_security_BlessingStoreImpl_nativeSetDefaultBlessings
func Java_io_v_core_veyron2_security_BlessingStoreImpl_nativeSetDefaultBlessings(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong, jBlessings C.jobject) {
	blessings, err := GoBlessings(env, jBlessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := (*(*security.BlessingStore)(jutil.Ptr(goPtr))).SetDefault(blessings); err != nil {
		jutil.JThrowV(env, err)
	}
}

//export Java_io_v_core_veyron2_security_BlessingStoreImpl_nativeDefaultBlessings
func Java_io_v_core_veyron2_security_BlessingStoreImpl_nativeDefaultBlessings(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong) C.jobject {
	blessings := (*(*security.BlessingStore)(jutil.Ptr(goPtr))).Default()
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return jBlessings
}

//export Java_io_v_core_veyron2_security_BlessingStoreImpl_nativePublicKey
func Java_io_v_core_veyron2_security_BlessingStoreImpl_nativePublicKey(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong) C.jobject {
	key := (*(*security.BlessingStore)(jutil.Ptr(goPtr))).PublicKey()
	jKey, err := JavaPublicKey(env, key)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return jKey
}

//export Java_io_v_core_veyron2_security_BlessingStoreImpl_nativePeerBlessings
func Java_io_v_core_veyron2_security_BlessingStoreImpl_nativePeerBlessings(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong) C.jobject {
	blessingsMap := (*(*security.BlessingStore)(jutil.Ptr(goPtr))).PeerBlessings()
	bmap := make(map[interface{}]interface{})
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
	return C.jobject(jBlessingsMap)
}

//export Java_io_v_core_veyron2_security_BlessingStoreImpl_nativeDebugString
func Java_io_v_core_veyron2_security_BlessingStoreImpl_nativeDebugString(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong) C.jstring {
	debug := (*(*security.BlessingStore)(jutil.Ptr(goPtr))).DebugString()
	return C.jstring(jutil.JString(env, debug))
}

//export Java_io_v_core_veyron2_security_BlessingStoreImpl_nativeToString
func Java_io_v_core_veyron2_security_BlessingStoreImpl_nativeToString(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong) C.jstring {
	str := fmt.Sprintf("%s", (*(*security.BlessingStore)(jutil.Ptr(goPtr))))
	return C.jstring(jutil.JString(env, str))
}

//export Java_io_v_core_veyron2_security_BlessingStoreImpl_nativeFinalize
func Java_io_v_core_veyron2_security_BlessingStoreImpl_nativeFinalize(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*security.BlessingStore)(jutil.Ptr(goPtr)))
}

//export Java_io_v_core_veyron2_security_BlessingPatternWrapper_nativeWrap
func Java_io_v_core_veyron2_security_BlessingPatternWrapper_nativeWrap(env *C.JNIEnv, jBlessingPatternWrapperClass C.jclass, jPattern C.jobject) C.jobject {
	pattern, err := GoBlessingPattern(env, jPattern)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jWrapper, err := JavaBlessingPatternWrapper(env, pattern)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return jWrapper
}

//export Java_io_v_core_veyron2_security_BlessingPatternWrapper_nativeIsMatchedBy
func Java_io_v_core_veyron2_security_BlessingPatternWrapper_nativeIsMatchedBy(env *C.JNIEnv, jBlessingPatternWrapper C.jobject, goPtr C.jlong, jBlessings C.jobjectArray) C.jboolean {
	blessings := jutil.GoStringArray(env, jBlessings)
	matched := (*(*security.BlessingPattern)(jutil.Ptr(goPtr))).MatchedBy(blessings...)
	if matched {
		return C.JNI_TRUE
	}
	return C.JNI_FALSE
}

//export Java_io_v_core_veyron2_security_BlessingPatternWrapper_nativeIsValid
func Java_io_v_core_veyron2_security_BlessingPatternWrapper_nativeIsValid(env *C.JNIEnv, jBlessingPatternWrapper C.jobject, goPtr C.jlong) C.jboolean {
	valid := (*(*security.BlessingPattern)(jutil.Ptr(goPtr))).IsValid()
	if valid {
		return C.JNI_TRUE
	}
	return C.JNI_FALSE
}

//export Java_io_v_core_veyron2_security_BlessingPatternWrapper_nativeMakeGlob
func Java_io_v_core_veyron2_security_BlessingPatternWrapper_nativeMakeGlob(env *C.JNIEnv, jBlessingPatternWrapper C.jobject, goPtr C.jlong) C.jobject {
	glob := (*(*security.BlessingPattern)(jutil.Ptr(goPtr))).MakeGlob()
	jWrapper, err := JavaBlessingPatternWrapper(env, glob)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return jWrapper
}

//export Java_io_v_core_veyron2_security_BlessingPatternWrapper_nativeFinalize
func Java_io_v_core_veyron2_security_BlessingPatternWrapper_nativeFinalize(env *C.JNIEnv, jBlessingPatternWrapper C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*security.BlessingPattern)(jutil.Ptr(goPtr)))
}
