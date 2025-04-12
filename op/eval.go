package op

import (
	"github.com/sekiguchi-nagisa/guniset/set"
	"io"
)

type EvalContext struct {
	catSet map[GeneralCategory]set.UniSet
	eawSet map[EastAsianWidth]set.UniSet
}

func NewEvalContext(unicodeData io.Reader, eastAsianWidth io.Reader) (*EvalContext, error) {
	_ = unicodeData
	_ = eastAsianWidth
	return &EvalContext{}, nil
}
