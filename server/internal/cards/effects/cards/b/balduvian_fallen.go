package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Balduvian Fallen — Creature — Zombie {3}{B}, 3/5 (Coldsnap, S82 OLD b558):
//
//	"Cumulative upkeep {1}
//	 Whenever this creature's cumulative upkeep is paid, it gets +1/+0 until
//	 end of turn for each {B} or {R} spent this way."
//
// The payment announces the mana it spent (EventCumulativeUpkeepPaid), so the
// bonus counts the black and red symbols of that payment only; generic mana
// paid with another colour adds nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f60a7ce-c714-4e87-a2d9-3710f2d610fa",
		Name:         "Balduvian Fallen",
		Completeness: effects.CompletenessFull,
		// The mana it reads is a pay-unless payment, made from the pool at the
		// upkeep prompt, never a cast or an activation a waiver could let
		// through short: the declaration that never asks (#3855).
		ReadsManaSpent: game.ReadsPoolManaOnly,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Balduvian Fallen — cumulative upkeep {1}", "{1}"),
			effects.WhenCumulativeUpkeepIsPaid("Balduvian Fallen — it gets +1/+0 until end of turn for each {B} or {R} spent this way",
				func(g *game.Game, item *game.StackItem) error {
					n := game.SpentSymbolsOf(item.Trigger.Event, "B", "R")
					if n <= 0 {
						return nil
					}
					return effects.ThisGetsUntilEndOfTurn(n, 0, "Balduvian Fallen — +1/+0 for each {B} or {R} spent")(g, item)
				}),
		},
	})
}
