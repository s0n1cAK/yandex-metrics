package customtype

import (
	"errors"
	"strconv"
	"time"

	"github.com/s0n1cAK/yandex-metrics/internal/lib"
)

var (
	ErrInvalidDurationFormat = errors.New("invalid duration format")
	ErrInvalidNumericFormat  = errors.New("invalid numeric duration format")
)

type Time time.Duration

func formatTime(value string) (Time, error) {
	var duration time.Duration
	var err error

	if lib.HasLetter(value) {
		duration, err = time.ParseDuration(value)
		if err != nil {
			return 0, ErrInvalidDurationFormat
		}
	} else {
		seconds, err := strconv.Atoi(value)
		if err != nil {
			return 0, ErrInvalidNumericFormat
		}
		duration = time.Duration(seconds) * time.Second
	}

	return Time(duration), nil
}

func (ct *Time) String() string {
	return time.Duration(*ct).String()
}

func (ct *Time) Type() string {
	return "time"
}

func (ct Time) Duration() time.Duration {
	return time.Duration(ct)
}

func (ct *Time) Set(value string) error {
	gValue, err := formatTime(value)
	if err != nil {
		return err
	}
	*ct = gValue
	return nil
}

func (ct *Time) UnmarshalText(text []byte) error {
	gValue, err := formatTime(string(text[:]))
	if err != nil {
		return err
	}
	*ct = gValue
	return nil
}
