package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flicker of Fate — Instant {1}{W} (Jumpstart 2022, #5948):
//
//	"Exile target creature or enchantment, then return it to the
//	 battlefield under its owner's control."
//
// Acrobatic Maneuver's flicker (CR 400.7: it returns as a new object).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d5b9d250-92f5-40d8-b962-4ae4adb466c8",
		Name:         "Flicker of Fate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or enchantment", effects.Or(effects.Creature(), effects.Enchantment())),
		OnResolve: effects.ResolveSteps(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
			return effects.Flicker{Target: id}.Apply(ctx)
		})),
	})
}
