package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glitch Ghost Surveyor — Creature — Spirit Scout {2}{U}, 2/2:
//
//	"Flying
//	 Start your engines!
//	 Max speed — {3}, Exile this card from your graveyard: Draw a card."
//
// ADR 0138 (#2711). The graveyard ability is the shared Surveyor body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f0f83ab8-6185-4139-b41e-a11500036915",
		Name:            "Glitch Ghost Surveyor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", effects.StartYourEngines},
		Activated:       []effects.ActivatedAbility{effects.MaxSpeedSurveyorDraw()},
	})
}
