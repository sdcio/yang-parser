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
// Copyright (c) 2015 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package parse_test

import (
	"testing"
)

func TestIdentityAccepted(t *testing.T) {

	input := `
  identity schema-format {
    description
      "Base identity for data model schema languages.";
  }

  identity xsd {
    base schema-format;
  }

  identity yang {
    base schema-format;
  }

  identity yin {
    base schema-format;
  }

  leaf foo {
    type identityref {
        base schema-format;
    }
  }
`

	expected := LeafNodeChecker{
		Name: "foo",
		Typ:  "identityref",
	}

	tree := createParseTreeFromYang(t, input)
	actual := findNode(tree, "foo")
	if actual == nil {
		t.Errorf("Failed to find expected schema node\n")
		return
	}

	expected.check(t, actual)
}
