package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "21d700e8-0255-418e-96a0-6fe05b3f836d",
		Name:         "Coldsteel Heart",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		AsEnters:     ChooseColorAsEnters("Coldsteel Heart"),
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: ProducedChosenColor(),
			Label:        "Add one mana of the chosen color",
		}},
	})
}
