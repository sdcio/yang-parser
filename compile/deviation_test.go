// Copyright (c) 2018-2019, AT&T Intellectual Property.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package compile_test

import (
	"testing"

	"github.com/steiler/yang-parser/testutils"
)

const deviationSchema = `
container remotecontainer {
	description "Test container";
	must "not(anotherleaf = mandatoryleaf)";
	leaf remoteleaf {
		type string;
	}
	leaf anotherleaf {
		type uint8;
	}
	leaf defaultleaf {
		type uint8;
		default 8;
		units "dollars";
		config true;
	}
	leaf mandatoryleaf {
		type uint8;
		mandatory true;
		units "dollars";
		config true;
	}
	leaf-list aleaflist {
		type uint32;
		min-elements 1;
		max-elements 4;
	}
	list alist {
		unique "listdata";

		key listkey;

		leaf listkey {
			type string;
		}

		leaf liststringdata {
			type string;
		}

		leaf listdata {
			type uint16;
		}
		leaf moredata {
			type string;
		}
	}
}`

var deviationTestSchema = testutils.TestSchema{
	Name: testutils.NameDef{
		Namespace: "prefix-remote",
		Prefix:    "remote",
	},
	SchemaSnippet: deviationSchema,
}

func runDeviateTestCases(t *testing.T, testCases []testutils.TestCase) {
	for _, tc := range testCases {
		applyAndVerifySchemas(t, &tc, false)
	}
}

func TestDeviateNotSupported(t *testing.T) {
	var tc = []testutils.TestCase{
		{
			Description: "Deviate: not-supported is allowed ",
			ExpResult:   true,
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `deviation /remote:remotecontainer {
					description "remote container not " +
						"supported on this platform";
					deviate not-supported;
					reference "RFC 6020";
				}`,
				},
				deviationTestSchema,
			},
		},
	}
	runDeviateTestCases(t, tc)
}

func TestNotSupportedNoSubstatements(t *testing.T) {
	var tc = []testutils.TestCase{
		{
			Description: "Deviate: not-supported is allowed but no sub-statements allowed",
			ExpResult:   false,
			ExpErrMsg:   "deviate not-supported: Property not allowed in deviate not-supported 'default'",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `deviation /remote:remotecontainer {
					deviate not-supported {
						default 8;
					}
				}`,
				},
				deviationTestSchema,
			},
		},
		{
			Description: "Deviate: not-supported is allowed ",
			ExpResult:   false,
			ExpErrMsg:   "deviate not-supported: Property not allowed in deviate not-supported 'type'",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `deviation /remote:remotecontainer {
					deviate not-supported {
						type string;
					}
				}`,
				},
				deviationTestSchema,
			},
		},
		{
			Description: "Deviate: not-supported is allowed ",
			ExpResult:   false,
			ExpErrMsg:   "deviate not-supported: Property not allowed in deviate not-supported 'mandatory'",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `deviation /remote:remotecontainer {
					deviate not-supported {
						mandatory true;
					}
				}`,
				},
				deviationTestSchema,
			},
		},
		{
			Description: "Deviate: not-supported is allowed ",
			ExpResult:   false,
			ExpErrMsg:   "deviate not-supported: Property not allowed in deviate not-supported 'max-elements'",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `deviation /remote:remotecontainer {
					deviate not-supported {
						max-elements 5;
					}
				}`,
				},
				deviationTestSchema,
			},
		},
	}
	runDeviateTestCases(t, tc)
}

func TestDeviateOnlyNotSupported(t *testing.T) {
	var tc = []testutils.TestCase{
		{
			Description: "Deviate: not-present can't co-exist with add",
			ExpResult:   false,
			ExpErrMsg:   "No other deviate statements allowed with not-supported",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `deviation /remote:remotecontainer/remote:remoteleaf {
					deviate not-supported;
					deviate add {
						mandatory true;
					}
				}`,
				},
				deviationTestSchema,
			},
		},
		{
			Description: "Deviate: not-present can't co-exist with replace",
			ExpResult:   false,
			ExpErrMsg:   "No other deviate statements allowed with not-supported",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `deviation /remote:remotecontainer/remote:remoteleaf {
					deviate not-supported;
					deviate replace {
						type uint8;
					}
				}`,
				},
				deviationTestSchema,
			},
		},
		{
			Description: "Deviate: not-present can't co-exist with delete",
			ExpResult:   false,
			ExpErrMsg:   "No other deviate statements allowed with not-supported",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `deviation /remote:remotecontainer/remote:defaultleaf {
					deviate not-supported;
					deviate delete {
						default 8;
					}
				}`,
				},
				deviationTestSchema,
			},
		},
		{
			Description: "Deviate: not-supported can't co-exist with other deviate statements",
			ExpResult:   false,
			ExpErrMsg:   "No other deviate statements allowed with not-supported",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `deviation /remote:remotecontainer/remote:defaultleaf {
					deviate not-supported;
					deviate delete {
						default 8;
					}
					deviate replace {
						type string;
					}
					deviate add {
						units "euros";
					}
				}`,
				},
				deviationTestSchema,
			},
		},
	}
	runDeviateTestCases(t, tc)
}

