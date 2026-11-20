package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clarion Cathars — Creature — Human Knight {3}{W}, 3/3 (Innistrad: Midnight Hunt, S82):
//
//	"When this creature enters, create a 1/1 white Human creature token."
//
// A stack trigger that makes one Human token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "193f9b12-ee34-40c2-9315-6794416a4f62",
		Name:         "Clarion Cathars",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Clarion Cathars — create a 1/1 white Human creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human"), N: 1})),
		},
	})
}
