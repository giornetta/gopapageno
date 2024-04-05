package gopapageno

type Token struct {
	Type       TokenType
	Precedence Precedence

	Value any
	// Lexeme string

	Next  *Token
	Child *Token
}

type TokenType uint16

const (
	TokenEmpty TokenType = 0
	TokenTerm  TokenType = 0x8000
)

func (t TokenType) IsTerminal() bool {
	return t >= 0x8000
}

func (t TokenType) Value() uint16 {
	return uint16(0x7FFF & t)
}
