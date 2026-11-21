package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Call for Unity — Enchantment {3}{W}{W} (Tales of Middle-earth Commander):
//
//	"Revolt — At the beginning of your end step, if a permanent left the
//	 battlefield under your control this turn, put a unity counter on this
//	 enchantment.
//	 Creatures you control get +1/+1 for each unity counter on this
//	 enchantment."
//
// Revolt is the shared end-step trigger (an intervening "if", checked as the
// step begins and again on resolution, CR 603.4); the static is a layer-7c
// modification read off the counters on the enchantment as the layers are
// recomputed (CR 613.4c), so it follows the count as it changes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "288d5415-ef44-400f-ab88-7c8a4deaaf4b",
		Name:         "Call for Unity",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStepIfRevolt("Call for Unity — put a unity counter on this enchantment (revolt)",
				effects.Do(effects.CounterOnThis{Kind: "unity", N: 1})),
		},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			Label:    "Creatures you control get +1/+1 for each unity counter on this enchantment.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && source.Counters["unity"] > 0
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				n := source.Counters["unity"]
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}
