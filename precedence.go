package gopapageno

type Precedence uint8

const (
	PrecEquals Precedence = iota
	PrecYields
	PrecTakes
	PrecAssociative
	PrecEmpty
)

func (p Precedence) String() string {
	switch p {
	case PrecYields:
		return "Yields"
	case PrecEquals:
		return "Equals"
	case PrecTakes:
		return "Takes"
	case PrecEmpty:
		return "Empty"
	case PrecAssociative:
		return "Associative"
	default:
		return "Unknown"
	}
}
