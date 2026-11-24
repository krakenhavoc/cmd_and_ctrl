package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kezzerdrix — Creature — Rabbit Beast {2}{B}{B}, 4/4 (Secret Lair Drop,
// #5462):
//
//	"First strike
//	 At the beginning of your upkeep, if your opponents control no
//	 creatures, this creature deals 4 damage to you."
//
// An intervening "if" (CR 603.4): asked as the upkeep begins and again as the
// ability resolves, over every creature an opponent controls. The damage is
// dealt by Kezzerdrix as it last existed if it has left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1523671c-0cf4-442c-a5f8-c74ae31016e6",
		Name:            "Kezzerdrix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "your opponents control no creatures",
				Test: func(ctx *effects.Context) bool {
					for _, c := range ctx.Game.BattlefieldCardsForEffect() {
						if c.IsCreature() && c.Controller != ctx.Controller() {
							return false
						}
					}
					return true
				},
			}, effects.AtYourUpkeep("Kezzerdrix — this creature deals 4 damage to you",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 4}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
