package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silhana Starfletcher — Creature — Elf Druid Archer {2}{G}, 1/3
// (Guildpact, #5599):
//
//	"Reach
//	 As this creature enters, choose a color.
//	 {T}: Add one mana of the chosen color."
//
// Coldsteel Heart's colour choice and mana ability on a reach creature:
// the colour is chosen as it enters (CR 614.12) and the ability adds one
// mana of it. Until a colour is chosen it adds nothing, the weaker
// reading. {T} needs it to have been under its controller's control
// since their most recent turn began (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3b4b6af2-f929-4082-a5ae-61a4e074bf14",
		Name:            "Silhana Starfletcher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		AsEnters:        ChooseColorAsEnters(game.ColorForMana, "Silhana Starfletcher"),
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: ProducedChosenColor(),
			Label:        "Add one mana of the chosen color",
		}},
	})
}
