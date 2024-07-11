%cut <

IDENTIFIER [a-zA-Z0-9_\-:]+
INFO [^<]+
SPACE [ \t]
NEWLINE [\r\n]
VALUE ([^"\\]|\\.)*

%%

<{IDENTIFIER}({SPACE}+{IDENTIFIER}="{VALUE}")*>
{
	id, attributesList := getIdAndAttributesListFrom(text)
	semanticValue := xpath.NewOpenTagSemanticValue(id, attributesList, start, end)

	token.Type = OPENTAG
	token.Value = *semanticValue
}
<{IDENTIFIER}({SPACE}+{IDENTIFIER}="{VALUE}")*/>
{
	id, attributesList := getIdAndAttributesListFrom(text)
	semanticValue := xpath.NewOpenCloseTagSemanticValue(id, attributesList, start, end)

    token.Type = OPENCLOSETAG
    token.Value = *semanticValue
}
</{IDENTIFIER}>
{
	id, _ := getIdAndAttributesListFrom(text)
	semanticValue := xpath.NewCloseTagSemanticValue(id, start, end)

    token.Type = CLOSETAG
    token.Value = *semanticValue
}
({SPACE}|{NEWLINE})+
{
	return gopapageno.LexSkip
}
<?[^?]+?>
{
	return gopapageno.LexSkip
}
<![^>]+>
{
	return gopapageno.LexSkip
}
{INFO}
{
	semanticValue := xpath.NewTextSemanticValue(text, start, end)

	token.Type = TEXT
	token.Value = *semanticValue
}

%%

import (
	"regexp"
	"github.com/giornetta/gopapageno/ext/xpath"
)

//A regexp.Regexp is safe for concurrent use by multiple goroutines, except for configuration methods, such as Longest.
var generalTagRegexp *(regexp.Regexp)
var attributesRegexp *(regexp.Regexp)

func getIdAndAttributesListFrom(text string) (id string, attributesList []*xpath.Attribute){
	tagMatch := generalTagRegexp.FindStringSubmatch(text)

	id = tagMatch[1]
	attributesString := tagMatch[2]

	if(attributesString != ""){
		attributesMatches := attributesRegexp.FindAllStringSubmatch(attributesString, -1)
		for _, attributeMatch := range attributesMatches{
			attribute := xpath.NewAttribute(attributeMatch[1], attributeMatch[2])
			attributesList = append(attributesList, attribute)
		}
	}
	return
}

func LexerPreallocMem(inputSize int, numThreads int){
	generalTagRegexp = regexp.MustCompile(`^<\/?([a-zA-Z0-9_\-:]+)(?:\s*)([^\/>]*)\/?>$`)
	attributesRegexp = regexp.MustCompile(`([a-zA-Z0-9_\-:]+)(?:=")([^"]+)(?:")`)
}