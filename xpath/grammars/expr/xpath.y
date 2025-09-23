// Copyright (c) 2019-2021, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This XPATH parser owes its inspiration to 2 sources:
//
// (a) HOC (high order calculator) described in The Unix Programming
//     Environment provided the basis for the initial grammar specification
//     and the run-later machine.  This was of course written in C.
//
// (b) The 'expr' basic calculator provided with the Go language source
//     provided an example YACC implementation in Go.
//

%{

package expr

import (
    "encoding/xml"

    "github.com/sdcio/yang-parser/xpath"
    "github.com/sdcio/yang-parser/xpath/xutils"
)

%}

%union {
	sym  *xpath.Symbol     /* Symbol table entry */
	val  float64     /* Numeric value */
	name string      /* NodeType or AxisName */
	xmlname xml.Name /* For NameTest */
}

%token	<val>			NUM DOTDOT DBLSLASH DBLCOLON ERR
%token	<sym>			FUNC TEXTFUNC
%token	<name>			NODETYPE AXISNAME LITERAL
%token	<xmlname>		NAMETEST

%token CURRENTFUNC DEREFFUNC COUNTFUNC


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
					getProgBldr(exprlex).CodeFn(
						getProgBldr(exprlex).Store, "store");
				}
		;
Expr:
				OrExpr
		;
OrExpr:
				AndExpr
		|		OrExpr OR AndExpr
				{
					getProgBldr(exprlex).CodeFn(
						getProgBldr(exprlex).Or, "or");
				}
		;
AndExpr:
				EqualityExpr
		|		AndExpr AND EqualityExpr
				{
					getProgBldr(exprlex).CodeFn(
						getProgBldr(exprlex).And, "and");
				}
		;
EqualityExpr:
				RelationalExpr
		|		EqualityExpr EQ RelationalExpr
				{
					getProgBldr(exprlex).CodeFn(
						getProgBldr(exprlex).Eq, "eq");
				}
		|		EqualityExpr NE RelationalExpr
				{
					getProgBldr(exprlex).CodeFn(
						getProgBldr(exprlex).Ne, "ne");
				}
		;
RelationalExpr:
				AdditiveExpr
		|	 	RelationalExpr LT AdditiveExpr
				{
					getProgBldr(exprlex).CodeFn(
						getProgBldr(exprlex).Lt, "lt");
				}
		|	 	RelationalExpr GT AdditiveExpr
				{
					getProgBldr(exprlex).CodeFn(
						getProgBldr(exprlex).Gt, "gt");
				}
		|	 	RelationalExpr LE AdditiveExpr
				{
					getProgBldr(exprlex).CodeFn(
						getProgBldr(exprlex).Le, "le");
				}
		|	 	RelationalExpr GE AdditiveExpr
				{
					getProgBldr(exprlex).CodeFn(
						getProgBldr(exprlex).Ge, "ge");
				}
		;
AdditiveExpr:
				MultiplicativeExpr
		|		AdditiveExpr '+' MultiplicativeExpr
				{
					getProgBldr(exprlex).CodeFn(
						   getProgBldr(exprlex).Add, "add");
				}
		|		AdditiveExpr '-' MultiplicativeExpr
				{
					getProgBldr(exprlex).CodeFn(
						   getProgBldr(exprlex).Sub, "sub");
				}
		;
MultiplicativeExpr:
				UnaryExpr
		|		MultiplicativeExpr '*' UnaryExpr
				{
					getProgBldr(exprlex).CodeFn(
						   getProgBldr(exprlex).Mul, "mul");
				}
		| 		MultiplicativeExpr DIV UnaryExpr
				{
					getProgBldr(exprlex).CodeFn(
						   getProgBldr(exprlex).Div, "div");
				}
		| 		MultiplicativeExpr MOD UnaryExpr
				{
					getProgBldr(exprlex).CodeFn(
						   getProgBldr(exprlex).Mod, "mod");
				}
		;
UnaryExpr:
				UnionExpr
		|		'-' UnaryExpr %prec UNARYMINUS
				{
					getProgBldr(exprlex).CodeFn(
						getProgBldr(exprlex).Negate, "negate");
				}
		;
UnionExpr:
				PathExpr
		|		UnionExpr '|' PathExpr
				{
					getProgBldr(exprlex).CodeFn(
						getProgBldr(exprlex).Union, "union");
				}
		;
PathExpr:
				LocationPath
				{
					getProgBldr(exprlex).CodeFn(
						getProgBldr(exprlex).EvalLocPath, "evalLocPath");
				}
		|		FilterExpr
		|		CompoundFilterExpr '/' RelativeLocationPath
				{
					getProgBldr(exprlex).CodeFn(
						getProgBldr(exprlex).EvalLocPath, "evalLocPath");
				}
		|		CompoundFilterExpr DoubleSlash RelativeLocationPath
				{
					getProgBldr(exprlex).CodeFn(
						getProgBldr(exprlex).EvalLocPath, "evalLocPath");
				}
		;
