package auth

import (
	"testing"
)

func TestCreateToken(t *testing.T) {
	// arrange
	user := "foo@bar.com"
	// act
	result := CreateToken(&user)
	// assert
	if len(result) == 0 {
		t.Error("Invalid token")
	}
}

var invalidTokenTable = []struct {
	actual string
}{
	{"Bearer some-token"},
	{"some-token"},
	{"some-token Bearer"},
}

func TestParseToken_InvalidToken(t *testing.T) {
	for _, entry := range invalidTokenTable {
		var _, err = ParseToken(entry.actual)

		if err == nil {
			t.Error("Expected to get an error")
		}
	}
}

func TestParseToken_ValidToken(t *testing.T) {
	// arrange
	user := "foo@bar.com"
	token := CreateToken(&user)
	authHeader := "Bearer " + token
	// act
	result, err := ParseToken(authHeader)
	// assert
	if err != nil {
		t.Fatal("Unexpected token: {0}", err)
	}

	if result.AuthToken != token {
		t.Error("Auth token is not set on user profile correctly")
	}

	if result.UserName != user {
		t.Error("Username is not set on user profile correctly")
	}
}

func TestRegisterAndRetriveClient(t *testing.T) {
	user := "foo@bar.com"
	var pubKey [32]byte

	for i := 0; i < len(pubKey); i++ {
		pubKey[i] = 1
	}

	RegisterClient(user, pubKey)
	result, err := RetrieveClient(user)

	if err != nil {
		t.Fatal("Could not retrive registered client. Error: {0}", err)
	}

	if result != pubKey {
		t.Error("Client was not registered correctly")
	}
}
