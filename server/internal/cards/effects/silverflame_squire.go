package effects

// Silverflame Squire // On Alert — Creature — Human Soldier {1}{W}, 2/1 // Instant — Adventure {2}{W} (Throne of Eldraine, #5516):
//
//	Silverflame Squire — no rules text.
//	On Alert — "Target creature gets +2/+2 until end of turn. Untap it."
//
// Both faces register: the creature under the oracle ID and the Adventure
// under "<oracle_id>#1". The Adventure lifecycle (exile on resolution, cast
// the creature later, CR 715) is the engine's.
//
// No simplification.
const silverflameSquireOracleID = "93e4308e-e701-440e-861a-a30324654095"

func init() {
	Register(Spec{
		OracleID:     silverflameSquireOracleID,
		Name:         "Silverflame Squire",
		Completeness: CompletenessFull,
	})
	Register(Spec{
		OracleID:     silverflameSquireOracleID + "#1",
		Name:         "On Alert",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return boostTheTargetUntilEOT(2, 2, "On Alert — +2/+2 until end of turn")(ctx.Item, ctx)
			}),
			Instruction(func(ctx *Context) error { return untapTheTarget(ctx.Game, ctx.Item) }),
		),
	})
}
