package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Uglúk of the White Hand — Legendary Creature — Orc Soldier {2}{B}{R}, 3/3
// (The Lord of the Rings: Tales of Middle-earth):
//
//	"Whenever another creature you control dies, put a +1/+1 counter on
//	 Uglúk. If that creature was a Goblin or Orc, put two +1/+1 counters on
//	 Uglúk instead."
//
// A dies trigger over another creature the controller controlled as it left
// (diedCreature, leftUnderControlOf, CR 603.10a). Whether it was a Goblin or
// Orc is read off the card as it last existed, through the source's text
// (CR 612.2), as the ability resolves; two counters replace the one, so
// there is one counter placement. A source that has left the battlefield
// puts no counters (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ebd6a3a0-ae8d-44e6-8d0b-048da6dd1f03",
		Name:         "Uglúk of the White Hand",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && dead.InstanceID != source.InstanceID && effects.LeftUnderControlOf(ev, dead) == source.Controller
			}, "Uglúk of the White Hand — put a +1/+1 counter on Uglúk, or two if that creature was a Goblin or Orc",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					n := 1
					if dead, ok := g.LookupCardForEffect(item.Trigger.Event.CardID); ok &&
						(dead.HasSubtype(ctx.TextSubtype("Goblin")) || dead.HasSubtype(ctx.TextSubtype("Orc"))) {
						n = 2
					}
					if !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
				}),
		},
	})
}
