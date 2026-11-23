package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cathars' Crusade — Enchantment {3}{W}{W} (EDHREC rank 553):
//
//	"Whenever a creature you control enters, put a +1/+1 counter on
//	 each creature you control."
//
// The token deck's exponential anthem: N creatures entering is N
// triggers, each growing the whole board. Every entering creature —
// token or card, cast or reanimated — fires it, and the set of
// creatures receiving counters is read when each trigger RESOLVES,
// so the creature that caused the trigger gets its own counter, as
// printed. Counters go through AddCounter, so Doubling Season and
// Hardened Scales apply to each one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc65ac73-5bef-4ecb-ad8e-39199084c027",
		Name:         "Cathars' Crusade",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.CreatureEnteredUnderYourControl, "Cathars' Crusade — a +1/+1 counter on each creature you control", func(g *game.Game, item *game.StackItem) error {
				return effects.B11PutCountersOnEachCreatureYouControl(g, item, 1)
			}),
		},
	})
}
