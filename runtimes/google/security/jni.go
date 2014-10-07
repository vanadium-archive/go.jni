// +build android

package security

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"time"
	"unsafe"

	"veyron.io/jni/runtimes/google/util"
	isecurity "veyron.io/veyron/veyron/runtimes/google/security"
	"veyron.io/veyron/veyron2/security"
	"veyron.io/veyron/veyron2/security/wire"
	"veyron.io/veyron/veyron2/vom"
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
	jPublicIDImplClass = C.jclass(util.JFindClassPtrOrDie(env, "io/veyron/veyron/veyron/runtimes/google/security/PublicID"))
	jContextImplClass = C.jclass(util.JFindClassPtrOrDie(env, "io/veyron/veyron/veyron/runtimes/google/security/Context"))
	jBlessingPatternClass = C.jclass(util.JFindClassPtrOrDie(env, "io/veyron/veyron/veyron2/security/BlessingPattern"))
	jDurationClass = C.jclass(util.JFindClassPtrOrDie(env, "org/joda/time/Duration"))
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
	return C.jobjectArray(util.JStringArrayPtr(env, names))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_PublicID_nativePublicKey
func Java_io_veyron_veyron_veyron_runtimes_google_security_PublicID_nativePublicKey(env *C.JNIEnv, jPublicID C.jobject, goPublicIDPtr C.jlong) C.jbyteArray {
	key := (*(*security.PublicID)(util.Ptr(goPublicIDPtr))).PublicKey()
	encoded, err := key.MarshalBinary()
	if err != nil {
		util.JThrowV(env, err)
		return C.jbyteArray(nil)
	}
	return C.jbyteArray(util.JByteArrayPtr(env, encoded))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_PublicID_nativeAuthorize
func Java_io_veyron_veyron_veyron_runtimes_google_security_PublicID_nativeAuthorize(env *C.JNIEnv, jPublicID C.jobject, goPublicIDPtr C.jlong, jContext C.jobject) C.jlong {
	id, err := (*(*security.PublicID)(util.Ptr(goPublicIDPtr))).Authorize(newContext(env, jContext))
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
	chains, err := getChains(pubID)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	jsonChains := make([]string, len(chains))
	for i, chain := range chains {
		enc, err := json.Marshal(chain)
		if err != nil {
			util.JThrowV(env, err)
			return nil
		}
		jsonChains[i] = string(enc)
	}
	return C.jobjectArray(util.JStringArrayPtr(env, jsonChains))
}

func getChains(id security.PublicID) ([]wire.ChainPublicID, error) {
	m := reflect.ValueOf(id).MethodByName("VomEncode")
	if !m.IsValid() {
		return nil, fmt.Errorf("type %T doesn't implement VomEncode()", id)
	}
	results := m.Call(nil)
	if len(results) != 2 {
		return nil, fmt.Errorf("type %T has wrong number of return arguments for VomEncode()", id)
	}
	if !results[1].IsNil() {
		err, ok := results[1].Interface().(error)
		if !ok {
			return nil, fmt.Errorf("second return argument must be an error, got %T", results[1].Interface())
		}
		return nil, fmt.Errorf("error invoking VomEncode(): %v", err)
	}
	if results[0].IsNil() {
		return nil, fmt.Errorf("VomEncode() returned nil encoding value")
	}
	switch result := results[0].Interface().(type) {
	case *wire.ChainPublicID:
		return []wire.ChainPublicID{*result}, nil
	case []security.PublicID:
		var ret []wire.ChainPublicID
		for _, childID := range result {
			chains, err := getChains(childID)
			if err != nil {
				return nil, err
			}
			ret = append(ret, chains...)
		}
		return ret, nil
	default:
		return nil, fmt.Errorf("unexpected return value of type %T for VomEncode", result)
	}
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
func Java_io_veyron_veyron_veyron_runtimes_google_security_PrivateID_nativeBless(env *C.JNIEnv, jPrivateID C.jobject, goPrivateIDPtr C.jlong, jBlesseeChains C.jobjectArray, name C.jstring, jDuration C.jobject) C.jlong {
	blessee, err := idFromChains(env, jBlesseeChains)
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
func Java_io_veyron_veyron_veyron_runtimes_google_security_PrivateID_nativeDerive(env *C.JNIEnv, jPrivateID C.jobject, goPrivateIDPtr C.jlong, jPublicIDChains C.jobjectArray) C.jlong {
	publicID, err := idFromChains(env, jPublicIDChains)
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

func idFromChains(env *C.JNIEnv, jPublicIDChains C.jobjectArray) (security.PublicID, error) {
	chainStrs := util.GoStringArray(env, jPublicIDChains)
	if len(chainStrs) == 0 {
		return nil, fmt.Errorf("Empty public id chains")
	}
	// JSON-decode chains.
	chains := make([]wire.ChainPublicID, len(chainStrs))
	for i, str := range chainStrs {
		if err := json.Unmarshal([]byte(str), &chains[i]); err != nil {
			return nil, fmt.Errorf("Couldn't JSON-decode chain %q: %v", str, err)
		}
	}
	// Create PublicIDs.
	ids := make([]security.PublicID, len(chains))
	for i, chain := range chains {
		// Total hack to obtain a PublicID from wire.ChainPublicID.
		// TODO(spetrovic): make sure this goes away when we switch to Principal/Blessing API.
		var buf bytes.Buffer
		if err := vom.NewEncoder(&buf).Encode(&chain); err != nil {
			return nil, fmt.Errorf("Couldn't VOM-encode chain: %v", err)
		}
		privID, err := isecurity.NewPrivateID("dummy", nil)
		if err != nil {
			return nil, fmt.Errorf("Couldn't mint new private id")
		}
		ids[i] = privID.PublicID()
		if err := vom.NewDecoder(&buf).Decode(&ids[i]); err != nil {
			return nil, fmt.Errorf("Couldn't VOM-decode chain: %v", err)
		}
	}
	return isecurity.NewSetPublicID(ids...)
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_PrivateID_nativeFinalize
func Java_io_veyron_veyron_veyron_runtimes_google_security_PrivateID_nativeFinalize(env *C.JNIEnv, jPrivateID C.jobject, goPrivateIDPtr C.jlong) {
	util.GoUnref((*security.PrivateID)(util.Ptr(goPrivateIDPtr)))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeMethod
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeMethod(env *C.JNIEnv, jContext C.jobject, goContextPtr C.jlong) C.jstring {
	return C.jstring(util.JStringPtr(env, (*(*security.Context)(util.Ptr(goContextPtr))).Method()))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeName
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeName(env *C.JNIEnv, jServerCall C.jobject, goContextPtr C.jlong) C.jstring {
	return C.jstring(util.JStringPtr(env, (*(*security.Context)(util.Ptr(goContextPtr))).Name()))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeSuffix
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeSuffix(env *C.JNIEnv, jServerCall C.jobject, goContextPtr C.jlong) C.jstring {
	return C.jstring(util.JStringPtr(env, (*(*security.Context)(util.Ptr(goContextPtr))).Suffix()))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLabel
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLabel(env *C.JNIEnv, jServerCall C.jobject, goContextPtr C.jlong) C.jint {
	return C.jint((*(*security.Context)(util.Ptr(goContextPtr))).Label())
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLocalID
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLocalID(env *C.JNIEnv, jServerCall C.jobject, goContextPtr C.jlong) C.jlong {
	id := (*(*security.Context)(util.Ptr(goContextPtr))).LocalID()
	util.GoRef(&id) // Un-refed when the Java PublicID object is finalized.
	return C.jlong(util.PtrValue(&id))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeRemoteID
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeRemoteID(env *C.JNIEnv, jServerCall C.jobject, goContextPtr C.jlong) C.jlong {
	id := (*(*security.Context)(util.Ptr(goContextPtr))).RemoteID()
	util.GoRef(&id)
	return C.jlong(util.PtrValue(&id))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLocalEndpoint
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeLocalEndpoint(env *C.JNIEnv, jServerCall C.jobject, goContextPtr C.jlong) C.jstring {
	return C.jstring(util.JStringPtr(env, (*(*security.Context)(util.Ptr(goContextPtr))).LocalEndpoint().String()))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeRemoteEndpoint
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeRemoteEndpoint(env *C.JNIEnv, jServerCall C.jobject, goContextPtr C.jlong) C.jstring {
	return C.jstring(util.JStringPtr(env, (*(*security.Context)(util.Ptr(goContextPtr))).RemoteEndpoint().String()))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeFinalize
func Java_io_veyron_veyron_veyron_runtimes_google_security_Context_nativeFinalize(env *C.JNIEnv, jServerCall C.jobject, goContextPtr C.jlong) {
	util.GoUnref((*security.Context)(util.Ptr(goContextPtr)))
}
