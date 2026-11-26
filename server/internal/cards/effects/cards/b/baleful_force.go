package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Baleful Force — Creature — Elemental {5}{B}{B}{B}, 7/7 (Commander 2013, S82
// OLD b380, #5560):
//
//	"At the beginning of each upkeep, you draw a card and you lose 1 life."
//
// Oath of Druids' every-upkeep trigger (AtEachUpkeep, anyone's), with
// Black Widow's two-instruction body: the draw, then the life loss.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f26fa9e4-f528-4e22-8f7c-329fc3b633e4",
		Name:         "Baleful Force",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Baleful Force — you draw a card and you lose 1 life",
				effects.Do(effects.DrawCards{N: 1}, effects.GainLife{Amount: -1})),
		},
	})
}
