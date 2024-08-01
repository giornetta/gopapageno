package generator

import (
	"fmt"
)

func skipSpaces(input string, index *int) {
	for *index < len(input) &&
		(input[*index] == ' ' ||
			input[*index] == '\t' ||
			input[*index] == '\r' ||
			input[*index] == '\n') {
		*index++
	}
}

func getIdentifier(input string, index *int) string {
	startIndex := *index

	if *index < len(input) &&
		((input[*index] >= 'a' && input[*index] <= 'z') ||
			(input[*index] >= 'A' && input[*index] <= 'Z') ||
			(input[*index] == '_')) {
		*index++

		for *index < len(input) &&
			((input[*index] >= 'a' && input[*index] <= 'z') ||
				(input[*index] >= 'A' && input[*index] <= 'Z') ||
				(input[*index] == '_') ||
				(input[*index] >= '9' && input[*index] <= '0')) {
			*index++
		}
	}
	return input[startIndex:*index]
}

func getAlternatives(input string, index *int) ([]string, [][]string, error) {
	startIndex := *index

	// Check if the current substring stars with an open parenthesis.
	if *index < len(input) && input[*index] == '(' {
		*index++
	} else {
		return nil, nil, fmt.Errorf("input string %s does not start with '('", input[*index:])
	}
	skipSpaces(input, index)

	rhsTokens := make([]string, 0)
	alternatives := make([][]string, 1)

	// Loop until we found a closed parenthesis
	for input[*index] != ')' {
		// If the next character is another open bracket, look for nested alternatives.
		if *index < len(input) && input[*index] == '(' {
			flattened, nestedAlternatives, err := getAlternatives(input, index)
			if err != nil {
				return nil, nil, fmt.Errorf("could not parse nested alternatives: %w", err)
			}

			rhsTokens = append(rhsTokens, flattened...)

			// Add each nested alternative to every alternative found so far.
			newAlternatives := make([][]string, len(alternatives)*len(nestedAlternatives), len(alternatives)*len(nestedAlternatives))
			for i := 0; i < len(alternatives); i++ {
				for j := 0; j < len(nestedAlternatives); j++ {
					newAlternatives[i*len(alternatives)+j] = append(alternatives[i], nestedAlternatives[j]...)
				}
			}
			alternatives = newAlternatives
		} else {
			// Instead it should be just a normal identifier
			identifier := getIdentifier(input, index)

			rhsTokens = append(rhsTokens, identifier)
			for i := 0; i < len(alternatives); i++ {
				alternatives[i] = append(alternatives[i], identifier)
			}
		}

		skipSpaces(input, index)
	}

	*index++

	if input[*index] != '+' {
		return nil, nil, fmt.Errorf("alternative string %s does not end with '+'", input[startIndex:*index+1])
	}
	*index++

	// Duplicate alternatives
	newAlternatives := make([][]string, len(alternatives)*2)
	for i, alt := range alternatives {
		newAlternatives[i] = alt
		newAlternatives[i+len(alternatives)] = append(alt, alt...)
	}

	return rhsTokens, newAlternatives, nil
}

func getSemanticFunction(input string, index *int) string {
	startIndex := *index

	numBraces := 0
	for *index < len(input) {
		if input[*index] == '\'' {
			*index++
			escape := false
			for *index < len(input) {
				if input[*index] == '\\' {
					escape = !escape
				} else if input[*index] == '\'' {
					if !escape {
						break
					}
					escape = false
				} else {
					escape = false
				}
				*index++
			}
		} else if input[*index] == '"' {
			*index++
			escape := false
			for *index < len(input) {
				if input[*index] == '\\' {
					escape = !escape
				} else if input[*index] == '"' {
					if !escape {
						break
					}
					escape = false
				} else {
					escape = false
				}
				*index++
			}
		} else if input[*index] == '`' {
			*index++
			for *index < len(input) {
				if input[*index] == '`' {
					break
				}
				*index++
			}
		} else if input[*index] == '/' && *index < len(input)-1 {
			*index++
			if input[*index] == '*' {
				*index++
				foundStar := false
				for *index < len(input) {
					if input[*index] == '*' {
						foundStar = true
					} else if input[*index] == '/' {
						if foundStar {
							break
						}
						foundStar = false
					} else {
						foundStar = false
					}
					*index++
				}
			} else if input[*index] == '/' {
				*index++
				for *index < len(input) {
					if input[*index] == '\n' {
						break
					}
					*index++
				}
			}
		} else if input[*index] == '{' {
			numBraces++
		} else if input[*index] == '}' {
			numBraces--
			if numBraces == 0 {
				*index++
				break
			}
		}
		*index++
	}

	return input[startIndex:*index]
}
