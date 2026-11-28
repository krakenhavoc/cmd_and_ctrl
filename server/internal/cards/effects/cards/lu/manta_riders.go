package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Manta Riders — Creature — Merfolk {U}, 1/1 (Tempest, S82 OLD b754, #5628):
//
//	"{U}: This creature gains flying until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d04f1f62-5363-4bf2-bef2-fb3387fc5f48",
		Name:         "Manta Riders",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.ConB1ThisGainsRow("{U}", "flying", "Manta Riders", game.AnswerCombatGrant)},
	})
}
