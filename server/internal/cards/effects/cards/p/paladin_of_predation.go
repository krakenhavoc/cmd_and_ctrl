package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Paladin of Predation — Creature — Phyrexian Knight {5}{G}{G}, 6/7
// (Phyrexia: All Will Be One, #5479):
//
//	"Toxic 6
//	 This creature can't be blocked by creatures with power 2 or less."
//
// Toxic is the engine's damage tail (CR 702.164); the evasion is a block
// rule on itself (Legolas Greenleaf's shape), reading each blocker's
// power live as the block is declared (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fb5a02ce-6080-47f9-815a-fef3758d753b",
		Name:            "Paladin of Predation",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"toxic 6"},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerLE(2), "creatures with power 2 or less"),
		},
	})
}
