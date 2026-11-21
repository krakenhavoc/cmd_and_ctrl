package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Poet's Quill — Artifact — Equipment {1}{B} (EDHREC rank 17438):
//
//	"When this Equipment enters, learn. (You may reveal a Lesson card
//	 you own from outside the game and put it into your hand, or discard
//	 a card to draw a card.)
//	 Equipped creature gets +1/+1 and has lifelink.
//	 Equip {1}{B}"
//
// The enters trigger learns (CR 701.48a, learn.go); the bonus is an
// attached static like Batterskull's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f474d421-5fd0-4d76-95cb-4824a2381a29",
		Name:         "Poet's Quill",
		Purpose:      game.Purpose{Learn: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Poet's Quill — learn", effects.Do(effects.Learn{})),
		},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("lifelink"),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{1}{B}"),
		},
	})
}
