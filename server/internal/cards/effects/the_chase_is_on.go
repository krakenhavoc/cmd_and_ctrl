package effects

// The Chase Is On — Instant {2}{R} (Murders at Karlov Manor, #4469):
//
//	"Target creature gets +3/+0 and gains first strike until end of turn.
//	 Investigate. (Create a Clue token. It's an artifact with '{2},
//	 Sacrifice this token: Draw a card.')"
//
// Auspicious Arrival's shape: the still-legal target (CR 608.2b) gets
// the pump and the keyword together, then one Clue (CR 701.16a).
//
// No simplification.
func init() {
	const label = "The Chase Is On — +3/+0 and first strike until end of turn"
	Register(Spec{
		OracleID:     "789899eb-3a33-4a41-b6b5-5e2f55a9af51",
		Name:         "The Chase Is On",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 3, 0, "first strike")),
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return pumpAndGrantFirstTarget(3, 0, label, "first strike")(ctx.Game, ctx.Item)
			}),
			CreateToken{Template: ClueToken(), N: 1},
		),
	})
}
