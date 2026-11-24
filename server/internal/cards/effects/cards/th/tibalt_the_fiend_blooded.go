package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tibalt, the Fiend-Blooded — Legendary Planeswalker — Tibalt {R}{R}, loyalty 2:
//
//	"+1: Draw a card, then discard a card at random.
//	 −4: Tibalt deals damage equal to the number of cards in target player's
//	 hand to that player.
//	 −6: Gain control of all creatures until end of turn. Untap them. They gain
//	 haste until end of turn."
//
// +1 draws first and then discards one card chosen at random from the whole
// hand, the drawn card included (CR 701.9b).
//
// −4 counts the hand as it resolves; a target player who left, or an empty
// hand, deals nothing (CR 120.8).
//
// −6 takes every creature on the battlefield as the ability resolves (CR
// 608.2c), then untaps them and grants haste, the Mob Rule shape.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ee9c57c6-41ef-4328-8417-e5a433eadf17",
		Name:            "Tibalt, the Fiend-Blooded",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 2,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Draw a card, then discard a card at random.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Draws: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.DrawCards{N: 1}).Apply(ctx); err != nil {
						return err
					}
					return g.DiscardRandomForEffect(item.Controller, 1)
				},
			},
			{
				Label:   "−4: Tibalt deals damage equal to the number of cards in target player's hand to that player.",
				Cost:    effects.LoyaltyCost(-4),
				Targets: effects.TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					pid, ok := effects.FirstLegalPlayerTarget(ctx)
					if !ok {
						return nil
					}
					p := g.PlayerByIDForEffect(pid)
					if p == nil || p.Hand == nil {
						return nil
					}
					return effects.DealDamage{Source: ctx.Source(), Target: pid, Amount: p.Hand.Size()}.Apply(ctx)
				},
			},
			{
				Label: "−6: Gain control of all creatures until end of turn. Untap them. They gain haste until end of turn.",
				Cost:  effects.LoyaltyCost(-6),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.RprB28StealAllMatching(ctx, "Tibalt, the Fiend-Blooded", effects.Creature())
				})),
			},
		},
	})
}
