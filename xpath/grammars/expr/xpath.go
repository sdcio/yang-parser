// Code generated by goyacc -o xpath.go -p expr xpath.y. DO NOT EDIT.

//line xpath.y:19

package expr

import __yyfmt__ "fmt"

//line xpath.y:20

import (
	"encoding/xml"

	"github.com/sdcio/yang-parser/xpath"
	"github.com/sdcio/yang-parser/xpath/xutils"
)

//line xpath.y:31
type exprSymType struct {
	yys     int
	sym     *xpath.Symbol /* Symbol table entry */
	val     float64       /* Numeric value */
	name    string        /* NodeType or AxisName */
	xmlname xml.Name      /* For NameTest */
}

const NUM = 57346
const DOTDOT = 57347
const DBLSLASH = 57348
const DBLCOLON = 57349
const ERR = 57350
const FUNC = 57351
const TEXTFUNC = 57352
const NODETYPE = 57353
const AXISNAME = 57354
const LITERAL = 57355
const NAMETEST = 57356
const CURRENTFUNC = 57357
const OR = 57358
const AND = 57359
const NE = 57360
const EQ = 57361
const GT = 57362
const GE = 57363
const LT = 57364
const LE = 57365
const DIV = 57366
const MOD = 57367
const UNARYMINUS = 57368

var exprToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"NUM",
	"DOTDOT",
	"DBLSLASH",
	"DBLCOLON",
	"ERR",
	"FUNC",
	"TEXTFUNC",
	"NODETYPE",
	"AXISNAME",
	"LITERAL",
	"NAMETEST",
	"CURRENTFUNC",
	"OR",
	"AND",
	"NE",
	"EQ",
	"GT",
	"GE",
	"LT",
	"LE",
	"'+'",
	"'-'",
	"'*'",
	"'/'",
	"DIV",
	"MOD",
	"UNARYMINUS",
	"'|'",
	"'('",
	"')'",
	"','",
	"'['",
	"']'",
	"'.'",
	"'@'",
}

var exprStatenames = [...]string{}

const exprEofCode = 1
const exprErrCode = 2
const exprInitialStackSize = 16

//line xpath.y:354

/* Code is in .go files so we get the benefit of gofmt etc.
 * What's above is formatted as best as emacs Bison-mode will allow,
 * with semi-colons added to help Bison-mode think the code is C!
 *
 * If anyone can come up with a better formatting model I'm all ears ... (-:
 */

//line yacctab:1
var exprExca = [...]int8{
	-1, 1,
	1, -1,
	-2, 0,
	-1, 14,
	6, 30,
	27, 30,
	-2, 27,
}

const exprPrivate = 57344

const exprLast = 208

var exprAct = [...]int8{
	2, 73, 35, 72, 16, 8, 6, 4, 9, 5,
	106, 61, 12, 20, 41, 113, 59, 104, 63, 65,
	58, 37, 99, 39, 110, 111, 68, 66, 107, 108,
	98, 7, 75, 70, 69, 57, 54, 45, 55, 56,
	74, 32, 67, 52, 53, 42, 40, 43, 49, 51,
	48, 50, 77, 79, 80, 78, 47, 46, 85, 86,
	44, 92, 39, 87, 88, 89, 64, 93, 94, 65,
	90, 101, 97, 71, 103, 102, 76, 65, 95, 96,
	81, 82, 83, 84, 42, 105, 91, 60, 38, 33,
	31, 21, 24, 23, 22, 18, 65, 65, 17, 19,
	65, 15, 14, 13, 103, 62, 10, 3, 1, 109,
	0, 0, 112, 27, 41, 42, 0, 0, 29, 28,
	30, 37, 26, 39, 36, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 11, 0, 34, 0, 0, 0,
	0, 25, 100, 27, 41, 42, 40, 43, 29, 28,
	30, 37, 26, 39, 36, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 11, 0, 34, 0, 0, 0,
	0, 25, 0, 27, 41, 42, 40, 43, 29, 28,
	30, 37, 26, 39, 36, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 34, 0, 0, 0,
	0, 25, 0, 0, 0, 0, 40, 43,
}

