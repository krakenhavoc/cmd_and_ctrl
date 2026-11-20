package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yarok's Fenlurker — Creature — Horror {B}{B}, 1/1 (Core Set 2020, #5518):
//
//	"When this creature enters, each opponent exiles a card from their hand.
//	 {2}{B}: This creature gets +1/+1 until end of turn."
//
// Mindleech Ghoul's exile as an enters trigger: each opponent picks a
// card from their own hand and it is exiled face up; an empty hand exiles
// nothing. The pump is the shared self-pump row, activatable any number
// of times.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6bb60f1f-9295-41ce-bdcd-3fe67f8e175c",
		Name:         "Yarok's Fenlurker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Yarok's Fenlurker — each opponent exiles a card from their hand", mindleechGhoulExile),
		},
		Activated: []ActivatedAbility{thisGetsRow("{2}{B}", 1, 1, "Yarok's Fenlurker")},
	})
}
