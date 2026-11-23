package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wave Goodbye — Sorcery {2}{U}{U} (EDHREC rank 1515):
//
//	"Return each creature without a +1/+1 counter on it to its
//	 owner's hand."
//
// The counters deck's one-sided Evacuation. One simultaneous bounce
// of every creature with no +1/+1 counter — every player's, the
// caster's own un-countered creatures included, as printed; tokens
// bounced this way cease to exist.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID: "06ef46c6-00ba-40e6-b866-d0095ab83749",
		Name:     "Wave Goodbye",
		// ADR 0126 §6: a creature with a +1/+1 counter stays.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepBounce, Partial: true}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BounceAllMatching{Match: effects.And(effects.Creature(), effects.B13WithoutPlusOneCounter())}.Apply(ctx)
		},
	})
}
