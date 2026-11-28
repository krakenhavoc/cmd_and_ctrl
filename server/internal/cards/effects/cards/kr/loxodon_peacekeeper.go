package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loxodon Peacekeeper — Creature — Elephant Soldier {1}{W}, 4/4 (Mirrodin,
// #5606):
//
//	"At the beginning of your upkeep, the player with the lowest life total
//	 gains control of this creature. If two or more players are tied for lowest
//	 life total, you choose one of them, and that player gains control of this
//	 creature."
//
// The lowest life total among the players still in the game is found as the
// ability resolves; one player holding it takes the creature, a tie is a
// choice of its controller among the tied players. The control change has no
// duration (CR 611.2a), so it lasts until another effect changes it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "89390fc2-17af-4754-bb1c-7af905d4a2a8",
		Name:         "Loxodon Peacekeeper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Loxodon Peacekeeper — the player with the lowest life total gains control of this creature", effects.OldB606PeacekeeperUpkeep),
		},
	})
}
