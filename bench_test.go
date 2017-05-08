// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func benchHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("handler"))
}

func BenchmarkMux_ServeHTTPBasic(b *testing.B) {
	a := assert.New(b)
	srv := New(false, false, nil, nil)

	srv.GetFunc("/blog/post/1", benchHandler)  // 1
	srv.GetFunc("/blog/posts/*", benchHandler) // 2
	srv.GetFunc("/api/v2/login", benchHandler) // 3

	r1, err := http.NewRequest("GET", "/blog/post/1", nil) // 1
	a.NotError(err).NotNil(r1)
	r2, err := http.NewRequest("GET", "/api/v2/login", nil)
	a.NotError(err).NotNil(r2)
	r3, err := http.NewRequest("GET", "/api/v2x/login", nil)
	a.NotError(err).NotNil(r3)
	r4, err := http.NewRequest("GET", "/blog/posts/4", nil) // 2
	a.NotError(err).NotNil(r4)
	reqs := []*http.Request{r1, r2, r3, r4}

	w := httptest.NewRecorder()

	srvfun := func(reqIndex int) {
		srv.ServeHTTP(w, reqs[reqIndex])
	}
	for i := 0; i < b.N; i++ {
		srvfun(i % len(reqs))
	}
}

func BenchmarkMux_ServeHTTPNamed(b *testing.B) {
	a := assert.New(b)
	srv := New(false, false, nil, nil)

	srv.GetFunc("/blog/post/{id}", benchHandler)   // 1
	srv.GetFunc("/blog/tags/{id}/*", benchHandler) // 2
	srv.GetFunc("/api/v2/", benchHandler)          // 3

	r1, err := http.NewRequest("GET", "/blog/post/1", nil) // 1
	a.NotError(err).NotNil(r1)
	r2, err := http.NewRequest("GET", "/api/v2/login", nil)
	a.NotError(err).NotNil(r2)
	r3, err := http.NewRequest("GET", "/api/v2x/login", nil)
	a.NotError(err).NotNil(r3)
	r4, err := http.NewRequest("GET", "/blog/tags/5/list", nil) // 2
	a.NotError(err).NotNil(r4)
	reqs := []*http.Request{r1, r2, r3, r4}

	w := httptest.NewRecorder()

	srvfun := func(reqIndex int) {
		srv.ServeHTTP(w, reqs[reqIndex])
	}
	for i := 0; i < b.N; i++ {
		srvfun(i % len(reqs))
	}
}

func BenchmarkMux_ServeHTTPRegexp(b *testing.B) {
	a := assert.New(b)
	srv := New(false, false, nil, nil)

	srv.GetFunc("/blog/post/{id:\\d+}/*", benchHandler)     // 1
	srv.GetFunc("/api/v{version:\\d+}/login", benchHandler) // 2

	r1, err := http.NewRequest("GET", "/blog/post/1/list", nil) // 1
	a.NotError(err).NotNil(r1)
	r2, err := http.NewRequest("GET", "/api/v2/login", nil) // 2
	a.NotError(err).NotNil(r2)
	r3, err := http.NewRequest("GET", "/api/v2x/login", nil)
	a.NotError(err).NotNil(r3)
	reqs := []*http.Request{r1, r2, r3}

	w := httptest.NewRecorder()

	srvfun := func(reqIndex int) {
		srv.ServeHTTP(w, reqs[reqIndex])
	}
	for i := 0; i < b.N; i++ {
		srvfun(i % len(reqs))
	}
}

func BenchmarkMux_ServeHTTPAll(b *testing.B) {
	a := assert.New(b)
	srv := New(false, false, nil, nil)

	srv.GetFunc("/blog/basic/1", benchHandler)
	srv.GetFunc("/blog/{id}/*", benchHandler)
	srv.GetFunc("/api/v{version:\\d+}/login", benchHandler)

	r1, err := http.NewRequest("GET", "/blog/1/list", nil)
	a.NotError(err).NotNil(r1)
	r2, err := http.NewRequest("GET", "/blog/basic/1", nil)
	a.NotError(err).NotNil(r2)
	r3, err := http.NewRequest("GET", "/api/v2/login", nil)
	a.NotError(err).NotNil(r3)
	r4, err := http.NewRequest("GET", "/api/v2x/login", nil)
	a.NotError(err).NotNil(r4)
	reqs := []*http.Request{r1, r2, r3, r4}

	w := httptest.NewRecorder()

	srvfun := func(reqIndex int) {
		srv.ServeHTTP(w, reqs[reqIndex])
	}
	for i := 0; i < b.N; i++ {
		srvfun(i % len(reqs))
	}
}

func BenchmarkCleanPath(b *testing.B) {
	a := assert.New(b)

	paths := []string{
		"/api//",
		"api//",
		"/api/",
		"/api/./",
		"/api/..",
		"/api/../",
		"/api/../../",
		"/api../",
	}

	for i := 0; i < b.N; i++ {
		ret := cleanPath(paths[i%len(paths)])
		a.True(len(ret) > 0)
	}
}
