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

// Copyright (c) 2020-2021, AT&T Intellectual Property.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package compile_test

import (
	"testing"
)

func TestChoiceBasic(t *testing.T) {
	input := `container c1 {
			choice foo {
				case case-a {
					leaf foo-a {
						type string;
					}
					leaf bar-a {
						type string;
					}
				}
				case case-b {
					leaf foo-b {
						type string;
					}
					leaf bar-b {
						type string;
					}
				}
			}
			choice bar {
				leaf barbaz {
					type string;
				}
			}
		}

		augment /c1 {
			choice foobar {
				case foo {
					leaf foo {
						type string;
					}
				}
				case bar {
					leaf bar {
						type string;
					}
				}
			}
		}`

	expected := NewContainerChecker(
		"c1",
		[]NodeChecker{
			NewLeafChecker("foo-a"),
			NewLeafChecker("bar-a"),
			NewLeafChecker("foo-b"),
			NewLeafChecker("bar-b"),
			NewLeafChecker("foo"),
			NewLeafChecker("bar"),
			NewLeafChecker("barbaz"),
		})

	ExpectSuccess(t, expected, input)
}

func TestChoiceAugmentChoiceWithCase(t *testing.T) {
	input := `container c1 {
			choice achoice {
				case a {
					leaf foo-a {
						type string;
					}
					leaf bar-a {
						type string;
					}
				}
				case b {
					leaf foo-b {
						type string;
					}
					leaf bar-b {
						type string;
					}
				}
			}
		}

		augment /c1/achoice {
			case c {
				leaf foo-c {
					type string;
				}
				leaf bar-c {
					type string;
				}
			}
		}`

	expected := NewContainerChecker(
		"c1",
		[]NodeChecker{
			NewLeafChecker("foo-a"),
			NewLeafChecker("bar-a"),
			NewLeafChecker("foo-b"),
			NewLeafChecker("bar-b"),
			NewLeafChecker("foo-c"),
			NewLeafChecker("bar-c"),
		})

	ExpectSuccess(t, expected, input)
}

func TestChoiceAugmentANonChoiceWithCaseFail(t *testing.T) {
	input := `container c1 {
			choice achoice {
				case a {
					leaf foo-a {
						type string;
					}
					leaf bar-a {
						type string;
					}
				}
				case b {
					leaf foo-b {
						type string;
					}
					leaf bar-b {
						type string;
					}
				}
			}
		}

		augment /c1 {
			case c {
				leaf foo-c {
					type string;
				}
				leaf bar-c {
					type string;
				}
			}
		}`

	expected := "augment /c1: invalid refinement case for statement container"

	ExpectFailure(t, expected, input)

}

func TestChoiceAugmentCase(t *testing.T) {
	input := `container c1 {
			choice achoice {
				case a {
					leaf foo-a {
						type string;
					}
					leaf bar-a {
						type string;
					}
				}
				case b {
					leaf foo-b {
						type string;
					}
					leaf bar-b {
						type string;
					}
				}
			}
		}

		augment /c1/achoice/b {
			leaf foo-c {
				type string;
			}
			leaf bar-c {
				type string;
			}
		}`

	expected := NewContainerChecker(
		"c1",
		[]NodeChecker{
			NewLeafChecker("foo-a"),
			NewLeafChecker("bar-a"),
			NewLeafChecker("foo-b"),
			NewLeafChecker("bar-b"),
			NewLeafChecker("foo-c"),
			NewLeafChecker("bar-c"),
		})

	ExpectSuccess(t, expected, input)
}

func TestChoiceAugmentImplicitCase(t *testing.T) {
	input := `container c1 {
			choice achoice {
				case a {
					leaf foo-a {
						type string;
					}
					leaf bar-a {
						type string;
					}
				}
				case b {
					leaf foo-b {
						type string;
					}
					leaf bar-b {
						type string;
					}
				}
				container c {
					// This is an implicit case, augment it using
					// path '/c1/achoice/c/c'
					presence "";
				}
			}
		}

		// container c above is an implicit case
		// and has implicit "case c"
		augment /c1/achoice/c/c {
			leaf foo-c {
				type string;
			}
			leaf bar-c {
				type string;
			}
		}`

	expected := NewContainerChecker(
		"c1",
		[]NodeChecker{
			NewLeafChecker("foo-a"),
			NewLeafChecker("bar-a"),
			NewLeafChecker("foo-b"),
			NewLeafChecker("bar-b"),
			NewContainerChecker(
				"c",
				[]NodeChecker{
					NewLeafChecker("foo-c"),
					NewLeafChecker("bar-c"),
				}),
		})

	ExpectSuccess(t, expected, input)
}

func TestChoiceAugmentCaseWithCaseFail(t *testing.T) {
	input := `container c1 {
			choice achoice {
				case a {
					leaf foo-a {
						type string;
					}
					leaf bar-a {
						type string;
					}
				}
				case b {
					leaf foo-b {
						type string;
					}
					leaf bar-b {
						type string;
					}
				}
			}
		}

		augment /c1/achoice/b {
			case c {
				leaf foo-c {
					type string;
				}
				leaf bar-c {
					type string;
				}
			}
		}`

	expected := "augment /c1/achoice/b: invalid refinement case for statement case"

	ExpectFailure(t, expected, input)

}

