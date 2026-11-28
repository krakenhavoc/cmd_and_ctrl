package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Taniwha — Legendary Creature — Serpent {3}{U}{U}, 7/7 (Mirage, S82 OLD batch
// 812, #5633):
//
//	"Trample
//	 Phasing (This phases in or out before you untap during each of your untap
//	 steps. While it's phased out, it's treated as though it doesn't exist.)
//	 At the beginning of your upkeep, all lands you control phase out. (They
//	 phase in before you untap during your next untap step.)"
//
// Phasing is the engine's own printed keyword (CR 702.26a). The upkeep trigger
// phases out every land its controller controls as it resolves, together
// (CR 702.26g); they come back in that player's next untap step (CR 502.1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "371c993a-69b8-4a51-af89-982b79940b2d",
		Name:            "Taniwha",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", game.KeywordPhasing},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Taniwha — all lands you control phase out", effects.S82OldB812LandsYouControlPhaseOut),
		},
	})
}
