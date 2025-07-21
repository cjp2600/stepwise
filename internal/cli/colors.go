package cli

import (
	"github.com/cjp2600/stepwise/internal/colors"
)

type Colors struct {
	*colors.Colors
}

func NewColors() *Colors {
	return &Colors{colors.NewColors()}
}
