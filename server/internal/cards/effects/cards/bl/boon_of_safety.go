package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boon of Safety — Instant {W} (Streets of New Capenna):
//
//	"Put a shield counter on target creature. (If it would be dealt
//	 damage or destroyed, remove a shield counter from it instead.)
//	 Scry 1."
//
// The counter protects through the CR 122.1c built-ins
// (game/shield_counters.go). With its only target illegal the spell
// does nothing on resolution, the scry included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "34f0beae-a8e2-4474-9f7f-4735d3f26863",
		Name:         "Boon of Safety",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Scry: 1, Answers: game.AnswerSurvive},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := effects.ShieldCounterOnFirstTarget(ctx); err != nil {
				return err
			}
			return effects.Scry{N: 1}.Apply(ctx)
		},
	})
}
