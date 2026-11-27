package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Beetleform Mage — Creature — Human Insect Wizard {1}{G}{U}, 2/2 (Dragon's Maze,
// S82 OLD batch 404, #5564):
//
//	"{G}{U}: This creature gets +2/+2 and gains flying until end of turn. Activate
//	 only once each turn."
//
// Rootwalla's once-per-turn pump (CR 602.5b gates the offer) with a layer-6 flying
// grant beside the layer-7 modification, both ending at cleanup (CR 514.2).
//
// No simplification.
func init() {
	const label = "{G}{U}: This creature gets +2/+2 and gains flying until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "227cbba1-700b-4255-b730-b324ab3dec34",
		Name:         "Beetleform Mage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{G}{U}"),
			Condition: effects.OncePerTurnActivation(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.ThisGetsUntilEndOfTurn(2, 2, "Beetleform Mage — +2/+2 until end of turn")(g, item); err != nil {
					return err
				}
				return effects.ThisGainsKeywordUntilEndOfTurn("flying", "Beetleform Mage — flying until end of turn")(g, item)
			},
		}},
	})
}
