package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Midnight Charm — Instant {B} (Planar Chaos, #5595):
//
//	"Choose one —
//	 • Midnight Charm deals 1 damage to target creature and you gain 1 life.
//	 • Target creature gains first strike until end of turn.
//	 • Tap target creature."
//
// Three bullets on the shared mode bodies. The first is two steps, so the
// life gain waits for anything the damage asks (CR 608.2c); a target gone
// by resolution makes the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed228f67-3adf-46d4-ac45-e0278592850d",
		Name:         "Midnight Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeSteps("Midnight Charm deals 1 damage to target creature and you gain 1 life.",
				effects.TargetCreature("target creature"),
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.DealFixedDamageToModesTarget(1)(ctx.Item, ctx, ctx.Occurrence())
				}),
				effects.GainLife{Amount: 1}), effects.ForTargets(effects.DamageToTarget(0, 1))),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gains first strike until end of turn.",
				effects.TargetCreature("target creature"),
				effects.PlcGrantTheModesTarget("Midnight Charm — first strike until end of turn", "first strike")),
				effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "first strike"))),
			effects.ModeDoing("Tap target creature.", effects.TargetCreature("target creature"), effects.TapTheModesTarget),
		),
	})
}