func TestDeviateAdd(t *testing.T) {
	var tc = []testutils.TestCase{
		{
			Description: "Deviation: Add is supported",
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
					deviation /remote:remotecontainer/remote:remoteleaf {
					    deviate add {
						// Check default, units, config and must are allowed
						default "newdefault";
						units "borgs";
						config true;
						must ".";
					    }
				        }
					deviation /remote:remotecontainer/remote:anotherleaf {
					    deviate add {
						// Check mandatory, units, config and must are allowed
						mandatory "true";
						units "euros";
						config true;
						must ".";
						configd:help "abc def";
					    }
					}
					deviation /remote:remotecontainer/remote:alist {
						deviate add {
							// check min/max-elements and unique
							min-elements 1;
							max-elements 5;
							unique "moredata";
						}
				        }`,
				},
				deviationTestSchema,
			},
		},
	}
	runDeviateTestCases(t, tc)
}

func TestDeviateAddExistsErrors(t *testing.T) {
	var tc = []testutils.TestCase{
		{
			Description: "Deviation: Can't add a default if one exists",
			ExpResult:   false,
			ExpErrMsg:   "Property being added to node already exists",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `
					deviation /remote:remotecontainer/remote:defaultleaf {
					    deviate add {
						default "newdefault";
					    }
				        }`,
				},
				deviationTestSchema,
			},
		},
		{
			Description: "Deviation: Can't add a mandatory if one exists",
			ExpResult:   false,
			ExpErrMsg:   "Property being added to node already exists",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `
					deviation /remote:remotecontainer/remote:mandatoryleaf {
					    deviate add {
						mandatory true;
					    }
				        }`,
				},
				deviationTestSchema,
			},
		},
		{
			Description: "Deviation: Can't add a min-elements if one exists",
			ExpResult:   false,
			ExpErrMsg:   "Property being added to node already exists",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `
					deviation /remote:remotecontainer/remote:aleaflist {
					    deviate add {
						min-elements 3;
					    }
				        }`,
				},
				deviationTestSchema,
			},
		},
		{
			Description: "Deviation: Can't add a max-elements if one exists",
			ExpResult:   false,
			ExpErrMsg:   "Property being added to node already exists",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `
					deviation /remote:remotecontainer/remote:aleaflist {
					    deviate add {
						max-elements 12;
					    }
				        }`,
				},
				deviationTestSchema,
			},
		},
		{
			Description: "Deviation: Can't add a config if one exists",
			ExpResult:   false,
			ExpErrMsg:   "Property being added to node already exists",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `
					deviation /remote:remotecontainer/remote:defaultleaf {
					    deviate add {
						config true;
					    }
				        }`,
				},
				deviationTestSchema,
			},
		},
	}
	runDeviateTestCases(t, tc)
}

func TestDeviateAddErrorsContainer(t *testing.T) {
	var tc = []testutils.TestCase{
		{
			Description: "Deviation: Can't add a default to container",
			ExpResult:   false,
			ExpErrMsg:   "Property 'default' not allowed on node of type container",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `
					deviation /remote:remotecontainer {
					    deviate add {
						default "newdefault";
					    }
				        }`,
				},
				deviationTestSchema,
			},
		},
		{
			Description: "Deviation: Can't add a unique to container",
			ExpResult:   false,
			ExpErrMsg:   "Property 'unique' not allowed on node of type container",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `
					deviation /remote:remotecontainer {
					    deviate add {
						unique "remoteleaf";
					    }
				        }`,
				},
				deviationTestSchema,
			},
		},
		{
			Description: "Deviation: Can't add a min-elements to container",
			ExpResult:   false,
			ExpErrMsg:   "Property 'min-elements' not allowed on node of type container",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `
					deviation /remote:remotecontainer {
					    deviate add {
						min-elements 3;
					    }
				        }`,
				},
				deviationTestSchema,
			},
		},
		{
			Description: "Deviation: Can't add a max-elements to container",
			ExpResult:   false,
			ExpErrMsg:   "Property 'max-elements' not allowed on node of type container",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `
					deviation /remote:remotecontainer {
					    deviate add {
						max-elements 12;
					    }
				        }`,
				},
				deviationTestSchema,
			},
		},
	}
	runDeviateTestCases(t, tc)
}

func TestDeviateAddErrorsLeaf(t *testing.T) {
	var tc = []testutils.TestCase{
		{
			Description: "Deviation: Can't add a max-elements to a leaf",
			ExpResult:   false,
			ExpErrMsg:   "Property 'max-elements' not allowed on node of type leaf",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `
					deviation /remote:remotecontainer/remote:remoteleaf {
					    deviate add {
						max-elements 3;
					    }
				        }`,
				},
				deviationTestSchema,
			},
		},
		{
			Description: "Deviation: Can't add a min-elements to a leaf",
			ExpResult:   false,
			ExpErrMsg:   "Property 'min-elements' not allowed on node of type leaf",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `
					deviation /remote:remotecontainer/remote:remoteleaf {
					    deviate add {
						min-elements 1;
					    }
				        }`,
				},
				deviationTestSchema,
			},
		},
		{
			Description: "Deviation: Can't add a unique to a leaf",
			ExpResult:   false,
			ExpErrMsg:   "Property 'unique' not allowed on node of type leaf",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `
					deviation /remote:remotecontainer/remote:remoteleaf {
					    deviate add {
						unique anotherleaf;
					    }
				        }`,
				},
				deviationTestSchema,
			},
		},
	}
	runDeviateTestCases(t, tc)
}

func TestDeviateAddErrorsLeafList(t *testing.T) {
	var tc = []testutils.TestCase{
		{
			Description: "Deviation: Can't add a unique to a leaf-list",
			ExpResult:   false,
			ExpErrMsg:   "Property 'unique' not allowed on node of type leaf-list",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `
					deviation /remote:remotecontainer/remote:aleaflist {
					    deviate add {
						unique anotherleaf;
					    }
				        }`,
				},
				deviationTestSchema,
			},
		},
		{
			Description: "Deviation: Can't add a mandatory to a leaf-list",
			ExpResult:   false,
			ExpErrMsg:   "Property 'mandatory' not allowed on node of type leaf-list",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `
					deviation /remote:remotecontainer/remote:aleaflist {
					    deviate add {
						mandatory false;
					    }
				        }`,
				},
				deviationTestSchema,
			},
		},
	}
	runDeviateTestCases(t, tc)
}

func TestDeviateAddErrorsList(t *testing.T) {
	var tc = []testutils.TestCase{
		{
			Description: "Deviation: Can't add a default to a list",
			ExpResult:   false,
			ExpErrMsg:   "Property 'default' not allowed on node of type list",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `
					deviation /remote:remotecontainer/remote:alist {
					    deviate add {
						default "avalue";
					    }
				        }`,
				},
				deviationTestSchema,
			},
		},
		{
			Description: "Deviation: Can't add a mandatory to a list",
			ExpResult:   false,
			ExpErrMsg:   "Property 'mandatory' not allowed on node of type list",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `
					deviation /remote:remotecontainer/remote:alist {
					    deviate add {
						mandatory false;
					    }
				        }`,
				},
				deviationTestSchema,
			},
		},
	}
	runDeviateTestCases(t, tc)
}

func TestDeviateReplace(t *testing.T) {
	var tc = []testutils.TestCase{
		{
			Description: "Deviation: Replace is supported",
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
					deviation /remote:remotecontainer/remote:defaultleaf {
					    deviate replace {
						// Check default, units, config and must are allowed
						default "6800";
						units "borgs";
						config false;
						type uint64;
					    }
				        }
					deviation /remote:remotecontainer/remote:mandatoryleaf {
					    deviate replace {
						// Check mandatory, units, config and must are allowed
						mandatory "false";
						units "borgs";
						config false;
					    }
					}
					deviation /remote:remotecontainer/remote:aleaflist {
                                            deviate replace {
						min-elements 0;
						max-elements unbounded;
					}
				}`,
				},
				deviationTestSchema,
			},
		},
	}
	runDeviateTestCases(t, tc)
}

