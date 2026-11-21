package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Craftwork Crusher — Artifact Creature — Boar Construct {3}{R}{R}{G}{G},
// 7/5:
//
//	"Trample
//	 When this creature enters, choose two —
//	 • This creature deals 4 damage to target creature or planeswalker.
//	 • Create a 2/2 colorless Wizard Soldier creature token named Cadet.
//	 • Draw a card."
//
// A modal enters trigger (CR 603.3c): the two bullets are picked as the
// trigger goes on the stack and a bullet with no legal target (the first,
// on an empty board) cannot be picked. They resolve in printed order
// whichever order they were chosen in. The damage is dealt by the Crusher
// and does nothing if it has left or the target is gone.
//
// No simplifications.
func init() {
	enters := effects.WhenThisEnters("Craftwork Crusher — choose two", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = effects.ChooseN("Choose two", 2, 2,
		effects.ModeWithPurpose(effects.ModeDoing("This creature deals 4 damage to target creature or planeswalker.",
			effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 4}.Apply(ctx)
			}), effects.ForTargets(effects.DamageToTarget(0, 4))),
		effects.ModeDoing("Create a 2/2 colorless Wizard Soldier creature token named Cadet.",
			nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("2/2 colorless Wizard Soldier named Cadet"), N: 1}.Apply(ctx)
			}),
		effects.ModeWithPurpose(effects.ModeDoing("Draw a card.",
			nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.DrawCards{N: 1}.Apply(ctx)
			}), game.Purpose{Draws: 1}),
	)
	effects.Register(effects.Spec{
		OracleID:        "d25284b9-50e4-4dcb-9aaf-6f6478c4a90a",
		Name:            "Craftwork Crusher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
