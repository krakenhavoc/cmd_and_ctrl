package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Weather the Storm — Instant {1}{G} (Modern Horizons, #5519):
//
//	"You gain 3 life.
//	 Storm (When you cast this spell, copy it for each spell cast before it
//	 this turn.)"
//
// The shared storm trigger over a 3-life gain; each copy gains 3 more.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a091f7b-efc9-4aaa-9da8-157ef8099a36",
		Name:         "Weather the Storm",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Storm()},
		OnResolve:    effects.ResolveSteps(effects.GainLife{Amount: 3}),
	})
}
