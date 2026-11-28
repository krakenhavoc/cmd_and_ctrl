package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wiitigo — Creature — Yeti {3}{G}{G}{G}, 0/0 (Ice Age, S82 OLD batch 845, #5636):
//
//	"This creature enters with six +1/+1 counters on it.
//	 At the beginning of your upkeep, put a +1/+1 counter on this creature if
//	 it has blocked or been blocked since your last upkeep. Otherwise, remove
//	 a +1/+1 counter from it."
//
// The six counters are the CR 614.1c entry clause. "Since your last upkeep"
// is read off the event log as the ability resolves (OldB845BlockedSince
// YourLastUpkeep), so no unprinted bookkeeping trigger is needed; the choice
// between the two halves is one instruction, not an intervening "if".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7f956a21-854e-41cc-96f9-ad774e69fa8e",
		Name:         "Wiitigo",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.B10EntersWithCounters(game.CounterPlusOne, 6, "Enters with six +1/+1 counters on it.")},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Wiitigo — put a +1/+1 counter on it if it has blocked or been blocked since your last upkeep, otherwise remove one",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
						return nil
					}
					if effects.OldB845BlockedSinceYourLastUpkeep(g, item.SourceCardID, item.Controller) {
						return effects.PutACounterOnThis(game.CounterPlusOne)(g, item)
					}
					return effects.RemoveACounterFromThis(game.CounterPlusOne)(g, item)
				}),
		},
	})
}
