// Copyright (c) 2019, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package compile_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/iptecharch/yang-parser/schema"
	"github.com/iptecharch/yang-parser/testutils"
)

type RpcChecker struct {
	Script string
	Input  NodeChecker
	Output NodeChecker
}

func (expected RpcChecker) check(t *testing.T, actual schema.Rpc) {
	expected.Input.check(t, actual.Input())
	expected.Output.check(t, actual.Output())
}

func getRpcSchemaNode(
	t *testing.T,
	schema_text *bytes.Buffer,
	namespace, name string,
) schema.Rpc {
	st, err := testutils.GetConfigSchema(schema_text.Bytes())
	if err != nil {
		t.Errorf("Unexpected error when parsing RPC schema: %s", err)
	}

	if actual := st.Rpcs()[namespace][name]; actual != nil {
		return actual
	}

	t.Errorf("Unable to find RPC for NS: %s, Name %s", namespace, name)
	return nil
}

func TestPingRpcAugmentExplicitInputOutput(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`rpc ping {
		description "Generates Ping and return response";

			input {
				leaf host {
					type string;
					mandatory true;
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
			}
		}

		augment /ping/output {
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

		augment /ping/input {
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

	    }`))

	expected := RpcChecker{
		Input: NewTreeChecker(
			"Input",
			[]NodeChecker{
				NewLeafChecker("host", CheckType("string")),
				NewLeafChecker("count"),
				NewLeafChecker("ttl"),
			}),
		Output: NewTreeChecker(
			"Output",
			[]NodeChecker{
				NewLeafChecker("tx-packet-count", CheckType("uint32")),
				NewLeafChecker("rx-packet-count"),
				NewLeafChecker("min-delay"),
				NewLeafChecker("average-delay"),
				NewLeafChecker("max-delay"),
			}),
	}

	if actual := getRpcSchemaNode(t, schema_text,
		"urn:vyatta.com:test:yang-compile", "ping"); actual != nil {
		expected.check(t, actual)
	}
}
func TestPingRpcAugmentImplicitInputOutput(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`rpc ping {
			description "Generates Ping and return response";
		}

		augment /ping/output {
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
		}

		augment /ping/input {
			leaf host {
				type string;
				mandatory true;
			}
			leaf count {
				type uint32;
				default 3;
				description "Number of ping echo request message to send";
			}
		}

		augment /ping/output {
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

		augment /ping/input {
			leaf ttl {
				type uint8;
				default "255";
				description "IP Time To Live";
			}

	    }`))

	expected := RpcChecker{
		Input: NewTreeChecker(
			"Input",
			[]NodeChecker{
				NewLeafChecker("host", CheckType("string")),
				NewLeafChecker("count"),
				NewLeafChecker("ttl"),
			}),
		Output: NewTreeChecker(
			"Output",
			[]NodeChecker{
				NewLeafChecker("tx-packet-count", CheckType("uint32")),
				NewLeafChecker("rx-packet-count"),
				NewLeafChecker("min-delay"),
				NewLeafChecker("average-delay"),
				NewLeafChecker("max-delay"),
			}),
	}

	if actual := getRpcSchemaNode(t, schema_text,
		"urn:vyatta.com:test:yang-compile", "ping"); actual != nil {
		expected.check(t, actual)
	}
}

func TestPingRpcSuccess(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`rpc ping {
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
	    }`))

	expected := RpcChecker{
		Input: NewTreeChecker(
			"Input",
			[]NodeChecker{
				NewLeafChecker("host", CheckType("string")),
				NewLeafChecker("count"),
				NewLeafChecker("ttl"),
			}),
		Output: NewTreeChecker(
			"Output",
			[]NodeChecker{
				NewLeafChecker("tx-packet-count", CheckType("uint32")),
				NewLeafChecker("rx-packet-count"),
				NewLeafChecker("min-delay"),
				NewLeafChecker("average-delay"),
				NewLeafChecker("max-delay"),
			}),
	}

	if actual := getRpcSchemaNode(t, schema_text,
		"urn:vyatta.com:test:yang-compile", "ping"); actual != nil {
		expected.check(t, actual)
	}
}

func TestPingRpcMissingInput(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`rpc ping {
		description "Generates Ping and return response";
		output {
			leaf tx-packet-count {
				type uint32;
				description "Transmitted Packet count";
			}
		}
	    }`))

	expected := RpcChecker{
		Input: NewTreeChecker("Input", []NodeChecker{}),
		Output: NewTreeChecker("Output", []NodeChecker{
			NewLeafChecker("tx-packet-count"),
		}),
	}

	if actual := getRpcSchemaNode(t, schema_text,
		"urn:vyatta.com:test:yang-compile", "ping"); actual != nil {
		expected.check(t, actual)
	}
}

func TestPingRpcRepeatedInput(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`rpc ping {
		description "Generates Ping and return response";
		input {
			leaf host {
				type string;
				mandatory true;
			}
		}
		input {
			leaf host {
				type string;
				mandatory true;
			}
		}
		output {
			leaf tx-packet-count {
				type uint32;
				description "Transmitted Packet count";
			}
		}
	    }`))
	expected := "only one 'input' statement is allowed"

	_, err := testutils.GetConfigSchema(schema_text.Bytes())
	assertErrorContains(t, err, expected)
}

func TestPingRpcMissingOutput(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`rpc ping {
		description "Generates Ping and return response";
		input {
			leaf host {
				type string;
				mandatory true;
			}
		}
	    }`))
	st, err := testutils.GetConfigSchema(schema_text.Bytes())
	if err != nil {
		t.Errorf("Unexpected error when parsing RPC schema: %s", err)
		t.FailNow()
	}

	actual := st.Rpcs()["urn:vyatta.com:test:yang-compile"]["ping"]
	expected := RpcChecker{
		Input: NewTreeChecker("Input", []NodeChecker{
			NewLeafChecker("host"),
		}),
		Output: NewTreeChecker("Output", []NodeChecker{}),
	}

	expected.check(t, actual)
}

func TestPingRpcRepeatedOutput(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`rpc ping {
		description "Generates Ping and return response";
		input {
			leaf host {
				type string;
				mandatory true;
			}
		}
		output {
			leaf tx-packet-count {
				type uint32;
				description "Transmitted Packet count";
			}
		}
		output {
			leaf tx-packet-count {
				type uint32;
				description "Transmitted Packet count";
			}
		}
	    }`))

	expected := "only one 'output' statement is allowed"

	_, err := testutils.GetConfigSchema(schema_text.Bytes())
	assertErrorContains(t, err, expected)
}
