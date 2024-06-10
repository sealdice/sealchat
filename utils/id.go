package utils

import (
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/samber/lo"
)

var defaultAlphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func NewID() string {
	// 62 ** 22 > 64 ** 21
	return lo.Must1(gonanoid.Generate(defaultAlphabet, 22))
}

func NewIDWithLength(length int) string {
	return lo.Must1(gonanoid.Generate(defaultAlphabet, length))
}
