package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dwarven Forge-Chanter — Creature — Dwarf Wizard {1}{R}, 1/3 (The Brothers' War, #5482):
//
//	"Ward—Pay 2 life.
//	 Prowess"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e69a8970-3ead-479d-860a-26ed65f9c846",
		Name:            "Dwarven Forge-Chanter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardLife(2), "Dwarven Forge-Chanter — ward—pay 2 life"),
		},
	})
}
