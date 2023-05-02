// Copyright (c) 2019-2021, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// XPATH parser for the leaf-ref / path statements.  Subset of full XPATH
// functionality.

%{

package leafref

import (
    "encoding/xml"

    "github.com/iptecharch/yang-parser/xpath"
    "github.com/iptecharch/yang-parser/xpath/xutils"
)

%}

%union {
	sym  *xpath.Symbol /* Symbol table entry */
	val  float64       /* Numeric value */
	xmlname xml.Name   /* For NameTest */
}

%token	<val>			DOTDOT EQ ERR
%token	<sym>			FUNC
%token	<xmlname>		NAMETEST

%%

top:
				Expr
				{
					getProgBldr(leafreflex).CodeFn(
						getProgBldr(leafreflex).Store, "store");
				}
		;
Expr:
				AbsolutePath
				{
					getProgBldr(leafreflex).CodeFn(
						getProgBldr(leafreflex).EvalLocPath, "evalLocPath");
				}
		|		RelativePath
				{
					getProgBldr(leafreflex).CodeFn(
						getProgBldr(leafreflex).EvalLocPath, "evalLocPath");
				}
		;
AbsolutePath:	Root NodeIdentifier PathPredicate1Plus AbsolutePathStep
		|		Root NodeIdentifier PathPredicate1Plus
		|		Root NodeIdentifier AbsolutePathStep
		|		Root NodeIdentifier
		;
AbsolutePathStep:
				'/' NodeIdentifier PathPredicate1Plus AbsolutePathStep
		|		'/' NodeIdentifier PathPredicate1Plus
		|		'/' NodeIdentifier AbsolutePathStep
		|		'/' NodeIdentifier
		;
/*
 * Record '/' only when at the start of an absolute path, so we can
 * differentiate from a relative path.  At any other point in the path
 * it is implicit.
 */
Root:           '/'
				{
					getProgBldr(leafreflex).CodePathOper('/');
				}
		;
RelativePath:	DotDot '/' RelativePath
		|		DotDot '/' DescendantPath
		;
DescendantPath:	NodeIdentifier PathPredicate1Plus AbsolutePathStep
		|		NodeIdentifier AbsolutePathStep
		|		NodeIdentifier
		;
PathPredicate1Plus:
				StartPred PathEqualityExpr EndPred PathPredicate1Plus
		|		StartPred PathEqualityExpr EndPred
		;
/*
 * We actually just call EvalLocPath() for '[' (predicate start) but to aid
 * in debugging machines, we give it a different name so the printed machine
 * shows 'LRefPredStart'.
 */
StartPred:		'['
				{
					getProgBldr(leafreflex).CodeFn(
						getProgBldr(leafreflex).EvalLocPath, "lrefPredStart");
				}
		;
EndPred:		']'
				{
					getProgBldr(leafreflex).CodeFn(
						getProgBldr(leafreflex).LRefPredEnd, "lrefPredEnd");
				}
		;
/*
 * We use EQ as it works better with the common lexer code than an explicit
 * '=' here.  Comes to the same thing eventually.
 */
PathEqualityExpr:
				NodeIdentifier Equals PathKeyExpr
		;
Equals:			EQ
				{
					getProgBldr(leafreflex).CodeFn(
						getProgBldr(leafreflex).LRefEquals, "lrefEquals");
				}
		;
PathKeyExpr:	CurrentFnInvocation '/' RelPathKeyExpr
		;
RelPathKeyExpr:	UpDir1Plus NodeIdentifierSlash1Plus NodeIdentifier
		|		UpDir1Plus NodeIdentifier
		;
/*
 * Up one or more directories.
 */
UpDir1Plus:		UpDir1Plus DotDot '/'
		|		DotDot '/'
		;
/*
 * One or more sets of a node-identifier followed by slash
 */
NodeIdentifierSlash1Plus:
				NodeIdentifierSlash1Plus NodeIdentifier '/'
		|		NodeIdentifier '/'
		;
DotDot:			DOTDOT
				{
					getProgBldr(leafreflex).CodePathOper(xutils.DOTDOT);
				}
		;
/*
 * Essentially a NAMETEST, but a more restrictive set of characters.
 *
 * First char is [a-zA-Z_].  Subsequent may be those, and additionally
 * '-', '.', or 0-9.  First 3 characters may not be 'XML' (any case, can be
 * mixed).
 */
NodeIdentifier:	NAMETEST
				{
					getProgBldr(leafreflex).CodeNameTest($1);
				}
		;
/*
 * As we know current is at the start of a path, we know that our context
 * will be the current node, and so can replace with '.'.
 *
 * Additionally, the lexer will reject any FUNC that is not 'current'.
 */
CurrentFnInvocation:
				FUNC '(' ')'
				{
					getProgBldr(leafreflex).CodePathOper('.');
				}
		;
%%

/* Code is in .go files so we get the benefit of gofmt etc.
 * What's above is formatted as best as emacs Bison-mode will allow,
 * with semi-colons added to help Bison-mode think the code is C!
 *
 * If anyone can come up with a better formatting model I'm all ears ... (-:
 */





