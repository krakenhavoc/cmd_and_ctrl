package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Domri's Nodorog — Creature — Beast {3}{R}{G}, 5/2 (Ravnica Allegiance):
//
//	"Trample
//	 When this creature enters, you may search your library and/or graveyard for a
//	 card named Domri, City Smasher, reveal it, and put it into your hand. If you
//	 search your library this way, shuffle."
//
// Ajani's Aid's search: the zones are asked first and the shuffle follows only a
// library search.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3992209a-0a28-449d-8ba5-f75e861a79b4",
		Name:            "Domri's Nodorog",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Domri's Nodorog — search for Domri, City Smasher",
				func(g *game.Game, item *game.StackItem) error {
					return effects.S82OldB105MayFetchNamed("Domri's Nodorog", "Domri, City Smasher")(effects.NewContext(g, item))
				}),
		},
	})
}
