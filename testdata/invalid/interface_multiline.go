package invalid

// Archive stores component state.
type Archive struct{}

// Device transfers the mutable fields that affect emulation.
type Device interface{ TransferState(*Archive) error }
