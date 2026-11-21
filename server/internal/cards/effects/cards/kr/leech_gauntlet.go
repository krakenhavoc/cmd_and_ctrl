package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "1aaddaeb-baab-42b6-a442-d97339da54db",
		Name:            "Leech Gauntlet",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Static:          []game.StaticAbility{effects.GrantToAttached("lifelink")},
		Activated:       effects.Reconfigure("{4}"),
	})
}
