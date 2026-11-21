package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alms of the Vein — Sorcery {2}{B} (Shadows over Innistrad, #5543):
//
//	"Target opponent loses 3 life and you gain 3 life.
//	 Madness {B} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your graveyard.)"
//
// Brush with Death's drain at three (targetPlayerLosesLifeYouGain), aimed
// at an opponent. Madness is grown by the engine from the one string (CR
// 702.35a, #657).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2013a774-e6ab-4f2d-a84f-fd277a637d7e",
		Name:         "Alms of the Vein",
		Completeness: effects.CompletenessFull,
		Madness:      "{B}",
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve:    effects.TargetPlayerLosesLifeYouGain(3),
	})
}
