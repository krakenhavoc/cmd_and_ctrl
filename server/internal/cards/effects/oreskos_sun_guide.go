package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oreskos Sun Guide — Creature — Cat Monk {1}{W}, 2/2 (Born of the Gods,
// #5559):
//
//	"Inspired — Whenever this creature becomes untapped, you gain 2 life."
//
// Inspired is an ability word: the trigger is Key to the City's
// (EventUntapCard on itself), so it fires in the untap step and on
// any other untap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "45d939d1-e9ca-4b4b-a0a9-59b936210a5b",
		Name:         "Oreskos Sun Guide",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventUntapCard, Self, "Oreskos Sun Guide — you gain 2 life (inspired)", Do(GainLife{Amount: 2})),
		},
	})
}
