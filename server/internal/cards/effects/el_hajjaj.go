package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// El-Hajjâj — Creature — Human Wizard {1}{B}{B}, 1/1 (Arabian Nights,
// #5641):
//
//	"Whenever this creature deals damage, you gain that much life."
//
// Exalted Angel's reading of "deals damage": each damage event El-Hajjâj
// is the source of, combat or not, with the amount actually dealt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f92c9a5d-853f-4157-89ba-8d8c8033c533",
		Name:         "El-Hajjâj",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Amount > 0 && ev.Source == source.InstanceID
			},
			Key: "El-Hajjâj — you gain that much life",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GainLife{Player: item.Controller, Amount: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
			},
		}},
	})
}
