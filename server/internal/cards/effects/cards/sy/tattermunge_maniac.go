package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tattermunge Maniac — Creature — Goblin Warrior {R/G}, 2/1 (Shadowmoor,
// #5591):
//
//	"This creature attacks each combat if able."
//
// AttacksEachCombat, Amarant Coral's requirement (CR 508.1d): the attack
// declaration is refused while it could attack and stays home.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69ba7864-0d09-4517-ad1a-7efaf437ded0",
		Name:         "Tattermunge Maniac",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AttacksEachCombat()},
	})
}
