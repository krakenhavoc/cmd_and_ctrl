package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heraldic Banner — Artifact {3}:
//
//	"As this artifact enters, choose a color.
//	 Creatures you control of the chosen color get +1/+0.
//	 {T}: Add one mana of the chosen color."
//
// Coldsteel Heart with an anthem and without the tapped entry. The
// anthem is a layer 7c static whose AppliesTo reads the colour stored
// on the Banner (ChosenColorAnthem): a creature's colour is its
// effective colour, so a creature that is two colours including the
// chosen one gets the bonus, and an unchosen colour pumps nothing.
// Resolving the prompt bumps the layer version, so the +1/+0 appears
// the moment the answer lands.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3525e263-e29a-49bf-a29f-fb3ce43bbd33",
		Name:         "Heraldic Banner",
		Completeness: CompletenessFull,
		AsEnters:     ChooseColorAsEnters(game.ColorForBenefit, "Heraldic Banner"),
		Static:       []game.StaticAbility{ChosenColorAnthem(1, 0)},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: ProducedChosenColor(),
			Label:        "Add one mana of the chosen color",
		}},
	})
}
