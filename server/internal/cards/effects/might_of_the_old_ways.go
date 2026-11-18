package effects

// Might of the Old Ways — Instant {1}{G} (Midnight Hunt, #5497):
//
//	"Target creature gets +2/+2 until end of turn.
//	 Coven — Then if you control three or more creatures with different
//	 powers, draw a card."
//
// The coven check is made after the pump, as the spell resolves; the
// pump can't change anyone's power but the target's, and that is read
// live. A target that is gone makes the whole spell fizzle (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "266d3f06-c982-4427-b421-7520e5da5c2e",
		Name:         "Might of the Old Ways",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				return BoostUntilEOT{Target: ts[0].ID, Power: 2, Toughness: 2, Label: "Might of the Old Ways — +2/+2"}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				if len(ctx.LegalTargets()) == 0 || !trekB48Coven(ctx.Game, ctx.Controller(), ctx.Item.SourceCardID) {
					return nil
				}
				return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
			}),
		),
	})
}
