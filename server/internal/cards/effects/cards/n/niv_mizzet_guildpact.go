package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Niv-Mizzet, Guildpact — Legendary Creature — Dragon Avatar
// {W}{U}{B}{R}{G} (Murders at Karlov Manor, #4469):
//
//	"Flying, hexproof from multicolored
//	 Whenever Niv-Mizzet deals combat damage to a player, it deals X
//	 damage to any target, target player draws X cards, and you gain X
//	 life, where X is the number of different color pairs among
//	 permanents you control that are exactly two colors."
//
// Hexproof from multicolored is the colour-count quality protection has
// (game.HexproofFromMulticolored, two or more colours, CR 105.2b). The
// trigger has two targets, each judged on its own as it resolves (CR
// 608.2b), and X is read once, as the ability resolves (the ruling), from
// the layered colours of the permanents you control. Niv-Mizzet is the
// damage's source; the draw is the chosen player's and the life gain is
// yours.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "46570366-5e58-40b7-b725-833321a3b7b1",
		Name:            "Niv-Mizzet, Guildpact",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.HexproofFromMulticolored},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisDealsCombatDamageToAPlayer(
					"Niv-Mizzet — X damage to any target, target player draws X cards, you gain X life",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						x := effects.DifferentColorPairsAmong(g, ctx.Controller())
						if x <= 0 {
							return nil
						}
						if t, ok := ctx.ClauseTarget(0); ok {
							if err := (effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: x}).Apply(ctx); err != nil {
								return err
							}
						}
						if t, ok := ctx.ClauseTarget(1); ok {
							if err := (effects.DrawCards{Player: t.ID, N: x}).Apply(ctx); err != nil {
								return err
							}
						}
						return effects.GainLife{Player: ctx.Controller(), Amount: x}.Apply(ctx)
					}),
				effects.TargetAny().Then(effects.TargetPlayer("target player"))),
		},
	})
}
