// +build android

package security

import (
	"fmt"

	"v.io/v23/security"
	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
	vsecurity "v.io/x/ref/security"
)

// #include "jni.h"
import "C"

var (
	principalSign       = jutil.ClassSign("io.v.v23.security.Principal")
	blessingsSign       = jutil.ClassSign("io.v.v23.security.Blessings")
	wireBlessingsSign   = jutil.ClassSign("io.v.v23.security.WireBlessings")
	blessingStoreSign   = jutil.ClassSign("io.v.v23.security.BlessingStore")
	blessingRootsSign   = jutil.ClassSign("io.v.v23.security.BlessingRoots")
	blessingPatternSign = jutil.ClassSign("io.v.v23.security.BlessingPattern")
	signerSign          = jutil.ClassSign("io.v.v23.security.Signer")
	caveatSign          = jutil.ClassSign("io.v.v23.security.Caveat")
	callSign            = jutil.ClassSign("io.v.v23.security.Call")
	signatureSign       = jutil.ClassSign("io.v.v23.security.Signature")
	publicKeySign       = jutil.ClassSign("java.security.interfaces.ECPublicKey")

	// Global reference for io.v.v23.security.Blessings class.
	jBlessingsClass C.jclass
	// Global reference for io.v.v23.security.Caveat class.
	jCaveatClass C.jclass
	// Global reference for io.v.v23.security.WireBlessings class.
	jWireBlessingsClass C.jclass
	// Global reference for io.v.v23.security.PrincipalImpl class.
	jPrincipalImplClass C.jclass
	// Global reference for io.v.v23.security.BlessingStoreImpl class.
	jBlessingStoreImplClass C.jclass
	// Global reference for io.v.v23.security.BlessingRootsImpl class.
	jBlessingRootsImplClass C.jclass
	// Global reference for io.v.v23.security.CallImpl class.
	jCallImplClass C.jclass
	// Global reference for io.v.v23.security.BlessingPatternWrapper class.
	jBlessingPatternWrapperClass C.jclass
	// Global reference for io.v.v23.security.CaveatRegistry class.
	jCaveatRegistryClass C.jclass
	// Global reference for io.v.v23.security.Util class.
	jUtilClass C.jclass
	// Global reference for java.lang.Object class.
	jObjectClass C.jclass
)

// Init initializes the JNI code with the given Java evironment. This method
// must be called from the main Java thread.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java environment is passed in an empty
// interface and then cast into the package-local environment type.
func Init(jEnv interface{}) error {
	security.SetCaveatValidator(caveatValidator)

	// Cache global references to all Java classes used by the package.  This is
	// necessary because JNI gets access to the class loader only in the system
	// thread, so we aren't able to invoke FindClass in other threads.
	class, err := jutil.JFindClass(jEnv, "io/v/v23/security/Blessings")
	if err != nil {
		return err
	}
	jBlessingsClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/security/Caveat")
	if err != nil {
		return err
	}
	jCaveatClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/security/WireBlessings")
	if err != nil {
		return err
	}
	jWireBlessingsClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/security/PrincipalImpl")
	if err != nil {
		return err
	}
	jPrincipalImplClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/security/BlessingStoreImpl")
	if err != nil {
		return err
	}
	jBlessingStoreImplClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/security/BlessingRootsImpl")
	if err != nil {
		return err
	}
	jBlessingRootsImplClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/security/CallImpl")
	if err != nil {
		return err
	}
	jCallImplClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/security/BlessingPatternWrapper")
	if err != nil {
		return err
	}
	jBlessingPatternWrapperClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/security/CaveatRegistry")
	if err != nil {
		return err
	}
	jCaveatRegistryClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/security/Util")
	if err != nil {
		return err
	}
	jUtilClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "java/lang/Object")
	if err != nil {
		return err
	}
	jObjectClass = C.jclass(class)
	return nil
}

//export Java_io_v_v23_security_CallImpl_nativeTimestamp
func Java_io_v_v23_security_CallImpl_nativeTimestamp(env *C.JNIEnv, jCall C.jobject, goPtr C.jlong) C.jobject {
	t := (*(*security.Call)(jutil.Ptr(goPtr))).Timestamp()
	jTime, err := jutil.JTime(env, t)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jTime)
}

