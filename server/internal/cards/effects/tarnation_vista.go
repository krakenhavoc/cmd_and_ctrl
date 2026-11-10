package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tarnation Vista — Land (The Big Score, #4468):
//
//	"This land enters tapped. As it enters, choose a color.
//	 {T}: Add one mana of the chosen color.
//	 {1}, {T}: For each color among monocolored permanents you control,
//	 add one mana of that color."
//
// Coldsteel Heart's entry (tapped, colour chosen as it enters) and first
// mana ability, and Faeburrow Elder's "for each color among" mana, read
// over monocolored permanents only (CR 105.2a: exactly one color). Both
// are mana abilities (CR 605.1a); the {1} is paid from the pool.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b0be3f25-edb1-4299-959c-9aad909730ca",
		Name:         "Tarnation Vista",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		AsEnters:     ChooseColorAsEnters(game.ColorForMana, "Tarnation Vista"),
		ManaAbilities: []ManaAbility{
			{
				Cost:         ManaAbilityCost{Tap: true},
				ProducedFunc: ProducedChosenColor(),
				Label:        "Add one mana of the chosen color",
			},
			{
				Cost:  ManaAbilityCost{Mana: "{1}", Tap: true},
				Label: "For each color among monocolored permanents you control, add one mana of that color",
				ProducedFunc: func(g *game.Game, controller, _ uuid.UUID) string {
					var sb strings.Builder
					for _, col := range colorsAmongMonocoloredPermanentsYouControl(g, controller) {
						sb.WriteString("{" + col + "}")
					}
					return sb.String()
				},
			},
		},
	})
}
