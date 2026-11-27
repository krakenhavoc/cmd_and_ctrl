package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Boiling Blood — Instant {2}{R} (Weatherlight, S82 older sets batch 777, #5629):
//
//	"Target creature attacks this turn if able.
//	 Draw a card."
//
// Into the Fray's requirement, then the draw, in printed order (CR 608.2c). A target that is
// illegal by resolution does nothing, and the draw is still the caster's (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd55a003-da0d-4522-b08a-d0731b1050c6",
		Name:         "Boiling Blood",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if t, ok := effects.FirstLegalTarget(ctx); ok && t.Kind == game.TargetCard {
				if err := (effects.ScopedEffectFor{
					Target:   t.ID,
					Mods:     []game.Mod{game.AddAttackRequirementMod(uuid.Nil)},
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Boiling Blood — attacks this turn if able",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
