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

// Export functions for test only
package xpath

func GetCustomFunctionInfo(plugins []XpathPlugin) []CustomFunctionInfo {
	return getCustomFunctionInfo(plugins)
}

func (sym *Symbol) CustomFunc() CustomFn { return sym.customFunc }
