%%

%axiom RE

%%

RE : UNION
{
	$$.Value = $1.Value
} | SIMPLE_RE
{
	$$.Value = $1.Value
};

UNION : RE Pipe SIMPLE_RE
{
	leftNfa := $1.Value.(*Nfa)
	rightNfa := $3.Value.(*Nfa)
	
	leftNfa.Unite(rightNfa)
	
	$$.Value = leftNfa
};

SIMPLE_RE : CONCATENATION
{
	$$.Value = $1.Value
} | Any
{
	$$.Value = $1.Value
} | Char
{
	newNfa := newNfaFromChar($1.Value.(byte))
	
	$$.Value = &newNfa
} | SquareLPar SET_ITEMS SquareRPar
{
	newNfa := newNfaFromCharClass($2.Value.([256]bool))
	
	$$.Value = &newNfa
} | SquareLPar Caret SET_ITEMS SquareRPar
{
	chars := $3.Value.([256]bool)
	
	//Skip the first Char (empty transition)
	for i := 1; i < len(chars); i++ {
		chars[i] = !chars[i]
	}
	
	newNfa := newNfaFromCharClass(chars)
	
	$$.Value = &newNfa
} | LPar RE RPar
{
	$$.Value = $2.Value
} | Any Star
{
	nfaAny := $1.Value.(*Nfa)
	nfaAny.KleeneStar()
	
	$$.Value = nfaAny
} | Char Star
{
	newNfa := newNfaFromChar($1.Value.(byte))
	newNfa.KleeneStar()
	
	$$.Value = &newNfa
} | SquareLPar SET_ITEMS SquareRPar Star
{
	newNfa := newNfaFromCharClass($2.Value.([256]bool))
	newNfa.KleeneStar()
	
	$$.Value = &newNfa
} | SquareLPar Caret SET_ITEMS SquareRPar Star
{
	chars := $3.Value.([256]bool)
	
	//Skip the first Char (empty transition)
	for i := 1; i < len(chars); i++ {
		chars[i] = !chars[i]
	}
	
	newNfa := newNfaFromCharClass(chars)
	
	newNfa.KleeneStar()
	
	$$.Value = &newNfa
} | LPar RE RPar Star
{
	nfaEnclosed := $2.Value.(*Nfa)
	
	nfaEnclosed.KleeneStar()
	
	$$.Value = nfaEnclosed
} | Any Plus
{
	nfaAny := $1.Value.(*Nfa)
	nfaAny.KleenePlus()
	
	$$.Value = nfaAny
} | Char Plus
{
	newNfa := newNfaFromChar($1.Value.(byte))
	newNfa.KleenePlus()
	
	$$.Value = &newNfa
} | SquareLPar SET_ITEMS SquareRPar Plus
{
	newNfa := newNfaFromCharClass($2.Value.([256]bool))
	newNfa.KleenePlus()
	
	$$.Value = &newNfa
} | SquareLPar Caret SET_ITEMS SquareRPar Plus
{
	chars := $3.Value.([256]bool)
	
	//Skip the first Char (empty transition)
	for i := 1; i < len(chars); i++ {
		chars[i] = !chars[i]
	}
	
	newNfa := newNfaFromCharClass(chars)
	
	newNfa.KleenePlus()
	
	$$.Value = &newNfa
} | LPar RE RPar Plus
{
	nfaEnclosed := $2.Value.(*Nfa)
	
	nfaEnclosed.KleenePlus()
	
	$$.Value = nfaEnclosed
};


