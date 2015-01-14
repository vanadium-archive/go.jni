package security

import (
	"v.io/core/veyron2/security"
	"v.io/core/veyron2/vom2"
	jutil "v.io/jni/util"
)

// isJNICaveat returns true iff the provided caveat encodes a jniCaveat.
func isJNICaveat(caveat security.Caveat) (jniCaveat, bool) {
	var validator security.CaveatValidator
	if err := vom2.Decode(caveat.ValidatorVOM, &validator); err != nil {
		return jniCaveat{}, false
	}
	jni, ok := validator.(jniCaveat)
	return jni, ok
}

func (c jniCaveat) Validate(context security.Context) error {
	env, freeFunc := jutil.GetEnv(javaVM)
	defer freeFunc()
	jCaveat, err := JavaCaveat(env, c.Caveat)
	if err != nil {
		return err
	}
	jValidator, err := jutil.CallStaticObjectMethod(env, jCaveatCoderClass, "decode", []jutil.Sign{caveatSign}, caveatValidatorSign, jCaveat)
	if err != nil {
		return err
	}
	jContext, err := JavaContext(env, context)
	if err != nil {
		return err
	}
	if err := jutil.CallVoidMethod(env, jValidator, "validate", []jutil.Sign{contextSign}, jContext); err != nil {
		return err
	}
	return nil
}
