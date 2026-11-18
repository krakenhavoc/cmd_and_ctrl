package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clarion Cathars — Creature — Human Knight {3}{W}, 3/3 (Innistrad: Midnight Hunt, S82):
//
//	"When this creature enters, create a 1/1 white Human creature token."
//
// A stack trigger that makes one Human token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "193f9b12-ee34-40c2-9315-6794416a4f62",
		Name:         "Clarion Cathars",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Clarion Cathars — create a 1/1 white Human creature token",
				Do(CreateToken{Template: TokenCard("1/1 white Human"), N: 1})),
		},
	})
}