// This represents a FilterExpr followed by further expression(s).  We need
// to perform certain processing at the end of the filter expression in
// preparation for what follows.
CompoundFilterExpr:
				FilterExpr
				{
					getProgBldr(exprlex).CodeFn(
						getProgBldr(exprlex).FilterExprEnd, "filterExprEnd");
				}
		;
FilterExpr:
				PrimaryExpr
		|		FilterExpr Predicate
		;
PrimaryExpr:
				'(' Expr ')'
		|		'(' ')'
		|		LITERAL
				{
					getProgBldr(exprlex).CodeLiteral($1);
				}
		|		NUM
				{
					getProgBldr(exprlex).CodeNum($1);
				}
		|		TEXTFUNC '(' ')'
				{
					getProgBldr(exprlex).Text();
				}
		|		FUNC '(' ')'
				{
					getProgBldr(exprlex).CodeBltin($1, 0);
				}
		|		FUNC '(' Expr ')'
				{
					getProgBldr(exprlex).CodeBltin($1, 1);
				}
		|		FUNC '(' Expr ',' Expr ')'
				{
					getProgBldr(exprlex).CodeBltin($1, 2);
				}
		|		FUNC '(' Expr ',' Expr ',' Expr ')'
				{
					getProgBldr(exprlex).CodeBltin($1, 3);
 				}
		|		NODETYPE
				{
					getProgBldr(exprlex).UnsupportedName(xutils.NODETYPE, $1);
				}
		;
LocationPath:
				RelativeLocationPath
		|		AbsoluteLocationPath
		|       CurrentRelativeLocationPath
		|       DerefRelativeLocationPath
		|       CountRelativeLocationPath
		;
AbsoluteLocationPath:
				Root
		|		Root RelativeLocationPath
		|		AbbreviatedAbsoluteLocationPath
		;
CurrentRelativeLocationPath:
                CurrentFunc
        |       CurrentFunc '/' RelativeLocationPath
            ;
CurrentFunc:
                CURRENTFUNC '(' ')'
                {
                    getProgBldr(exprlex).CodePathSetCurrent();
                }
                ;

DerefRelativeLocationPath:
                DerefFunc
         |      DerefFunc '/' RelativeLocationPath
         ;

DerefFunc:
		DEREFFUNC '(' LocationPath ')'
				{
					getProgBldr(exprlex).Deref();
				}
				;

CountRelativeLocationPath:
                CountFunc
         |      CountFunc '/' RelativeLocationPath
         ;

CountFunc:
		COUNTFUNC '(' LocationPath ')'
				{
					getProgBldr(exprlex).Count();
				}
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
					getProgBldr(exprlex).CodePathOper('/');
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
				AxisSpecifier NodeTest PredicatesStart PredicateSet PredicatesEnd
		|		AxisSpecifier NodeTest
		|		NodeTest PredicatesStart PredicateSet PredicatesEnd
		|		NodeTest
		|		AbbreviatedStep
		;
AxisSpecifier:
				AXISNAME DBLCOLON
				{
					getProgBldr(exprlex).UnsupportedName(xutils.AXISNAME, $1);
				}
		|		AbbreviatedAxisSpecifier
		;
NodeTest:		NAMETEST
				{
					getProgBldr(exprlex).CodeNameTest($1);
				}
		;
PredicateSet:
				Predicate 
		|		PredicateSet Predicate
		;
PredicatesStart:
				{
					getProgBldr(exprlex).PredicatesStart();
				}

PredicatesEnd:
				{
					getProgBldr(exprlex).PredicatesEnd();
				}
Predicate:
				PredicateStart PredicateExpr PredicateEnd
		;
PredicateStart:	'['
				{
					getProgBldr(exprlex).CodePredStart();
				}
		;
PredicateExpr:
				Expr
		;
PredicateEnd:	']'
				{
					getProgBldr(exprlex).CodePredEnd();
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
					getProgBldr(exprlex).CodePathOper('.');
				}
		|		DOTDOT
				{
					getProgBldr(exprlex).CodePathOper(xutils.DOTDOT);
				}
		;
AbbreviatedAxisSpecifier: // 0 or 1 instances
				'@'
				{
					getProgBldr(exprlex).UnsupportedName(
						'@', "not yet implemented");
				}
		;
DoubleSlash: //	 Called out into own production so stored in correct order.
				DBLSLASH
				{
					getProgBldr(exprlex).UnsupportedName(
						xutils.DBLSLASH, "not yet implemented");
				}
		;
%%

/* Code is in .go files so we get the benefit of gofmt etc.
 * What's above is formatted as best as emacs Bison-mode will allow,
 * with semi-colons added to help Bison-mode think the code is C!
 *
 * If anyone can come up with a better formatting model I'm all ears ... (-:
 */





