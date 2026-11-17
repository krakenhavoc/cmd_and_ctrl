package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Defabricate — Instant {1}{U} (The Brothers' War, #5482):
//
//	"Choose one —
//	 • Counter target artifact or enchantment spell. If that spell is
//	   countered this way, exile it instead of putting it into its
//	   owner's graveyard.
//	 • Counter target activated or triggered ability."
//
// The first bullet is No More Lies' exile destination on a counter (CR
// 701.6a): the spell goes to exile instead of its owner's graveyard, and a
// spell that can't be countered stays on the stack. The second is Stifle's
// ability counter (CounterTheModesTarget, which takes a stack item of either
// kind). An ability copy of a spell is not an "activated or triggered
// ability" and is not offered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "60fc5eb0-95d6-49b0-a538-945ef599e484",
		Name:         "Defabricate",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Counter target artifact or enchantment spell. If that spell is countered this way, exile it instead of putting it into its owner's graveyard.",
				TargetSpell("target artifact or enchantment spell", Or(Artifact(), Enchantment())),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return CounterTarget{StackID: t.ID, Dest: game.ZoneRef{Kind: game.ZoneExile}}.Apply(ctx)
				}),
			ModeDoing("Counter target activated or triggered ability.",
				AbilityOnStack("target activated or triggered ability"),
				CounterTheModesTarget),
		),
	})
}
