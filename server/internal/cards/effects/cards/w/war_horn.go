package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// War Horn — Artifact {3} (Magic Origins, #5549):
//
//	"Attacking creatures you control get +1/+0."
//
// Goblin Oriflamme's static.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "99bc739e-c745-47d8-8401-2f34b6becd85",
		Name:         "War Horn",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AttackingCreaturesYouControlGetPower(1)},
	})
}
