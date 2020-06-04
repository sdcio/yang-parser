// Copyright (c) 2018-2021, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package compile_test

import (
	"bytes"
	"testing"

	"github.com/danos/yang/testutils"
)

func TestAugmentImplicitLocalRef(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment: implicit local module reference",
		ExpResult:   true,
		Template:    BlankTemplate,
		Schema: `container testcontainer {
				description "Test";
			}

			augment /testcontainer {
				leaf testleaf {
					type string;
				}
			} `,
	}
	applyAndVerifySchema(t, &tc, false)
}

func TestAugmentExplicitLocalRef(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment: explicit local module reference",
		ExpResult:   true,
		Template:    BlankTemplate,
		Schema: `container testcontainer {
			}

			augment /test:testcontainer {
				leaf testleaf {
					type string;
				}
			} `,
	}
	applyAndVerifySchema(t, &tc, false)
}

const remoteAugmentSchema = `
container remotecontainer {
	description "Test container";
	leaf remoteleaf {
		type string;
	}
	container remoteInnerCont {
		description "Inner container";
		leaf remoteInnerLeaf {
			type string;
		}
	}
}`

var remoteTestSchema = testutils.TestSchema{
	Name: testutils.NameDef{
		Namespace: "prefix-remote",
		Prefix:    "remote",
	},
	SchemaSnippet: remoteAugmentSchema,
}

func TestAugmentRemoteRef(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment: remote module reference",
		ExpResult:   true,
		Schemas: []testutils.TestSchema{
			{
				Name: testutils.NameDef{
					Namespace: "prefix-test",
					Prefix:    "test",
				},
				Imports: []testutils.NameDef{
					{"prefix-remote", "remote"}},
				SchemaSnippet: `augment /remote:remotecontainer {
					leaf testleaf {
						type string;
					}
				}`,
			},
			remoteTestSchema,
		},
	}
	applyAndVerifySchemas(t, &tc, false)
}

func TestAugmentLocalRefWithMandatoryInNPCont(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment of local with mandatory in container allowed",
		ExpResult:   true,
		Template:    BlankTemplate,
		Schema: `container testcontainer {
				description "Test";
			}

			augment /test:testcontainer {
				leaf testleaf {
					type string;
				}
				container nonpresence {
					leaf mandatoryleaf {
						type string;
						mandatory true;
					}
				}
			} `,
	}
	applyAndVerifySchema(t, &tc, false)
}

func TestAugmentLocalRefMandatoryInPresenceCont(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment with mandatory in presence container allowed",
		ExpResult:   true,
		Template:    BlankTemplate,
		Schema: `container testcontainer {
				description "Test";
			}

			augment /test:testcontainer {
				leaf testleaf {
					type string;
				}
				container presence {
					presence "Allow mandatory in augment";
					leaf mandatoryleaf {
						type string;
						mandatory true;
					}
				}
			} `,
	}
	applyAndVerifySchema(t, &tc, false)
}

func TestAugmentLocalRefMandatoryInList(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment with mandatory in list allowed",
		ExpResult:   true,
		Template:    BlankTemplate,
		Schema: `container testcontainer {
				description "Test";
			}

			augment /test:testcontainer {
				leaf testleaf {
					type string;
				}
				list newlist {
					key name;
					leaf name { type string; }
					leaf mandatoryleaf {
						type string;
						mandatory true;
					}
				}
			} `,
	}
	applyAndVerifySchema(t, &tc, false)
}

func TestAugmentLocalMandatoryInPresenceCont(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment with mandatory in presence container allowed",
		ExpResult:   true,
		Template:    BlankTemplate,
		Schema: `grouping testgroup {
				leaf groupleaf {
					type string;
				}
			}
			container testcontainer {
				description "Test";
			}

			augment /test:testcontainer {
				leaf testleaf {
					type string;
				}
				container presence {
					presence "Allow mandatory in augment";
					leaf mandatoryleaf {
						type string;
						mandatory true;
					}
				}
				uses testgroup;
			} `,
	}
	applyAndVerifySchema(t, &tc, false)
}

