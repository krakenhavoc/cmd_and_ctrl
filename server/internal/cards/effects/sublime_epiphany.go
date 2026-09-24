package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sublime Epiphany — Instant {4}{U}{U} (EDHREC rank 1707):
//
//	"Choose one or more —
//	 • Counter target spell.
//	 • Counter target activated or triggered ability.
//	 • Return target nonland permanent to its owner's hand.
//	 • Create a token that's a copy of target creature you control.
//	 • Target player draws a card."
//
// The "choose one or more" card (CR 700.2, Min 1 / Max n). Before
// #764 the shape was declarable but useless: at most one chosen
// bullet could target, so the only legal "one or more" was one.
// Now every chosen bullet is its own occurrence with its own target
// group, and they resolve in the order chosen (CR 608.2c) — the
// order that decides whether the token copy sees the creature the
// bounce is about to return.
//
// The SECOND BULLET is the card the targeting seam was measured by:
// it shipped with four bullets because "counter target activated or
// triggered ability" could not be written, and #1211 gives it the
// fifth. The bullet is `AbilityOnStack` plus the same CounterTarget
// the first bullet uses — the primitive discriminates on
// StackItem.Kind, so the two bullets differ only in their clause.
//
// The two counter bullets are NOT one bullet with a wider clause,
// even though `TargetSpellOrAbility` could express that in one line.
// CR 700.2 makes each bullet an independent choice with its own
// target, so "choose one or more" lets a player take BOTH — counter a
// spell and an ability with one Epiphany — and a merged bullet would
// silently take that away. Printed structure wins over expressible
// structure.
//
// An unfillable bullet is not offered (#764), so the second bullet
// simply does not appear when no ability is on the stack, which is
// the ordinary case and needs nothing of its own.
func init() {
	Register(Spec{
		OracleID:     "56148ae7-a9df-4771-8d53-d9ffb815c884",
		Name:         "Sublime Epiphany",
		Completeness: CompletenessFull,
		Modes: ChooseOneOrMore(
			ModeDoing("Counter target spell.",
				TargetSpell("target spell"),
				CounterTheModesTarget),
			ModeDoing("Counter target activated or triggered ability.",
				AbilityOnStack("target activated or triggered ability"),
				CounterTheModesTarget),
			ModeDoing("Return target nonland permanent to its owner's hand.",
				TargetPermanent("target nonland permanent", Nonland()),
				BounceTheModesTarget),
			ModeDoing("Create a token that's a copy of target creature you control.",
				TargetCreature("target creature you control", YouControl()),
				TokenCopyTheModesTarget),
			ModeDoing("Target player draws a card.",
				TargetPlayer("target player"),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return DrawCards{Player: t.ID, N: 1}.Apply(ctx)
				}),
		),
	})
}
