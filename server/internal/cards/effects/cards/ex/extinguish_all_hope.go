package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Extinguish All Hope — Sorcery {4}{B}{B} (EDHREC rank 3833):
//
//	"Destroy all nonenchantment creatures."
//
// The enchantress deck's one-sided Wrath: every creature that is not
// also an enchantment, read post-layer, so a Theros god that is a
// creature survives and an animated Opalescence-style enchantment
// does too. One simultaneous destruction event, so every dies
// trigger sees the whole batch; indestructible is honoured on the
// mass path since #470.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID: "c7ca25f3-7a22-477b-8546-4c2597d5d1ff",
		Name:     "Extinguish All Hope",
		// ADR 0126 §6: an enchantment creature stays.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.And(effects.Creature(), effects.Not(effects.Enchantment()))}.Apply(ctx)
		},
	})
}