func TestAugmentMandatoryLocalRef(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment with a mandatory against a local target",
		ExpResult:   true,
		Template:    BlankTemplate,
		Schema: `container testcontainer {
				description "Test";
			}

			augment /test:testcontainer {
				leaf testleaf {
					type string;
					mandatory true;
				}
			} `,
	}
	applyAndVerifySchema(t, &tc, false)
}

func TestAugmentListLocalRef(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment of a list target is allowed ",
		ExpResult:   true,
		Template:    BlankTemplate,
		Schema: `container testcontainer {
				list testlist {
					key testkey;
					leaf testkey {
						type string;
					}
				}
			}

			augment /test:testcontainer/testlist {
				leaf testleaf {
					type string;
				}
			} `,
	}
	applyAndVerifySchema(t, &tc, false)
}

func TestAugmentLocalRefMandatoryInUses(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment with a mandatory in a uses, local target",
		ExpResult:   true,
		Template:    BlankTemplate,
		Schema: `grouping testgrouping {
				container groupcontainer {
					container subone {
						container subtwo {
							leaf groupleaf {
								type string;
								mandatory "true";
							}
						}
					}
				}
			}
			container testcontainer {
				description "Test";
			}

			augment /test:testcontainer {
				leaf testleafPaul {
					type string;
				}
				uses testgrouping;
			} `,
	}
	applyAndVerifySchema(t, &tc, false)
}

func TestAugmentMandatoryLocalMultiPartPath(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment with a mandatory, multi-part local path",
		ExpResult:   true,
		Template:    BlankTemplate,
		Schema: `container testcontainer {
				description "Test";
			}

			augment /test:testcontainer {
				leaf testleaf {
					type string;
					mandatory false;
				}
				container innercontainer {
					description "local target node";
				}
			}

			augment /test:testcontainer/innercontainer {
				leaf anotherleaf {
					type string;
					mandatory true;
				}
			}
			`,
	}
	applyAndVerifySchema(t, &tc, false)
}

func TestAugmentRPCInput(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment an RPC input statement",
		ExpResult:   true,
		Template:    BlankTemplate,
		Schema: `rpc dosomething {
				description "Test";
				input {
					leaf one {
						type string;
					}
				}
			}

			augment /dosomething/input {
				leaf two {
					type string;
				}
			}`,
	}
	applyAndVerifySchema(t, &tc, false)
}

func TestAugmentRPCInputImplicit(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment an RPC input statement",
		ExpResult:   true,
		Template:    BlankTemplate,
		Schema: `rpc dosomething {
				description "Test";
			}

			augment /dosomething/input {
				leaf two {
					type string;
				}
			}`,
	}
	applyAndVerifySchema(t, &tc, false)
}

func TestAugmentRPCOutput(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment an RPC output statement",
		ExpResult:   true,
		Template:    BlankTemplate,
		Schema: `rpc dosomething {
				description "Test";
				output {
					leaf one {
						type string;
					}
				}
			}

			augment /dosomething/output {
				leaf two {
					type string;
				}
			}`,
	}
	applyAndVerifySchema(t, &tc, false)
}

func TestAugmentRPCOutputImplicit(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment an RPC output statement",
		ExpResult:   true,
		Template:    BlankTemplate,
		Schema: `rpc dosomething {
				description "Test";
			}

			augment /dosomething/output {
				leaf two {
					type string;
				}
			}`,
	}
	applyAndVerifySchema(t, &tc, false)
}

