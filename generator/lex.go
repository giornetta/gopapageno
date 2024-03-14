package generator

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var (
	separatorRegex  = regexp.MustCompile("^%%\\s*$")
	cutPointsRegex  = regexp.MustCompile("^%cut\\s*([^\\s].*)$")
	definitionRegex = regexp.MustCompile("^([a-zA-Z][a-zA-Z0-9]*)\\s*(.+)$")
)

type lexRule struct {
	Regex  string
	Action string
}

func (r lexRule) String() string {
	return fmt.Sprintf("%s: %s", r.Regex, r.Action)
}

func parseLexer(r io.Reader) ([]lexRule, string, string) {
	scanner := bufio.NewScanner(r)

	/****************************
	 * DEFINITIONS + CUT POINTS *
	 ****************************/
	cutPoints := ""
	definitions := make(map[string]string)

	for scanner.Scan() {
		l := scanner.Text()

		if separatorRegex.MatchString(l) {
			break
		}
		defMatch := definitionRegex.FindStringSubmatch(l)
		cutPointsMatch := cutPointsRegex.FindStringSubmatch(l)
		if defMatch != nil {
			definitions[defMatch[1]] = strings.TrimSpace(defMatch[2])
		} else if cutPointsMatch != nil {
			cutPoints = cutPointsMatch[1] // TODO: Does this only support a single cutpoints definition?
		}
	}

	// TODO: Log Definitions

	/*********
	 * RULES *
	 *********/
	var ruleBuilder strings.Builder
	for scanner.Scan() {
		l := scanner.Text()
		if separatorRegex.MatchString(l) {
			break
		}
		ruleBuilder.WriteString(l)
		ruleBuilder.WriteString("\n")
	}

	lexRules := parseLexRules(ruleBuilder.String(), definitions)

	/********
	 * CODE *
	 ********/
	var codeBuilder strings.Builder
	for scanner.Scan() {
		l := scanner.Text()
		codeBuilder.WriteString(l)
		codeBuilder.WriteString("\n") // TODO: Consider reading directly from r? Or find a way to read until end of r in 1 instruction
	}

	return lexRules, cutPoints, codeBuilder.String()
}

func parseLexRules(input string, definitions map[string]string) []lexRule {
	bytes := []byte(input)
	lexRules := make([]lexRule, 0)
	pos := 0
	pos = skipSpaces(bytes, pos)
	curRegex := ""

	for pos < len(bytes) {
		startingPos := pos

		//Read anything until a { is reached
		for pos < len(bytes) && bytes[pos] != '{' {
			pos++
		}

		if pos >= len(bytes) {
			break
		}

		//When a { is reached, try to read a definition and then a }
		//If it's not possible then the regex part is over
		curlyLParPos := pos

		pos++

		var identifier string
		identifier, pos = getIdentifier(bytes, pos)
		curlyRParPos := pos
		if bytes[curlyRParPos] == '}' && identifier != "" {
			foundInDefinitions := false
			for key, value := range definitions {
				if key == identifier {
					foundInDefinitions = true
					curRegex += string(bytes[startingPos:curlyLParPos]) + value
					pos++
					break
				}
			}

			if !foundInDefinitions {
				fmt.Println(lexRules)
				panic(fmt.Sprintf("Missing definition \"%s\"", identifier))
			}
		} else {
			semFun, pos := getSemanticFunction(bytes, curlyLParPos)
			curRegex += string(bytes[startingPos:curlyLParPos])
			curRegex = strings.Trim(curRegex, " \t\r\n")
			lexRules = append(lexRules, lexRule{curRegex, semFun})
			curRegex = ""
			pos = skipSpaces(bytes, pos)
		}
	}

	return lexRules
}
