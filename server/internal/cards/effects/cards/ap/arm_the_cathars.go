package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arm the Cathars — Sorcery {1}{W}{W} (Innistrad: Crimson Vow, #5495):
//
//	"Until end of turn, target creature gets +3/+3, up to one other
//	 target creature gets +2/+2, and up to one other target creature gets
//	 +1/+1. Those creatures gain vigilance until end of turn."
//
// Three target clauses, the second and third marked Distinct ("other":
// different from every earlier pick). Each pick still legal as the spell
// resolves gets its own pump and vigilance; an empty or illegal clause
// affects nothing and does not stop the others (CR 608.2b). Cast with
// nothing but the first target, it is legal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6a1a91f-46b9-4e41-a3d3-3423b02c42a5",
		Name:         "Arm the Cathars",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.Distinct(effects.TargetCreature("up to one other target creature").WithCount(0, 1)),
			effects.Distinct(effects.TargetCreature("up to one other target creature").WithCount(0, 1)),
		),
		Purpose: effects.ForTargets(
			effects.BenefitUntilEOT(0, 3, 3, "vigilance"),
			effects.BenefitUntilEOT(1, 2, 2, "vigilance"),
			effects.BenefitUntilEOT(2, 1, 1, "vigilance"),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.S82VowB02ArmSlot(ctx, 0, 3) }),
			effects.Instruction(func(ctx *effects.Context) error { return effects.S82VowB02ArmSlot(ctx, 1, 2) }),
			effects.Instruction(func(ctx *effects.Context) error { return effects.S82VowB02ArmSlot(ctx, 2, 1) }),
		),
	})
}
