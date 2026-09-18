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
// group, and they resolve in the order chosen (CR 700.2c) — the
// order that decides whether the token copy sees the creature the
// bounce is about to return.
//
// Declared simplification, weaker than printed: the second bullet,
// "counter target activated or triggered ability", is NOT offered.
// An ability on the stack is a StackMeta item rather than a card in
// the stack zone, and nothing in the targeting vocabulary can name
// one — a target clause walks ZONES. Offering the bullet and then
// failing to find anything to point at would be a mode that can
// never be taken; leaving it out is four bullets instead of five,
// and never a line the printed card forbids.
func init() {
	Register(Spec{
		OracleID:     "56148ae7-a9df-4771-8d53-d9ffb815c884",
		Name:         "Sublime Epiphany",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The \"counter target activated or triggered ability\" mode isn't offered — abilities on the stack can't be targeted yet."},
		Modes: ChooseOneOrMore(
			ModeDoing("Counter target spell.",
				TargetSpell("target spell"),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return CounterTarget{StackID: t.ID}.Apply(ctx)
				}),
			ModeDoing("Return target nonland permanent to its owner's hand.",
				TargetPermanent("target nonland permanent", Nonland()),
				BounceTheModesTarget),
			ModeDoing("Create a token that's a copy of target creature you control.",
				TargetCreature("target creature you control", YouControl()),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return CreateTokenCopy{Controller: item.Controller, Copy: t.ID, N: 1}.Apply(ctx)
				}),
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
