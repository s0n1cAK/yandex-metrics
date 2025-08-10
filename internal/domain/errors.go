package domain

import "errors"

var (
	ErrNotFound       = errors.New("metric not found")
	ErrInvalidType    = errors.New("invalid metric type")
	ErrInvalidPayload = errors.New("invalid payload")
	ErrZeroCounter    = errors.New("counter cannot be zero")
)
