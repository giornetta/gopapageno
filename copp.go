package gopapageno

type CToken struct {
	*Token
	CurrentConstruction  []TokenType
	PreviousConstruction []TokenType
}
