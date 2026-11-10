package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ferocification — Enchantment {2}{R} (Outlaws of Thunder Junction):
//
//	"At the beginning of combat on your turn, choose one —
//	 • Target creature you control gets +2/+0 until end of turn.
//	 • Target creature you control gains menace and haste until end of turn."
//
// A modal triggered ability (CR 700.2b): the mode, and the target of the
// mode chosen, are picked as the trigger goes on the stack, and only a
// mode with a legal target can be chosen (Giant-Sized Flying Ant's shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "814e06c0-9b61-4a64-8c67-73a8eb00a768",
		Name:         "Ferocification",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{func() game.TriggeredAbility {
			t := AtBeginningOfYourCombat("Ferocification — choose one: +2/+0, or menace and haste",
				func(*game.Game, *game.StackItem) error { return nil })
			t.Modes = ChooseOne(
				ModeDoing("Target creature you control gets +2/+0 until end of turn.",
					TargetCreature("target creature you control", YouControl()),
					func(_ *game.StackItem, ctx *Context, occ int) error {
						t, ok := ModeTarget(ctx, occ)
						if !ok {
							return nil
						}
						return BoostUntilEOT{Target: t.ID, Power: 2, Label: "Ferocification — +2/+0"}.Apply(ctx)
					}),
				ModeDoing("Target creature you control gains menace and haste until end of turn.",
					TargetCreature("target creature you control", YouControl()),
					func(_ *game.StackItem, ctx *Context, occ int) error {
						t, ok := ModeTarget(ctx, occ)
						if !ok {
							return nil
						}
						return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"menace", "haste"}, Label: "Ferocification — menace and haste"}.Apply(ctx)
					}),
			)
			return t
		}()},
	})
}
