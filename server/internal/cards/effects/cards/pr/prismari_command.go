package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
// The draw-then-discard bullet is a real CR 701.9a choice: the
// targeted player draws two, then picks which two to pitch through
// the ordinary discard prompt. The discard is not "then draw a card
// for each" — nothing downstream reads how many were discarded, so a
// fire-and-forget QueueDiscardChoiceForEffect is exactly the printed
// card (contrast Mishra's Command's first bullet, which does need
// the count).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa3e28b1-131c-4223-81e0-18dfbab22c26",
		Name:         "Prismari Command",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose two", 2, 2,
			effects.ModeWithPurpose(effects.ModeDoing("Prismari Command deals 2 damage to any target.",
				effects.TargetAny(),
				effects.DealFixedDamageToModesTarget(2)),
				effects.ForTargets(effects.DamageToTarget(0, 2))),
			// Two instructions, so the discard waits for a prompt the draw
			// asks (CR 608.2c, #4155).
			effects.ModeWithPurpose(effects.ModeSteps("Target player draws two cards, then discards two cards.",
				effects.TargetPlayer("target player"),
				effects.Instruction(func(ctx *effects.Context) error {
					t, ok := effects.ModeTarget(ctx, ctx.Occurrence())
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return effects.DrawCards{Player: t.ID, N: 2}.Apply(ctx)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					t, ok := effects.ModeTarget(ctx, ctx.Occurrence())
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
						Player:   t.ID,
						Source:   ctx.Source(),
						N:        2,
						Question: "Prismari Command — discard two cards",
					})
					return nil
				})),
				effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 2, Discards: 2})),
			effects.ModeWithPurpose(effects.ModeDoing("Target player creates a Treasure token.",
				effects.TargetPlayer("target player"),
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return effects.CreateToken{Controller: t.ID, Template: effects.TreasureToken(), N: 1}.Apply(ctx)
				}),
				effects.ForTargets(game.TargetPurpose{Slot: 0, Tokens: 1})),
			effects.ModeDoing("Destroy target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact()),
				effects.DestroyTheModesTarget),
		),
	})
}
