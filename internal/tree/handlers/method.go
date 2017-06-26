// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"net/http"
	"sort"
	"strings"
)

type methodType int16

// 各个请求方法的值
const (
	get methodType = 1 << iota
	post
	del
	put
	patch
	options
	head
	connect
	trace

	none methodType = 0
	max             = trace // 最大值
)

var (
	methodMap = map[string]methodType{
		http.MethodGet:     get,
		http.MethodPost:    post,
		http.MethodDelete:  del,
		http.MethodPut:     put,
		http.MethodPatch:   patch,
		http.MethodOptions: options,
		http.MethodHead:    head,
		http.MethodConnect: connect,
		http.MethodTrace:   trace,
	}

	methodStringMap = make(map[methodType]string, len(methodMap))

	// 当前支持的所有请求方法
	supported = make([]string, 0, len(methodMap))

	// any 调用 *.Any 时添加所使用的请求方法列表，
	// 默认为除 http.MethodOptions 之外的所有 supported 中的元素
	any = make([]string, 0, len(methodMap)-1)

	// 所有的 OPTIONS 请求的 allow 报头字符串
	optionsStrings = make(map[methodType]string, max)
)

func init() {
	// 生成 methodStringMap
	for typ, str := range methodMap {
		methodStringMap[str] = typ
	}

	// 生成 supported 和 any
	for typ := range methodMap {
		supported = append(supported, typ)
		if typ != http.MethodOptions {
			any = append(any, typ)
		}
	}

	makeOptionsStrings()
}

func makeOptionsStrings() {
	methods := make([]string, 0, len(supported))
	for i := methodType(0); i <= max; i++ {
		if i&get == get {
			methods = append(methods, methodStringMap[get])
		}
		if i&post == post {
			methods = append(methods, methodStringMap[post])
		}
		if i&del == del {
			methods = append(methods, methodStringMap[del])
		}
		if i&put == put {
			methods = append(methods, methodStringMap[put])
		}
		if i&patch == patch {
			methods = append(methods, methodStringMap[patch])
		}
		if i&options == options {
			methods = append(methods, methodStringMap[options])
		}
		if i&head == head {
			methods = append(methods, methodStringMap[head])
		}
		if i&connect == connect {
			methods = append(methods, methodStringMap[connect])
		}
		if i&trace == trace {
			methods = append(methods, methodStringMap[trace])
		}

		sort.Strings(methods) // 防止每次从 map 中读取的顺序都不一样
		optionsStrings[i] = strings.Join(methods, ", ")
		methods = methods[:0]
	}
}