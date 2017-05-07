// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package entry

// 最基本的字符串匹配，只能全字符串匹配。
type basic struct {
	*base
}

func (b *basic) Type() int {
	return TypeBasic
}

func (b *basic) Params(url string) map[string]string {
	return nil
}

func (b *basic) Match(url string) int {
	if url == b.pattern {
		return 0
	}
	return -1
}

func (b *basic) URL(params map[string]string) (string, error) {
	return b.pattern, nil
}
