package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hookhand Mariner // Riphook Raider — {3}{G} Creature — Human Werewolf
// 4/4 // Creature — Werewolf 6/4 (#2586, ADR 0132):
//
//	Front: "Daybound"
//	Back:  "This creature can't be blocked by creatures with power 2 or
//	        less.
//	        Nightbound"
//
// The back face's evasion is a pair rule on the creature itself,
// Legolas Greenleaf's: CantBeBlockedBy(OnSelf(), PowerLE(2)). The
// blocker's power is read live when blockers are declared.
//
// No simplification.
func init() {
	const oracle = "52def237-0374-4ab0-8c01-d0d00aaa324f"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Hookhand Mariner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"daybound"},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Riphook Raider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerLE(2), "creatures with power 2 or less"),
		},
	})
}
