package helper

import "testing"

func TestGenKey(t *testing.T) {
	prv1, _ := GenKey("")
	prv2, _ := LoadKey("")
	b1,_ := prv1.Bytes()
	b2,_ := prv2.Bytes()
	t.Log(string(b1)==string(b2))
}
