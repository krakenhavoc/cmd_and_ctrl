package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "45d939d1-e9ca-4b4b-a0a9-59b936210a5b",
		Name:         "Oreskos Sun Guide",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventUntapCard, effects.Self, "Oreskos Sun Guide — you gain 2 life (inspired)", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
