package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Angel of Destiny — Creature — Angel Cleric {3}{W}{W}, 2/6 (Zendikar Rising,
// S82 OLD #5508):
//
//	"Flying, double strike
//	 Whenever a creature you control deals combat damage to a player, you and
//	 that player each gain that much life.
//	 At the beginning of your end step, if you have at least 15 life more
//	 than your starting life total, each player this creature attacked this
//	 turn loses the game."
//
// The life gain is the damage dealt, to the damaged player and to you. The
// end step trigger is an intervening "if" (CR 603.4): the life test is made
// as the step begins and again as it resolves, and the players are the ones
// this very object attacked this turn (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "02f46271-c90d-4318-9004-74e7bee64a89",
		Name:            "Angel of Destiny",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "double strike"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.GapB103CreatureYouControlDealtCombatDamageToAPlayer,
				"Angel of Destiny — you and that player each gain that much life",
				func(g *game.Game, item *game.StackItem) error {
					ev := item.Trigger.Event
					ctx := effects.NewContext(g, item)
					if err := (effects.GainLife{Player: item.Controller, Amount: ev.Amount}).Apply(ctx); err != nil {
						return err
					}
					return effects.GainLife{Player: ev.Target, Amount: ev.Amount}.Apply(ctx)
				}),
			effects.InterveningIfYou("you have at least 15 life more than your starting life total",
				effects.S82OldB091FifteenAboveStartingLife,
				effects.AtYourEndStep("Angel of Destiny — each player this creature attacked this turn loses the game",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						if ctx.IsNewSourceObject(item.SourceCardID) {
							return nil
						}
						for _, victim := range g.AttackedPlayersThisTurn(item.SourceCardID) {
							if err := (effects.LoseTheGame{Player: victim}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					})),
		},
	})
}
