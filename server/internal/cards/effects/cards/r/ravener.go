package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ravener — Creature — Tyranid {X}{G}{U}, 0/0 (Warhammer 40,000 Commander,
// S82 GAP batch 193, #5948):
//
//	"Flash
//	 Ravenous (This creature enters with X +1/+1 counters on it. If X is 5 or
//	 more, draw a card when it enters.)
//	 When this creature enters, target creature attacks target opponent this
//	 turn if able."
//
// X is the value announced for the spell (CR 107.3m). The counters are the
// cast-time entry counters (Morlun's XCounters); the ravenous draw is its own
// enters trigger that reads X off the permanent as it is built, so a copy or a
// cheat onto the battlefield (X = 0) draws nothing. The attack requirement is
// Dulcet Sirens': two target clauses (the creature, then the opponent), the
// creature bound to attack that player until end of turn if able.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "6c4ac289-edca-4dce-8572-817efbdd08b2",
		Name:                       "Ravener",
		Completeness:               effects.CompletenessFull,
		PrintedKeywords:            []string{"flash"},
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterPlusOne)},
		Triggered: []game.TriggeredAbility{
			effects.GapB193RavenousDraw("Ravener"),
			effects.Targeting(effects.WhenThisEnters("Ravener — target creature attacks target opponent this turn if able", effects.GapB193TargetCreatureAttacksTargetOpponent),
				effects.Clauses(effects.TargetCreature("target creature"), effects.TargetPlayer("target opponent", effects.Opponent()))),
		},
	})
}
