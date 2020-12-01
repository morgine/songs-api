package crypt

import (
	"testing"
)

func TestAesCBCDecrypt(t *testing.T) {
	key := []byte("change this pass")
	message := []byte("hello world!")
	data, err := AesCBCEncrypt(message, key)
	if err != nil {
		t.Fatal(err)
	}
	decrypt, err := AesCBCDecrypt(data, key)
	if err != nil {
		t.Fatal(err)
	}
	if need, got := string(message), string(decrypt); need != got {
		t.Errorf("need: %s, got: %s\n", need, got)
	}
}
