package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glimmering Angel (Invasion, S82 OLD batch 646, #5616):
//
//	Creature — Angel {3}{W}, 2/2:
//	"Flying
//	{U}: This creature gains shroud until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9629ab26-f231-481d-8392-e2450ca02b67",
		Name:            "Glimmering Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.ConB1ThisGainsRow("{U}", "shroud", "Glimmering Angel", game.AnswerProtect)},
	})
}
