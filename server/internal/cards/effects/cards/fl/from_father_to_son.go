package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// From Father to Son — Sorcery {1}{W} (Final Fantasy, #4462):
//
//	"Search your library for a Vehicle card, reveal it, and put it into
//	 your hand. If this spell was cast from a graveyard, put that card
//	 onto the battlefield instead. Then shuffle.
//	 Flashback {4}{W}{W}{W}"
//
// Anchor to Reality's search with the destination read from the spell's
// announce record (CastFromZone): a graveyard cast (flashback or any
// other) puts the Vehicle onto the battlefield and doesn't reveal it;
// a hand cast reveals it into the hand. The shuffle follows either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "87864421-505c-4bdb-b1e3-3e44e05f2beb",
		Name:             "From Father to Son",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Tutors: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{4}{W}{W}{W}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			s := effects.SearchLibrary{
				Player:    item.Controller,
				Predicate: func(c game.Card) bool { return c.HasSubtype("Vehicle") },
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "From Father to Son — a Vehicle card",
			}
			if item.CastFromZone == game.ZoneGraveyard {
				s.Dest = game.ZoneBattlefield
				s.Reveal = false
			}
			return s.Apply(ctx)
		},
	})
}
