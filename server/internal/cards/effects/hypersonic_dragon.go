package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hypersonic Dragon — Creature — Dragon {3}{U}{R}, 4/4 (Ravnica: Clue
// Edition, #5929):
//
//	"Flying, haste
//	 You may cast sorcery spells as though they had flash."
//
// Najal's static: a cast-timing rule for sorcery spells.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e1bf20b0-2bd7-4a0f-9a02-41e43bd44641",
		Name:            "Hypersonic Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		CastTimings: []game.CastTimingRule{
			CastKindAsThoughFlash(game.PermissionFilter{SorceryOnly: true}, "You may cast sorcery spells as though they had flash."),
		},
	})
}
