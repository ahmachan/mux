// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mux

import (
	"errors"
	"net/http"
	"regexp"
)

// 支持正则表达式的匹配内容。
type entry struct {
	pattern string
	expr    *regexp.Regexp
	handler http.Handler
}

// 新建一个Entry实例。
//
// pattern为路由的匹配模式，包含以下三种情况：
// - 以?字符开头：表示一个正则匹配，将?所有的字符串转换成正则表达式；
// - 其它情况：表示一个普通的字符串匹配；
//
// 当pattern为空时，将返回错误信息。
// 调用方必须保证参数h不为空值！
func newEntry(pattern string, h http.Handler) (*entry, error) {
	if len(pattern) == 0 {
		return nil, errors.New("newEntry:参数pattern不能为空值")
	}

	entry := &entry{
		pattern: pattern,
		handler: h,
	}

	if pattern[0] == '?' { // 若以？开头，则表示是一个正则表达式
		if len(pattern) == 1 {
			return nil, errors.New("newEntry:pattern正则内容不能为空")
		}
		entry.pattern = pattern[1:]
		entry.expr = regexp.MustCompile(pattern[1:])
	}

	return entry, nil
}

// 当前实例是否与参数匹配。
// 若是匹配，还将返回符合正则表达式的命名匹配，如果存在的话。
func (entry *entry) match(pattern string) (bool, map[string]string) {
	if entry.expr == nil { // 简单的字符串匹配
		return entry.pattern == pattern, nil
	}

	if !entry.expr.MatchString(pattern) {
		return false, nil
	}

	// 获取命名匹配变量。
	ret := make(map[string]string)
	subexps := entry.expr.SubexpNames()
	args := entry.expr.FindStringSubmatch(pattern)
	for index, name := range subexps {
		if len(name) > 0 {
			ret[name] = args[index]
		}
	}

	return true, ret
}
