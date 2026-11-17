package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moment of Defiance — Instant for {2}{B} (The Brothers' War):
//
//	"Target creature gets +2/+1 and gains lifelink until end of turn.
//	Draw a card."
func init() {
	const label = "Moment of Defiance — +2/+1 and lifelink until end of turn"
	p := ForTargets(BenefitUntilEOT(0, 2, 1, "lifelink"))
	p.Draws = 1
	Register(Spec{
		OracleID:     "257f81fb-b3c9-4ebe-9889-dd0ef960bba3",
		Name:         "Moment of Defiance",
		Purpose:      p,
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
					return nil
				}
				if err := (BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 1, Label: label}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"lifelink"}, Label: label}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
			}),
		),
	})
}
