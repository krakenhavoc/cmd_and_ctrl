package cardssy

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Take Heart — Instant {W} (Jumpstart, #5510):
//
//	"Target creature gets +2/+2 until end of turn. You gain 1 life for each
//	 attacking creature you control."
//
// The life is counted as the spell resolves: every attacking creature you
// control then, the target's pump included. A spell whose only target has
// become illegal does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c5e30c3-0b9f-4d3e-af44-be7fcd34df25",
		Name:         "Take Heart",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Take Heart — +2/+2"}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				n := 0
				for _, id := range effects.AttackingCreatures(ctx.Game) {
					if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.Controller == ctx.Controller() {
						n++
					}
				}
				return effects.GainLife{Amount: n}.Apply(ctx)
			}),
		),
	})
}
