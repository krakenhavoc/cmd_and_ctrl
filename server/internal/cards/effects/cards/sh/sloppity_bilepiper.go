package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sloppity Bilepiper — Creature — Demon {3}{B}, 3/3 (Warhammer 40,000
// Commander, #5948):
//
//	"Jolly Gutpipes — {2}, {T}, Sacrifice a creature: The next creature spell
//	 you cast this turn has cascade."
//
// Smoldering Stagecoach's "next spell you cast this turn has cascade" watch,
// narrowed to creature spells, behind Birthing Pod's cost of mana, tap and a
// creature sacrifice. Each activation sets up one watch that cascades once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6bd38092-7788-4bb6-9447-16c56e11e603",
		Name:         "Sloppity Bilepiper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice a creature: The next creature spell you cast this turn has cascade.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeACreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.WhenYouNextCast("Sloppity Bilepiper — cascade: exile until a nonland card with lesser mana value",
					game.CastFilter{Types: []string{"Creature"}}, effects.Cmdr2S41CascadeTheSpellBody).Apply(effects.NewContext(g, item))
			},
		}},
	})
}
