package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Infiltrator il-Kor — Creature — Kor Rogue {4}{U}, 3/1 (Future Sight, #5594):
//
//	"Shadow (This creature can block or be blocked by only creatures
//	 with shadow.)
//	 Suspend 2—{1}{U} (Rather than cast this card from your hand, you
//	 may pay {1}{U} and exile it with two time counters on it. At the
//	 beginning of your upkeep, remove a time counter. When the last is
//	 removed, you may cast it without paying its mana cost. It has
//	 haste.)"
//
// Shadow (CR 702.28) and suspend (CR 702.62a), both the engine's from
// the declarations.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dd491ac7-3c6c-4a47-94af-071d1edc5eb2",
		Name:            "Infiltrator il-Kor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		SpecialActions: []game.SpecialAction{
			Suspend(2, "{1}{U}"),
		},
	})
}
