package conn

import "testing"

func TestChacha20Decode(t *testing.T) {

	msg := "admin:Yamei-cjds1023@tcp(10.170.0.2:3306)/win88?charset=utf8&parseTime=True"
	pass := "0Yoo4LxSbMo64MbD3G4ICjnU"

	ciphertext := Chacha20Encode(msg, pass)

	t.Log(ciphertext)

	dst, err := Chacha20Decode(ciphertext, pass)
	if err != nil {
		t.Error(err)
	}

	t.Log(string(dst))
}
