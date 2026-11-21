package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loxodon Surveyor — Creature — Elephant Scout {2}{G}, 3/3:
//
//	"Start your engines!
//	 Max speed — {3}, Exile this card from your graveyard: Draw a card."
//
// ADR 0138 (#2711). The graveyard ability is the shared Surveyor body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "970f26d4-c916-444a-a2d2-4c1f5bb5556b",
		Name:            "Loxodon Surveyor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{effects.StartYourEngines},
		Activated:       []effects.ActivatedAbility{effects.MaxSpeedSurveyorDraw()},
	})
}
