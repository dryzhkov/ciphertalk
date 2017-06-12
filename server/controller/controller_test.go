package controller

import (
	"bytes"
	"ciphertalk/common/models"
	"ciphertalk/server/auth"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockResponseWriter struct {
	header http.Header
}

func (mrw MockResponseWriter) Header() http.Header {
	return mrw.header
}

func (mrw MockResponseWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

func (mrw MockResponseWriter) WriteHeader(code int) {
}

func TestLogin(t *testing.T) {
	// arrange
	var userKey [32]byte
	var controller APIController
	var responseWriter MockResponseWriter
	responseWriter.header = make(map[string][]string)

	user := "foo@bar.com"
	loginReq := models.LoginRequest{UserName: user, PublicKey: userKey}
	payload, _ := json.Marshal(loginReq)
	httpRequest := httptest.NewRequest("GET", "/login", bytes.NewReader(payload))

	// act
	controller.Login(responseWriter, httpRequest)

	// assert
	if responseWriter.header.Get("Content-Type") != "application/json; charset=UTF-8" {
		t.Error("Content-Type header was not set")
	}
}

func TestLogin_BadRequest(t *testing.T) {
	var invalidLoginTable [][]byte
	body, _ := json.Marshal(models.LoginRequest{UserName: "", PublicKey: [32]byte{}})
	invalidLoginTable = append(invalidLoginTable, body)
	body = make([]byte, 1)
	invalidLoginTable = append(invalidLoginTable, body)

	for _, entry := range invalidLoginTable {
		// arrange
		var controller APIController
		req := httptest.NewRequest("GET", "/login", bytes.NewReader(entry))

		rr := httptest.NewRecorder()
		// act
		controller.Login(rr, req)

		// assert
		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Unexpected response. Actual: %v Expected: %v", status, http.StatusBadRequest)
		}

		if body := rr.Body.String(); body == "" {
			t.Error("Response body shouldn't be empty")
		}
	}
}

func TestSecureChannel_BadRequest(t *testing.T) {
	// arrange
	var controller APIController
	wr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/body", nil)
	// act
	controller.SecureChannel(wr, req)
	// assert
	if wr.Code != http.StatusBadRequest {
		t.Errorf("Unexpected status code. expected: %v, actual %v", http.StatusBadRequest, wr.Code)
	}
}

func TestSecureChannel_NotFound(t *testing.T) {
	// arrange
	var controller APIController
	wr := httptest.NewRecorder()
	payload := []byte("{\"userName\":\"foo\"}")
	req := httptest.NewRequest("GET", "/body", bytes.NewReader(payload))
	// act
	controller.SecureChannel(wr, req)
	// assert
	if wr.Code != http.StatusNotFound {
		t.Errorf("Unexpected status code. expected: %v, actual %v", http.StatusNotFound, wr.Code)
	}
}

func TestSecureChannel(t *testing.T) {
	// arrange
	var controller APIController
	wr := httptest.NewRecorder()
	payload := []byte("{\"userName\":\"foo\"}")
	req := httptest.NewRequest("GET", "/body", bytes.NewReader(payload))
	auth.RegisterClient("foo", [32]byte{})
	// act
	controller.SecureChannel(wr, req)
	// assert
	if wr.Code != http.StatusOK {
		t.Errorf("Unexpected status code. expected: %v, actual %v", http.StatusOK, wr.Code)
	}
}
