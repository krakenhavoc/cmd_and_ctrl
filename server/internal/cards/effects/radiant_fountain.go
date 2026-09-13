package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Radiant Fountain — Land (EDHREC rank 1573):
//
//	"When this land enters, you gain 2 life.
//	 {T}: Add {C}."
//
// The life is a real ETB trigger with a response window, as printed
// — the gain-land shape with the number doubled and no colour. The
// land enters untapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6db442e5-fbcc-4456-a4c5-bea1aee3fc8e",
		Name:         "Radiant Fountain",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Radiant Fountain — you gain 2 life",
					func(g *game.Game, item *game.StackItem) error {
						return GainLife{Player: item.Controller, Amount: 2}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
