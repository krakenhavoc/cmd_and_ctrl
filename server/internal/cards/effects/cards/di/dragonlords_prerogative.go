package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragonlord's Prerogative — Instant {4}{U}{U} (Dragons of Tarkir, S82 old
// batch 329):
//
//	"As an additional cost to cast this spell, you may reveal a Dragon card
//	 from your hand.
//	 If you revealed a Dragon card or controlled a Dragon as you cast this
//	 spell, this spell can't be countered.
//	 Draw four cards."
//
// Orator of Ojutai's optional Dragon reveal and its "controlled a Dragon as
// you cast" number fixed at announcement (CR 601.2i); the conditional
// uncounterable is Banefire's CantBeCounteredIf, read off the stack item.
//
// No simplification.
func init() {
	const dragonKey = "controlled_a_dragon"
	effects.Register(effects.Spec{
		OracleID:      "46bc0aaf-987f-4c8d-b98b-50b90f97c9a3",
		Name:          "Dragonlord's Prerogative",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Draws: 4},
		OptionalCosts: []game.AdditionalCost{effects.OptionalBehold("a", "Dragon")},
		CastTimeValues: []game.CastTimeValue{{
			Key:   dragonKey,
			Value: effects.ControlsADragonAsCast,
		}},
		CantBeCounteredIf: func(g *game.Game, item *game.StackItem) bool {
			if effects.NewContext(g, item).WasBeheld() {
				return true
			}
			n, _ := item.CastValue(dragonKey)
			return n > 0
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{N: 4}.Apply(ctx)
		},
	})
}
