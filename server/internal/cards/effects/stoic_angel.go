package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stoic Angel — Creature — Angel {1}{G}{W}{U}, 3/4 (Shards of Alara,
// #5588):
//
//	"Flying, vigilance
//	 Players can't untap more than one creature during their untap
//	 steps."
//
// Smoke's untap cap (ADR 0070) on a flier with vigilance: each player's
// untap step offers their tapped creatures and untaps at most one; other
// permanents untap as normal (CR 502.3). Vigilance is why its own
// controller rarely minds.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b50c865b-705f-489e-9ffc-ba8060ba262f",
		Name:            "Stoic Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		UntapCaps: []game.UntapCap{
			cantUntapMoreThan("Stoic Angel — no more than one creature", 1, Creature()),
		},
	})
}
