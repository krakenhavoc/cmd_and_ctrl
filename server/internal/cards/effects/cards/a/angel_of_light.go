package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angel of Light — Creature — Angel {4}{W}, 3/3
// (Starter 1999, #5620):
//
//	"Flying, vigilance"
//
// Serra Angel's two keywords, carried as printed data: it can be
// blocked only by creatures with flying or reach (CR 702.9b), and
// attacking doesn't tap it (CR 702.20b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d12cf640-ff60-4935-a2c7-28ab788de42c",
		Name:            "Angel of Light",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
}
