package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prismari Command — Instant {1}{U}{R}:
//
//	"Choose two —
//	 • Prismari Command deals 2 damage to any target.
//	 • Target player draws two cards, then discards two cards.
//	 • Target player creates a Treasure token.
//	 • Destroy target artifact."
//
// The Izzet charm on the "choose two" template — Kolaghan's
// Command's shape (#764, ADR 0065 §3): three of the four bullets
// target, each chosen bullet gets its own target group, picked in
// the order chosen and resolved in that order (CR 608.2c), and each
// bullet's target is re-checked against ITS OWN clause at resolution
// (CR 608.2b) — an artifact that stopped being one is skipped while
// the other chosen bullet still happens.
//
// The draw-then-discard bullet is a real CR 701.8a choice: the
// targeted player draws two, then picks which two to pitch through
// the ordinary discard prompt. The discard is not "then draw a card
// for each" — nothing downstream reads how many were discarded, so a
// fire-and-forget QueueDiscardChoiceForEffect is exactly the printed
// card (contrast Mishra's Command's first bullet, which does need
// the count).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa3e28b1-131c-4223-81e0-18dfbab22c26",
		Name:         "Prismari Command",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose two", 2, 2,
			ModeDoing("Prismari Command deals 2 damage to any target.",
				TargetAny(),
				DealFixedDamageToModesTarget(2)),
			ModeDoing("Target player draws two cards, then discards two cards.",
				TargetPlayer("target player"),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					if err := (DrawCards{Player: t.ID, N: 2}).Apply(ctx); err != nil {
						return err
					}
					ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
						Player:   t.ID,
						Source:   item.SourceCardID,
						N:        2,
						Question: "Prismari Command — discard two cards",
					})
					return nil
				}),
			ModeDoing("Target player creates a Treasure token.",
				TargetPlayer("target player"),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return CreateToken{Controller: t.ID, Template: TreasureToken(), N: 1}.Apply(ctx)
				}),
			ModeDoing("Destroy target artifact.",
				TargetPermanent("target artifact", Artifact()),
				DestroyTheModesTarget),
		),
	})
}
