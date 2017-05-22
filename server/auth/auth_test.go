package auth

import "testing"

func TestParseToken(t *testing.T) {
	var _, err = ParseToken("fdasfdas")

	if err != nil {
		t.Error("Expected user, got ", err)
	}
}