var exprPact = [...]int16{
	139, -1000, -1000, 44, 20, 38, 28, 19, 10, -1000,
	4, 139, -1000, -1000, -24, 78, 39, -1000, -1000, -1000,
	-1000, -1000, 9, -1000, 15, 139, -1000, -1000, 2, 1,
	-1000, 48, -24, -1000, -1000, 9, 0, 69, -1000, -1000,
	-1000, -1000, -1000, -1000, 139, 139, 139, 139, 139, 139,
	139, 139, 139, 139, 139, 139, 139, 169, -1000, -1000,
	139, -1000, 9, 9, 9, 9, 39, 9, -3, -11,
	109, -24, -24, -1000, 39, -16, -1000, 20, 38, 28,
	28, 19, 19, 19, 19, 10, 10, -1000, -1000, -1000,
	-1000, -26, -1000, 39, 39, -1000, -1000, 39, -1000, -1000,
	-1000, -5, -24, -1000, -1000, -1000, -1000, -1000, 139, -9,
	-1000, 139, -18, -1000,
}

var exprPgo = [...]int8{
	0, 108, 0, 107, 7, 9, 6, 31, 5, 8,
	106, 12, 103, 102, 101, 4, 2, 99, 1, 98,
	95, 94, 93, 92, 13, 91, 90, 41, 3, 89,
	88, 87, 86, 85,
}

var exprR1 = [...]int8{
	0, 1, 2, 3, 3, 4, 4, 5, 5, 5,
	6, 6, 6, 6, 6, 7, 7, 7, 8, 8,
	8, 8, 9, 9, 10, 10, 11, 11, 11, 11,
	14, 13, 13, 17, 17, 17, 17, 17, 17, 17,
	17, 17, 12, 12, 12, 19, 19, 19, 20, 23,
	21, 15, 15, 15, 24, 24, 24, 24, 24, 26,
	26, 27, 28, 28, 18, 31, 32, 33, 22, 25,
	29, 29, 30, 16,
}

var exprR2 = [...]int8{
	0, 1, 1, 1, 3, 1, 3, 1, 3, 3,
	1, 3, 3, 3, 3, 1, 3, 3, 1, 3,
	3, 3, 1, 2, 1, 3, 1, 1, 3, 3,
	1, 1, 2, 3, 1, 1, 3, 3, 4, 6,
	8, 1, 1, 1, 1, 1, 2, 1, 3, 3,
	1, 1, 3, 1, 3, 2, 2, 1, 1, 2,
	1, 1, 1, 2, 3, 1, 1, 1, 2, 3,
	1, 1, 1, 1,
}

var exprChk = [...]int16{
	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, 25, -11, -12, -13, -14, -15, -19, -20, -17,
	-24, -25, -21, -22, -23, 32, 13, 4, 10, 9,
	11, -26, -27, -29, 27, -16, 15, 12, -30, 14,
	37, 5, 6, 38, 16, 17, 19, 18, 22, 20,
	23, 21, 24, 25, 26, 28, 29, 31, -9, -18,
	-31, 35, 27, -16, 27, -16, -15, 27, -2, 32,
	32, -27, -28, -18, -15, 32, 7, -4, -5, -6,
	-6, -7, -7, -7, -7, -8, -8, -9, -9, -9,
	-11, -32, -2, -15, -15, -24, -24, -15, 33, 33,
	33, -2, -28, -18, 33, -33, 36, 33, 34, -2,
	33, 34, -2, 33,
}

