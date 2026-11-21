package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "6bb60f1f-9295-41ce-bdcd-3fe67f8e175c",
		Name:         "Yarok's Fenlurker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Yarok's Fenlurker — each opponent exiles a card from their hand", effects.MindleechGhoulExile),
		},
		Activated: []effects.ActivatedAbility{effects.ThisGetsRow("{2}{B}", 1, 1, "Yarok's Fenlurker")},
	})
}