//export Java_io_v_v23_security_CallImpl_nativeMethod
func Java_io_v_v23_security_CallImpl_nativeMethod(env *C.JNIEnv, jCall C.jobject, goPtr C.jlong) C.jstring {
	method := (*(*security.Call)(jutil.Ptr(goPtr))).Method()
	return C.jstring(jutil.JString(env, jutil.CamelCase(method)))
}

//export Java_io_v_v23_security_CallImpl_nativeMethodTags
func Java_io_v_v23_security_CallImpl_nativeMethodTags(env *C.JNIEnv, jCall C.jobject, goPtr C.jlong) C.jobjectArray {
	tags := (*(*security.Call)(jutil.Ptr(goPtr))).MethodTags()
	jTags, err := jutil.JVDLValueArray(env, tags)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobjectArray(jTags)
}

//export Java_io_v_v23_security_CallImpl_nativeSuffix
func Java_io_v_v23_security_CallImpl_nativeSuffix(env *C.JNIEnv, jCall C.jobject, goPtr C.jlong) C.jstring {
	return C.jstring(jutil.JString(env, (*(*security.Call)(jutil.Ptr(goPtr))).Suffix()))
}

//export Java_io_v_v23_security_CallImpl_nativeLocalEndpoint
func Java_io_v_v23_security_CallImpl_nativeLocalEndpoint(env *C.JNIEnv, jCall C.jobject, goPtr C.jlong) C.jstring {
	return C.jstring(jutil.JString(env, (*(*security.Call)(jutil.Ptr(goPtr))).LocalEndpoint().String()))
}

//export Java_io_v_v23_security_CallImpl_nativeRemoteEndpoint
func Java_io_v_v23_security_CallImpl_nativeRemoteEndpoint(env *C.JNIEnv, jCall C.jobject, goPtr C.jlong) C.jstring {
	return C.jstring(jutil.JString(env, (*(*security.Call)(jutil.Ptr(goPtr))).RemoteEndpoint().String()))
}

//export Java_io_v_v23_security_CallImpl_nativeLocalPrincipal
func Java_io_v_v23_security_CallImpl_nativeLocalPrincipal(env *C.JNIEnv, jCall C.jobject, goPtr C.jlong) C.jobject {
	principal := (*(*security.Call)(jutil.Ptr(goPtr))).LocalPrincipal()
	jPrincipal, err := JavaPrincipal(env, principal)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jPrincipal)
}

//export Java_io_v_v23_security_CallImpl_nativeLocalBlessings
func Java_io_v_v23_security_CallImpl_nativeLocalBlessings(env *C.JNIEnv, jCall C.jobject, goPtr C.jlong) C.jobject {
	blessings := (*(*security.Call)(jutil.Ptr(goPtr))).LocalBlessings()
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jBlessings)
}

//export Java_io_v_v23_security_CallImpl_nativeRemoteBlessings
func Java_io_v_v23_security_CallImpl_nativeRemoteBlessings(env *C.JNIEnv, jCall C.jobject, goPtr C.jlong) C.jobject {
	blessings := (*(*security.Call)(jutil.Ptr(goPtr))).RemoteBlessings()
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jBlessings)
}

//export Java_io_v_v23_security_CallImpl_nativeContext
func Java_io_v_v23_security_CallImpl_nativeContext(env *C.JNIEnv, jCall C.jobject, goPtr C.jlong) C.jobject {
	ctx := (*(*security.Call)(jutil.Ptr(goPtr))).Context()
	jCtx, err := jcontext.JavaContext(env, ctx, nil)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jCtx)
}

//export Java_io_v_v23_security_CallImpl_nativeFinalize
func Java_io_v_v23_security_CallImpl_nativeFinalize(env *C.JNIEnv, jCall C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*security.Call)(jutil.Ptr(goPtr)))
}

//export Java_io_v_v23_security_PrincipalImpl_nativeCreate
func Java_io_v_v23_security_PrincipalImpl_nativeCreate(env *C.JNIEnv, jPrincipalImplClass C.jclass) C.jobject {
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
	return C.jobject(jPrincipal)
}

