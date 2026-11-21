package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vagrant Plowbeasts — Creature — Beast {5}{G}{W}, 6/6 (Conflux, #5587):
//
//	"{1}: Regenerate target creature with power 5 or greater."
//
// Asceticism's regeneration row (CR 701.19a) behind Spearbreaker
// Behemoth's power-5 target: the power is read as the ability is
// activated and again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e97870a3-855a-4701-8344-91a90095f70a",
		Name:         "Vagrant Plowbeasts",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: Regenerate target creature with power 5 or greater.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{1}"),
			Targets: effects.TargetCreature("target creature with power 5 or greater", effects.PowerGE(5)),
			Effect:  effects.RegenerateTheTargetPermanent,
		}},
	})
}
