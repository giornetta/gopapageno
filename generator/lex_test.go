package generator

import (
	"strings"
	"testing"
)

const definitions = `
%cut \n

LPAR    \(
RPAR    \)
PLUS    \+
TIMES   \*
DIGIT   [0-9]
SPACE   [ \t]
NEWLINE [\r\n]

%%
`

func TestParseLexer(t *testing.T) {
	r := strings.NewReader(definitions)
	rules, cutPoints, lexCode := parseLexer(r)

	t.Logf("Rules: %+v\nCutPoints: %+v\nLexCode: %+v\n", rules, cutPoints, lexCode)
}
