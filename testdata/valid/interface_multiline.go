package valid

// Archive stores component state.
type Archive struct{}

// Device transfers the mutable fields that affect emulation.
type Device interface {
	TransferState(*Archive) error
}

// Empty marks a value that carries no method set.
type Empty interface{}
