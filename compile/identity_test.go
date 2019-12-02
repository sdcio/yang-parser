// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package compile_test

import (
	"testing"

	"github.com/danos/yang/testutils"
)

//
//  Test Cases
//
func TestIdentitySuccessSimple(t *testing.T) {
	schema_snippet := `
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
	st := buildSchema(t, schema_snippet)
	assertLeafMatches(t, st, "foo", "identityref")
}

const identitySchema = `
        identity numbers;

        identity one {
	        base numbers;
        }
	identity two {
		base numbers;
	}
	identity three {
		base one;
	}
	identity four {
		base two;
	}

	typedef number {
		type identityref {
			base numbers;
		}
	}
	grouping number-value {
		leaf a-number {
			type number;
		}
	}
`

var identityTestSchema = testutils.TestSchema{
	Name: testutils.NameDef{
		Namespace: "prefix-remote",
		Prefix:    "remote",
	},
	SchemaSnippet: identitySchema,
}

func runIdentityTestCases(t *testing.T, testCases []testutils.TestCase) {
	for _, tc := range testCases {
		applyAndVerifySchemas(t, &tc, false)
	}
}

func TestIdentityRefValid(t *testing.T) {
	var tc = []testutils.TestCase{
		{
			Description: "Valid identities",
			ExpResult:   true,
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `
						// one through four match identities define in
						// remote, but have different namespace
						identity one {
							base remote:numbers;
						}
						identity two {
							base remote:numbers;
						}
						identity three {
							base remote:numbers;
						}
						identity four {
							base remote:numbers;
						}

						identity twentys;
						identity twenty {
							base twentys;
						}
						identity twentyone {
							base twentys;
						}

						uses remote:number-value;

						leaf all {
							type union {
								type identityref {
									base twentys;
								}
								type identityref {
									base remote:numbers;
								}
							}
						}
						leaf numbers {
							type identityref {
								base remote:numbers;
							}
				}`,
				},
				identityTestSchema,
			},
		},
	}
	runIdentityTestCases(t, tc)
}
func TestIdentityDuplicateIdentity(t *testing.T) {
	var tc = []testutils.TestCase{
		{
			Description: "Detect an identically named identity in a module",
			ExpResult:   false,
			ExpErrMsg:   "identity ninety: Duplicate identity ninety in module prefix-test",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `identity ninety;
							identity ninety {
								base remote:numbers;
				}`,
				},
				identityTestSchema,
			},
		},
	}
	runIdentityTestCases(t, tc)
}

func TestIdentityCyclicReference(t *testing.T) {
	var tc = []testutils.TestCase{
		{
			Description: "Detect an identity cyclic reference",
			ExpResult:   false,
			ExpErrMsg:   "Identity cyclic reference",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `// Intentional cycle
						identity start {
							base end;
						}
						identity middle {
							base start;
						}
						identity end {
							base middle;
				}`,
				},
				identityTestSchema,
			},
		},
	}
	runIdentityTestCases(t, tc)
}

func TestIdentityBaseIdentityNotValid(t *testing.T) {
	var tc = []testutils.TestCase{
		{
			Description: "identity has non-existent base in local module",
			ExpResult:   false,
			ExpErrMsg:   "identity not valid",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `
						identity foo {
							base bogus;
				}`,
				},
				identityTestSchema,
			},
		},
	}
	runIdentityTestCases(t, tc)
}

func TestIdentityBaseRemoteIdentityNotValid(t *testing.T) {
	var tc = []testutils.TestCase{
		{
			Description: "Identity has non-existent base in remote module",
			ExpResult:   false,
			ExpErrMsg:   "identity not valid",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `
						identity foo {
							base remote:bogus;
				}`,
				},
				identityTestSchema,
			},
		},
	}
	runIdentityTestCases(t, tc)
}

func TestIdentityRefBaseIdentityNotValid(t *testing.T) {
	var tc = []testutils.TestCase{
		{
			Description: "An identityref has an invalid base in local module",
			ExpResult:   false,
			ExpErrMsg:   "identity not valid",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `
						identity foo;

						leaf test {
							type identityref {
								base bogus;
							}
				}`,
				},
				identityTestSchema,
			},
		},
	}
	runIdentityTestCases(t, tc)
}

func TestIdentityRefBaseRemoteIdentityNotValid(t *testing.T) {
	var tc = []testutils.TestCase{
		{
			Description: "An identityref has a non-existent base, in remote module",
			ExpResult:   false,
			ExpErrMsg:   "identity not valid",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `
						identity bogus;
						leaf test {
							type identityref {
								base remote:bogus;
							}
				}`,
				},
				identityTestSchema,
			},
		},
	}
	runIdentityTestCases(t, tc)
}
