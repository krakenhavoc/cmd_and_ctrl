package effects

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
	Register(Spec{
		OracleID:     "ed228f67-3adf-46d4-ac45-e0278592850d",
		Name:         "Midnight Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeSteps("Midnight Charm deals 1 damage to target creature and you gain 1 life.",
				TargetCreature("target creature"),
				Instruction(func(ctx *Context) error {
					return DealFixedDamageToModesTarget(1)(ctx.Item, ctx, ctx.Occurrence())
				}),
				GainLife{Amount: 1}), ForTargets(DamageToTarget(0, 1))),
			ModeWithPurpose(ModeDoing("Target creature gains first strike until end of turn.",
				TargetCreature("target creature"),
				plcGrantTheModesTarget("Midnight Charm — first strike until end of turn", "first strike")),
				ForTargets(BenefitUntilEOT(0, 0, 0, "first strike"))),
			ModeDoing("Tap target creature.", TargetCreature("target creature"), TapTheModesTarget),
		),
	})
}