func TestChoiceAugmentCaseWithChoice(t *testing.T) {
	input := `container c1 {
			choice achoice {
				case a {
					leaf foo-a {
						type string;
					}
					leaf bar-a {
						type string;
					}
				}
				case b {
					leaf foo-b {
						type string;
					}
					leaf bar-b {
						type string;
					}
				}
			}
		}

		augment /c1/achoice/b {
			choice foobar {
				case foo {
					leaf foo {
						type string;
					}
				}
				case bar {
					leaf bar {
						type string;
					}
				}
			}
		}`

	expected := NewContainerChecker(
		"c1",
		[]NodeChecker{
			NewLeafChecker("foo-a"),
			NewLeafChecker("bar-a"),
			NewLeafChecker("foo-b"),
			NewLeafChecker("bar-b"),
			NewLeafChecker("foo"),
			NewLeafChecker("bar"),
		})

	ExpectSuccess(t, expected, input)
}

func TestChoiceAugmentCaseWithMandatoryChoice(t *testing.T) {
	input := `container c1 {
			choice achoice {
				case a {
					leaf foo-a {
						type string;
					}
					leaf bar-a {
						type string;
					}
				}
				case b {
					leaf foo-b {
						type string;
					}
					leaf bar-b {
						type string;
					}
				}
			}
		}

		augment /c1/achoice/b {
			choice foobar {
				// Mandatory should be allowed
				// as local module is being augmented
				mandatory true;
				case foo {
					leaf foo {
						type string;
					}
				}
				case bar {
					leaf bar {
						type string;
					}
				}
			}
		}`

	expected := NewContainerChecker(
		"c1",
		[]NodeChecker{
			NewLeafChecker("foo-a"),
			NewLeafChecker("bar-a"),
			NewLeafChecker("foo-b"),
			NewLeafChecker("bar-b"),
			NewLeafChecker("foo"),
			NewLeafChecker("bar"),
		})

	ExpectSuccess(t, expected, input)
}

func TestChoiceWithDuplicateCaseFail(t *testing.T) {
	input := `container c1 {
			choice achoice {
				mandatory true;
				default a;

				case a {
					leaf foo-a {
						type string;
						default "abc";
					}
					leaf bar-a {
						type string;
					}
				}
				case b {
					leaf foo-b {
						type string;
					}
					leaf bar-b {
						type string;
					}
				}
			}
		}`

	expected := "choice achoice: Choice cannot have default and be mandatory"

	ExpectFailure(t, expected, input)

}

func TestChoiceWithDuplicateNodesUnderCaseFail(t *testing.T) {
	input := `container c1 {
			choice achoice {
				case a {
					leaf foo-a {
						type string;
					}
					leaf bar-a {
						type string;
					}
				}
				case b {
					leaf foo-b {
						type string;
					}
					leaf bar-b {
						type string;
					}
					leaf bar-a {
						// bar-a appears under cases a and b
						type string;
					}
				}
			}
		}`

	expected := "choice achoice: redefinition of name bar-a"

	ExpectFailure(t, expected, input)
}

func TestChoiceWithDuplicateCaseNamesFail(t *testing.T) {
	input := `container c1 {
			choice achoice {
				case a {
					leaf foo-a {
						type string;
					}
					leaf bar-a {
						type string;
					}
				}
				case a {
					leaf foo-b {
						type string;
					}
					leaf bar-b {
						type string;
					}
				}
			}
		}`

	expected := "choice achoice: redefinition of name a"

	ExpectFailure(t, expected, input)
}

func TestChoiceWithDuplicateCaseNamesAugmentedFail(t *testing.T) {
	input := `container c1 {
			choice achoice {
				case a {
					leaf foo-a {
						type string;
					}
					leaf bar-a {
						type string;
					}
				}
			}
		}

		augment /c1/achoice {
			case a {
				leaf foo-b {
					type string;
				}
				leaf bar-b {
					type string;
				}
			}
		}`

	expected := "choice achoice: redefinition of name a"

	ExpectFailure(t, expected, input)
}

func TestChoiceWithDuplicateChoiceNameFail(t *testing.T) {
	input := `container c1 {
			choice achoice {
				case a {
					leaf foo-a {
						type string;
					}
					leaf bar-a {
						type string;
					}
				}
			}
			choice achoice {
				case b {
					leaf foo-b {
						type string;
					}
					leaf bar-b {
						type string;
					}
				}
			}
		}`

	expected := "container c1: redefinition of name achoice"

	ExpectFailure(t, expected, input)
}

func TestChoiceWithDuplicateChoiceNameAugmentedFail(t *testing.T) {
	input := `container c1 {
			choice achoice {
				case a {
					leaf foo-a {
						type string;
					}
					leaf bar-a {
						type string;
					}
				}
			}
		}

		augment /c1 {
			choice achoice {
				case b {
					leaf foo-b {
						type string;
					}
					leaf bar-b {
						type string;
					}
				}
			}
		}`

	expected := "container c1: redefinition of name achoice"

	ExpectFailure(t, expected, input)
}

func TestChoiceWithMissingDefaultFail(t *testing.T) {
	input := `container c1 {
			choice achoice {
				default c;

				case a {
					leaf foo-a {
						type string;
					}
					leaf bar-a {
						type string;
					}
				}
			}
		}

		augment /c1 {
			choice achoice {
				case b {
					leaf foo-b {
						type string;
					}
					leaf bar-b {
						type string;
					}
				}
			}
		}`

	expected := "choice achoice: Choice default c not found"

	ExpectFailure(t, expected, input)
}
