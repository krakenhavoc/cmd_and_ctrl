package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zedruu the Greathearted — Legendary Creature — Minotaur Monk
// {1}{U}{R}{W}, 2/4 (Commander Legends):
//
//	"At the beginning of your upkeep, you gain X life and draw X cards,
//	 where X is the number of permanents you own that your opponents
//	 control.
//	 {U}{R}{W}: Target opponent gains control of target permanent you
//	 control."
//
// The upkeep trigger counts the battlefield as it resolves (CR 608.2c): every
// permanent the controller owns that another player controls. The activated
// ability is Harmless Offering's: both targets are needed, the gift lasts for
// good (CR 611.2a), and a permanent that left and came back is a new object
// that is not handed over (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "514179c0-a50b-4565-8e71-9dad256edd85",
		Name:         "Zedruu the Greathearted",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Zedruu the Greathearted — you gain X life and draw X cards, where X is the number of permanents you own that your opponents control",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					me := ctx.Controller()
					x := 0
					for _, c := range ctx.Game.BattlefieldCardsForEffect() {
						if c.Owner == me && c.Controller != me {
							x++
						}
					}
					if x == 0 {
						return nil
					}
					if err := (effects.GainLife{Player: me, Amount: x}).Apply(ctx); err != nil {
						return err
					}
					return effects.DrawCards{Player: me, N: x}.Apply(ctx)
				}))),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{U}{R}{W}: Target opponent gains control of target permanent you control.",
			Cost:  effects.ManaCost("{U}{R}{W}"),
			Targets: effects.Clauses(
				effects.TargetPlayer("target opponent", effects.Opponent()),
				effects.TargetPermanent("target permanent you control", effects.YouControl()),
			),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				opp, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				gift, ok := ctx.ClauseTarget(1)
				if !ok {
					return nil
				}
				return effects.GainControl{
					Target:     gift.ID,
					Controller: opp.ID,
					Duration:   game.IndefiniteDuration(),
					Label:      "Zedruu the Greathearted — target opponent gains control of it",
				}.Apply(ctx)
			},
		}},
	})
}
