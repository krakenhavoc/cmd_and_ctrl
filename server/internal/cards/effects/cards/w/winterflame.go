package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Winterflame — Instant {1}{U}{R} (Khans of Tarkir):
//
//	"Choose one or both —
//	 • Tap target creature.
//	 • Winterflame deals 2 damage to target creature."
//
// Molten Rebuke's "choose one or both" (CR 700.2): each bullet reads its own
// target, in printed order (CR 608.2c), and a target that is no longer legal
// skips only its own bullet (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8d3a0b81-1895-47a8-87a3-3e4093cbd951",
		Name:         "Winterflame",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeDoing("Tap target creature.",
				effects.TargetCreature("target creature"),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.TapTarget{Target: id}.Apply(ctx)
				})),
			effects.ModeDoing("Winterflame deals 2 damage to target creature.",
				effects.TargetCreature("target creature"),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 2}.Apply(ctx)
				})),
		),
	})
}

var _ = game.ZoneHand