var exprDef = [...]int8{
	0, -2, 1, 2, 3, 5, 7, 10, 15, 18,
	22, 0, 24, 26, -2, 0, 42, 43, 44, 31,
	51, 53, 45, 47, 0, 0, 34, 35, 0, 0,
	41, 0, 57, 58, 50, 0, 0, 0, 60, 61,
	70, 71, 73, 72, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 23, 32,
	0, 65, 0, 0, 0, 0, 46, 0, 0, 0,
	0, 55, 56, 62, 68, 0, 59, 4, 6, 8,
	9, 11, 12, 13, 14, 16, 17, 19, 20, 21,
	25, 0, 66, 28, 29, 52, 69, 48, 33, 36,
	37, 0, 54, 63, 49, 64, 67, 38, 0, 0,
	39, 0, 0, 40,
}

var exprTok1 = [...]int8{
	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	32, 33, 26, 24, 34, 25, 37, 27, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 38, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 35, 3, 36, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 31,
}

var exprTok2 = [...]int8{
	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 28, 29, 30,
}

var exprTok3 = [...]int8{
	0,
}

var exprErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	exprDebug        = 0
	exprErrorVerbose = false
)

type exprLexer interface {
	Lex(lval *exprSymType) int
	Error(s string)
}

type exprParser interface {
	Parse(exprLexer) int
	Lookahead() int
}

type exprParserImpl struct {
	lval  exprSymType
	stack [exprInitialStackSize]exprSymType
	char  int
}

func (p *exprParserImpl) Lookahead() int {
	return p.char
}

func exprNewParser() exprParser {
	return &exprParserImpl{}
}

const exprFlag = -1000

