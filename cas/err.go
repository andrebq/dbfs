package cas

type (
	// Err represents errors that don't carry any extra information
	Err string
)

const (
	ErrNotFound = Err("cas reference could not be found")
)

func (e Err) Error() string { return string(e) }
