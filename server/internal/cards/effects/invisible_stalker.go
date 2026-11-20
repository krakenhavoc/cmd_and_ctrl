package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invisible Stalker — Creature — Human Rogue {1}{U}, 1/1 (Innistrad, #5572):
//
//	"Hexproof (This creature can't be the target of spells or abilities
//	 your opponents control.)
//	 This creature can't be blocked."
//
// Hexproof is an engine keyword carried as printed data (CR 702.11b),
// and "can't be blocked" is Talas Warrior's RestrictSelf(CantBeBlocked).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1a137a67-4156-45c8-aae9-838f688e12bd",
		Name:            "Invisible Stalker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Static:          []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
	})
}