func TestAugmentRemoteWithMandatoryPresenceCont(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment remote with a mandatory presence container",
		ExpResult:   true,
		Template:    BlankTemplate,
		Schemas: []testutils.TestSchema{
			{
				Name: testutils.NameDef{
					Namespace: "prefix-test",
					Prefix:    "test",
				},
				Imports: []testutils.NameDef{
					{"prefix-remote", "remote"}},
				SchemaSnippet: `container testcontainer {
					description "Test";
				}

				augment /remote:remotecontainer {
					list innercontainer {
						key name;
						leaf name { type string; }
						leaf testleaf {
							type string;
							mandatory true;
						}
					}
				}

				augment /remote:remotecontainer/innercontainer {
					container presence {
						presence "I exist therefore I am";
						leaf anotherleaf {
							type string;
							mandatory true;
						}
					}
				}`,
			},
			remoteTestSchema,
		},
	}
	applyAndVerifySchemas(t, &tc, false)
}

// Ensure we augment data node (container) not must node.
func TestAugmentNonDataDefNodeSharesName(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment: non-data-def-node shares name",
		ExpResult:   true,
		Template:    BlankTemplate,
		Schema: `container testcontainer {
				description "Test";
                must "subCont"; // This line must be before container.
                container subCont {
                    presence "true";
                    leaf subLeaf {
                        type string;
                    }
                }
			}

			augment /testcontainer/subCont {
				leaf testleaf {
					type string;
				}
			} `,
	}
	applyAndVerifySchema(t, &tc, false)
}

func TestAugmentInvalidLocalNodeImplicitPrefix(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Implicit local module prefix, invalid node ID",
		ExpResult:   false,
		ExpErrMsg: "augment /badtestcontainer: Invalid path: " +
			"badtestcontainer",
		Template: BlankTemplate,
		Schema: `container testcontainer {
					description "Test";
				}

				augment /badtestcontainer {
					leaf testleaf {
						type string;
					}
				} `,
	}
	applyAndVerifySchema(t, &tc, false)
}

func TestAugmentInvalidLocalNodeExplicitPrefix(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Explicit local module prefix, invalid node ID",
		ExpResult:   false,
		ExpErrMsg: "augment /test:badtestcontainer: Invalid path: " +
			"test:badtestcontainer",
		Template: BlankTemplate,
		Schema: `container testcontainer {
					description "Test";
				}

				augment /test:badtestcontainer {
					leaf testleaf {
						type string;
					}
				} `,
	}
	applyAndVerifySchema(t, &tc, false)
}

func TestAugmentInvalidPrefix(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Unknown import",
		ExpResult:   false,
		ExpErrMsg:   "unknown import testprefix",
		Template:    BlankTemplate,
		Schema: ` augment /testprefix:testcontainer {
					leaf testleaf {
						type string;
					}
				} `,
	}
	applyAndVerifySchema(t, &tc, false)
}

// Augment paths must be absolute
func TestAugmentRelativePathFails(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Not an absolute schema",
		ExpResult:   false,
		ExpErrMsg:   "expected absolute schema id",
		Template:    BlankTemplate,
		Schema: ` augment testprefix:testcontainer {
				 	leaf testleaf {
						type string;
					}
				} `,
	}
	applyAndVerifySchema(t, &tc, false)
}

func TestAugmentLeafNotAllowed(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment of a leaf target node is not allowed",
		ExpResult:   false,
		ExpErrMsg:   "Augment not permitted for target leaf",
		Template:    BlankTemplate,
		Schema: `grouping testgrouping {
					leaf groupleaf {
						type string;
					}
				}
				container testcontainer {
					description "Test";
					leaf targetleaf {
						type string;
					}
				}

				augment /test:testcontainer/targetleaf {
					uses testgrouping;
				} `,
	}
	applyAndVerifySchema(t, &tc, false)
}

func TestAugmentLeafListNotAllowed(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment of a leaf-list target node is not allowed",
		ExpResult:   false,
		ExpErrMsg:   "Augment not permitted for target leaf-list",
		Template:    BlankTemplate,
		Schema: `grouping testgrouping {
					leaf groupleaf {
						type string;
					}
				}
				container testcontainer {
					description "Test";
					leaf-list targetleaf {
						type string;
					}
				}

				augment /test:testcontainer/targetleaf {
					uses testgrouping;
				} `,
	}
	applyAndVerifySchema(t, &tc, false)
}

