// +build android

package ipc

import (
	"encoding/json"
	"fmt"

	jutil "veyron.io/jni/util"
	"veyron.io/veyron/veyron2/ipc"
)

// #include <stdlib.h>
// #include <jni.h>
import "C"

func newStream(s ipc.Stream, mArgs *methodArgs) stream {
	return stream{
		stream: s,
		mArgs:  mArgs,
	}
}

type stream struct {
	stream ipc.Stream
	mArgs  *methodArgs
}

func (s *stream) Send(env *C.JNIEnv, jItem C.jstring) error {
	argStr := jutil.GoString(env, jItem)
	argptr := s.mArgs.StreamSendPtr()
	if argptr == nil {
		return fmt.Errorf("nil stream input argument, expected a non-nil type for argument %q", argStr)
	}
	if err := json.Unmarshal([]byte(argStr), argptr); err != nil {
		return err
	}
	return s.stream.Send(jutil.DerefOrDie(argptr))
}

func (s *stream) Recv(env *C.JNIEnv) (C.jstring, error) {
	argptr := s.mArgs.StreamRecvPtr()
	if argptr == nil {
		return nil, fmt.Errorf("nil stream output argument")
	}
	if err := s.stream.Recv(argptr); err != nil {
		return nil, err
	}
	// JSON encode the result.
	result, err := json.Marshal(jutil.DerefOrDie(argptr))
	if err != nil {
		return nil, err
	}
	return C.jstring(jutil.JString(env, string(result))), nil
}
