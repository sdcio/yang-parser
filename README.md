# SDC YANG Parser

![sdc logo](https://docs.sdcio.dev/assets/logos/SDC-transparent-withname-100x133.png)

This repository is part of Schema Driven Configuration (SDC)

The paradigm of schema-driven API approaches is gaining increasing popularity as it facilitates programmatic interaction with systems by both machines and humans. While OpenAPI schema stands out as a widely embraced system, there are other notable schema approaches like YANG, among others. This project endeavors to empower users with a declarative and idempotent method for seamless interaction with API systems, providing a robust foundation for effective system configuration."

## Packages

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

## Functional Description

* Schema / Lexer / Parser: parse Yang, and validate at the individual node
  type level ranges, cardinality etc.  Output is a tree of parse nodes.
* Compile: takes the tree of parse nodes and converts to a Schema tree which
  only contains a subset of nodes.  Type / typedef nodes (what one might
  call data types) become part of the encommpassing node (eg a leaf node).

## Join us

Have questions, ideas, bug reports or just want to chat? Come join [our discord server](https://discord.com/channels/1240272304294985800/1311031796372344894).

## License and Code of Conduct

Code is under the [Apache License 2.0](LICENSE), documentation is [CC BY 4.0](LICENSE-documentation).

The SDC project is following the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/main/code-of-conduct.md). More information and links about the CNCF Code of Conduct are [here](code-of-conduct.md).
