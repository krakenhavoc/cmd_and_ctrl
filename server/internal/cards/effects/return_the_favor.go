package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Return the Favor — Instant {R}{R}:
//
//	"Spree (Choose one or more additional costs.)
//	 + {1} — Copy target instant spell, sorcery spell, activated
//	   ability, or triggered ability. You may choose new targets for
//	   the copy.
//	 + {1} — Change the target of target spell or ability with a
//	   single target."
//
// Two Spree bullets (CR 702.172a) over the stack, each with its own
// "spell or ability" clause (CR 115.4, #1211).
//
// The first bullet's clause admits every activated and triggered
// ability and, among spells, only instants and sorceries — a creature
// spell is not a legal target. Which copy it makes depends on what the
// target turned out to be: a spell is copied by CopySpell (CR 707.10,
// the Reverberate path) and an ability by CopyAbility (the Lithoform
// Engine path). Both offer the CR 707.10c "you may choose new targets"
// prompt, and both give the copy to Return the Favor's controller
// (CR 707.10b), whoever controlled the original.
//
// The second bullet is Bolt Bend's sentence exactly: CR 115.7b's
// mandatory change of the single target, with "with a single target"
// as the clause predicate so a two-target spell is never offered.
//
// Chosen together, the bullets run in printed order: the copy first,
// then the change of target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "30377bd5-99ab-4888-9ea7-e5a069b527e8",
		Name:         "Return the Favor",
		Completeness: CompletenessFull,
		Modes: Spree(
			SpreeModeDoing("Copy target instant spell, sorcery spell, activated ability, or triggered ability. You may choose new targets for the copy.", "{1}",
				TargetSpellOrAbility("target instant spell, sorcery spell, activated ability, or triggered ability",
					AnyStackItem(AnAbilityItem(), ASpellItem(Or(Instant(), Sorcery())))),
				copyTheModesSpellOrAbility),
			SpreeModeDoing("Change the target of target spell or ability with a single target.", "{1}",
				TargetSpellOrAbility("target spell or ability with a single target", ItemHasASingleTarget()),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return ChangeTargets{
						StackID: t.ID,
						Policy:  game.RetargetChangeOne,
						Reason:  "Return the Favor — change the target",
					}.Apply(ctx)
				}),
		),
	})
}

// copyTheModesSpellOrAbility is "copy target spell or ability. You may
// choose new targets for the copy" as a modal bullet's body: the
// mode's still-legal target is copied by whichever primitive fits what
// it is — CopySpell for a spell, CopyAbility for an activated or
// triggered ability.
func copyTheModesSpellOrAbility(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	target := ctx.Game.StackItemForEffect(t.ID)
	if target == nil {
		return nil
	}
	if target.Kind == game.StackItemSpell {
		return CopySpell{StackID: t.ID, ChooseNewTargets: true}.Apply(ctx)
	}
	return CopyAbility{ItemID: t.ID, ChooseNewTargets: true}.Apply(ctx)
}
