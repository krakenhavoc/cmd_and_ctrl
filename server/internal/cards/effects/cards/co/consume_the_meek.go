package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Consume the Meek — Instant {3}{B}{B} (Duel Decks: Zendikar vs. Eldrazi,
// #5548):
//
//	"Destroy each creature with mana value 3 or less. They can't be
//	 regenerated."
//
// Mana value is read through the layered view, so a token (mana value 0) and
// a face-down creature (mana value 0) are destroyed, as the rules have it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2555c38-1bfc-48c1-85aa-7e40117db78b",
		Name:         "Consume the Meek",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreaturesManaValue3OrLess, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.And(effects.Creature(), effects.ManaValueLE(3)), CantBeRegenerated: true}.Apply(ctx)
		},
	})
}
