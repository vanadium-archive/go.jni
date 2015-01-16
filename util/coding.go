package util

import (
	"bytes"

	"v.io/core/veyron2/vdl"
	"v.io/core/veyron2/vom2"
)

// VomDecodeToValue VOM-decodes the provided value into *vdl.Value using a new
// instance of a VOM decoder.
func VomDecodeToValue(data []byte) (*vdl.Value, error) {
	decoder, err := vom2.NewDecoder(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	var value *vdl.Value
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

func VomCopy(src interface{}, dstptr interface{}) error {
	data, err := vom2.Encode(src)
	if err != nil {
		return err
	}
	return vom2.Decode(data, dstptr)
}
