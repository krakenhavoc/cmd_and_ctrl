package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Breath of Darigaaz — Sorcery {1}{R} (Commander 2015, S82 OLD #5546):
//
//	"Kicker {2} (You may pay an additional {2} as you cast this spell.)
//	 Breath of Darigaaz deals 1 damage to each creature without flying and
//	 each player. If this spell was kicked, it deals 4 damage to each
//	 creature without flying and each player instead."
//
// One damage instance over every creature without flying (through the
// layers, as it resolves) and every player (Squallmonger's sweep), for 1, or
// for 4 instead when kicked, never both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "1954994f-17bf-4ea5-af72-60f9bfcb6569",
		Name:          "Breath of Darigaaz",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{2}")},
		Purpose:       game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := 1
			if ctx.WasKicked() {
				n = 4
			}
			return effects.ThisDealsDamageToEachCreatureMatchingAndEachPlayer(effects.And(effects.Creature(), effects.Not(effects.HasKeyword("flying"))), n)(ctx.Game, item)
		},
	})
}
