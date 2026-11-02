package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grow Extra Arms — Instant {1}{G}:
//
//	"This spell costs {1} less to cast if it targets a Spider.
//	 Target creature gets +4/+4 until end of turn."
//
// The discount is a self cost modifier that reads the announced target
// (CostsLessIfItTargets, as Depower's does), judged as the spell is
// cast. No simplification.
func init() {
	Register(Spec{
		OracleID:     "52f03e73-3dba-44c2-b256-10e7e208b961",
		Name:         "Grow Extra Arms",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 4, 4)),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(1, "This spell costs {1} less to cast if it targets a Spider.",
				And(Creature(), OfCreatureType("Spider"))),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return BoostUntilEOT{Target: t.ID, Power: 4, Toughness: 4, Label: "Grow Extra Arms — +4/+4"}.Apply(ctx)
			}
			return nil
		},
	})
}
