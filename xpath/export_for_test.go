// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// Export functions for test only
package xpath

func GetCustomFunctionInfo(plugins []XpathPlugin) []CustomFunctionInfo {
	return getCustomFunctionInfo(plugins)
}

func (sym *Symbol) CustomFunc() CustomFn { return sym.customFunc }
