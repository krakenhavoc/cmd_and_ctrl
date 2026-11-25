package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sun-Blessed Mount — Creature — Dinosaur {3}{R}{W}, 4/4 (Ixalan):
//
//	"When this creature enters, you may search your library and/or graveyard
//	 for a card named Huatli, Dinosaur Knight, reveal it, then put it into your
//	 hand. If you search your library this way, shuffle."
//
// Domri's Nodorog's search, for another name: the zones are asked first and the
// shuffle follows only a library search.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "139b0fd5-8aa7-4b36-ab5d-cbf398381c6d",
		Name:         "Sun-Blessed Mount",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sun-Blessed Mount — search for Huatli, Dinosaur Knight",
				func(g *game.Game, item *game.StackItem) error {
					return effects.S82OldB105MayFetchNamed("Sun-Blessed Mount", "Huatli, Dinosaur Knight")(effects.NewContext(g, item))
				}),
		},
	})
}
