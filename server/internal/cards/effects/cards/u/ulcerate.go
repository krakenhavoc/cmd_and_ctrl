package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ulcerate — Instant {B} (Foundations Jumpstart, #5929):
//
//	"Target creature gets -3/-3 until end of turn. You lose 3 life."
//
// A spell whose only target is gone fizzles (CR 608.2b), so no life is lost.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a9e73b7-8c8b-4489-b670-ce49c2c8a46a",
		Name:         "Ulcerate",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -3, -3)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.BoostUntilEOT{Target: id, Power: -3, Toughness: -3, Label: "Ulcerate — -3/-3"}.Apply(ctx)
			}),
			effects.GainLife{Amount: -3},
		),
	})
}
