package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loxodon Hierarch — Creature — Elephant Cleric {2}{G}{W}, 4/4 (Ravnica:
// City of Guilds, #5600):
//
//	"When this creature enters, you gain 4 life.
//	 {G}{W}, Sacrifice this creature: Regenerate each creature you
//	 control."
//
// An enters life gain, and Golgari Charm's "regenerate each creature you
// control" on a sacrifice row: the shields go on the creatures its
// controller controls as it resolves (CR 701.19a). The Hierarch is gone
// by then, so it gets none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2df404eb-fbfb-4902-b52b-71225b898dd4",
		Name:         "Loxodon Hierarch",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Loxodon Hierarch — you gain 4 life", effects.Do(effects.GainLife{Amount: 4})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}{W}, Sacrifice this creature: Regenerate each creature you control.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.Plus(effects.ManaCost("{G}{W}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.RegenerateEachCreatureControlledBy(effects.NewContext(g, item), item.Controller)
			},
		}},
	})
}
