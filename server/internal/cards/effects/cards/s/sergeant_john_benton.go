package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sergeant John Benton — Legendary Creature — Human Soldier {1}{G}{W}, 2/4
// (Doctor Who):
//
//	"Trample, haste
//	 Share Intelligence — Whenever Sergeant John Benton deals combat damage to a
//	 player, you and that player each draw that many cards."
//
// "That many" is the combat damage dealt, read off the damage event; you draw first,
// then the damaged player (APNAP order does not matter for a pair of draws).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "987a0b61-2ce4-47bd-8f5e-1e7bf01d0869",
		Name:            "Sergeant John Benton",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Source == source.InstanceID && effects.CombatDamageToPlayerBy(ev, source.Controller, g)
			}, "Share Intelligence — you and that player each draw that many cards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					ev := ctx.Trigger().Event
					n := ev.Amount
					if n <= 0 {
						return nil
					}
					if err := (effects.DrawCards{Player: item.Controller, N: n}).Apply(ctx); err != nil {
						return err
					}
					return effects.DrawCards{Player: ev.Target, N: n}.Apply(ctx)
				}),
		},
	})
}
