package generator

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
		// TODO: Consider regex?
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
