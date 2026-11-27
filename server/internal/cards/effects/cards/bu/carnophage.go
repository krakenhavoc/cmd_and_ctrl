package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Carnophage — Creature — Zombie {B}, 2/2 (Exodus, S82 OLD batch 743, #5626):
//
//	"At the beginning of your upkeep, tap this creature unless you pay 1
//	 life."
//
// Morgul-Knife Wound's pay-or-else with the penalty a tap: the controller
// chooses to pay 1 life or tap it; a player who can't pay (CR 119.4) taps it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ac77cfd-7d78-4c8a-bffc-17649a852114",
		Name:         "Carnophage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Carnophage — tap this creature unless you pay 1 life", effects.S82OldB743CarnophageUpkeep),
		},
	})
}