func exprTokname(c int) string {
	if c >= 1 && c-1 < len(exprToknames) {
		if exprToknames[c-1] != "" {
			return exprToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func exprStatname(s int) string {
	if s >= 0 && s < len(exprStatenames) {
		if exprStatenames[s] != "" {
			return exprStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func exprErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !exprErrorVerbose {
		return "syntax error"
	}

	for _, e := range exprErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + exprTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := int(exprPact[state])
	for tok := TOKSTART; tok-1 < len(exprToknames); tok++ {
		if n := base + tok; n >= 0 && n < exprLast && int(exprChk[int(exprAct[n])]) == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if exprDef[state] == -2 {
		i := 0
		for exprExca[i] != -1 || int(exprExca[i+1]) != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; exprExca[i] >= 0; i += 2 {
			tok := int(exprExca[i])
			if tok < TOKSTART || exprExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if exprExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += exprTokname(tok)
	}
	return res
}

func exprlex1(lex exprLexer, lval *exprSymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = int(exprTok1[0])
		goto out
	}
	if char < len(exprTok1) {
		token = int(exprTok1[char])
		goto out
	}
	if char >= exprPrivate {
		if char < exprPrivate+len(exprTok2) {
			token = int(exprTok2[char-exprPrivate])
			goto out
		}
	}
	for i := 0; i < len(exprTok3); i += 2 {
		token = int(exprTok3[i+0])
		if token == char {
			token = int(exprTok3[i+1])
			goto out
		}
	}

out:
	if token == 0 {
		token = int(exprTok2[1]) /* unknown char */
	}
	if exprDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", exprTokname(token), uint(char))
	}
	return char, token
}

func exprParse(exprlex exprLexer) int {
	return exprNewParser().Parse(exprlex)
}

func (exprrcvr *exprParserImpl) Parse(exprlex exprLexer) int {
	var exprn int
	var exprVAL exprSymType
	var exprDollar []exprSymType
	_ = exprDollar // silence set and not used
	exprS := exprrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	exprstate := 0
	exprrcvr.char = -1
	exprtoken := -1 // exprrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		exprstate = -1
		exprrcvr.char = -1
		exprtoken = -1
	}()
	exprp := -1
	goto exprstack

ret0:
	return 0

ret1:
	return 1

exprstack:
	/* put a state and value onto the stack */
	if exprDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", exprTokname(exprtoken), exprStatname(exprstate))
	}

	exprp++
	if exprp >= len(exprS) {
		nyys := make([]exprSymType, len(exprS)*2)
		copy(nyys, exprS)
		exprS = nyys
	}
	exprS[exprp] = exprVAL
	exprS[exprp].yys = exprstate

exprnewstate:
	exprn = int(exprPact[exprstate])
	if exprn <= exprFlag {
		goto exprdefault /* simple state */
	}
	if exprrcvr.char < 0 {
		exprrcvr.char, exprtoken = exprlex1(exprlex, &exprrcvr.lval)
	}
	exprn += exprtoken
	if exprn < 0 || exprn >= exprLast {
		goto exprdefault
	}
	exprn = int(exprAct[exprn])
	if int(exprChk[exprn]) == exprtoken { /* valid shift */
		exprrcvr.char = -1
		exprtoken = -1
		exprVAL = exprrcvr.lval
		exprstate = exprn
		if Errflag > 0 {
			Errflag--
		}
		goto exprstack
	}

exprdefault:
	/* default state action */
	exprn = int(exprDef[exprstate])
	if exprn == -2 {
		if exprrcvr.char < 0 {
			exprrcvr.char, exprtoken = exprlex1(exprlex, &exprrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if exprExca[xi+0] == -1 && int(exprExca[xi+1]) == exprstate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			exprn = int(exprExca[xi+0])
			if exprn < 0 || exprn == exprtoken {
				break
			}
		}
		exprn = int(exprExca[xi+1])
		if exprn < 0 {
			goto ret0
		}
	}
	if exprn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			exprlex.Error(exprErrorMessage(exprstate, exprtoken))
			Nerrs++
			if exprDebug >= 1 {
				__yyfmt__.Printf("%s", exprStatname(exprstate))
				__yyfmt__.Printf(" saw %s\n", exprTokname(exprtoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for exprp >= 0 {
				exprn = int(exprPact[exprS[exprp].yys]) + exprErrCode
				if exprn >= 0 && exprn < exprLast {
					exprstate = int(exprAct[exprn]) /* simulate a shift of "error" */
					if int(exprChk[exprstate]) == exprErrCode {
						goto exprstack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if exprDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", exprS[exprp].yys)
				}
				exprp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if exprDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", exprTokname(exprtoken))
			}
			if exprtoken == exprEofCode {
				goto ret1
			}
			exprrcvr.char = -1
			exprtoken = -1
			goto exprnewstate /* try again in the same state */
		}
	}

	/* reduction by production exprn */
	if exprDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", exprn, exprStatname(exprstate))
	}

	exprnt := exprn
	exprpt := exprp
	_ = exprpt // guard against "declared and not used"

	exprp -= int(exprR2[exprn])
	// exprp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if exprp+1 >= len(exprS) {
		nyys := make([]exprSymType, len(exprS)*2)
		copy(nyys, exprS)
		exprS = nyys
	}
	exprVAL = exprS[exprp+1]

	/* consult goto table to find next state */
	exprn = int(exprR1[exprn])
	exprg := int(exprPgo[exprn])
	exprj := exprg + exprS[exprp].yys + 1

	if exprj >= exprLast {
		exprstate = int(exprAct[exprg])
	} else {
		exprstate = int(exprAct[exprj])
		if int(exprChk[exprstate]) != -exprn {
			exprstate = int(exprAct[exprg])
		}
	}
	// dummy call; replaced with literal code
	switch exprnt {

	case 1:
		exprDollar = exprS[exprpt-1 : exprpt+1]
//line xpath.y:63
		{
			getProgBldr(exprlex).CodeFn(
				getProgBldr(exprlex).Store, "store")
		}
	case 2:
		exprDollar = exprS[exprpt-1 : exprpt+1]
//line xpath.y:70
		{
			getProgBldr(exprlex).CurrentFix()
		}
	case 4:
		exprDollar = exprS[exprpt-3 : exprpt+1]
//line xpath.y:77
		{
			getProgBldr(exprlex).CodeFn(
				getProgBldr(exprlex).Or, "or")
		}
	case 6:
		exprDollar = exprS[exprpt-3 : exprpt+1]
//line xpath.y:85
		{
			getProgBldr(exprlex).CodeFn(
				getProgBldr(exprlex).And, "and")
		}
	case 8:
		exprDollar = exprS[exprpt-3 : exprpt+1]
//line xpath.y:93
		{
			getProgBldr(exprlex).CodeFn(
				getProgBldr(exprlex).Eq, "eq")
		}
	case 9:
		exprDollar = exprS[exprpt-3 : exprpt+1]
//line xpath.y:98
		{
			getProgBldr(exprlex).CodeFn(
				getProgBldr(exprlex).Ne, "ne")
		}
	case 11:
		exprDollar = exprS[exprpt-3 : exprpt+1]
//line xpath.y:106
		{
			getProgBldr(exprlex).CodeFn(
				getProgBldr(exprlex).Lt, "lt")
		}
	case 12:
		exprDollar = exprS[exprpt-3 : exprpt+1]
//line xpath.y:111
		{
			getProgBldr(exprlex).CodeFn(
				getProgBldr(exprlex).Gt, "gt")
		}
	case 13:
		exprDollar = exprS[exprpt-3 : exprpt+1]
//line xpath.y:116
		{
			getProgBldr(exprlex).CodeFn(
				getProgBldr(exprlex).Le, "le")
		}
	case 14:
		exprDollar = exprS[exprpt-3 : exprpt+1]
//line xpath.y:121
		{
			getProgBldr(exprlex).CodeFn(
				getProgBldr(exprlex).Ge, "ge")
		}
	case 16:
		exprDollar = exprS[exprpt-3 : exprpt+1]
//line xpath.y:129
		{
			getProgBldr(exprlex).CodeFn(
				getProgBldr(exprlex).Add, "add")
		}
	case 17:
		exprDollar = exprS[exprpt-3 : exprpt+1]
//line xpath.y:134
		{
			getProgBldr(exprlex).CodeFn(
				getProgBldr(exprlex).Sub, "sub")
		}
	case 19:
		exprDollar = exprS[exprpt-3 : exprpt+1]
//line xpath.y:142
		{
			getProgBldr(exprlex).CodeFn(
				getProgBldr(exprlex).Mul, "mul")
		}
	case 20:
		exprDollar = exprS[exprpt-3 : exprpt+1]
//line xpath.y:147
		{
			getProgBldr(exprlex).CodeFn(
				getProgBldr(exprlex).Div, "div")
		}
	case 21:
		exprDollar = exprS[exprpt-3 : exprpt+1]
//line xpath.y:152
		{
			getProgBldr(exprlex).CodeFn(
				getProgBldr(exprlex).Mod, "mod")
		}
	case 23:
		exprDollar = exprS[exprpt-2 : exprpt+1]
//line xpath.y:160
		{
			getProgBldr(exprlex).CodeFn(
				getProgBldr(exprlex).Negate, "negate")
		}
	case 25:
		exprDollar = exprS[exprpt-3 : exprpt+1]
//line xpath.y:168
		{
			getProgBldr(exprlex).CodeFn(
				getProgBldr(exprlex).Union, "union")
		}
	case 26:
		exprDollar = exprS[exprpt-1 : exprpt+1]
//line xpath.y:175
		{
			getProgBldr(exprlex).CodeFn(
				getProgBldr(exprlex).EvalLocPath, "evalLocPath")
		}
	case 28:
		exprDollar = exprS[exprpt-3 : exprpt+1]
//line xpath.y:181
		{
			getProgBldr(exprlex).CodeFn(
				getProgBldr(exprlex).EvalLocPath, "evalLocPath")
		}
	case 29:
		exprDollar = exprS[exprpt-3 : exprpt+1]
//line xpath.y:186
		{
			getProgBldr(exprlex).CodeFn(
				getProgBldr(exprlex).EvalLocPath, "evalLocPath")
		}
	case 30:
		exprDollar = exprS[exprpt-1 : exprpt+1]
//line xpath.y:196
		{
			getProgBldr(exprlex).CodeFn(
				getProgBldr(exprlex).FilterExprEnd, "filterExprEnd")
		}
	case 34:
		exprDollar = exprS[exprpt-1 : exprpt+1]
//line xpath.y:208
		{
			getProgBldr(exprlex).CodeLiteral(exprDollar[1].name)
		}
	case 35:
		exprDollar = exprS[exprpt-1 : exprpt+1]
//line xpath.y:212
		{
			getProgBldr(exprlex).CodeNum(exprDollar[1].val)
		}
	case 36:
		exprDollar = exprS[exprpt-3 : exprpt+1]
//line xpath.y:216
		{
			getProgBldr(exprlex).CodeFn(
				getProgBldr(exprlex).EvalLocPath, "evalLocPath")
			getProgBldr(exprlex).CodeBltin(exprDollar[1].sym, 0)
		}
	case 37:
		exprDollar = exprS[exprpt-3 : exprpt+1]
//line xpath.y:222
		{
			getProgBldr(exprlex).CodeBltin(exprDollar[1].sym, 0)
		}
	case 38:
		exprDollar = exprS[exprpt-4 : exprpt+1]
//line xpath.y:226
		{
			getProgBldr(exprlex).CodeBltin(exprDollar[1].sym, 1)
		}
	case 39:
		exprDollar = exprS[exprpt-6 : exprpt+1]
//line xpath.y:230
		{
			getProgBldr(exprlex).CodeBltin(exprDollar[1].sym, 2)
		}
	case 40:
		exprDollar = exprS[exprpt-8 : exprpt+1]
//line xpath.y:234
		{
			getProgBldr(exprlex).CodeBltin(exprDollar[1].sym, 3)
		}
	case 41:
		exprDollar = exprS[exprpt-1 : exprpt+1]
//line xpath.y:238
		{
			getProgBldr(exprlex).UnsupportedName(xutils.NODETYPE, exprDollar[1].name)
		}
	case 49:
		exprDollar = exprS[exprpt-3 : exprpt+1]
//line xpath.y:257
		{
			getProgBldr(exprlex).CodePathSetCurrent()
		}
	case 50:
		exprDollar = exprS[exprpt-1 : exprpt+1]
//line xpath.y:269
		{
			getProgBldr(exprlex).CodePathOper('/')
		}
	case 59:
		exprDollar = exprS[exprpt-2 : exprpt+1]
//line xpath.y:294
		{
			getProgBldr(exprlex).UnsupportedName(xutils.AXISNAME, exprDollar[1].name)
		}
	case 61:
		exprDollar = exprS[exprpt-1 : exprpt+1]
//line xpath.y:300
		{
			getProgBldr(exprlex).CodeNameTest(exprDollar[1].xmlname)
		}
	case 65:
		exprDollar = exprS[exprpt-1 : exprpt+1]
//line xpath.y:312
		{
			getProgBldr(exprlex).CodePredStart()
		}
	case 67:
		exprDollar = exprS[exprpt-1 : exprpt+1]
//line xpath.y:320
		{
			getProgBldr(exprlex).CodePredEnd()
		}
	case 70:
		exprDollar = exprS[exprpt-1 : exprpt+1]
//line xpath.y:332
		{
			getProgBldr(exprlex).CodePathOper('.')
		}
	case 71:
		exprDollar = exprS[exprpt-1 : exprpt+1]
//line xpath.y:336
		{
			getProgBldr(exprlex).CodePathOper(xutils.DOTDOT)
		}
	case 72:
		exprDollar = exprS[exprpt-1 : exprpt+1]
//line xpath.y:342
		{
			getProgBldr(exprlex).UnsupportedName(
				'@', "not yet implemented")
		}
	case 73:
		exprDollar = exprS[exprpt-1 : exprpt+1]
//line xpath.y:349
		{
			getProgBldr(exprlex).UnsupportedName(
				xutils.DBLSLASH, "not yet implemented")
		}
	}
	goto exprstack /* stack new state and value */
}
