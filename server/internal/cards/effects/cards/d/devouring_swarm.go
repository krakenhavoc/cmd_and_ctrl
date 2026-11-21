package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Devouring Swarm — Creature — Insect {1}{B}{B}, 2/1:
//
//	"Flying
//	 Sacrifice a creature: This creature gets +1/+1 until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "11656b2a-b459-4b1c-b505-8547c3507856",
		Name:            "Devouring Swarm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a creature: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost:    effects.SacrificeACreature(),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Devouring Swarm — +1/+1 until end of turn"),
		}},
	})
}
