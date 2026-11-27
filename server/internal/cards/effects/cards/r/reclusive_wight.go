package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reclusive Wight — Creature — Zombie Minion {3}{B}, 4/4 (Urza's Saga, S82
// OLD batch 731, #5624):
//
//	"At the beginning of your upkeep, if you control another nonland
//	 permanent, sacrifice this creature."
//
// The "if" is intervening (CR 603.4), checked as the upkeep begins and again as
// the trigger resolves: another nonland permanent its controller controls, of
// any type, read through the layers (a land that is also a creature is a
// land, so it does not count). Losing the other permanent in response stops
// the sacrifice.
//
// No simplification.
func init() {
	another := func(g *game.Game, controller, self game.Card) bool {
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == controller.Controller && c.InstanceID != self.InstanceID && !c.IsLand() {
				return true
			}
		}
		return false
	}
	effects.Register(effects.Spec{
		OracleID:     "dbde53c1-c6b7-4775-87d3-e95166392e60",
		Name:         "Reclusive Wight",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control another nonland permanent",
				Test: func(ctx *effects.Context) bool {
					src, ok := ctx.Game.LookupCardForEffect(ctx.Source())
					return ok && another(ctx.Game, src, src)
				},
			}, effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.ByYou(ev, source, lki, g) && another(g, *source, *source)
			}, "Reclusive Wight — sacrifice this creature", effects.SacrificeThisIfStillOnBattlefield)),
		},
	})
}
