package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Defender of Law — Creature — Human Knight {2}{W}, 2/1 (Urza's Legacy,
// #5623):
//
//	"Flash
//	 Protection from red"
//
// Both are printed keywords: flash lets it be cast any time an instant could
// be (CR 702.8a), and protection from red is read by targeting, blocking and
// damage (CR 702.16a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "656513cf-1bf0-431e-8d3a-2d1ec79d0bec",
		Name:            "Defender of Law",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "protection from red"},
	})
}