func TestAugmentRemoteWithMandatoryLeafFails(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment with a mandatory leaf should be rejected",
		ExpResult:   false,
		ExpErrMsg:   "Cannot add mandatory nodes to another module: remote",
		Schemas: []testutils.TestSchema{
			{
				Name: testutils.NameDef{
					Namespace: "prefix-test",
					Prefix:    "test",
				},
				Imports: []testutils.NameDef{
					{"prefix-remote", "remote"}},
				SchemaSnippet: `container testcontainer {
					description "Test";
				}

				augment /remote:remotecontainer {
					leaf testleaf {
						type string;
						mandatory true;
					}
				} `,
			},
			remoteTestSchema,
		},
	}
	applyAndVerifySchemas(t, &tc, false)
}

func TestAugmentRemoteWithMandatoryListFails(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment with a mandatory list should be rejected",
		ExpResult:   false,
		ExpErrMsg:   "Cannot add mandatory nodes to another module: remote",
		Schemas: []testutils.TestSchema{
			{
				Name: testutils.NameDef{
					Namespace: "prefix-test",
					Prefix:    "test",
				},
				Imports: []testutils.NameDef{
					{"prefix-remote", "remote"}},
				SchemaSnippet: `container testcontainer {
					description "Test";
				}

				augment /remote:remotecontainer {
					list testlist {
						key testkey;
						leaf testkey {
							type string;
						}
						min-elements 1;
					}
				} `,
			},
			remoteTestSchema,
		},
	}
	applyAndVerifySchemas(t, &tc, false)
}

func TestAugmentRemoteWithMandatoryLeafListFails(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment with a mandatory leaf-list should be rejected",
		ExpResult:   false,
		ExpErrMsg:   "Cannot add mandatory nodes to another module: remote",
		Schemas: []testutils.TestSchema{
			{
				Name: testutils.NameDef{
					Namespace: "prefix-test",
					Prefix:    "test",
				},
				Imports: []testutils.NameDef{
					{"prefix-remote", "remote"}},
				SchemaSnippet: `container testcontainer {
					description "Test";
				}

				augment /remote:remotecontainer {
					leaf-list testleaflist {
						type string;
						min-elements 1;
					}
				} `,
			},
			remoteTestSchema,
		},
	}
	applyAndVerifySchemas(t, &tc, false)
}

func TestAugmentRemoteWithMinElementListFails(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment with a mandatory container should be rejected",
		ExpResult:   false,
		ExpErrMsg:   "Cannot add mandatory nodes to another module: remote",
		Schemas: []testutils.TestSchema{
			{
				Name: testutils.NameDef{
					Namespace: "prefix-test",
					Prefix:    "test",
				},
				Imports: []testutils.NameDef{
					{"prefix-remote", "remote"}},
				SchemaSnippet: `container testcontainer {
					description "Test";
				}

				augment /remote:remotecontainer {
					container one {
						container two {
							list testlist {
								key testkey;
								leaf testkey {
									type string;
								}
								min-elements 1;
							}
						}
					}
				} `,
			},
			remoteTestSchema,
		},
	}
	applyAndVerifySchemas(t, &tc, false)
}

func TestAugmentRemoteWithMandatoryChoiceFails(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment with a mandatory choice should be rejected",
		ExpResult:   false,
		ExpErrMsg:   "Cannot add mandatory nodes to another module: remote",
		Schemas: []testutils.TestSchema{
			{
				Name: testutils.NameDef{
					Namespace: "prefix-test",
					Prefix:    "test",
				},
				Imports: []testutils.NameDef{
					{"prefix-remote", "remote"}},
				SchemaSnippet: `container testcontainer {
					description "Test";
				}

				augment /remote:remotecontainer {
					choice testchoice {
						mandatory true;

						leaf testleaf {
							type string;
							mandatory true;
						}

						case container {
							container testcontainer {
								leaf testleaf {
									type string;
								}
							}
						}
					}
				} `,
			},
			remoteTestSchema,
		},
	}
	applyAndVerifySchemas(t, &tc, false)
}

