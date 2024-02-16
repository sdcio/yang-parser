// Copyright 2024 Nokia
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2014 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package parse

import "errors"
import "fmt"

var ErrNoShadow = errors.New("cannot shadow")

type TEnv struct {
	prev *TEnv
	syms map[string]Node
}

func (e *TEnv) Get(s string) (Node, bool) {
	if e == nil {
		return nil, false
	}
	sym, ok := e.syms[s]
	if !ok {
		sym, ok = e.prev.Get(s)
	}
	return sym, ok
}

func (e *TEnv) Copy() *TEnv {
	t := &TEnv{
		prev: e.prev,
		syms: make(map[string]Node),
	}
	for k, v := range e.syms {
		t.syms[k] = v
	}
	return t
}

func (e *TEnv) Put(s string, sym Node) error {
	//no shadowing
	if _, ok := e.Get(s); !ok {
		e.syms[s] = sym
		return nil
	}
	return errors.New(ErrNoShadow.Error() + " typedef " + s)
}

func NewTEnv(p *TEnv) *TEnv {
	return &TEnv{
		prev: p,
		syms: make(map[string]Node),
	}
}

type GEnv struct {
	prev *GEnv
	syms map[string]Node
}

func (e *GEnv) Copy() *GEnv {
	t := &GEnv{
		prev: e.prev,
		syms: make(map[string]Node),
	}
	for k, v := range e.syms {
		t.syms[k] = v
	}
	return t
}

func (e *GEnv) String() string {
	if e.prev == nil {
		return "<end>"
	}
	return fmt.Sprintf("{%s, %s}", e.syms, e.prev)
}

func (e *GEnv) Get(s string) (Node, bool) {
	if e == nil {
		return nil, false
	}
	sym, ok := e.syms[s]
	if !ok {
		sym, ok = e.prev.Get(s)
	}
	return sym, ok
}

func (e *GEnv) Put(s string, sym Node) error {
	//no shadowing
	if _, ok := e.Get(s); !ok {
		e.syms[s] = sym
		return nil
	}
	return errors.New(ErrNoShadow.Error() + " grouping " + s)
}

func NewGEnv(p *GEnv) *GEnv {
	return &GEnv{
		prev: p,
		syms: make(map[string]Node),
	}
}
