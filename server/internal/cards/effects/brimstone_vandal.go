package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "7af23925-6d32-407d-9a99-a57b2ebb63ac",
		Name:            "Brimstone Vandal",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		AsEnters:        BecomesDayAsEnters(),
		Triggered: []game.TriggeredAbility{
			WheneverDayBecomesNightOrNightBecomesDay("Brimstone Vandal — 1 damage to each opponent",
				func(g *game.Game, item *game.StackItem) error {
					return damageToEachOpponent(g, item, 1)
				}),
		},
	})
}
