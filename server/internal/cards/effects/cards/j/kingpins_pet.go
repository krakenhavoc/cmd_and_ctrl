package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kingpin's Pet — Creature — Thrull {1}{W}{B}, 2/2 (Gatecrash, #5565):
//
//	"Flying
//	 Extort (Whenever you cast a spell, you may pay {W/B}. If you do, each
//	 opponent loses 1 life and you gain that much life.)"
//
// Flying is the engine keyword; extort is the shared Extort trigger
// (Basilica Screecher's).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1c4561d1-377c-4a6a-b7a2-8ce81efa39c6",
		Name:            "Kingpin's Pet",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{effects.Extort("Kingpin's Pet")},
	})
}
