package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shore Up — Instant {U} (Bloomburrow, #4467):
//
//	"Target creature you control gets +1/+1 and gains hexproof until end
//	 of turn. Untap it. (It can't be the target of spells or abilities
//	 your opponents control.)"
//
// Acrobatic Leap's body on a creature you control: the pump and the
// keyword are locked to the target at resolution (CR 611.2c) and end in
// the cleanup step (CR 514.2), then it untaps (CR 701.26b). A target that
// left in response, or that you no longer control, counters the spell
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c9d8c983-2845-42e6-b6b8-bf7cf2729b23",
		Name:         "Shore Up",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 1, "hexproof")),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if err := (BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1, Label: "Shore Up — +1/+1"}).Apply(ctx); err != nil {
				return err
			}
			if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"hexproof"}, Label: "Shore Up — hexproof"}).Apply(ctx); err != nil {
				return err
			}
			return UntapTarget{Target: t.ID}.Apply(ctx)
		},
	})
}
