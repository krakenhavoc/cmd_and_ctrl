package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ferocification — Enchantment {2}{R} (Outlaws of Thunder Junction,
// #4468):
//
//	"At the beginning of combat on your turn, choose one —
//	 • Target creature you control gets +2/+0 until end of turn.
//	 • Target creature you control gains menace and haste until end of
//	   turn."
//
// A modal triggered ability (CR 700.2b, Appa, Loyal Sky Bison's shape):
// the mode and its target are chosen as the trigger is put on the stack,
// and the chosen bullet's target is re-checked as it resolves (CR
// 608.2b).
//
// No simplification.
func init() {
	trigger := AtBeginningOfYourCombat("Ferocification — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = ChooseOne(
		ModeWithPurpose(ModeDoing("Target creature you control gets +2/+0 until end of turn.",
			TargetCreature("target creature you control", YouControl()),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return BoostUntilEOT{Target: t.ID, Power: 2, Label: "Ferocification — +2/+0"}.Apply(ctx)
			}), ForTargets(BenefitUntilEOT(0, 2, 0))),
		ModeWithPurpose(ModeDoing("Target creature you control gains menace and haste until end of turn.",
			TargetCreature("target creature you control", YouControl()),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"menace", "haste"}, Label: "Ferocification — menace and haste"}.Apply(ctx)
			}), ForTargets(BenefitUntilEOT(0, 0, 0, "menace", "haste"))),
	)
	Register(Spec{
		OracleID:     "814e06c0-9b61-4a64-8c67-73a8eb00a768",
		Name:         "Ferocification",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
