package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Peppersmoke — Kindred Instant — Faerie {B} (Lorwyn, S82 OLD b546, #5593):
//
//	"Target creature gets -1/-1 until end of turn. If you control a Faerie,
//	 draw a card."
//
// The shrink, then the draw, decided as the spell resolves (CR 608.2c): a
// Faerie you control by then counts, a changeling included. A creature that
// is an illegal target by resolution stops the whole spell, draw included
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "071733a2-6dc6-4b6e-8ed6-a130d14657fd",
		Name:         "Peppersmoke",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -1, -1)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.BoostUntilEOT{Target: id, Power: -1, Toughness: -1, Label: "Peppersmoke — -1/-1"}.Apply(ctx)
			}),
			effects.S82OldB546DrawIfYouControl("Faerie"),
		),
	})
}
