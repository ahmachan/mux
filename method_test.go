// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func TestMethod_Add(t *testing.T) {
	a := assert.New(t)
	m := NewMethod()
	a.NotNil(m)

	// handler不能为空
	a.Error(m.Add("abc", nil, "GET"))

	fn := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})

	// methods为空
	a.Error(m.Add("abc", fn))

	a.NotError(m.Add("abc", fn, "GET", "POST"))
	_, found := m.entries["GET"]
	a.True(found)
	_, found = m.entries["POST"]
	a.True(found)
	_, found = m.entries["DELETE"]
	a.False(found)

	a.NotError(m.Get("def", fn))
	es, found := m.entries["GET"]
	a.True(found).Equal(2, len(es.list))
}

func TestMethod_ServeHTTP(t *testing.T) {
	a := assert.New(t)

	newMethod := func(pattern string) http.Handler {
		h := NewMethod()
		a.NotError(h.AddFunc(pattern, defaultHandler, "GET"))
		return h
	}

	tests := []*handlerTester{
		&handlerTester{
			name:       "普通匹配",
			h:          newMethod("/abc"),
			query:      "/abc",
			statusCode: 200,
		},
		&handlerTester{
			name:       "普通不匹配",
			h:          newMethod("/abc"),
			query:      "/abcd",
			statusCode: 404,
		},
		&handlerTester{
			name:       "正则匹配数字",
			h:          newMethod("?/api/(?P<version>\\d+)"),
			query:      "/api/2",
			statusCode: 200,
			ctxName:    "params",
			ctxMap:     map[string]string{"version": "2"},
		},
		&handlerTester{
			name:       "正则匹配多个名称",
			h:          newMethod("?/api/(?P<version>\\d+)/(?P<name>\\w+)"),
			query:      "/api/2/login",
			statusCode: 200,
			ctxName:    "params",
			ctxMap:     map[string]string{"version": "2", "name": "login"},
		},
		&handlerTester{
			name:       "正则不匹配多个名称",
			h:          newMethod("?/api/(?P<version>\\d+)/(?P<name>\\w+)"),
			query:      "/api/2.0/login",
			statusCode: 404,
		},
		&handlerTester{
			name:       "带域名的字符串不匹配", //无法匹配端口信息
			h:          newMethod("127.0.0.1/abc"),
			query:      "/abc",
			statusCode: 404,
		},
		&handlerTester{
			name:       "带域名的正则匹配", //无法匹配端口信息
			h:          newMethod("?127.0.0.1:\\d+/abc"),
			query:      "/abc",
			statusCode: 200,
		},
		&handlerTester{
			name:       "带域名的命名正则匹配", //无法匹配端口信息
			h:          newMethod("?127.0.0.1:\\d+/api/v(?P<version>\\d+)/login"),
			query:      "/api/v2/login",
			statusCode: 200,
			ctxName:    "params",
			ctxMap:     map[string]string{"version": "2"},
		},
	}

	runHandlerTester(a, tests)
}