func TestDeviateDelete(t *testing.T) {
	var tc = []testutils.TestCase{
		{
			Description: "Deviation: delete is supported",
			ExpResult:   true,
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `deviation /remote:remotecontainer/remote:defaultleaf {
					    deviate delete {
						// Check default, units, config and must are allowed
						default "8";
						units "dollars";
					    }
					}
					deviation /remote:remotecontainer {
						deviate delete {
							must "not(anotherleaf = mandatoryleaf)";
						}
					}
					deviation /remote:remotecontainer/remote:alist {
					    deviate delete {
						unique "listdata";
					    }
				}`,
				},
				deviationTestSchema,
			},
		},
	}
	runDeviateTestCases(t, tc)
}

func TestDeviateDeleteReject(t *testing.T) {
	var tc = []testutils.TestCase{
		{
			Description: "Deviation: delete of a type is not allowed",
			ExpResult:   false,
			ExpErrMsg:   "deviate delete: Property not allowed in deviate delete 'type'",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `deviation /remote:remotecontainer/remote:remoteleaf {
					deviate delete {
						type uint32;
					}
				}`,
				},
				deviationTestSchema,
			},
		},
		{
			Description: "Deviation: delete of a default is not allowed if not present",
			ExpResult:   false,
			ExpErrMsg:   "Property being deleted by deviation must exist",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `deviation /remote:remotecontainer/remote:remoteleaf {
					deviate delete {
						default 8;
					}
				}`,
				},
				deviationTestSchema,
			},
		},
		{
			Description: "Deviation: delete of a default is not allowed if does not match",
			ExpResult:   false,
			ExpErrMsg:   "Property being deleted by deviation must exist",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `deviation /remote:remotecontainer/remote:defaultleaf {
					deviate delete {
						default 18;
					}
				}`,
				},
				deviationTestSchema,
			},
		},
		{
			Description: "Deviation: delete of a mandatory property is not allowed if not present",
			ExpResult:   false,
			ExpErrMsg:   "deviate delete: Property not allowed in deviate delete 'mandatory'",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `deviation /remote:remotecontainer/remote:mandatoryleaf {
					// mandatory can not be deleted, only added or replaced
					deviate delete {
						mandatory true;
					}
				}`,
				},
				deviationTestSchema,
			},
		},
		{
			Description: "Deviation: delete of a min-elements is not allowed if not present",
			ExpResult:   false,
			ExpErrMsg:   "deviate delete: Property not allowed in deviate delete 'min-elements'",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `deviation /remote:remotecontainer/remote:aleaflist {
					deviate delete {
						min-elements 1;
					}
				}`,
				},
				deviationTestSchema,
			},
		},
		{
			Description: "Deviation: delete of a max-elements is not allowed if not present",
			ExpResult:   false,
			ExpErrMsg:   "deviate delete: Property not allowed in deviate delete 'max-elements'",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `deviation /remote:remotecontainer/remote:aleaflist {
					deviate delete {
						max-elements 4;
					}
				}`,
				},
				deviationTestSchema,
			},
		},
		{
			Description: "Deviation: delete of a must is not allowed if not identical",
			ExpResult:   false,
			ExpErrMsg:   "Property being deleted by deviation must exist",
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `deviation /remote:remotecontainer {
					deviate delete {
						must "not(mandatoryleaf == anotherleaf)";
					}
				}`,
				},
				deviationTestSchema,
			},
		},
	}
	runDeviateTestCases(t, tc)
}

