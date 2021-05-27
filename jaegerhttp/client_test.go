package jaegerhttp

import (
	"bytes"
	"fmt"
	jaegerModule "github.com/blinkbean/jaeger-module"
	"github.com/opentracing/opentracing-go"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeRequest(t *testing.T, url string) {

	client := WrapClient(&http.Client{})
	req, err := http.NewRequest("GET", url, nil)
	span, ctx := opentracing.StartSpanFromContext(req.Context(), "top1")
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(ctx)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	span.Finish()
}

func TestHttpClientTrace(t *testing.T) {
	closer := jaegerModule.InitJaeger("jaeger_http2")
	defer closer.Close()
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(writer http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ok", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("/fail", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "failure", http.StatusInternalServerError)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	tests := []struct {
		url string
	}{
		{url: "/ok"},
		{url: "/redirect"},
		{url: "/fail"},
		{url: "/ok"},
	}
	for _, tt := range tests {
		t.Log(tt.url)
		makeRequest(t, srv.URL+tt.url)
	}
}

func TestWriteCloserFromRequest(t *testing.T) {
	closer := jaegerModule.InitJaeger("jaeger_http3")
	defer closer.Close()
	wait := make(chan bool, 0)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			wait <- true
		}()

		w.Header().Set("Upgrade", "websocket")
		w.Header().Set("Connection", "Upgrade")
		w.WriteHeader(http.StatusSwitchingProtocols)

		hijacker := w.(http.Hijacker)
		_, rw, err := hijacker.Hijack()

		if err != nil {
			t.Fatal("Failed to hijack connection")
		}

		line, _, err := rw.ReadLine()
		if string(line) != "ping" {
			t.Fatalf("Expected 'ping' received %q", string(line))
		}

		if err != nil {
			t.Fatal(err)
		}
	}))

	var buf bytes.Buffer
	req, err := http.NewRequest("POST", srv.URL, &buf)
	req.Header.Set("Connection", "upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Proto = "HTTP/1.1"
	req.ProtoMajor = 1
	req.ProtoMinor = 1
	if err != nil {
		t.Fatal(err)
	}
	span, ctx := opentracing.StartSpanFromContext(req.Context(), "top1")
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(ctx)
	client := WrapClient(&http.Client{})
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	rw, ok := resp.Body.(io.ReadWriteCloser)
	if !ok {
		t.Fatal("resp.Body is not a io.ReadWriteCloser")
	}

	fmt.Fprint(rw, "ping\n")
	<-wait
	rw.Close()
	span.Finish()
}
