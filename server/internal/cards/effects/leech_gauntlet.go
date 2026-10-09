package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leech Gauntlet — Artifact Creature — Equipment Leech {1}{B}, 2/2:
//
//	"Lifelink
//	 Equipped creature has lifelink.
//	 Reconfigure {4}"
//
// Reconfigure is reconfigure.go (#2639).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1aaddaeb-baab-42b6-a442-d97339da54db",
		Name:            "Leech Gauntlet",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Static:          []game.StaticAbility{GrantToAttached("lifelink")},
		Activated:       Reconfigure("{4}"),
	})
}
