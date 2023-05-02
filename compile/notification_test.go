// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2017 by Brocade Communications Systems, Inc.
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

type NotificationChecker struct {
	Notification NodeChecker
}

func (expected NotificationChecker) check(t *testing.T, actual schema.Notification) {
	expected.Notification.check(t, actual.Schema())
}

func getNotificationSchemaNode(
	t *testing.T,
	schema_text *bytes.Buffer,
	namespace, name string,
) schema.Notification {
	st, err := testutils.GetConfigSchema(schema_text.Bytes())
	if err != nil {
		t.Errorf("Unexpected error when parsing Notification schema: %s", err)
	}

	if actual := st.Notifications()[namespace][name]; actual != nil {
		return actual
	}

	t.Errorf("Unable to find Notification for NS: %s, Name %s", namespace, name)
	return nil
}

func TestNotificationSuccess(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`notification interface-event {
		description "Interface state change";
			leaf interface {
				type string;
			}
			leaf state {
				type enumeration {
					enum unplugged;
					enum down;
					enum up;
				}
			}
	    }`))
	expected := NotificationChecker{
		NewTreeChecker(
			"Notification",
			[]NodeChecker{
				NewLeafChecker("interface", CheckType("string")),
				NewLeafChecker("state", CheckType("enumeration")),
			}),
	}

	if actual := getNotificationSchemaNode(t, schema_text,
		"urn:vyatta.com:test:yang-compile", "interface-event"); actual != nil {
		expected.check(t, actual)
	}
}

func TestNotificationListSuccess(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`notification interface-event {
		description "Interface state change";
			list interface {
				key name;
				leaf name {
					type string;
				}
				leaf state {
					type enumeration {
						enum unplugged;
						enum down;
						enum up;
					}
				}
			}
	    }`))
	expected := NotificationChecker{
		NewTreeChecker(
			"Notification",
			[]NodeChecker{
				NewListChecker("interface",
					[]NodeChecker{
						NewKeyChecker("name"),
						NewLeafChecker("state", CheckType("enumeration")),
					}),
			}),
	}

	if actual := getNotificationSchemaNode(t, schema_text,
		"urn:vyatta.com:test:yang-compile", "interface-event"); actual != nil {
		expected.check(t, actual)
	}
}
