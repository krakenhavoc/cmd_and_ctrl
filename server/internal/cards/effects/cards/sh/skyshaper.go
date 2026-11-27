package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skyshaper — Artifact {2} (Exodus, S82 OLD b737, #5626):
//
//	"Sacrifice this artifact: Creatures you control gain flying until end of
//	 turn."
//
// Frostwind Invoker's team flying, paid by sacrificing the artifact.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b5642460-883c-4e38-8552-9bea4e50d8e8",
		Name:         "Skyshaper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this artifact: Creatures you control gain flying until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Keywords: []string{"flying"}}, Answers: game.AnswerCombatGrant},
			Cost:    effects.SacrificeThis(),
			Effect: effects.Do(effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"flying"},
				Label:    "Skyshaper — creatures you control gain flying",
			}),
		}},
	})
}
