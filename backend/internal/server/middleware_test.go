package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS_SetsHeaders(t *testing.T) {
	handler := cors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/chat", nil)
	handler.ServeHTTP(w, r)

	h := w.Header()
	if h.Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("expected '*', got %q", h.Get("Access-Control-Allow-Origin"))
	}
	if h.Get("Access-Control-Allow-Methods") != "GET, POST, OPTIONS" {
		t.Fatalf("unexpected methods: %q", h.Get("Access-Control-Allow-Methods"))
	}
	if h.Get("Access-Control-Allow-Headers") != "Content-Type" {
		t.Fatalf("unexpected headers: %q", h.Get("Access-Control-Allow-Headers"))
	}
}

func TestCORS_OptionsReturnsNoContent(t *testing.T) {
	handler := cors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called for OPTIONS")
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("OPTIONS", "/chat", nil)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestCORS_ProxiesGET(t *testing.T) {
	called := false
	handler := cors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/chat", nil)
	handler.ServeHTTP(w, r)

	if !called {
		t.Fatal("expected next handler to be called")
	}
}

func TestLogRequest_ResponseWriterWrapsStatus(t *testing.T) {
	handler := logRequest(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/chat", nil)
	handler(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected response status 404, got %d", w.Code)
	}
}

func TestResponseWriter_Flush(t *testing.T) {
	rw := &responseWriter{ResponseWriter: httptest.NewRecorder()}
	rw.WriteHeader(http.StatusOK)
	rw.Flush()
	if rw.status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rw.status)
	}
}

func TestResponseWriter_FlushPassthrough(t *testing.T) {
	rw := &responseWriter{ResponseWriter: httptest.NewRecorder()}
	rw.Flush()
}