//export Java_io_v_v23_security_PrincipalImpl_nativeCreateForSigner
func Java_io_v_v23_security_PrincipalImpl_nativeCreateForSigner(env *C.JNIEnv, jPrincipalImplClass C.jclass, jSigner C.jobject) C.jobject {
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

//export Java_io_v_v23_security_PrincipalImpl_nativeCreateForAll
func Java_io_v_v23_security_PrincipalImpl_nativeCreateForAll(env *C.JNIEnv, jPrincipalImplClass C.jclass, jSigner C.jobject, jStore C.jobject, jRoots C.jobject) C.jobject {
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

//export Java_io_v_v23_security_PrincipalImpl_nativeCreatePersistent
func Java_io_v_v23_security_PrincipalImpl_nativeCreatePersistent(env *C.JNIEnv, jPrincipalImplClass C.jclass, jPassphrase C.jstring, jDir C.jstring) C.jobject {
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

//export Java_io_v_v23_security_PrincipalImpl_nativeCreatePersistentForSigner
func Java_io_v_v23_security_PrincipalImpl_nativeCreatePersistentForSigner(env *C.JNIEnv, jPrincipalImplClass C.jclass, jSigner C.jobject, jDir C.jstring) C.jobject {
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

//export Java_io_v_v23_security_PrincipalImpl_nativeBless
func Java_io_v_v23_security_PrincipalImpl_nativeBless(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong, jKey C.jobject, jWith C.jobject, jExtension C.jstring, jCaveat C.jobject, jAdditionalCaveats C.jobjectArray) C.jobject {
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
	return C.jobject(jBlessings)
}

//export Java_io_v_v23_security_PrincipalImpl_nativeBlessSelf
func Java_io_v_v23_security_PrincipalImpl_nativeBlessSelf(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong, jName C.jstring, jCaveats C.jobjectArray) C.jobject {
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
	return C.jobject(jBlessings)
}

//export Java_io_v_v23_security_PrincipalImpl_nativeSign
func Java_io_v_v23_security_PrincipalImpl_nativeSign(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong, jMessage C.jbyteArray) C.jobject {
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
	return C.jobject(jSig)
}

//export Java_io_v_v23_security_PrincipalImpl_nativePublicKey
func Java_io_v_v23_security_PrincipalImpl_nativePublicKey(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong) C.jobject {
	key := (*(*security.Principal)(jutil.Ptr(goPtr))).PublicKey()
	jKey, err := JavaPublicKey(env, key)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jKey)
}

//export Java_io_v_v23_security_PrincipalImpl_nativeBlessingsByName
func Java_io_v_v23_security_PrincipalImpl_nativeBlessingsByName(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong, jPattern C.jobject) C.jobjectArray {
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

//export Java_io_v_v23_security_PrincipalImpl_nativeBlessingsInfo
func Java_io_v_v23_security_PrincipalImpl_nativeBlessingsInfo(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong, jBlessings C.jobject) C.jobject {
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

//export Java_io_v_v23_security_PrincipalImpl_nativeBlessingStore
func Java_io_v_v23_security_PrincipalImpl_nativeBlessingStore(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong) C.jobject {
	store := (*(*security.Principal)(jutil.Ptr(goPtr))).BlessingStore()
	jStore, err := JavaBlessingStore(env, store)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jStore)
}

//export Java_io_v_v23_security_PrincipalImpl_nativeRoots
func Java_io_v_v23_security_PrincipalImpl_nativeRoots(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong) C.jobject {
	roots := (*(*security.Principal)(jutil.Ptr(goPtr))).Roots()
	jRoots, err := JavaBlessingRoots(env, roots)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jRoots)
}

//export Java_io_v_v23_security_PrincipalImpl_nativeAddToRoots
func Java_io_v_v23_security_PrincipalImpl_nativeAddToRoots(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong, jBlessings C.jobject) {
	blessings, err := GoBlessings(env, jBlessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := (*(*security.Principal)(jutil.Ptr(goPtr))).AddToRoots(blessings); err != nil {
		jutil.JThrowV(env, err)
	}
}

//export Java_io_v_v23_security_PrincipalImpl_nativeFinalize
func Java_io_v_v23_security_PrincipalImpl_nativeFinalize(env *C.JNIEnv, jPrincipalImpl C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*security.Principal)(jutil.Ptr(goPtr)))
}

//export Java_io_v_v23_security_Blessings_nativeCreate
func Java_io_v_v23_security_Blessings_nativeCreate(env *C.JNIEnv, jBlessingsClass C.jclass, jWire C.jobject) C.jobject {
	var blessings security.Blessings
	if err := jutil.GoVomCopy(env, jWire, jWireBlessingsClass, &blessings); err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jBlessings)
}

//export Java_io_v_v23_security_Blessings_nativeCreateUnion
func Java_io_v_v23_security_Blessings_nativeCreateUnion(env *C.JNIEnv, jBlessingsClass C.jclass, jBlessingsArr C.jobjectArray) C.jobject {
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
	return C.jobject(jBlessings)
}

//export Java_io_v_v23_security_Blessings_nativeForCall
func Java_io_v_v23_security_Blessings_nativeForCall(env *C.JNIEnv, jBlessings C.jobject, goPtr C.jlong, jCall C.jobject) C.jobjectArray {
	call, err := GoCall(env, jCall)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	blessingStrs, _ := (*(*security.Blessings)(jutil.Ptr(goPtr))).ForCall(call)
	return C.jobjectArray(jutil.JStringArray(env, blessingStrs))
}

//export Java_io_v_v23_security_Blessings_nativePublicKey
func Java_io_v_v23_security_Blessings_nativePublicKey(env *C.JNIEnv, jBlessings C.jobject, goPtr C.jlong) C.jobject {
	key := (*(*security.Blessings)(jutil.Ptr(goPtr))).PublicKey()
	jPublicKey, err := JavaPublicKey(env, key)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jPublicKey)
}

