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
// SPDX-License-Identifier: MPL-2.0

package xpath

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
	"unicode"
)

var pluginsLoaded bool

// Plugins need to provide a table, exported as 'RegistrationData', which is
// a slice of the following structure.
type CustomFunctionInfo struct {
	// Function name as used in XPATH must statement.  Lower case and hyphen
	// are only allowed characters.
	Name string
	// Implementation of the function.
	FnPtr CustomFn
	// Arguments (type and number) taken by the function
	Args []DatumTypeChecker
	// Return value type
	RetType DatumTypeChecker
	// Default return value, for use if function panics for any reason.
	DefaultRetVal Datum
}

// init() - set up logger to discard
var dlog *log.Logger

func init() {
	dlog = log.New(ioutil.Discard, "", 0)
}

func SetDebugLogger(logger *log.Logger) {
	dlog = logger
}

func RegisterCustomFunctions(customFunctionInfoTbl []CustomFunctionInfo) {

	pluginsLoaded = true

	for _, regInfo := range customFunctionInfoTbl {

		if !validateName(regInfo.Name) {
			dlog.Printf("Invalid name for XPATH plugin function: %s\n",
				regInfo.Name)
			continue
		}

		xpathFunctionTable[regInfo.Name] = NewCustomFnSym(
			regInfo.Name,
			wrapFnWithRecover(regInfo.FnPtr, regInfo.DefaultRetVal),
			regInfo.Args,
			regInfo.RetType)

		dlog.Printf("Adding XPATH custom function: %s\n",
			regInfo.Name)
	}
}

// validateName - can only use lower case letters and hyphen.
//
// Somewhat arbitrary, but matches standard naming convention for XPATH
// functions and avoids any possible parsing issues.
func validateName(name string) bool {
	if len(name) == 0 {
		return false
	}

	if unicode.IsNumber(rune(name[0])) {
		return false
	}

	for _, c := range name {
		if !unicode.IsLower(c) && !unicode.IsNumber(c) && c != '-' {
			return false
		}
	}
	return true
}

// If internal functions panic, we allow the context to catch and recover.
// For external functions, we catch the panic more locally and return a
// default value based on the expected return type.
func wrapFnWithRecover(fnPtr CustomFn, defRetVal Datum) CustomFn {
	return func(args []Datum) (res Datum) {
		defer func() {
			if r := recover(); r != nil {
				res = defRetVal
			}
		}()
		return fnPtr(args)
	}
}

type XpathPlugin interface {
	Lookup(name string) (plugin.Symbol, error)
	Name() string
}

type xpathPlugin struct {
	provider *plugin.Plugin
	name     string
}

func (xp *xpathPlugin) Name() string { return xp.name }
func (xp *xpathPlugin) Lookup(name string) (plugin.Symbol, error) {
	return xp.provider.Lookup(name)
}

const (
	xpathPluginDir = "/lib/xpath/plugins/"
)

// NOT covered by unit test.
func openPlugins() []CustomFunctionInfo {
	// Find plugins in well-known location and return
	dir, err := os.Open(xpathPluginDir)
	if err != nil {
		return nil
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return nil
	}

	var plugins []XpathPlugin
	for _, file := range files {
		if !file.Mode().IsRegular() || filepath.Ext(file.Name()) != ".so" {
			continue
		}
		xpathFnPlugin, err := plugin.Open(xpathPluginDir + file.Name())
		if err != nil {
			dlog.Printf("Unable to load XPATH plugin %s: %s\n",
				file.Name(), err)
			continue
		}
		plugins = append(plugins,
			&xpathPlugin{provider: xpathFnPlugin, name: file.Name()})
	}

	return getCustomFunctionInfo(plugins)
}

func getCustomFunctionInfo(plugins []XpathPlugin) []CustomFunctionInfo {

	var customFnInfoTbl []CustomFunctionInfo

	for _, provider := range plugins {

		pluginData, err := provider.Lookup("RegistrationData")
		if err != nil {
			dlog.Printf("Unable to register XPATH plugin %s: %s\n",
				provider.Name(), err)
			continue
		}

		if pluginInfoTbl, ok := pluginData.(*[]CustomFunctionInfo); ok {
			customFnInfoTbl = append(customFnInfoTbl, (*pluginInfoTbl)...)
			dlog.Printf("Loaded XPATH plugin: %s\n", provider.Name())
		} else {
			dlog.Printf("%s XPATH plugin data is of wrong type:  %v\n",
				provider.Name(), reflect.TypeOf(pluginInfoTbl))
		}
	}

	return customFnInfoTbl
}
