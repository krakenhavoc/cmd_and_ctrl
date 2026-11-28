package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Breezekeeper — Creature — Djinn {3}{U}, 4/4 (Visions, S82 older sets batch
// 791, #5631):
//
//	"Flying
//	 Phasing (This phases in or out before you untap during each of your untap
//	 steps. While it's phased out, it's treated as though it doesn't exist.)"
//
// Both are printed keywords, as on Teferi's Drake: phasing is the engine's own
// (CR 702.26a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "138439ad-76c6-4d75-8025-e68b94f68fa3",
		Name:            "Breezekeeper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordPhasing},
	})
}
