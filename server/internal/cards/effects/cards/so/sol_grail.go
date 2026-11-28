package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sol Grail — Artifact {3} (Alliances, S82 OLD batch 819):
//
//	"As this artifact enters, choose a color.
//	 {T}: Add one mana of the chosen color."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0396ef40-3774-4684-9971-160aaccf6ac6",
		Name:         "Sol Grail",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseColorAsEnters(game.ColorForMana, "Sol Grail"),
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			ProducedFunc: effects.ProducedChosenColor(),
			Label:        "Add one mana of the chosen color",
		}},
	})
}
