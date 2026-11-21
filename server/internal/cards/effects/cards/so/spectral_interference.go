package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spectral Interference — Instant {1}{U} (Aetherdrift, #4464):
//
//	"Counter target artifact or creature spell unless its controller
//	 pays {4}."
//
// Spell Pierce's pay-unless (CounterUnlessPaid) with Hurl into
// History's "artifact or creature spell" target. The prompt goes to the
// spell's controller, and the spell cannot resolve while the {4} is
// outstanding.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb5ed146-5079-46ef-a3e0-e584fa4467f6",
		Name:         "Spectral Interference",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target artifact or creature spell", effects.Or(effects.Artifact(), effects.Creature())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.CounterUnlessPaid{
				StackID:  item.Targets[0].ID,
				Cost:     "{4}",
				Question: "Spectral Interference — pay {4} or your spell is countered",
			}.Apply(ctx)
		},
	})
}
