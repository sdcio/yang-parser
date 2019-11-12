Packages
========

  * compile / compile_test
	* compiletest
  * data
	* datanode
	* encoding
  * parse / parse_test
	* parsetest
  * schema / schema_test
	* schematests
  * testutils
	* assert
  * xpath
	* grammars
	  * expr
	  * leafref
	  * lexertest
	* xpathtest
	* xutils / xutils_test

Functional Description
----------------------

  * Schema / Lexer / Parser: parse Yang, and validate at the individual node
    type level ranges, cardinality etc.  Output is a tree of parse nodes.
  * Compile: takes the tree of parse nodes and converts to a Schema tree which
    only contains a subset of nodes.  Type / typedef nodes (what one might
    call data types) become part of the encommpassing node (eg a leaf node).