//export Java_io_v_v23_security_Blessings_nativeFinalize
func Java_io_v_v23_security_Blessings_nativeFinalize(env *C.JNIEnv, jBlessings C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*security.Blessings)(jutil.Ptr(goPtr)))
}

//export Java_io_v_v23_security_BlessingRootsImpl_nativeAdd
func Java_io_v_v23_security_BlessingRootsImpl_nativeAdd(env *C.JNIEnv, jBlessingRootsImpl C.jobject, goPtr C.jlong, jRoot C.jobject, jPattern C.jobject) {
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

//export Java_io_v_v23_security_BlessingRootsImpl_nativeRecognized
func Java_io_v_v23_security_BlessingRootsImpl_nativeRecognized(env *C.JNIEnv, jBlessingRootsImpl C.jobject, goPtr C.jlong, jRoot C.jobject, jBlessing C.jstring) {
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

//export Java_io_v_v23_security_BlessingRootsImpl_nativeDebugString
func Java_io_v_v23_security_BlessingRootsImpl_nativeDebugString(env *C.JNIEnv, jBlessingRootsImpl C.jobject, goPtr C.jlong) C.jstring {
	debug := (*(*security.BlessingRoots)(jutil.Ptr(goPtr))).DebugString()
	return C.jstring(jutil.JString(env, debug))
}

//export Java_io_v_v23_security_BlessingRootsImpl_nativeToString
func Java_io_v_v23_security_BlessingRootsImpl_nativeToString(env *C.JNIEnv, jBlessingRootsImpl C.jobject, goPtr C.jlong) C.jstring {
	str := fmt.Sprintf("%v", (*(*security.BlessingRoots)(jutil.Ptr(goPtr))))
	return C.jstring(jutil.JString(env, str))
}

//export Java_io_v_v23_security_BlessingRootsImpl_nativeFinalize
func Java_io_v_v23_security_BlessingRootsImpl_nativeFinalize(env *C.JNIEnv, jBlessingRootsImpl C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*security.BlessingRoots)(jutil.Ptr(goPtr)))
}

//export Java_io_v_v23_security_BlessingStoreImpl_nativeSet
func Java_io_v_v23_security_BlessingStoreImpl_nativeSet(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong, jBlessings C.jobject, jForPeers C.jobject) C.jobject {
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
	return C.jobject(jOldBlessings)
}

//export Java_io_v_v23_security_BlessingStoreImpl_nativeForPeer
func Java_io_v_v23_security_BlessingStoreImpl_nativeForPeer(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong, jPeerBlessings C.jobjectArray) C.jobject {
	peerBlessings := jutil.GoStringArray(env, jPeerBlessings)
	blessings := (*(*security.BlessingStore)(jutil.Ptr(goPtr))).ForPeer(peerBlessings...)
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jBlessings)
}

