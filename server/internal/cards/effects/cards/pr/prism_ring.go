package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prism Ring — Artifact {1} (Magic Origins, #5549):
//
//	"As this artifact enters, choose a color.
//	 Whenever you cast a spell of the chosen color, you gain 1 life."
//
// Diamond Mare's two abilities on an artifact: the colour is chosen as it
// enters (CR 614.12), and the trigger reads the spell's colours.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "129ee7d6-e3df-40ef-9502-7aad10183031",
		Name:         "Prism Ring",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseColorAsEnters(game.ColorForBenefit, "Prism Ring"),
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastASpellOfTheChosenColor,
				"Prism Ring — you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
