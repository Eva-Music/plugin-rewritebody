// Package plugin_rewritebody a plugin to rewrite response body.
package plugin_rewritebody

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/tidwall/gjson"
	"log"
	"net"
	"net/http"
)

// Rewrite holds one rewrite body configuration.
//type Rewrite struct {
//	Regex       string `json:"regex,omitempty"`
//	Replacement string `json:"replacement,omitempty"`
//}

//type Path struct {
//	NestedName string `json:"nestedName,omitempty"`
//}

// Config holds the plugin configuration.
//type Config struct {
//	LastModified bool      `json:"lastModified,omitempty"`
//	Rewrites     []Rewrite `json:"rewrites,omitempty"`
//}

type Config struct {
	Path string `json:"path,omitempty"`
}

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

//type rewrite struct {
//	regex       *regexp.Regexp
//	replacement []byte
//}

//type path struct {
//	nestedName string
//}

//
//type rewriteBody struct {
//	name         string
//	next         http.Handler
//	rewrites     []rewrite
//	lastModified bool
//}

type rewrite struct {
	next   http.Handler
	config Config
}

// New creates and returns a new rewrite body plugin instance.
func New(_ context.Context, next http.Handler, config *Config, _ string) (http.Handler, error) {
	//paths := make([]path, len(config.Paths))
	//
	//for i, rewriteConfig := range config.Paths {
	//	paths[i] = path{
	//		nestedName: string(rewriteConfig.NestedName),
	//	}
	//}

	return &rewrite{
		next:   next,
		config: *config,
	}, nil
}

func (r *rewrite) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	wrappedWriter := &responseWriter{
		body:           &bytes.Buffer{},
		ResponseWriter: rw,
	}

	r.next.ServeHTTP(wrappedWriter, req)

	bodyBytes := wrappedWriter.body.Bytes()

	//contentEncoding := wrappedWriter.Header().Get("Content-Encoding")
	//
	//if contentEncoding != "" && contentEncoding != "identity" {
	//	if _, err := rw.Write(bodyBytes); err != nil {
	//		log.Printf("unable to write body: %v", err)
	//	}
	//
	//	return
	//}

	//resp := make(map[string]interface{})
	//err := json.Unmarshal(bodyBytes, &resp)
	//if err != nil {
	//	log.Printf("unable to write rewrited body: %v", err)
	//}

	//var jsonResp json.RawMessage
	//for _, rwt := range r.paths {
	//	jsonResp, _ = json.Marshal(resp[rwt.nestedName])
	//
	//}

	result := gjson.GetBytes(bodyBytes, r.config.Path)
	if _, err := rw.Write([]byte(result.Raw)); err != nil {
		log.Printf("unable to write rewrited body: %v", err)
	}
}

type responseWriter struct {
	body *bytes.Buffer
	http.ResponseWriter
}

func (r responseWriter) Write(b []byte) (int, error) {
	r.ResponseWriter.WriteHeader(http.StatusOK)
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

//func (r *responseWriter) Write(p []byte) (int, error) {
//	if !r.wroteHeader {
//		r.WriteHeader(http.StatusOK)
//	}
//
//	return r.buffer.Write(p)
//}

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
