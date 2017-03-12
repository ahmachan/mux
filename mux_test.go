// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

// 一些预定义的处理函数
var (
	f1 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(1)
	}
	f2 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(2)
	}
	f3 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(3)
	}
	f4 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(4)
	}
	f5 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(5)
	}
	f6 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(6)
	}
)

func request(a *assert.Assertion, srvmux *ServeMux, method, url string, status int) {
	w := httptest.NewRecorder()
	a.NotNil(w)

	r, err := http.NewRequest(method, url, nil)
	a.NotError(err).NotNil(r)

	srvmux.ServeHTTP(w, r)
	a.Equal(w.Code, status)
}

func TestClearPath(t *testing.T) {
	a := assert.New(t)

	a.Equal(cleanPath(""), "/")
	a.Equal(cleanPath("/api//"), "/api/")

	a.Equal(cleanPath("/api/"), "/api/")
	a.Equal(cleanPath("/api/./"), "/api/")

	a.Equal(cleanPath("/api/.."), "/")
	a.Equal(cleanPath("/api/../"), "/")

	a.Equal(cleanPath("/api/../../"), "/")
	a.Equal(cleanPath("/api../"), "/api../")
}

func TestServeMux_Add_Remove_1(t *testing.T) {
	a := assert.New(t)

	srvmux := NewServeMux(false)
	a.NotNil(srvmux)

	// 添加 delete /api/1
	a.NotPanic(func() {
		srvmux.DeleteFunc("/api/1", f1)
	})
	a.Equal(srvmux.entries.Len(), 1)

	// 添加 patch /api/1
	a.NotPanic(func() {
		srvmux.PatchFunc("/api/1", f1)
	})
	a.Equal(srvmux.entries.Len(), 1) // 加在同一个 Entry 下，所以数量不变

	// 添加 post /api/2
	a.NotPanic(func() {
		srvmux.PostFunc("/api/2", f1)
	})
	a.Equal(srvmux.entries.Len(), 2)

	// 删除 any /api/2
	srvmux.Remove("/api/2")
	a.Equal(srvmux.entries.Len(), 1)

	// 删除 delete /api/1
	srvmux.Remove("/api/1", http.MethodDelete)
	a.Equal(srvmux.entries.Len(), 1)

	// 删除 patch /api/1
	srvmux.Remove("/api/1", http.MethodPatch)
	a.Equal(srvmux.entries.Len(), 0)
}

func TestServeMux_Add_Remove_2(t *testing.T) {
	a := assert.New(t)
	srvmux := NewServeMux(false)
	a.NotNil(srvmux)

	// 添加 GET /api/1
	// 添加 PUT /api/1
	// 添加 GET /api/2
	a.NotError(srvmux.AddFunc("/api/1", f1, http.MethodGet))
	a.NotPanic(func() {
		srvmux.PutFunc("/api/1", f1)
	})
	a.NotPanic(func() {
		srvmux.GetFunc("/api/2", f2)
	})
	request(a, srvmux, http.MethodGet, "/api/1", 1)
	request(a, srvmux, http.MethodPut, "/api/1", 1)
	request(a, srvmux, http.MethodGet, "/api/2", 2)
	request(a, srvmux, http.MethodDelete, "/api/1", http.StatusMethodNotAllowed) // 未实现

	// 删除 GET /api/1
	srvmux.Remove("/api/1", http.MethodGet)
	request(a, srvmux, http.MethodGet, "/api/1", http.StatusMethodNotAllowed)
	request(a, srvmux, http.MethodPut, "/api/1", 1) // 不影响 PUT
	request(a, srvmux, http.MethodGet, "/api/2", 2)

	// 删除 GET /api/2，只有一个，所以相当于整个 Entry 被删除
	srvmux.Remove("/api/2", http.MethodGet)
	request(a, srvmux, http.MethodGet, "/api/1", http.StatusMethodNotAllowed)
	request(a, srvmux, http.MethodPut, "/api/1", 1)                   // 不影响 PUT
	request(a, srvmux, http.MethodGet, "/api/2", http.StatusNotFound) // 整个 entry 被删除

	// 添加 POST /api/1
	a.NotPanic(func() {
		srvmux.PostFunc("/api/1", f1)
	})
	request(a, srvmux, http.MethodPost, "/api/1", 1)

	// 删除 ANY /api/1
	srvmux.Remove("/api/1")
	request(a, srvmux, http.MethodPost, "/api/1", http.StatusNotFound) // 404 表示整个 entry 都没了
}

func TestServeMux_Options(t *testing.T) {
	a := assert.New(t)
	srvmux := NewServeMux(false)
	a.NotNil(srvmux)

	// TODO
}

func TestServeMux_Params(t *testing.T) {
	a := assert.New(t)
	srvmux := NewServeMux(false)
	a.NotNil(srvmux)

	// TODO
}

// 测试匹配顺序是否正确
func TestServeMux_ServeHTTP_Order(t *testing.T) {
	a := assert.New(t)
	serveMux := NewServeMux(false)
	a.NotNil(serveMux)

	a.NotError(serveMux.AddFunc("/post/", f1, "GET"))          // f1
	a.NotError(serveMux.AddFunc("/post/{id:\\d+}", f2, "GET")) // f2
	a.NotError(serveMux.AddFunc("/post/1", f3, "GET"))         // f3

	request(a, serveMux, "GET", "/post/1", 3)   // f3 静态路由项完全匹配
	request(a, serveMux, "GET", "/post/2", 2)   // f2 正则完全匹配
	request(a, serveMux, "GET", "/post/abc", 1) // f1 匹配度最高
}

func TestMethodIsSupported(t *testing.T) {
	a := assert.New(t)

	a.True(MethodIsSupported("get"))
	a.True(MethodIsSupported("POST"))
	a.False(MethodIsSupported("not exists"))
}
