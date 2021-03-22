// Copyright (c) 2019-2021, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This XPATH parser is intended to strip out everything except paths
// to allow for validation at compile time.  While invalid paths are
// perfectly acceptable in XPATH, it turns out some NETCONF clients are
// overly restrictive and won't compile them
//.
// Perhaps more relevant is the fact that if our YANG has invalid paths,
// it's highly likely someone made a mistake and we should point it out ...
//
// The grammar is essentially a stripped down version of the 'expr' variant
// as we need to parse exactly the same input; the only difference is that
// we want different output.
//
// In future, it may well make sense to merge this back with the
// full 'expr' grammar.  However, until full predicate and function support
// parsing is done here, it's not clear how easy it will be and for expediency
// a copy has been taken.  No productions have been removed or moved, so it
// is easy to compare the two.

%{

package path_eval

import (
    "encoding/xml"

    "github.com/danos/yang/xpath"
    "github.com/danos/yang/xpath/xutils"
)

%}

%union {
	sym  *xpath.Symbol     /* Symbol table entry */
	val  float64     /* Numeric value */
	name string      /* NodeType or AxisName */
	xmlname xml.Name /* For NameTest */
}

%token	<val>			NUM DOTDOT DBLSLASH DBLCOLON
%token	<sym>			FUNC
%token	<name>			NODETYPE AXISNAME LITERAL
%token	<xmlname>		NAMETEST

/* Set associativity (left or right) and precedence.  Items on one line
 * (eg '+' and  '-') are of equal precedence, but lower than line(s)
 *  below (eg '*' and  '/')
 */

%left OR
%left AND
%left NE EQ
%left GT GE LT LE
%left '+' '-'
%left '*' '/' DIV MOD
%left UNARYMINUS

%%

top:
				Expr
				{
					getProgBldr(pathEvallex).CodeFn(
						getProgBldr(pathEvallex).StorePathEval,
							"storePathEval");
					return 1;
				}
		;
Expr:
				OrExpr
		;
OrExpr:
				AndExpr
		|		OrExpr OR AndExpr
		;
AndExpr:
				EqualityExpr
		|		AndExpr AND EqualityExpr
		;
EqualityExpr:
				RelationalExpr
		|		EqualityExpr EQ RelationalExpr
		|		EqualityExpr NE RelationalExpr
		;
RelationalExpr:
				AdditiveExpr
		|	 	RelationalExpr LT AdditiveExpr
		|	 	RelationalExpr GT AdditiveExpr
		|	 	RelationalExpr LE AdditiveExpr
		|	 	RelationalExpr GE AdditiveExpr
		;
AdditiveExpr:
				MultiplicativeExpr
		|		AdditiveExpr '+' MultiplicativeExpr
		|		AdditiveExpr '-' MultiplicativeExpr
		;
MultiplicativeExpr:
				UnaryExpr
		|		MultiplicativeExpr '*' UnaryExpr
		| 		MultiplicativeExpr DIV UnaryExpr
		| 		MultiplicativeExpr MOD UnaryExpr
		;
UnaryExpr:
				UnionExpr
		|		'-' UnaryExpr %prec UNARYMINUS
		;
UnionExpr:
				PathExpr
		|		UnionExpr '|' PathExpr
		;
PathExpr:
				LocationPath
				{
					getProgBldr(pathEvallex).CodeEvalLocPathExists()
				}
		|		FilterExpr
		|		CompoundFilterExpr '/' RelativeLocationPath
				{
					getProgBldr(pathEvallex).CodeEvalLocPathExists()
				}
		|		CompoundFilterExpr DoubleSlash RelativeLocationPath
				{
					getProgBldr(pathEvallex).CodeEvalLocPathExists()
				}
		;
// This represents a FilterExpr followed by further expression(s).  We need
// to perform certain processing at the end of the filter expression in
// preparation for what follows.
CompoundFilterExpr:
				FilterExpr
		;
FilterExpr:
				PrimaryExpr
		|		FilterExpr Predicate
		;
PrimaryExpr:
				'(' Expr ')'
		|		LITERAL
		|		NUM
		|		FUNC '(' ')'
		|		FUNC '(' Expr ')'
		|		FUNC '(' Expr ',' Expr ')'
		|		FUNC '(' Expr ',' Expr ',' Expr ')'
		|		NODETYPE
				{
					getProgBldr(pathEvallex).UnsupportedName(xutils.NODETYPE, $1);
				}
		;
LocationPath:
				RelativeLocationPath
		|		AbsoluteLocationPath
		;
AbsoluteLocationPath:
				Root
		|		Root RelativeLocationPath
		|		AbbreviatedAbsoluteLocationPath
		;
/*
 * '/' called out into own production so stored in correct order.  Only stored
 * when it indicates an absolute path.  Otherwise there's no point storing
 * it as it provides no extra information over and above individual path
 * elements.
 */
Root:
				'/'
				{
					getProgBldr(pathEvallex).CodePathOper('/');
				}
		;
RelativeLocationPath:
				Step
		|		RelativeLocationPath '/' Step
		|		AbbreviatedRelativeLocationPath
		;
Step:
				// AxisSpecifier can resolve to AbbreviatedAxisSpecifier
				// which in turn resolves to 0 or 1 '@' symbols.  So, we
				// need to allow for this by having NodeTest w/o the
				// AxisSpecifier preceding it as an option here.
				// To further complicate matters, we can have 0 or more
				// Predicates following NodeTest so we have to handle that as
				// well.
				AxisSpecifier NodeTest PredicateSet
		|		AxisSpecifier NodeTest
		|		NodeTest PredicateSet
		|		NodeTest
		|		AbbreviatedStep
		;
AxisSpecifier:
				AXISNAME DBLCOLON
				{
					getProgBldr(pathEvallex).UnsupportedName(xutils.AXISNAME, $1);
				}
		|		AbbreviatedAxisSpecifier
		;
NodeTest:		NAMETEST
				{
					getProgBldr(pathEvallex).CodeNameTest($1);
				}
		;
PredicateSet:
				Predicate
		|		PredicateSet Predicate
		;
Predicate:
				PredicateStart PredicateExpr PredicateEnd
		;
PredicateStart:	'['
				{
					getProgBldr(pathEvallex).CodePredStartIgnore();
				}
		;
PredicateExpr:
				Expr
		;
PredicateEnd:	']'
				{
					getProgBldr(pathEvallex).CodePredEndIgnore();
				}
		;
AbbreviatedAbsoluteLocationPath:
				DoubleSlash RelativeLocationPath
		;
AbbreviatedRelativeLocationPath:
				RelativeLocationPath DoubleSlash Step
		;
AbbreviatedStep:
				'.'
				{
					getProgBldr(pathEvallex).CodePathOper('.');
				}
		|		DOTDOT
				{
					getProgBldr(pathEvallex).CodePathOper(xutils.DOTDOT);
				}
		;
AbbreviatedAxisSpecifier: // 0 or 1 instances
				'@'
				{
					getProgBldr(pathEvallex).UnsupportedName(
						'@', "not yet implemented");
				}
		;
DoubleSlash: //	 Called out into own production so stored in correct order.
				DBLSLASH
				{
					getProgBldr(pathEvallex).UnsupportedName(
						xutils.DBLSLASH, "not yet implemented");
				}
		;
%%
