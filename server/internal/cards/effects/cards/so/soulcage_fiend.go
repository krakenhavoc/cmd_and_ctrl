package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soulcage Fiend — {1}{B}{B} Creature — Demon 3/2 (Jumpstart 2022, #5948):
//
//	"When this creature dies, each player loses 3 life."
//
// Deathcurse Ogre's trigger, every player in turn order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ecbb660a-4013-410a-b683-82c3ef1690cd",
		Name:         "Soulcage Fiend",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Soulcage Fiend — each player loses 3 life", effects.NemEachPlayerLoses(3)),
		},
	})
}
