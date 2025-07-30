package lib

import (
	"flag"
	"unicode"
)

func HasLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func FloatPtr(f float64) *float64 {
	return &f
}

func IntPtr(i int64) *int64 {
	return &i
}

func IsFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