CONCATENATION : Any SIMPLE_RE
{
	leftNfa := $1.Value.(*Nfa)
	rightNfa := $2.Value.(*Nfa)
	
	leftNfa.Concatenate(*rightNfa)
	
	$$.Value = leftNfa
} | Char SIMPLE_RE
{
	newNfa := newNfaFromChar($1.Value.(byte))
	rightNfa := $2.Value.(*Nfa)

	newNfa.Concatenate(*rightNfa)
	
	$$.Value = &newNfa
} | SquareLPar SET_ITEMS SquareRPar SIMPLE_RE
{
	newNfa := newNfaFromCharClass($2.Value.([256]bool))
	rightNfa := $4.Value.(*Nfa)
	
	newNfa.Concatenate(*rightNfa)
	
	$$.Value = &newNfa
} | SquareLPar Caret SET_ITEMS SquareRPar SIMPLE_RE
{
	chars := $3.Value.([256]bool)
	
	//Skip the first Char (empty transition)
	for i := 1; i < len(chars); i++ {
		chars[i] = !chars[i]
	}
	
	newNfa := newNfaFromCharClass(chars)
	rightNfa := $5.Value.(*Nfa)
	
	newNfa.Concatenate(*rightNfa)
	
	$$.Value = &newNfa
} | LPar RE RPar SIMPLE_RE
{
	nfaEnclosed := $2.Value.(*Nfa)
	rightNfa := $4.Value.(*Nfa)
	
	nfaEnclosed.Concatenate(*rightNfa)
	
	$$.Value = nfaEnclosed
} | Any Star SIMPLE_RE
{
	nfaAny := $1.Value.(*Nfa)
	nfaAny.KleeneStar()
	rightNfa := $3.Value.(*Nfa)
	
	nfaAny.Concatenate(*rightNfa)
	
	$$.Value = nfaAny
} | Char Star SIMPLE_RE
{
	newNfa := newNfaFromChar($1.Value.(byte))
	newNfa.KleeneStar()
	rightNfa := $2.Value.(*Nfa)

	newNfa.Concatenate(*rightNfa)
	
	$$.Value = &newNfa
} | SquareLPar SET_ITEMS SquareRPar Star SIMPLE_RE
{
	newNfa := newNfaFromCharClass($2.Value.([256]bool))
	newNfa.KleeneStar()
	rightNfa := $5.Value.(*Nfa)
	
	newNfa.Concatenate(*rightNfa)
	
	$$.Value = &newNfa
} | SquareLPar Caret SET_ITEMS SquareRPar Star SIMPLE_RE
{
	chars := $3.Value.([256]bool)
	
	//Skip the first Char (empty transition)
	for i := 1; i < len(chars); i++ {
		chars[i] = !chars[i]
	}
	
	newNfa := newNfaFromCharClass(chars)
	newNfa.KleeneStar()
	rightNfa := $6.Value.(*Nfa)
	
	newNfa.Concatenate(*rightNfa)
	
	$$.Value = &newNfa
} | LPar RE RPar Star SIMPLE_RE
{
	nfaEnclosed := $2.Value.(*Nfa)
	nfaEnclosed.KleeneStar()
	rightNfa := $5.Value.(*Nfa)
	
	nfaEnclosed.Concatenate(*rightNfa)
	
	$$.Value = nfaEnclosed
} | Any Plus SIMPLE_RE
{
	nfaAny := $1.Value.(*Nfa)
	nfaAny.KleenePlus()
	rightNfa := $3.Value.(*Nfa)
	
	nfaAny.Concatenate(*rightNfa)
	
	$$.Value = nfaAny
} | Char Plus SIMPLE_RE
{
	newNfa := newNfaFromChar($1.Value.(byte))
	newNfa.KleenePlus()
	rightNfa := $2.Value.(*Nfa)

	newNfa.Concatenate(*rightNfa)
	
	$$.Value = &newNfa
} | SquareLPar SET_ITEMS SquareRPar Plus SIMPLE_RE
{
	newNfa := newNfaFromCharClass($2.Value.([256]bool))
	newNfa.KleenePlus()
	rightNfa := $5.Value.(*Nfa)
	
	newNfa.Concatenate(*rightNfa)
	
	$$.Value = &newNfa
} | SquareLPar Caret SET_ITEMS SquareRPar Plus SIMPLE_RE
{
	chars := $3.Value.([256]bool)
	
	//Skip the first Char (empty transition)
	for i := 1; i < len(chars); i++ {
		chars[i] = !chars[i]
	}
	
	newNfa := newNfaFromCharClass(chars)
	newNfa.KleenePlus()
	rightNfa := $6.Value.(*Nfa)
	
	newNfa.Concatenate(*rightNfa)
	
	$$.Value = &newNfa
} | LPar RE RPar Plus SIMPLE_RE
{
	nfaEnclosed := $2.Value.(*Nfa)
	nfaEnclosed.KleenePlus()
	rightNfa := $5.Value.(*Nfa)
	
	nfaEnclosed.Concatenate(*rightNfa)
		
	$$.Value = nfaEnclosed
};
SET_ITEMS : CharInSet
{
	var charSet [256]bool
	charSet[$1.Value.(byte)] = true
	
	$$.Value = charSet
} | CharInSet Dash CharInSet
{
	charStart := $1.Value.(byte)
	charEnd := $3.Value.(byte)
	
	if charStart > charEnd {
		temp := charStart
		charStart = charEnd
		charEnd = temp
	}
	
	var charSet [256]bool
	for i := charStart; i <= charEnd; i++ {
		charSet[i] = true
	}
	
	$$.Value = charSet
} | CharInSet SET_ITEMS
{
	charSet := $2.Value.([256]bool)
	charSet[$1.Value.(byte)] = true
	
	$$.Value = charSet
} | CharInSet Dash CharInSet SET_ITEMS
{
	charStart := $1.Value.(byte)
	charEnd := $3.Value.(byte)
	charSet := $4.Value.([256]bool)
	
	if charStart > charEnd {
		temp := charStart
		charStart = charEnd
		charEnd = temp
	}
	
	for i := charStart; i <= charEnd; i++ {
		charSet[i] = true
	}
	
	$$.Value = charSet
};