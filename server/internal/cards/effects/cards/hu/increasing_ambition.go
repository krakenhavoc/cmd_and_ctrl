package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Increasing Ambition — Sorcery {4}{B}:
//
//	"Search your library for a card and put that card into your hand. If
//	 this spell was cast from a graveyard, instead search your library for
//	 two cards and put those cards into your hand. Then shuffle.
//	 Flashback {7}{B}"
func init() {
	effects.Register(effects.Spec{
		OracleID:         "09027657-bc9a-4702-88c1-e5b3534ad188",
		Name:             "Increasing Ambition",
		Purpose:          game.Purpose{Tutors: 1},
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.CostWithPurpose(effects.Flashback("{7}{B}"), game.Purpose{Tutors: 2})},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := 1
			if item.CastFromZone == game.ZoneGraveyard {
				n = 2
			}
			return effects.SearchLibrary{
				Player:  ctx.Controller(),
				Dest:    game.ZoneHand,
				Limit:   n,
				Shuffle: true,
				Reason:  "Increasing Ambition — search your library for a card",
			}.Apply(ctx)
		},
	})
}
