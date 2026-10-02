package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vug Lizard — Creature — Lizard, {1}{R}{R}, 3/4:
//
//	"Mountainwalk (This creature can't be blocked as long as defending player controls a Mountain.)
//	 Echo {1}{R}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ac9ec838-fbad-4f01-846b-3084e1d37a44",
		Name:            "Vug Lizard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"mountainwalk"},
		Triggered: []game.TriggeredAbility{
			Echo("Vug Lizard", "{1}{R}{R}"),
		},
	})
}
