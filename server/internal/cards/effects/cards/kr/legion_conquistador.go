package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Legion Conquistador — Creature — Vampire Soldier {2}{W}, 2/2:
//
//	"When this creature enters, you may search your library for any number of
//	 cards named Legion Conquistador, reveal them, put them into your hand,
//	 then shuffle."
//
// A search for a stated name may find any number of the cards (CR 701.23b);
// the controller chooses how many, none included, and the library is shuffled
// after a yes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b20671e6-d740-4ef6-b408-bf8a0465c0da",
		Name:         "Legion Conquistador",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Legion Conquistador — search for cards named Legion Conquistador",
				effects.S82OldB216SearchForAllNamed("Legion Conquistador")),
		},
	})
}
