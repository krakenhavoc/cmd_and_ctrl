package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grim Backwoods — Land (EDHREC rank 2067):
//
//	"{T}: Add {C}.
//	 {2}{B}{G}, {T}, Sacrifice a creature: Draw a card."
//
// A colourless land that is also a sacrifice outlet — the Golgari
// deck's way to turn a creature about to die into a card. The mana
// ability is Buried Ruin's; the draw is a CR 602 activation with
// three cost components, the creature paid at announce so its
// dies-triggers land above the ability and resolve first.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5effaa94-7f87-4485-8959-473d584c5034",
		Name:         "Grim Backwoods",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{2}{B}{G}, {T}, Sacrifice a creature: Draw a card.",
			Cost:  Plus(ManaCost("{2}{B}{G}"), TapCost(), SacrificeACreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
