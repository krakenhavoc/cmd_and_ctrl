package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brimstone Vandal — {2}{R} Creature — Devil 2/3 (#2561, ADR 0132):
//
//	"Menace
//	 If it's neither day nor night, it becomes day as this creature enters.
//	 Whenever day becomes night or night becomes day, this creature deals
//	 1 damage to each opponent."
//
// The as-enters clause is AsEnters (off the stack), the trigger fires on a
// flip only (the first designation a game gains is not one), and the
// damage is one damage instance across every opponent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7af23925-6d32-407d-9a99-a57b2ebb63ac",
		Name:            "Brimstone Vandal",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		AsEnters:        effects.BecomesDayAsEnters(),
		Triggered: []game.TriggeredAbility{
			effects.WheneverDayBecomesNightOrNightBecomesDay("Brimstone Vandal — 1 damage to each opponent",
				effects.Do(effects.DamageEachOpponent{N: 1})),
		},
	})
}
