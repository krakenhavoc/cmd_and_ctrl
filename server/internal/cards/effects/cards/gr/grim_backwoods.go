package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "5effaa94-7f87-4485-8959-473d584c5034",
		Name:         "Grim Backwoods",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}{G}, {T}, Sacrifice a creature: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}{G}"), effects.TapCost(), effects.SacrificeACreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
