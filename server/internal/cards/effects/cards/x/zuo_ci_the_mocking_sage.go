package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zuo Ci, the Mocking Sage — Legendary Creature — Human Advisor {1}{G}{G}, 1/2 (Portal Three Kingdoms, S82 OLD b714, #5622):
//
//	"Hexproof
//	Zuo Ci can't be blocked by creatures with horsemanship."
//
// Hexproof is a printed keyword; the block rule is Skyscythe Engulfer's
// (CantBeBlockedBy over a keyword), each blocker's horsemanship read as
// blockers are declared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ae8773a3-05f2-4074-9a53-033b0c127235",
		Name:            "Zuo Ci, the Mocking Sage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.HasKeyword("horsemanship"), "creatures with horsemanship"),
		},
	})
}
