package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sighted-Caste Sorcerer — Creature — Human Wizard Sorcerer {1}{W}, 1/1
// (Shards of Alara, S82 OLD batch 511, #5588):
//
//	"Exalted
//	 {U}: This creature gains shroud until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f9921c9e-3248-488d-ad7b-8b064c5387e7",
		Name:            "Sighted-Caste Sorcerer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"exalted"},
		Activated:       []effects.ActivatedAbility{effects.ConB1ThisGainsRow("{U}", "shroud", "Sighted-Caste Sorcerer", game.AnswerProtect)},
	})
}
