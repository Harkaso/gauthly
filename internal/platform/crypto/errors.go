package crypto

import "errors"

// ErrInvalidParamLength indicates that a fixed-size cryptographic parameter
// does not have the required length.
var ErrInvalidParamLength = errors.New("invalid parameter length")
