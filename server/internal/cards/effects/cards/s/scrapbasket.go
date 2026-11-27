package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scrapbasket — Artifact Creature — Scarecrow {4}, 3/2 (Shadowmoor, S82 OLD
// b531, #5591):
//
//	"{1}: This creature becomes all colors until end of turn."
//
// Tam's colour-setting effect on the source itself: a layer-5 set of all five
// colours until cleanup (CR 613.1e, 514.2). It adds nothing else, so the
// Scrapbasket stays an artifact creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "779269f9-78c1-4e65-bafe-b6dea40a2b24",
		Name:         "Scrapbasket",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: This creature becomes all colors until end of turn.",
			Cost:    effects.ManaCost("{1}"),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  effects.OldB531ThisBecomesAllColorsUntilEOT("Scrapbasket — becomes all colors until end of turn"),
		}},
	})
}
