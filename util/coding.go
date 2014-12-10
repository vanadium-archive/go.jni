package util

import (
	"bytes"

	"veyron.io/veyron/veyron2/vdl"
	"veyron.io/veyron/veyron2/vom2"
)

// VomEncode encodes the provided value using a new instance of a VOM encoder.
func VomEncode(value interface{}) ([]byte, error) {
	var buf bytes.Buffer
	encoder, err := vom2.NewBinaryEncoder(&buf)
	if err != nil {
		return nil, err
	}
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// VomDecode VOM-decodes the given data into the provided value using a new
// instance of a VOM decoder.
func VomDecode(data []byte, valptr interface{}) error {
	decoder, err := vom2.NewDecoder(bytes.NewReader(data))
	if err != nil {
		return err
	}
	return decoder.Decode(valptr)
}

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
