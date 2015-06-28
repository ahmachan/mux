// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"container/list"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/issue9/context"
)

// 支持的所有请求方法
var supportMethods = []string{
	"GET",
	"POST",
	"HEAD",
	"DELETE",
	"PUT",
	"OPTIONS",
	"TRACE",
	"PATCH",
}

// http.ServeMux的升级版，可处理对URL的正则匹配和根据METHOD进行过滤。
//
// 用法如下：
//  m := mux.NewServeMux()
//  m.Get("www.example.com/abc", h1).              // 只匹配www.example.com域名下的路径
//    Post("/abc/"", h2).                          // 不限定域名的路径匹配
//    Add("/api/{version:\\d+}",h3, "GET", "POST") // 只匹配GET和POST
//  http.ListenAndServe(m)
//
//
// 路由参数：
//
// 路由参数可通过context包获取：
//  ctx := context.Get(req)
//  params := ctx.Get("params") // 若不存在路由参数，则返回一个空值
// NOTE:记得在退出整个请求之前清除context中的内容：
//  context.Free(req)
//
//
// 匹配规则：
//
// 可能会出现多条记录与同一请求都匹配的情况，这种情况下，
// 系统会找到一条认为最匹配的路由来处理，判断规则如下：
//  1.当只有部分匹配时，以匹配字符最多的项为准。
//  2.当有多条完全匹配时，以静态路由优先。
//
// 正则匹配语法：
//  /post/{id}     // 匹配/post/开头的任意字符串，其后的字符串保存到id中；
//  /post/{id:\d+} // 同上，但id的值只能为\d+；
//  /post/{:\d+}   // 同上，但是没有命名；
type ServeMux struct {
	sync.Mutex

	// 路由列表，键名表示method。list中静态路由在前，正则路由在后。
	list map[string]*list.List

	// 路由的命名列表，方便查找。
	named map[string]map[string]*entry
}

// 声明一个新的ServeMux
func NewServeMux() *ServeMux {
	l := make(map[string]*list.List, len(supportMethods))
	n := make(map[string]map[string]*entry, len(supportMethods))
	for _, method := range supportMethods {
		l[method] = list.New()
		n[method] = map[string]*entry{}
	}

	return &ServeMux{
		list:  l,
		named: n,
	}
}

// 添加一个路由项。
func (mux *ServeMux) add(g *Group, pattern string, h http.Handler, methods ...string) *ServeMux {
	if h == nil {
		panic("add:参数h不能为空")
	}

	if len(pattern) == 0 {
		panic("add:pattern匹配内容不能为空")
	}

	if len(methods) == 0 {
		methods = supportMethods
	}

	e := newEntry(pattern, h, g)

	mux.Lock()
	defer mux.Unlock()

	for _, method := range methods {
		method = strings.ToUpper(method)

		es, found := mux.named[method]
		if !found {
			panic(fmt.Sprintf("add:不支持的request.Method:[%v]", method))
		}

		if _, found := es[pattern]; found {
			panic("add:该模式的路由项已经存在")
		}

		es[pattern] = e
		if e.expr == nil { // 静态路由，在前端插入
			mux.list[method].PushFront(e)
		} else { // 正则路由，在后端插入
			mux.list[method].PushBack(e)
		}
	}

	return mux
}

// 添加一条路由数据。
//
// pattern为路由匹配模式，可以是正则匹配也可以是字符串匹配，
// 可以带上域名，当第一个字符为'/'当作是一个路径，否则就将'/'之前的当作域名或IP。
// methods参数应该只能为supportMethods中的字符串，若不指定，默认为所有，
// 当h或是pattern为空时，将触发panic。
func (mux *ServeMux) Add(pattern string, h http.Handler, methods ...string) *ServeMux {
	return mux.add(nil, pattern, h, methods...)
}

// Get相当于ServeMux.Add(pattern, h, "GET")的简易写法
func (mux *ServeMux) Get(pattern string, h http.Handler) *ServeMux {
	return mux.Add(pattern, h, "GET")
}

// Post相当于ServeMux.Add(pattern, h, "POST")的简易写法
func (mux *ServeMux) Post(pattern string, h http.Handler) *ServeMux {
	return mux.Add(pattern, h, "POST")
}

// Delete相当于ServeMux.Add(pattern, h, "DELETE")的简易写法
func (mux *ServeMux) Delete(pattern string, h http.Handler) *ServeMux {
	return mux.Add(pattern, h, "DELETE")
}

