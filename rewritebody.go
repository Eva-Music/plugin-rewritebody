// Package plugin_rewritebody a plugin to rewrite response body.
package plugin_rewritebody

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
)

// Rewrite holds one rewrite body configuration.
type Config struct {
	obj string `json:"obj,omitempty"`
}

// Config holds the plugin configuration.
//type Config struct {
//	LastModified bool      `json:"lastModified,omitempty"`
//	Rewrites     []Rewrite `json:"rewrites,omitempty"`
//}

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

type rewriteBody struct {
	next   http.Handler
	config Config
}

// New creates and returns a new rewrite body plugin instance.
func New(_ context.Context, next http.Handler, config *Config, _ string) (http.Handler, error) {
	return &rewriteBody{
		next:   next,
		config: *config,
	}, nil
}

func (r *rewriteBody) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	wrappedWriter := &responseWriter{
		ResponseWriter: rw,
	}

	r.next.ServeHTTP(wrappedWriter, req)

	bodyBytes := wrappedWriter.buffer.Bytes()

	//contentEncoding := wrappedWriter.Header().Get("Content-Encoding")
	//
	//if contentEncoding != "" && contentEncoding != "identity" {
	//	if _, err := rw.Write(bodyBytes); err != nil {
	//		log.Printf("unable to write body: %v", err)
	//	}
	//
	//	return
	//}

	var resp map[string]interface{}
	err := json.Unmarshal(bodyBytes, &resp)
	if err != nil {
		log.Printf("unable to write rewrited body: %v", err)
	}
	jsonResp, _ := json.Marshal(resp[r.config.obj])
	rw.Write(jsonResp)
}

type responseWriter struct {
	buffer       bytes.Buffer
	lastModified bool
	wroteHeader  bool

	http.ResponseWriter
}

func (r *responseWriter) WriteHeader(statusCode int) {
	if !r.lastModified {
		r.ResponseWriter.Header().Del("Last-Modified")
	}

	r.wroteHeader = true

	// Delegates the Content-Length Header creation to the final body write.
	r.ResponseWriter.Header().Del("Content-Length")

	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseWriter) Write(p []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}

	return r.buffer.Write(p)
}

func (r *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("%T is not a http.Hijacker", r.ResponseWriter)
	}

	return hijacker.Hijack()
}

func (r *responseWriter) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
