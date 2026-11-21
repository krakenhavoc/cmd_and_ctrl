package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coldsteel Heart — Snow Artifact {2}:
//
//	"This artifact enters tapped.
//	 As this artifact enters, choose a color.
//	 {T}: Add one mana of the chosen color."
//
// The two-mana rock that fixes for exactly one colour. A tapped entry
// (SelfEntersTapped), the #742 colour prompt on AsEnters, and a mana
// ability that reads the stored colour at activation. Until the
// controller answers, the ability produces nothing — the weaker
// direction, and unobservable because the open prompt holds priority.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21d700e8-0255-418e-96a0-6fe05b3f836d",
		Name:         "Coldsteel Heart",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		AsEnters:     effects.ChooseColorAsEnters(game.ColorForMana, "Coldsteel Heart"),
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			ProducedFunc: effects.ProducedChosenColor(),
			Label:        "Add one mana of the chosen color",
		}},
	})
}
