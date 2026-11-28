package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ancient Kavu (Invasion, S82 OLD batch 645, #5616):
//
//	Creature — Kavu {3}{R}, 3/3:
//	"{2}: This creature becomes colorless until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "598a7347-b772-4839-a02f-b930076c1354",
		Name:         "Ancient Kavu",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: This creature becomes colorless until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.ManaCost("{2}"),
			Effect:  effects.S82OldB645ThisBecomesColorless("Ancient Kavu"),
		}},
	})
}
