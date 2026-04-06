package leakage

import "errors"

// ErrInvalidDays is returned when the requested period is outside the supported range.
var ErrInvalidDays = errors.New("days must be between 7 and 90")
