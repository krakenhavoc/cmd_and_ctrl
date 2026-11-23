package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kederekt Leviathan — Creature — Leviathan {6}{U}{U}, 5/5 (S82 GAP, #5948):
//
//	"When this creature enters, return all other nonland permanents to their
//	 owners' hands.
//	 Unearth {6}{U}"
//
// Evacuation's mass bounce, sparing the source; unearth is the shared keyword.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "edf459c0-53e4-47e8-8098-105656f9b49c",
		Name:         "Kederekt Leviathan",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepBounce}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Kederekt Leviathan — return all other nonland permanents to their owners' hands",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BounceAllMatching{
						Match: effects.And(effects.Nonland(), effects.OtherThan(item.SourceCardID)),
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{effects.Unearth("{6}{U}")},
	})
}