//export Java_io_v_v23_security_BlessingStoreImpl_nativeSetDefaultBlessings
func Java_io_v_v23_security_BlessingStoreImpl_nativeSetDefaultBlessings(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong, jBlessings C.jobject) {
	blessings, err := GoBlessings(env, jBlessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := (*(*security.BlessingStore)(jutil.Ptr(goPtr))).SetDefault(blessings); err != nil {
		jutil.JThrowV(env, err)
	}
}

//export Java_io_v_v23_security_BlessingStoreImpl_nativeDefaultBlessings
func Java_io_v_v23_security_BlessingStoreImpl_nativeDefaultBlessings(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong) C.jobject {
	blessings := (*(*security.BlessingStore)(jutil.Ptr(goPtr))).Default()
	jBlessings, err := JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jBlessings)
}

//export Java_io_v_v23_security_BlessingStoreImpl_nativePublicKey
func Java_io_v_v23_security_BlessingStoreImpl_nativePublicKey(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong) C.jobject {
	key := (*(*security.BlessingStore)(jutil.Ptr(goPtr))).PublicKey()
	jKey, err := JavaPublicKey(env, key)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jKey)
}

//export Java_io_v_v23_security_BlessingStoreImpl_nativePeerBlessings
func Java_io_v_v23_security_BlessingStoreImpl_nativePeerBlessings(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong) C.jobject {
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

//export Java_io_v_v23_security_BlessingStoreImpl_nativeDebugString
func Java_io_v_v23_security_BlessingStoreImpl_nativeDebugString(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong) C.jstring {
	debug := (*(*security.BlessingStore)(jutil.Ptr(goPtr))).DebugString()
	return C.jstring(jutil.JString(env, debug))
}

//export Java_io_v_v23_security_BlessingStoreImpl_nativeToString
func Java_io_v_v23_security_BlessingStoreImpl_nativeToString(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong) C.jstring {
	str := fmt.Sprintf("%s", (*(*security.BlessingStore)(jutil.Ptr(goPtr))))
	return C.jstring(jutil.JString(env, str))
}

//export Java_io_v_v23_security_BlessingStoreImpl_nativeFinalize
func Java_io_v_v23_security_BlessingStoreImpl_nativeFinalize(env *C.JNIEnv, jBlessingStoreImpl C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*security.BlessingStore)(jutil.Ptr(goPtr)))
}

//export Java_io_v_v23_security_BlessingPatternWrapper_nativeWrap
func Java_io_v_v23_security_BlessingPatternWrapper_nativeWrap(env *C.JNIEnv, jBlessingPatternWrapperClass C.jclass, jPattern C.jobject) C.jobject {
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
	return C.jobject(jWrapper)
}

//export Java_io_v_v23_security_BlessingPatternWrapper_nativeIsMatchedBy
func Java_io_v_v23_security_BlessingPatternWrapper_nativeIsMatchedBy(env *C.JNIEnv, jBlessingPatternWrapper C.jobject, goPtr C.jlong, jBlessings C.jobjectArray) C.jboolean {
	blessings := jutil.GoStringArray(env, jBlessings)
	matched := (*(*security.BlessingPattern)(jutil.Ptr(goPtr))).MatchedBy(blessings...)
	if matched {
		return C.JNI_TRUE
	}
	return C.JNI_FALSE
}

//export Java_io_v_v23_security_BlessingPatternWrapper_nativeIsValid
func Java_io_v_v23_security_BlessingPatternWrapper_nativeIsValid(env *C.JNIEnv, jBlessingPatternWrapper C.jobject, goPtr C.jlong) C.jboolean {
	valid := (*(*security.BlessingPattern)(jutil.Ptr(goPtr))).IsValid()
	if valid {
		return C.JNI_TRUE
	}
	return C.JNI_FALSE
}

//export Java_io_v_v23_security_BlessingPatternWrapper_nativeMakeNonExtendable
func Java_io_v_v23_security_BlessingPatternWrapper_nativeMakeNonExtendable(env *C.JNIEnv, jBlessingPatternWrapper C.jobject, goPtr C.jlong) C.jobject {
	p := (*(*security.BlessingPattern)(jutil.Ptr(goPtr))).MakeNonExtendable()
	jWrapper, err := JavaBlessingPatternWrapper(env, p)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jWrapper)
}

//export Java_io_v_v23_security_BlessingPatternWrapper_nativeFinalize
func Java_io_v_v23_security_BlessingPatternWrapper_nativeFinalize(env *C.JNIEnv, jBlessingPatternWrapper C.jobject, goPtr C.jlong) {
	jutil.GoUnref((*security.BlessingPattern)(jutil.Ptr(goPtr)))
}