// Put相当于ServeMux.Add(pattern, h, "PUT")的简易写法
func (mux *ServeMux) Put(pattern string, h http.Handler) *ServeMux {
	return mux.Add(pattern, h, "PUT")
}

// Patch相当于ServeMux.Add(pattern, h, "PATCH")的简易写法
func (mux *ServeMux) Patch(pattern string, h http.Handler) *ServeMux {
	return mux.Add(pattern, h, "PATCH")
}

// Any相当于ServeMux.Add(pattern, h)的简易写法
func (mux *ServeMux) Any(pattern string, h http.Handler) *ServeMux {
	return mux.Add(pattern, h)
}

func (mux *ServeMux) addFunc(g *Group, pattern string, fun func(http.ResponseWriter, *http.Request), methods ...string) *ServeMux {
	return mux.add(g, pattern, http.HandlerFunc(fun), methods...)
}

// 功能同ServeMux.Add()，但是将第二个参数从http.Handler换成了func(http.ResponseWriter, *http.Request)
func (mux *ServeMux) AddFunc(pattern string, fun func(http.ResponseWriter, *http.Request), methods ...string) *ServeMux {
	return mux.Add(pattern, http.HandlerFunc(fun), methods...)
}

// GetFunc相当于ServeMux.AddFunc(pattern, func, "GET")的简易写法
func (mux *ServeMux) GetFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.AddFunc(pattern, fun, "GET")
}

// PutFunc相当于ServeMux.AddFunc(pattern, func, "PUT")的简易写法
func (mux *ServeMux) PutFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.AddFunc(pattern, fun, "PUT")
}

// PostFunc相当于ServeMux.AddFunc(pattern, func, "POST")的简易写法
func (mux *ServeMux) PostFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.AddFunc(pattern, fun, "POST")
}

// DeleteFunc相当于ServeMux.AddFunc(pattern, func, "DELETE")的简易写法
func (mux *ServeMux) DeleteFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.AddFunc(pattern, fun, "DELETE")
}

// PatchFunc相当于ServeMux.AddFunc(pattern, func, "PATCH")的简易写法
func (mux *ServeMux) PatchFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.AddFunc(pattern, fun, "PATCH")
}

// AnyFunc相当于ServeMux.AddFunc(pattern, func)的简易写法
func (mux *ServeMux) AnyFunc(pattern string, fun func(http.ResponseWriter, *http.Request)) *ServeMux {
	return mux.AddFunc(pattern, fun)
}

// 移除指定的路由项，通过路由表达式和method来匹配。
// 当未指定methods时，将删除所有method匹配的项。
// 指定错误的method值，将自动忽略该值。
func (mux *ServeMux) Remove(pattern string, methods ...string) {
	if len(methods) == 0 { // 删除所有method下匹配的项
		methods = supportMethods
	}

	mux.Lock()
	defer mux.Unlock()

	for _, method := range methods {
		es, found := mux.named[method]
		if !found {
			continue
		}

		if _, found := es[pattern]; !found {
			continue
		}

		delete(es, pattern)
		for item := mux.list[method].Front(); item != nil; item = item.Next() {
			e := item.Value.(*entry)
			if e.pattern == pattern {
				mux.list[method].Remove(item)
				break
			}
		}
	}
}

// implement http.Handler.ServerHTTP()
func (mux *ServeMux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	hostURL := req.Host + req.URL.Path
	size := -1
	var e *entry
	var p string

	mux.Lock()
	for item := mux.list[req.Method].Front(); item != nil; item = item.Next() {
		entry := item.Value.(*entry)
		url := req.URL.Path
		if entry.pattern[0] != '/' {
			url = hostURL
		}

		s := entry.match(url)
		if s == -1 || (size > 0 && s > size) { // 完全不匹配
			continue
		}

		size = s
		e = entry
		p = url

		if s == 0 { // 完全匹配，可以中止匹配过程
			break
		}
	}
	mux.Unlock() // 没必要等ServeHTTP执行完之后才解锁。

	if size < 0 {
		panic(fmt.Sprintf("没有找到与之前匹配的路径，Host:[%v],Path:[%v]", req.Host, req.URL.Path))
	}

	if e.group != nil && !e.group.isRunning {
		panic("该路由已经被暂停！")
	}

	ctx := context.Get(req)
	ctx.Set("params", e.getParams(p))
	e.handler.ServeHTTP(w, req)
}