func TestAugmentOmmitChoiceCaseFails(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment missing choice and case from path is rejected",
		ExpResult:   false,
		ExpErrMsg:   "augment /testcontainer/container-one: Invalid path: testcontainer/container-one",
		Schemas: []testutils.TestSchema{
			{
				Name: testutils.NameDef{
					Namespace: "prefix-test",
					Prefix:    "test",
				},
				Imports: []testutils.NameDef{
					{"prefix-remote", "remote"}},
				SchemaSnippet: `container testcontainer {
					description "Test";
					choice one {
						container container-one {

						}
						case cont-two {
							container container-two {

							}
						}
					}
				}

				augment /testcontainer/container-one {
					// augment path should be:
					// /testcontainer/one/container-one/container-one
					choice testchoice {
						leaf testleaf {
							type string;
							mandatory true;
						}

						case container {
							container testcontainer {
								leaf testleaf {
									type string;
								}
							}
						}
					}
				} `,
			},
			remoteTestSchema,
		},
	}
	applyAndVerifySchemas(t, &tc, false)
}

func TestAugmentRemoteWithMandatoryLeafLocalPathFails(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment with a mandatory, mixed local remote path",
		ExpResult:   false,
		ExpErrMsg:   "Cannot add mandatory nodes to another module: remote",
		Schemas: []testutils.TestSchema{
			{
				Name: testutils.NameDef{
					Namespace: "prefix-test",
					Prefix:    "test",
				},
				Imports: []testutils.NameDef{
					{"prefix-remote", "remote"}},
				SchemaSnippet: `container testcontainer {
					description "Test";
				}

				augment /remote:remotecontainer {
					leaf testleaf {
						type string;
						mandatory false;
					}
					container innercontainer {
						description "local target node";
					}
				}

				augment /remote:remotecontainer/innercontainer {
					leaf anotherleaf {
						type string;
						mandatory true;
					}
				}
				`,
			},
			remoteTestSchema,
		},
	}
	applyAndVerifySchemas(t, &tc, false)
}

// Testing for implicit interpretation / checking of prefixes on paths
// after the first entry in the path - referring to these paths as
// 'compound' paths as they have >1 element.

// augment /remote:remoteCont/remote:localCont - should FAIL
func TestAugmentCompoundPathRemotePrefixOnLocalNode(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment: compound path, remote prefix on local node",
		ExpResult:   false,
		ExpErrMsg:   "Invalid path: remote:remotecontainer/remote:localCont",
		Schemas: []testutils.TestSchema{
			{
				Name: testutils.NameDef{
					Namespace: "prefix-local",
					Prefix:    "local",
				},
				Imports: []testutils.NameDef{
					{"prefix-remote", "remote"}},
				SchemaSnippet: `container testcontainer {
					description "Test";
				}

				augment /remote:remotecontainer {
					container localCont {
						description "local container";
					}
				}

                // Wrong prefix on localCont
				augment /remote:remotecontainer/remote:localCont {
					leaf localLeaf {
						type string;
					}
				}
				`,
			},
			remoteTestSchema,
		},
	}
	applyAndVerifySchemas(t, &tc, false)
}

// augment /remote:remoteCont/localCont - should PASS
func TestAugmentCompoundPathImplicitLocalPrefix(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment: compound path, implicit local prefix",
		ExpResult:   true,
		Schemas: []testutils.TestSchema{
			{
				Name: testutils.NameDef{
					Namespace: "prefix-local",
					Prefix:    "local",
				},
				Imports: []testutils.NameDef{
					{"prefix-remote", "remote"}},
				SchemaSnippet: `container testcontainer {
					description "Test";
				}

				augment /remote:remotecontainer {
					container localCont {
						description "local container";
					}
				}

				augment /remote:remotecontainer/localCont {
					leaf localLeaf {
						type string;
					}
				}
				`,
			},
			remoteTestSchema,
		},
	}
	applyAndVerifySchemas(t, &tc, false)
}

