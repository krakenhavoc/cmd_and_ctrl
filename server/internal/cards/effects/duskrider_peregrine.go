package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Duskrider Peregrine — Creature — Bird {5}{W}, 3/3 (Time Spiral, #5596):
//
//	"Flying, protection from black
//	 Suspend 3—{1}{W} (Rather than cast this card from your hand, you
//	 may pay {1}{W} and exile it with three time counters on it. At the
//	 beginning of your upkeep, remove a time counter. When the last is
//	 removed, you may cast it without paying its mana cost. It has
//	 haste.)"
//
// A 3/3 flier with protection from black whose only other text is suspend
// (CR 702.62a). The keyword is the engine's: the special action, the time
// counters, the upkeep countdown, the free cast when the last counter
// comes off and the haste it then has are all grown from the one
// declaration.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2fdde5f7-a235-43a4-88a5-2faaf5900bd6",
		Name:            "Duskrider Peregrine",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from black"},
		SpecialActions: []game.SpecialAction{
			Suspend(3, "{1}{W}"),
		},
	})
}
