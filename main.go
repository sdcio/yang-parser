package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/sdcio/yang-parser/xpath"
	"github.com/sdcio/yang-parser/xpath/grammars/expr"
	"github.com/sdcio/yang-parser/xpath/xutils"
	log "github.com/sirupsen/logrus"
)

// call this
// find /home/mava/projects/schema-server/lab/common/yang/ -type f -name "*.yang" | xargs perl -lne '/must \"(.*)\"/ && print $1'

type stats struct {
	total  int
	failed int
}

func (s *stats) String() string {
	pass_ratio := (float32(s.total) - float32(s.failed)) * 100 / float32(s.total)
	return fmt.Sprintf("Pass-Ratio: %.2f%%, Total: %d, Pass %d, Failed: %d", pass_ratio, s.total, s.total-s.failed, s.failed)
}

func main() {
	Other("/srl_nokia-if:interface[srl_nokia-if:name=substring-before(current(), '.')]/srl_nokia-if:subinterface[srl_nokia-if:index=substring-after(current(), '.')]")
}

func Other(exprStr string) error {
	prgbuilder := xpath.NewProgBuilder(exprStr)
	lexer := expr.NewExprLex(exprStr, prgbuilder, nil)

	lexer.Parse()
	prog, err := lexer.CreateProgram(exprStr)
	if err != nil {
		return err
	}

	xpm := xpath.NewMachine(exprStr, prog, "exprMachine")

	xpm.PrintMachine()

	return nil

}

func ProcessFileMain() {
	s := &stats{}

	processor := func(x string) {
		s.total += 1
		xpm, err := expr.NewExprMachine(x, nil)
		if err != nil {
			log.Error(err)
			s.failed += 1
		}
		_ = xpm

		fmt.Println(xpm.PrintMachine())
	}

	ProcessFile(processor, "/home/mava/projects/yang-parser/INPUTDATA")

	fmt.Println(s)
}

func foo(xpm *xpath.Machine, ctxNode xutils.XpathNode) {
	xpath.NewCtxFromMach(xpm, ctxNode)
}

func ProcessFile(f func(string), file string) {
	readFile, err := os.Open(file)
	if err != nil {
		fmt.Println(err)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		f(fileScanner.Text())
	}

	readFile.Close()
}

func ProcessStdIn(f func(string)) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		f(scanner.Text())
	}
}