// augment /remote:remoteCont/local:localCont - should PASS
func TestAugmentCompoundPathExplicitLocalPrefix(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment: compound path, explicit local prefix",
		ExpResult:   true,
		Schemas: []testutils.TestSchema{
			{
				Name: testutils.NameDef{
					Namespace: "prefix-local",
					Prefix:    "local",
				},
				Imports: []testutils.NameDef{
					{"prefix-remote", "remote"}},
				SchemaSnippet: `container testcontainer {
					description "Test";
				}

				augment /remote:remotecontainer {
					container localCont {
						description "local container";
					}
				}

				augment /remote:remotecontainer/local:localCont {
					leaf localLeaf {
						type string;
					}
				}
				`,
			},
			remoteTestSchema,
		},
	}
	applyAndVerifySchemas(t, &tc, false)
}

// augment /remote:remoteCont/remoteInnerCont - should FAIL
func TestAugmentCompoundPathImplicitRemotePrefix(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment: compound path, implicit remote prefix",
		ExpResult:   false,
		ExpErrMsg:   "Invalid path: remote:remotecontainer/remoteInnerCont",
		Schemas: []testutils.TestSchema{
			{
				Name: testutils.NameDef{
					Namespace: "prefix-local",
					Prefix:    "local",
				},
				Imports: []testutils.NameDef{
					{"prefix-remote", "remote"}},
				SchemaSnippet: `container testcontainer {
					description "Test";
				}

				augment /remote:remotecontainer/remoteInnerCont {
					container localCont {
						description "local container";
					}
				}
				`,
			},
			remoteTestSchema,
		},
	}
	applyAndVerifySchemas(t, &tc, false)
}

// augment /remote:remoteCont/remote:remoteInnerCont - should PASS
func TestAugmentCompoundPathExplicitRemotePrefix(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment: compound path, explicit remote prefix",
		ExpResult:   true,
		Schemas: []testutils.TestSchema{
			{
				Name: testutils.NameDef{
					Namespace: "prefix-local",
					Prefix:    "local",
				},
				Imports: []testutils.NameDef{
					{"prefix-remote", "remote"}},
				SchemaSnippet: `
				augment /remote:remotecontainer/remote:remoteInnerCont {
					container localCont {
						description "local container";
					}
				}
				`,
			},
			remoteTestSchema,
		},
	}
	applyAndVerifySchemas(t, &tc, false)
}

// augment /localCont/localCont - should PASS
func TestAugmentCompoundPathImplicitLocalPrefixes(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment: compound path, implicit local prefix",
		ExpResult:   true,
		Schemas: []testutils.TestSchema{
			{
				Name: testutils.NameDef{
					Namespace: "prefix-local",
					Prefix:    "local",
				},
				SchemaSnippet: `container localCont {
					description "Test";
					container innerCont {
						description "InnerCont";
					}
				}

				augment /localCont/innerCont {
					container localCont {
						description "local container";
					}
				}
				`,
			},
		},
	}
	applyAndVerifySchemas(t, &tc, false)
}

// augment /localCont/localCont - should PASS
func TestAugmentCompoundPathExplicitLocalPrefixes(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment: compound path, explicit local prefix",
		ExpResult:   true,
		Schemas: []testutils.TestSchema{
			{
				Name: testutils.NameDef{
					Namespace: "prefix-local",
					Prefix:    "local",
				},
				SchemaSnippet: `container localCont {
					description "Test";
					container innerCont {
						description "InnerCont";
					}
				}

				augment /local:localCont/local:innerCont {
					container innermostCont {
						description "local container";
					}
				}
				`,
			},
		},
	}
	applyAndVerifySchemas(t, &tc, false)
}

const moreRemoteAugmentSchema = `
	augment /remote:remotecontainer {
	container moreRemoteCont {
		description "more remote container";
	}
}`

