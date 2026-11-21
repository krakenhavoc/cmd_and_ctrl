package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Basri's Solidarity — Sorcery {1}{W} (Jumpstart Selesnya deck request,
// #4893):
//
//	"Put a +1/+1 counter on each creature you control."
//
// The set is the creatures the caster controls as the spell resolves, read
// before the first counter goes on (Abandoned Air Temple's body).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68281e06-64e3-490b-8691-8f4169614ec0",
		Name:         "Basri's Solidarity",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B11PutCounterOnEachCreatureYouControl(ctx.Game, item)
		},
	})
}
