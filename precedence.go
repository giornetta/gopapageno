package gopapageno

type Precedence int

const (
	PrecYields = iota
	PrecEquals
	PrecTakes
	PrecEmpty
	PrecUnknown
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
	default:
		return "Unknown"
	}
}
