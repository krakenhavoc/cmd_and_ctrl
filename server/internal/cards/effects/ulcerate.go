package effects

import "github.com/google/uuid"

// Ulcerate — Instant {B} (Foundations Jumpstart, #5929):
//
//	"Target creature gets -3/-3 until end of turn. You lose 3 life."
//
// A spell whose only target is gone fizzles (CR 608.2b), so no life is lost.
func init() {
	Register(Spec{
		OracleID:     "4a9e73b7-8c8b-4489-b670-ce49c2c8a46a",
		Name:         "Ulcerate",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(HarmUntilEOT(0, -3, -3)),
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return BoostUntilEOT{Target: id, Power: -3, Toughness: -3, Label: "Ulcerate — -3/-3"}.Apply(ctx)
			}),
			GainLife{Amount: -3},
		),
	})
}
