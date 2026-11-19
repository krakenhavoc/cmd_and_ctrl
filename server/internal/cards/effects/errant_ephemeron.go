package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Errant Ephemeron — Creature — Illusion {6}{U}, 4/4 (Time Spiral, #5596):
//
//	"Flying
//	 Suspend 4—{1}{U} (Rather than cast this card from your hand, you
//	 may pay {1}{U} and exile it with four time counters on it. At the
//	 beginning of your upkeep, remove a time counter. When the last is
//	 removed, you may cast it without paying its mana cost. It has
//	 haste.)"
//
// A 4/4 flier whose only other text is suspend (CR 702.62a). The keyword
// is the engine's: the special action, the time counters, the upkeep
// countdown, the free cast when the last counter comes off and the haste
// it then has are all grown from the one declaration.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c77b89ec-4514-4e62-a508-2012f38a3e15",
		Name:            "Errant Ephemeron",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SpecialActions: []game.SpecialAction{
			Suspend(4, "{1}{U}"),
		},
	})
}
