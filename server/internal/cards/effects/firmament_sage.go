package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Firmament Sage — {3}{U} Creature — Human Wizard 2/3 (#2561, ADR 0132):
//
//	"If it's neither day nor night, it becomes day as this creature enters.
//	 Whenever day becomes night or night becomes day, draw a card."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "57163ade-fe89-4ba0-98a9-daeeed5badb9",
		Name:         "Firmament Sage",
		Completeness: CompletenessFull,
		AsEnters:     BecomesDayAsEnters(),
		Triggered: []game.TriggeredAbility{
			WheneverDayBecomesNightOrNightBecomesDay("Firmament Sage — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
