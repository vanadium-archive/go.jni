// +build android

package security

import (
	"reflect"
	"time"
	"unsafe"

	"veyron.io/jni/util"
	jsecurity "veyron.io/jni/veyron2/security"
	isecurity "veyron.io/veyron/veyron/runtimes/google/security"
	"veyron.io/veyron/veyron2/security"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

var (
	// Global reference for io.veyron.veyron.veyron.runtimes.google.security.PublicID class.
	jPublicIDImplClass C.jclass
	// Global reference for io.veyron.veyron.veyron.runtimes.google.security.Context class.
	jContextImplClass C.jclass
	// Global reference for io.veyron.veyron.veyron2.security.BlessingPattern class.
	jBlessingPatternClass C.jclass
	// Global reference for org.joda.time.Duration class.
	jDurationClass C.jclass
	// Signature of the PublicID interface.
	publicIDSign = util.ClassSign("io.veyron.veyron.veyron2.security.PublicID")
	// Signature of the BlessingPattern class.
	principalPatternSign = util.ClassSign("io.veyron.veyron.veyron2.security.BlessingPattern")
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
	jPublicIDImplClass = C.jclass(util.JFindClassOrPrint(env, "io/veyron/veyron/veyron/runtimes/google/security/PublicID"))
	jContextImplClass = C.jclass(util.JFindClassOrPrint(env, "io/veyron/veyron/veyron/runtimes/google/security/Context"))
	jBlessingPatternClass = C.jclass(util.JFindClassOrPrint(env, "io/veyron/veyron/veyron2/security/BlessingPattern"))
	jDurationClass = C.jclass(util.JFindClassOrPrint(env, "org/joda/time/Duration"))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_PublicIDStore_nativeAdd
func Java_io_veyron_veyron_veyron_runtimes_google_security_PublicIDStore_nativeAdd(env *C.JNIEnv, jPublicIDStore C.jobject, goPublicIDStorePtr C.jlong, jID C.jobject, jPeerPattern C.jstring) {
	idPtr := util.CallLongMethodOrCatch(env, jID, "getNativePtr", nil)
	id := (*(*security.PublicID)(util.Ptr(idPtr)))
	peerPattern := security.BlessingPattern(util.GoString(env, jPeerPattern))
	if err := (*(*security.PublicIDStore)(util.Ptr(goPublicIDStorePtr))).Add(id, peerPattern); err != nil {
		util.JThrowV(env, err)
		return
	}
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_PublicIDStore_nativeGetPeerID
func Java_io_veyron_veyron_veyron_runtimes_google_security_PublicIDStore_nativeGetPeerID(env *C.JNIEnv, jPublicIDStore C.jobject, goPublicIDStorePtr C.jlong, jPeerID C.jobject) C.jlong {
	peerIDPtr := util.CallLongMethodOrCatch(env, jPeerID, "getNativePtr", nil)
	peerID := (*(*security.PublicID)(util.Ptr(peerIDPtr)))
	id, err := (*(*security.PublicIDStore)(util.Ptr(goPublicIDStorePtr))).ForPeer(peerID)
	if err != nil {
		util.JThrowV(env, err)
		return C.jlong(0)
	}
	util.GoRef(&id) // Un-refed when the Java PublicID is finalized.
	return C.jlong(util.PtrValue(&id))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_PublicIDStore_nativeDefaultPublicID
func Java_io_veyron_veyron_veyron_runtimes_google_security_PublicIDStore_nativeDefaultPublicID(env *C.JNIEnv, jPublicIDStore C.jobject, goPublicIDStorePtr C.jlong) C.jlong {
	id, err := (*(*security.PublicIDStore)(util.Ptr(goPublicIDStorePtr))).DefaultPublicID()
	if err != nil {
		util.JThrowV(env, err)
		return C.jlong(0)
	}
	util.GoRef(&id) // Un-refed when the Java PublicID is finalized.
	return C.jlong(util.PtrValue(&id))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_PublicIDStore_nativeSetDefaultBlessingPattern
func Java_io_veyron_veyron_veyron_runtimes_google_security_PublicIDStore_nativeSetDefaultBlessingPattern(env *C.JNIEnv, jPublicIDStore C.jobject, goPublicIDStorePtr C.jlong, jPattern C.jstring) {
	pattern := security.BlessingPattern(util.GoString(env, jPattern))
	if err := (*(*security.PublicIDStore)(util.Ptr(goPublicIDStorePtr))).SetDefaultBlessingPattern(pattern); err != nil {
		util.JThrowV(env, err)
		return
	}
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_PublicIDStore_nativeFinalize
func Java_io_veyron_veyron_veyron_runtimes_google_security_PublicIDStore_nativeFinalize(env *C.JNIEnv, jPublicIDStore C.jobject, goPublicIDStorePtr C.jlong) {
	util.GoUnref((*security.PublicIDStore)(util.Ptr(goPublicIDStorePtr)))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_PublicID_nativeNames
func Java_io_veyron_veyron_veyron_runtimes_google_security_PublicID_nativeNames(env *C.JNIEnv, jPublicID C.jobject, goPublicIDPtr C.jlong) C.jobjectArray {
	names := (*(*security.PublicID)(util.Ptr(goPublicIDPtr))).Names()
	return C.jobjectArray(util.JStringArray(env, names))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_PublicID_nativePublicKey
func Java_io_veyron_veyron_veyron_runtimes_google_security_PublicID_nativePublicKey(env *C.JNIEnv, jPublicID C.jobject, goPublicIDPtr C.jlong) C.jbyteArray {
	key := (*(*security.PublicID)(util.Ptr(goPublicIDPtr))).PublicKey()
	encoded, err := key.MarshalBinary()
	if err != nil {
		util.JThrowV(env, err)
		return C.jbyteArray(nil)
	}
	return C.jbyteArray(util.JByteArray(env, encoded))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_PublicID_nativeAuthorize
func Java_io_veyron_veyron_veyron_runtimes_google_security_PublicID_nativeAuthorize(env *C.JNIEnv, jPublicID C.jobject, goPublicIDPtr C.jlong, jContext C.jobject) C.jlong {
	context, err := jsecurity.GoContext(env, jContext)
	if err != nil {
		util.JThrowV(env, err)
		return C.jlong(0)
	}
	id, err := (*(*security.PublicID)(util.Ptr(goPublicIDPtr))).Authorize(context)
	if err != nil {
		util.JThrowV(env, err)
		return C.jlong(0)
	}
	util.GoRef(&id) // Un-refed when the Java PublicID is finalized.
	return C.jlong(util.PtrValue(&id))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_PublicID_nativeEncode
func Java_io_veyron_veyron_veyron_runtimes_google_security_PublicID_nativeEncode(env *C.JNIEnv, jPublicID C.jobject, goPublicIDPtr C.jlong) C.jobjectArray {
	pubID := *(*security.PublicID)(util.Ptr(goPublicIDPtr))
	chains, err := EncodeChains(pubID)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return C.jobjectArray(util.JStringArray(env, chains))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_PublicID_nativeEquals
func Java_io_veyron_veyron_veyron_runtimes_google_security_PublicID_nativeEquals(env *C.JNIEnv, jPublicID C.jobject, goPublicIDPtr, goOtherPublicIDPtr C.jlong) C.jboolean {
	id := *(*security.PublicID)(util.Ptr(goPublicIDPtr))
	other := *(*security.PublicID)(util.Ptr(goOtherPublicIDPtr))
	if reflect.DeepEqual(id, other) {
		return C.JNI_TRUE
	}
	return C.JNI_FALSE
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_PublicID_nativeFinalize
func Java_io_veyron_veyron_veyron_runtimes_google_security_PublicID_nativeFinalize(env *C.JNIEnv, jPublicID C.jobject, goPublicIDPtr C.jlong) {
	util.GoUnref((*security.PublicID)(util.Ptr(goPublicIDPtr)))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_PrivateID_nativeCreate
func Java_io_veyron_veyron_veyron_runtimes_google_security_PrivateID_nativeCreate(env *C.JNIEnv, jPrivateIDClass C.jclass, name C.jstring, jSigner C.jobject) C.jlong {
	signer := newSigner(env, jSigner)
	id, err := isecurity.NewPrivateID(util.GoString(env, name), signer)
	if err != nil {
		util.JThrowV(env, err)
		return C.jlong(0)
	}
	util.GoRef(&id) // Un-refed when the Java PrivateID is finalized.
	return C.jlong(util.PtrValue(&id))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_PrivateID_nativePublicID
func Java_io_veyron_veyron_veyron_runtimes_google_security_PrivateID_nativePublicID(env *C.JNIEnv, jPrivateID C.jobject, goPrivateIDPtr C.jlong) C.jlong {
	id := (*(*security.PrivateID)(util.Ptr(goPrivateIDPtr))).PublicID()
	util.GoRef(&id) // Un-refed when the Java PublicID is finalized.
	return C.jlong(util.PtrValue(&id))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_PrivateID_nativeBless
func Java_io_veyron_veyron_veyron_runtimes_google_security_PrivateID_nativeBless(env *C.JNIEnv, jPrivateID C.jobject, goPrivateIDPtr C.jlong, jBlesseeEncodedChains C.jobjectArray, name C.jstring, jDuration C.jobject) C.jlong {
	blessee, err := DecodeChains(util.GoStringArray(env, jBlesseeEncodedChains))
	if err != nil {
		util.JThrowV(env, err)
		return C.jlong(0)
	}
	duration := time.Duration(util.CallLongMethodOrCatch(env, jDuration, "getMillis", nil)) * time.Millisecond
	id, err := (*(*security.PrivateID)(util.Ptr(goPrivateIDPtr))).Bless(blessee, util.GoString(env, name), duration, nil)
	if err != nil {
		util.JThrowV(env, err)
		return C.jlong(0)
	}
	util.GoRef(&id) // Un-refed when the Java PublicID is finalized
	return C.jlong(util.PtrValue(&id))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_PrivateID_nativeDerive
func Java_io_veyron_veyron_veyron_runtimes_google_security_PrivateID_nativeDerive(env *C.JNIEnv, jPrivateID C.jobject, goPrivateIDPtr C.jlong, jEncodedChains C.jobjectArray) C.jlong {
	publicID, err := DecodeChains(util.GoStringArray(env, jEncodedChains))
	if err != nil {
		util.JThrowV(env, err)
		return C.jlong(0)
	}
	id, err := (*(*security.PrivateID)(util.Ptr(goPrivateIDPtr))).Derive(publicID)
	if err != nil {
		util.JThrowV(env, err)
		return C.jlong(0)
	}
	util.GoRef(&id) // Un-refed when the Java PrivateID is finalized.
	return C.jlong(util.PtrValue(&id))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_PrivateID_nativeFinalize
func Java_io_veyron_veyron_veyron_runtimes_google_security_PrivateID_nativeFinalize(env *C.JNIEnv, jPrivateID C.jobject, goPrivateIDPtr C.jlong) {
	util.GoUnref((*security.PrivateID)(util.Ptr(goPrivateIDPtr)))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeMethod
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeMethod(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jstring {
	return C.jstring(util.JString(env, (*(*security.Context)(util.Ptr(goContextPtr))).Method()))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeName
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeName(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jstring {
	return C.jstring(util.JString(env, (*(*security.Context)(util.Ptr(goContextPtr))).Name()))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeSuffix
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeSuffix(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jstring {
	return C.jstring(util.JString(env, (*(*security.Context)(util.Ptr(goContextPtr))).Suffix()))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLabel
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLabel(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jint {
	return C.jint((*(*security.Context)(util.Ptr(goContextPtr))).Label())
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLocalEndpoint
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLocalEndpoint(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jstring {
	return C.jstring(util.JString(env, (*(*security.Context)(util.Ptr(goContextPtr))).LocalEndpoint().String()))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeRemoteEndpoint
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeRemoteEndpoint(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jstring {
	return C.jstring(util.JString(env, (*(*security.Context)(util.Ptr(goContextPtr))).RemoteEndpoint().String()))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLocalPrincipal
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLocalPrincipal(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jobject {
	principal := (*(*security.Context)(util.Ptr(goContextPtr))).LocalPrincipal()
	jPrincipal, err := jsecurity.JavaPrincipal(env, principal)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return C.jobject(jPrincipal)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLocalBlessings
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLocalBlessings(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jobject {
	blessings := (*(*security.Context)(util.Ptr(goContextPtr))).LocalBlessings()
	jBlessings, err := jsecurity.JavaBlessings(env, blessings)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return C.jobject(jBlessings)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeRemoteBlessings
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeRemoteBlessings(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jobject {
	blessings := (*(*security.Context)(util.Ptr(goContextPtr))).RemoteBlessings()
	jBlessings, err := jsecurity.JavaBlessings(env, blessings)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return C.jobject(jBlessings)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLocalID
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLocalID(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jlong {
	id := (*(*security.Context)(util.Ptr(goContextPtr))).LocalID()
	util.GoRef(&id) // Un-refed when the Java PublicID object is finalized.
	return C.jlong(util.PtrValue(&id))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeRemoteID
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeRemoteID(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jlong {
	id := (*(*security.Context)(util.Ptr(goContextPtr))).RemoteID()
	util.GoRef(&id)
	return C.jlong(util.PtrValue(&id))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeFinalize
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeFinalize(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) {
	util.GoUnref((*security.Context)(util.Ptr(goContextPtr)))
}
