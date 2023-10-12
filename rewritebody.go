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
	"regexp"
)

// Rewrite holds one rewrite body configuration.
//type Rewrite struct {
//	Regex       string `json:"regex,omitempty"`
//	Replacement string `json:"replacement,omitempty"`
//}

type Path struct {
	NestedName string `json:"nestedName,omitempty"`
}

// Config holds the plugin configuration.
//type Config struct {
//	LastModified bool      `json:"lastModified,omitempty"`
//	Rewrites     []Rewrite `json:"rewrites,omitempty"`
//}

type Config struct {
	Paths []Path `json:"paths,omitempty"`
}

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

//type rewrite struct {
//	regex       *regexp.Regexp
//	replacement []byte
//}

type path struct {
	nestedName string
}

//
//type rewriteBody struct {
//	name         string
//	next         http.Handler
//	rewrites     []rewrite
//	lastModified bool
//}

type pathBody struct {
	next  http.Handler
	paths []path
}

// New creates and returns a new rewrite body plugin instance.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	paths := make([]path, len(config.Paths))

	for i, rewriteConfig := range config.Paths {
		paths[i] = path{
			nestedName: string(rewriteConfig.NestedName),
		}
	}

	return &pathBody{
		next:  next,
		paths: paths,
	}, nil
}

func (r *pathBody) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	wrappedWriter := &responseWriter{
		ResponseWriter: rw,
	}

	r.next.ServeHTTP(wrappedWriter, req)

	bodyBytes := wrappedWriter.buffer.Bytes()

	contentEncoding := wrappedWriter.Header().Get("Content-Encoding")

	if contentEncoding != "" && contentEncoding != "identity" {
		if _, err := rw.Write(bodyBytes); err != nil {
			log.Printf("unable to write body: %v", err)
		}

		return
	}

	var resp map[string]interface{}
	err := json.Unmarshal(bodyBytes, &resp)
	if err != nil {
		log.Printf("unable to write rewrited body: %v", err)
	}

	var jsonResp json.RawMessage
	for _, rwt := range r.paths {
		jsonResp, _ = json.Marshal(resp[rwt.nestedName])

	}

	if _, err := rw.Write(jsonResp); err != nil {
		log.Printf("unable to write rewrited body: %v", err)
	}
}

type responseWriter struct {
	buffer      bytes.Buffer
	wroteHeader bool

	http.ResponseWriter
}

func (r *responseWriter) WriteHeader(statusCode int) {

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
