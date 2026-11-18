package effects

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
	Register(Spec{
		OracleID:     "c6a1a91f-46b9-4e41-a3d3-3423b02c42a5",
		Name:         "Arm the Cathars",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature"),
			Distinct(TargetCreature("up to one other target creature").WithCount(0, 1)),
			Distinct(TargetCreature("up to one other target creature").WithCount(0, 1)),
		),
		Purpose: ForTargets(
			BenefitUntilEOT(0, 3, 3, "vigilance"),
			BenefitUntilEOT(1, 2, 2, "vigilance"),
			BenefitUntilEOT(2, 1, 1, "vigilance"),
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return s82VowB02ArmSlot(ctx, 0, 3) }),
			Instruction(func(ctx *Context) error { return s82VowB02ArmSlot(ctx, 1, 2) }),
			Instruction(func(ctx *Context) error { return s82VowB02ArmSlot(ctx, 2, 1) }),
		),
	})
}
