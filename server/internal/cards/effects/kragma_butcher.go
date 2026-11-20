package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kragma Butcher — Creature — Minotaur Warrior {2}{R}, 2/3 (Born of the Gods,
// #5559):
//
//	"Inspired — Whenever this creature becomes untapped, it gets +2/+0
//	 until end of turn."
//
// Inspired is an ability word: the trigger is Key to the City's
// (EventUntapCard on itself), so it fires in the untap step and on
// any other untap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3e6dad31-6192-4675-81df-09689906c12b",
		Name:         "Kragma Butcher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventUntapCard, Self, "Kragma Butcher — it gets +2/+0 until end of turn (inspired)",
				thisGetsUntilEndOfTurn(2, 0, "Kragma Butcher — +2/+0")),
		},
	})
}
