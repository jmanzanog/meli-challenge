package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewServerFail(t *testing.T) {
	server := NewServer("4500")
	err := server.Start()

	if err == nil {
		t.Error("Start server should fail")
	}
}

func TestHandler(t *testing.T) {
	const path = "/mock"

	const content = "server running on :4500"

	const targetMock = "http://example.com/foo"

	handlerMock := func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, content) }
	req := httptest.NewRequest(GetMethod, targetMock, nil)
	w := httptest.NewRecorder()
	server := NewServer(ServerPortDefaultValue)
	server.Handle(path, GetMethod, handlerMock)
	handlerMock(w, req)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	if server.Router.rules == nil {
		t.Error("Register handler filed")
	}
	if server.Router.rules[path] == nil {
		t.Error("path /mock does not registered")
	}
	if server.Router.rules[path][GetMethod] == nil {
		t.Error("method GET does not registered")
	}
	if resp.StatusCode != 200 {
		t.Errorf("got %v, want %v", resp.StatusCode, 200)
	}
	if bodyResponse := string(body); bodyResponse != content {
		t.Errorf("got %v, want %v", bodyResponse, content)
	}
}
