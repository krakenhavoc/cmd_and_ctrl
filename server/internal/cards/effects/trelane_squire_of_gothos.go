package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trelane, Squire of Gothos — Legendary Creature — Q {5}{U}, 4/5 (Star
// Trek Commander, #4974):
//
//	"Vigilance
//	 Ward {2}
//	 Trelane can't be blocked."
//
// Vigilance is the printed keyword, ward {2} the shared ward trigger
// (CR 702.21a), and "can't be blocked" the self restriction every
// unblockable creature carries.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3c516327-00a3-4d03-87a3-73b4a90719d3",
		Name:            "Trelane, Squire of Gothos",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static:          []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Trelane, Squire of Gothos — ward {2}"),
		},
	})
}
