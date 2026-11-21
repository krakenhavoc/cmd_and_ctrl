package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const FraDrawOracle = "test-fra-empower-a-draw"

func init() {
	effects.Register(effects.Spec{
		OracleID: FraDrawOracle,
		Name:     "Fixture Draw",
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
