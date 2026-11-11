package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Exalted Angel — Creature — Angel {4}{W}{W}, 4/5:
//
//	"Flying
//	 Whenever this creature deals damage, you gain that much life.
//	 Morph {2}{W}{W}"
//
// Phantom Nishoba's reading of "deals damage": each damage event the
// Angel is the source of, with the amount actually dealt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "2f6e9ab8-6951-4c5f-bcde-0e96e0fe87f2",
		Name:             "Exalted Angel",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{Morph("{2}{W}{W}")},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Amount > 0 && ev.Source == source.InstanceID
			},
			Key: "Exalted Angel — you gain that much life",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GainLife{Player: item.Controller, Amount: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
			},
		}},
	})
}
