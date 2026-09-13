package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Adventurer's Inn — Land — Town (EDHREC rank 1551):
//
//	"When this land enters, you gain 2 life.
//	 {T}: Add {C}."
//
// Radiant Fountain with a Town subtype, which the printed type line
// carries for anything that cares. The life is a real ETB trigger
// with a response window, as printed; the land enters untapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "232bd88c-ecdb-43dd-b34a-d381cb3bedf2",
		Name:         "Adventurer's Inn",
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
				return game.NewTriggeredItem(source, "Adventurer's Inn — you gain 2 life",
					func(g *game.Game, item *game.StackItem) error {
						return GainLife{Player: item.Controller, Amount: 2}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
