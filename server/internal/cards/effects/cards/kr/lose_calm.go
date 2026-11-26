package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lose Calm — Sorcery {3}{R} (Dragons of Tarkir):
//
//	"Gain control of target creature until end of turn. Untap that creature. It
//	 gains haste and menace until end of turn."
//
// The shared gain-control, untap and keyword grant (ThreatenAndGrant), in
// printed order (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "73d71731-f694-471c-b2ad-a36854bb06a1",
		Name:         "Lose Calm",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
			return effects.ThreatenAndGrant(ctx, id, "Lose Calm", "haste", "menace")
		})),
	})
}
