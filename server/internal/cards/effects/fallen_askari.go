package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fallen Askari — Creature — Human Knight {1}{B}, 2/2 (Visions, #5631):
//
//	"Flanking (Whenever a creature without flanking blocks this creature,
//	 the blocking creature gets -1/-1 until end of turn.)
//	 This creature can't block."
//
// Flanking is the shared trigger in flanking.go (CR 702.25a), one per
// blocker without flanking. "Can't block" is a static restriction on
// itself (RestrictSelf), read by the block check (CR 509.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4956008e-d912-48ba-9b82-176258281060",
		Name:            "Fallen Askari",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordFlanking},
		Triggered:       []game.TriggeredAbility{Flanking("Fallen Askari")},
		Static:          []game.StaticAbility{RestrictSelf(game.CantBlock)},
	})
}
