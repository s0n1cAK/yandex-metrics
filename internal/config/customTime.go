package config

import (
	"strconv"
	"time"

	"github.com/s0n1cAK/yandex-metrics/internal/lib"
)

type customTime time.Duration

func formatCustomTime(value string) (customTime, error) {
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

	return customTime(duration), nil
}

func (ct *customTime) String() string {
	return time.Duration(*ct).String()
}

func (ct customTime) Duration() time.Duration {
	return time.Duration(ct)
}

func (ct *customTime) Set(value string) error {
	gValue, err := formatCustomTime(value)
	if err != nil {
		return err
	}
	*ct = gValue
	return nil
}

func (ct *customTime) UnmarshalText(text []byte) error {
	gValue, err := formatCustomTime(string(text[:]))
	if err != nil {
		return err
	}
	*ct = gValue
	return nil
}
