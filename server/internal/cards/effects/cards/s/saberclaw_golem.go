package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Saberclaw Golem — Artifact Creature — Golem {5}, 4/2 (Scars of Mirrodin, #5577):
//
//	"{R}: This creature gains first strike until end of turn."
//
// Noble Panther's ability at {R}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4989d38e-332f-48f5-863b-1566921156c8",
		Name:         "Saberclaw Golem",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: This creature gains first strike until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"first strike"}}},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("first strike", "Saberclaw Golem — first strike until end of turn"),
		}},
	})
}