var moreRemoteTestSchema = testutils.TestSchema{
	Name: testutils.NameDef{
		Namespace: "prefix-moreRemote",
		Prefix:    "moreRemote",
	},
	Imports: []testutils.NameDef{
		{"prefix-remote", "remote"}},
	SchemaSnippet: moreRemoteAugmentSchema,
}

// augment /remote:remoteCont/moreRemoteCont - should FAIL
func TestAugmentCompoundPathImplicitMoreRemotePrefix(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment: compound path, implicit more remote prefix",
		ExpResult:   false,
		ExpErrMsg:   "Invalid path: remote:remotecontainer/moreRemoteCont",
		Schemas: []testutils.TestSchema{
			{
				Name: testutils.NameDef{
					Namespace: "prefix-local",
					Prefix:    "local",
				},
				Imports: []testutils.NameDef{
					{"prefix-remote", "remote"},
					{"prefix-moreRemote", "moreRemote"}},
				SchemaSnippet: `
				augment /remote:remotecontainer/moreRemoteCont {
					container localCont {
						description "local container";
					}
				}
				`,
			},
			remoteTestSchema,
			moreRemoteTestSchema,
		},
	}
	applyAndVerifySchemas(t, &tc, false)
}

// augment /remote:remoteCont/moreRemote:moreRemoteCont - should PASS
func TestAugmentCompoundPathExplicitMoreRemotePrefix(t *testing.T) {
	var tc = testutils.TestCase{
		Description: "Augment: compound path, explicit more remote prefix",
		ExpResult:   true,
		Schemas: []testutils.TestSchema{
			{
				Name: testutils.NameDef{
					Namespace: "prefix-local",
					Prefix:    "local",
				},
				Imports: []testutils.NameDef{
					{"prefix-remote", "remote"},
					{"prefix-moreRemote", "moreRemote"}},
				SchemaSnippet: `
				augment /remote:remotecontainer/moreRemote:moreRemoteCont {
					container localCont {
						description "local container";
					}
				}
				`,
			},
			remoteTestSchema,
			moreRemoteTestSchema,
		},
	}
	applyAndVerifySchemas(t, &tc, false)
}

func TestAugmentRestrictionOfChoice(t *testing.T) {
	t.Skipf("Choice not currently supported")
}

func TestAugmentRestictionsOfCase(t *testing.T) {
	t.Skipf("Case not currently supported")
}

func TestAugmentRestrictionsOfNotification(t *testing.T) {
	t.Skipf("Notification not currently supported")
}

func TestAugmentOfAugment(t *testing.T) {
	module1_text := bytes.NewBufferString(
		`module test-yang-compile1 {
		namespace "urn:vyatta.com:test:yang-compile1";
		prefix test;

		organization "Brocade Communications Systems, Inc.";
		revision 2014-12-29 {
			description "Test schema";
		}

		container one {
		}
	}`)

	module2_text := bytes.NewBufferString(
		`module test-yang-compile2 {
		namespace "urn:vyatta.com:test:yang-compile2";
		prefix test;

		import test-yang-compile1 { prefix compile1; }

		organization "Brocade Communications Systems, Inc.";
		revision 2014-12-29 {
			description "Test schema";
		}

		augment /compile1:one {
			container two {
			}
		}
	}`)

	module3_text := bytes.NewBufferString(
		`module test-yang-compile3 {
		namespace "urn:vyatta.com:test:yang-compile3";
		prefix test;

		import test-yang-compile1 { prefix compile1; }
		import test-yang-compile2 { prefix compile2; }

		organization "Brocade Communications Systems, Inc.";
		revision 2014-12-29 {
			description "Test schema";
		}

		augment /compile1:one/compile2:two {
			leaf three {
				type string;
			}
		}
	}`)

	expected := NewLeafChecker("three")

	st, err := testutils.GetConfigSchema(
		module1_text.Bytes(),
		module2_text.Bytes(),
		module3_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected error %s", err.Error())
	}

	if actual := findSchemaNodeInTree(t, st,
		[]string{"one", "two", "three"}); actual != nil {
		expected.check(t, actual)
	}
}
