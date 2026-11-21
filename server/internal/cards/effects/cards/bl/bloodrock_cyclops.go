package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodrock Cyclops — Creature — Cyclops {2}{R}, 3/3 (Weatherlight, #5629):
//
//	"This creature attacks each combat if able."
//
// The shared attack requirement (attack_requirements.go), checked as
// attackers are declared (CR 508.1d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "06a3fe24-7e1f-4f98-ba2f-dbb5d38f4b9d",
		Name:         "Bloodrock Cyclops",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AttacksEachCombat()},
	})
}
