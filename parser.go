package log_parser

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"bytes"
)

type ErrorDescr struct {
	Name          string
	Number        int
	FullErr       string
	completeValue bool
}

type Parser struct {
	debug  bool
	Errors map[string]*ErrorDescr
}

var (
	sensitivity = 50
	errTextRE   = regexp.MustCompile(fmt.Sprintf(`(\[ERROR\]|\[FATAL\]).{10,%v}`, sensitivity))
	ErrorsRE    = regexp.MustCompile(`.*(\[ERROR\]|\[FATAL\]).+((\n.+){0,2})`)
	Date_regExp = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2}`)
)

func NewParser(mode bool) *Parser {
	return &Parser{mode, make(map[string]*ErrorDescr)}
}

func (e *Parser) include(name string) bool {
	if _, ok := e.Errors[name]; ok {
		return true
	}
	return false
}

func (p *Parser) ParseLog(logFile io.Reader) {
	var lastErr bytes.Buffer
	var lastErrText bytes.Buffer

	scanner := bufio.NewScanner(logFile)
	for scanner.Scan() {
		ln := scanner.Text()
		errName := errTextRE.FindString(ln)
		lastErrName := lastErr.String()
		if ErrorsRE.MatchString(ln) {
			if p.include(lastErrName) && !p.Errors[lastErrName].completeValue {
				p.Errors[lastErrName].FullErr = lastErrText.String()
				p.Errors[lastErrName].completeValue = true
			}
			if p.include(errName) {
				p.Errors[errName].Number++
			} else {
				p.Errors[errName] = &ErrorDescr{errName, 1, "", false}
				lastErrText.Reset()
				lastErrText.WriteString(ln)
				lastErrText.WriteString("\n")
			}
			lastErr.Reset()
			lastErr.WriteString(errName)
			// fmt.Println("=========================", lastErr.String())
		} else {
			if p.include(lastErrName) {
				if !p.Errors[lastErrName].completeValue && !Date_regExp.MatchString(ln) {
					lastErrText.WriteString(ln)
					lastErrText.WriteString("\n")
				} else if Date_regExp.MatchString(ln) {
					p.Errors[lastErrName].FullErr = lastErrText.String()
					p.Errors[lastErrName].completeValue = true
				}
			}
		}
	}
	return
}