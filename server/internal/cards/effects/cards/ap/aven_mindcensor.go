package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aven Mindcensor — Creature — Bird Wizard {2}{W}, 2/1 (S82 GAP2, #5948,
// #6090):
//
//	"Flash
//	 Flying
//	 If an opponent would search a library, that player searches the top
//	 four cards of that library instead."
//
// The static is the engine's Spec.OpponentsSearchTopN
// (game/search_top_n.go), applied in the one search primitive: every
// search an opponent makes, of their own library or of another player's
// (Bribery), looks at the top four cards of the library searched. The search
// still happens, so "whenever a player searches" fires and "then shuffle"
// shuffles the whole library; the graveyard or hand half of an and/or
// search is not a library and is not narrowed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "d9517c5d-66d0-4178-96fb-a8c04f311ad8",
		Name:                "Aven Mindcensor",
		Completeness:        effects.CompletenessFull,
		PrintedKeywords:     []string{"flash", "flying"},
		OpponentsSearchTopN: 4,
	})
}
