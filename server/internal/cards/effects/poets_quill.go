package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f474d421-5fd0-4d76-95cb-4824a2381a29",
		Name:         "Poet's Quill",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Poet's Quill — learn", Do(Learn{})),
		},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("lifelink"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{1}{B}"),
		},
	})
}
