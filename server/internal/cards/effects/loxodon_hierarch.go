package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2df404eb-fbfb-4902-b52b-71225b898dd4",
		Name:         "Loxodon Hierarch",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Loxodon Hierarch — you gain 4 life", Do(GainLife{Amount: 4})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{G}{W}, Sacrifice this creature: Regenerate each creature you control.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    Plus(ManaCost("{G}{W}"), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return regenerateEachCreatureControlledBy(NewContext(g, item), item.Controller)
			},
		}},
	})
}
