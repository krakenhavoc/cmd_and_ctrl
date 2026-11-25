package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghoulsteed — Creature — Zombie Horse {4}{B}, 4/4 (Shadows over Innistrad,
// #5543):
//
//	"{2}{B}, Discard two cards: Return this card from your graveyard to the
//	 battlefield tapped."
//
// Stitchwing Skaab's graveyard activation without the flying: the discard
// is a cost, and the card returns tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dbc34afa-16bc-41f2-b74d-29bf2bfe24a9",
		Name:         "Ghoulsteed",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.S82OldB285ReturnThisDiscardingTwo("{2}{B}"),
		},
	})
}
