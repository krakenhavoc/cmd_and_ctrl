package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Biomass Mutation — Instant {X}{G/U}{G/U}:
//
//	"Creatures you control have base power and toughness X/X until end of
//	 turn."
//
// Mirror Entity's base power and toughness SET (layer 7b) without the
// creature types, on a spell: the creatures it affects are fixed as it
// resolves (CR 611.2c), so a creature that enters later this turn keeps
// its own base. A pump or +1/+1 counter still applies on top (layer 7c).
// X = 0 makes every creature you control a 0/0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2dbe8aaa-2e34-4552-8da7-cf8ecbbaa7b6",
		Name:         "Biomass Mutation",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			return effects.UntilEndOfTurn(ctx, uuid.Nil, effects.And(effects.Creature(), effects.YouControl()),
				"Biomass Mutation — base power and toughness X/X",
				game.SetBasePowerMod(x), game.SetBaseToughnessMod(x))
		},
	})
}
