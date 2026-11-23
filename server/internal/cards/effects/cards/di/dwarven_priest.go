package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dwarven Priest — Creature — Dwarf Cleric {3}{W}, 2/4
// (Core Set 2019, #5527):
//
//	"When this creature enters, you gain 1 life for each creature you control."
//
// Conclave Phalanx's enters trigger: 1 life per creature you control,
// counted on resolution (itself included).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6a077a9b-d725-4b14-957f-1586db2adf59",
		Name:         "Dwarven Priest",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Dwarven Priest — you gain 1 life for each creature you control", effects.Cmdr2S37GainLifePerCreature),
		},
	})
}
