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

func TestRpcAccepted(t *testing.T) {

	input := `rpc ping {
		description "Generates Ping and return response";
		input {
			leaf host {
				type string;
				mandatory true;
			}
			leaf count {
				type uint32;
				default 3;
				description "Number of ping echo request message to send";
			}
			leaf ttl {
				type uint8;
				default "255";
				description "IP Time To Live";
			}
		}
		output {
			leaf tx-packet-count {
				type uint32;
				description "Transmitted Packet count";
			}
			leaf rx-packet-count {
				type uint32;
				description "Received packet count";
			}
			leaf min-delay {
				type uint32;
				units "milliseconds";
				description "Minimum packet delay";
			}
			leaf average-delay {
				type uint32;
				units "milliseconds";
				description "Average packet delay";
			}
			leaf max-delay {
				type uint32;
				units "millisecond";
				description "Minimum packet delay";
			}
		}
	}`
	expected := RpcNodeChecker{
		Name: "ping",
		Input: InputNodeChecker{
			Body: []LeafNodeChecker{
				LeafNodeChecker{Name: "host"},
				LeafNodeChecker{Name: "count"},
				LeafNodeChecker{Name: "ttl"},
			},
		},
		Output: OutputNodeChecker{
			Body: []LeafNodeChecker{
				LeafNodeChecker{Name: "tx-packet-count"},
				LeafNodeChecker{Name: "rx-packet-count"},
				LeafNodeChecker{Name: "min-delay"},
				LeafNodeChecker{Name: "average-delay"},
				LeafNodeChecker{Name: "max-delay"},
			},
		},
	}

	verifyExpectedPass(t, input, "rpc ping", expected)
}

func TestInvalidStatement(t *testing.T) {

	input := `rpc ping {
		description "Generates Ping and return response";
		output {
			mandatory true;
		}
	}`

	expected := "cardinality mismatch: invalid substatement 'mandatory'"

	verifyExpectedFail(t, input, expected)
}

func TestRepeatedDescription(t *testing.T) {

	input := `rpc ping {
		description "First description";
		description "Second description";
	}`

	expected := "cardinality mismatch: only one 'description' statement is allowed"

	verifyExpectedFail(t, input, expected)
}
