package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Celestial Force — Creature — Elemental {5}{W}{W}{W}, 7/7 (Commander 2011,
// #5574):
//
//	"At the beginning of each upkeep, you gain 3 life."
//
// "Each upkeep" is every player's upkeep step, so its controller gains the 3
// life four times a turn cycle in a four-player game, and the life is the
// Force's controller's whoever's upkeep it is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8add6996-6c5c-4aa2-93d4-a39efcfd22b2",
		Name:         "Celestial Force",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Celestial Force — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
}
