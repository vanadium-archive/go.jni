package util

import "testing"

func TestSigns(t *testing.T) {
	tests := []struct {
		input  Sign
		output string
	}{
		{ClassSign("java.lang.Object"), "Ljava/lang/Object;"},
		{ClassSign("java.lang.String"), "Ljava/lang/String;"},
		{ArraySign(ObjectSign), "[Ljava/lang/Object;"},
		{ArraySign(StringSign), "[Ljava/lang/String;"},
		{ArraySign(IntSign), "[I"},
		{FuncSign(nil, IntSign), "()I"},
		{FuncSign([]Sign{}, IntSign), "()I"},
		{FuncSign([]Sign{BoolSign}, VoidSign), "(Z)V"},
		{FuncSign([]Sign{CharSign, ByteSign, ShortSign}, FloatSign), "(CBS)F"},
		{FuncSign([]Sign{ClassSign("io.veyron.veyron.veyron.testing.misc")}, ClassSign("io.veyron.veyron.veyron.ret")), "(Lio/veyron/veyron/veyron/testing/misc;)Lio/veyron/veyron/veyron/ret;"},
		{FuncSign([]Sign{ClassSign("io.veyron.veyron.veyron.testing.misc"), ClassSign("other")}, ClassSign("io.veyron.veyron.veyron.ret")), "(Lio/veyron/veyron/veyron/testing/misc;Lother;)Lio/veyron/veyron/veyron/ret;"},
	}
	for _, test := range tests {
		output := string(test.input)
		if output != test.output {
			t.Errorf("expected %v, got %v", test.output, output)
		}
	}
}
