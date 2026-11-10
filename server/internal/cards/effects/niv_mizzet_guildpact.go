package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "46570366-5e58-40b7-b725-833321a3b7b1",
		Name:            "Niv-Mizzet, Guildpact",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", game.HexproofFromMulticolored},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisDealsCombatDamageToAPlayer(
					"Niv-Mizzet — X damage to any target, target player draws X cards, you gain X life",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						x := differentColorPairsAmong(g, ctx.Controller())
						if x <= 0 {
							return nil
						}
						if t, ok := ctx.ClauseTarget(0); ok {
							if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: x}).Apply(ctx); err != nil {
								return err
							}
						}
						if t, ok := ctx.ClauseTarget(1); ok {
							if err := (DrawCards{Player: t.ID, N: x}).Apply(ctx); err != nil {
								return err
							}
						}
						return GainLife{Player: ctx.Controller(), Amount: x}.Apply(ctx)
					}),
				TargetAny().Then(TargetPlayer("target player"))),
		},
	})
}