func TestDeviateDeleteThenAdd(t *testing.T) {
	var tc = []testutils.TestCase{
		{
			Description: "Deviation: combined delete and add",
			ExpResult:   true,
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `deviation /remote:remotecontainer/remote:defaultleaf {
					    deviate delete {
						default "8";
						units "dollars";
					    }
					    deviate add {
						default "9";
						units "GBP";
					    }
					}
					deviation /remote:remotecontainer {
						deviate delete {
							must "not(anotherleaf = mandatoryleaf)";
						}
						deviate add {
							must "not(mandatoryleaf = anotherleaf)";
						}
					}
					deviation /remote:remotecontainer/remote:alist {
					    deviate delete {
						unique "listdata";
					    }
					    deviate add {
						unique "liststringdata";
					    }
				}`,
				},
				deviationTestSchema,
			},
		},
	}
	runDeviateTestCases(t, tc)
}

func TestDeviateUnknownStatement(t *testing.T) {
	var tc = []testutils.TestCase{
		{
			Description: "Deviation: Test unknown statements",
			ExpResult:   true,
			Schemas: []testutils.TestSchema{
				{
					Name: testutils.NameDef{
						Namespace: "prefix-test",
						Prefix:    "test",
					},
					Imports: []testutils.NameDef{
						{"prefix-remote", "remote"}},
					SchemaSnippet: `extension test {
						argument text;
					}
					deviation /remote:remotecontainer {
						deviate not-supported {
							test:test test-unknown {

							}
						}
					}
					deviation /remote:remotecontainer/remote:defaultleaf {
					    deviate delete {
						default "8";
						units "dollars";
						test:test unknown-in-delete;
					    }
					    deviate add {
						test:test unknown-in-add;
						default "9";
						units "GBP";
					    }
				}`,
				},
				deviationTestSchema,
			},
		},
	}
	runDeviateTestCases(t, tc)
}
