package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jace's Scrutiny — Instant {1}{U} (Shadows over Innistrad, #5543):
//
//	"Target creature gets -4/-0 until end of turn.
//	 Investigate. (Create a Clue token. It's an artifact with "{2},
//	 Sacrifice this token: Draw a card.")"
//
// Code of Constraint's -4/-0 until end of turn, then investigate (CR
// 701.16) in printed order (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8578c462-7b2d-4b69-ad10-a2eab44ac98d",
		Name:         "Jace's Scrutiny",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -4, 0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.BoostUntilEOT{Target: id, Power: -4, Label: "Jace's Scrutiny — -4/-0"}.Apply(ctx)
			}),
			effects.Investigate{N: 1},
		),
	})
}
