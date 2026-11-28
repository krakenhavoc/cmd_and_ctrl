package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Locust Swarm — Creature — Insect {3}{G}, 1/1 (Mirage, S82 OLD b808):
//
//	"Flying
//	 {G}: Regenerate this creature.
//	 {G}: Untap this creature. Activate only once each turn."
//
// Flying is engine data; the regeneration is the shared "<cost>: Regenerate
// this creature" row. The untap is gated by the per-object once-each-turn
// count of activations (CR 602.5b), which counts one still on the stack, and
// untaps only while the Swarm is still the permanent that was activated.
//
// No simplification.
func init() {
	const untap = "{G}: Untap this creature. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:        "9963b6a9-d715-430e-9b5e-576fb688d944",
		Name:            "Locust Swarm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			effects.RegenerateThisRow("{G}"),
			{
				Label:     untap,
				Purpose:   game.Purpose{Answers: game.AnswerValue},
				Cost:      effects.ManaCost("{G}"),
				Condition: effects.OncePerTurnActivation(untap),
				Effect:    effects.UntapThisPermanentEffect,
			},
		},
	})
}
