package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frantic Confrontation — Instant {X}{R} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"Target creature you control gets +X/+0 and gains first strike and
//	 trample until end of turn."
//
// X is announced with the cast (CR 107.3, 601.2b) and read back with
// ctx.X(); the pump (layer 7c) and the two keywords (layer 6) are both
// until-end-of-turn effects on the one target, applied as it resolves. A
// target that has become illegal gets neither (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a866cc05-69b0-45f2-8829-b6ef08193ed1",
		Name:         "Frantic Confrontation",
		Completeness: CompletenessFull,
		XMatters:     true,
		Purpose:      ForTargets(PowerXUntilEOT(0, "first strike", "trample")),
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if err := (BoostUntilEOT{Target: t.ID, Power: ctx.X(), Label: "Frantic Confrontation — +X/+0"}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"first strike", "trample"},
				Label: "Frantic Confrontation — first strike and trample"}.Apply(ctx)
		},
	})
}
