package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emberwilde Caliph — Creature — Djinn {2}{U}{R}, 4/4 (Mirage, #5633):
//
//	"Flying, trample
//	 This creature attacks each combat if able.
//	 Whenever this creature deals damage, you lose that much life."
//
// Any damage it deals, combat or not, to anything; the life lost is the amount
// dealt after replacement (CR 120.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8434bdb1-286b-484e-904e-946d074614e7",
		Name:            "Emberwilde Caliph",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Static:          []game.StaticAbility{effects.AttacksEachCombat()},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Amount > 0 && ev.Source == source.InstanceID
			}, "Emberwilde Caliph — you lose that much life", func(g *game.Game, item *game.StackItem) error {
				return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -item.Trigger.Event.Amount)
			}),
		},
	})
}
