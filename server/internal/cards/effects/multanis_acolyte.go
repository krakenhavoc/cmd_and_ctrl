package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Multani's Acolyte — Creature — Elf, {G}{G}, 2/1:
//
//	"Echo {G}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, draw a card."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d08725e6-ed9a-4c82-89d4-0dd037c376f6",
		Name:         "Multani's Acolyte",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Echo("Multani's Acolyte", "{G}{G}"),
			WhenThisEnters("Multani's Acolyte — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
