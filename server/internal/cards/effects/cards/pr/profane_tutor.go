package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Profane Tutor — Sorcery with no mana cost (Modern Horizons 2):
//
//	"Suspend 2—{1}{B} (Rather than cast this card from your hand, pay {1}{B} and
//	 exile it with two time counters on it. At the beginning of your upkeep,
//	 remove a time counter. When the last is removed, you may cast it without
//	 paying its mana cost.)
//	 Search your library for a card, put that card into your hand, then shuffle."
//
// Suspend is the whole of its playability (no mana cost, CR 118.6). The search is
// Demonic Tutor's: any card, to your hand, then a shuffle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27a1f42c-0b86-4609-9609-1fa9cab7e7c9",
		Name:         "Profane Tutor",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		SpecialActions: []game.SpecialAction{
			effects.Suspend(2, "{1}{B}"),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:  ctx.Controller(),
				Dest:    game.ZoneHand,
				Limit:   1,
				Shuffle: true,
				Reason:  "Profane Tutor — search your library for a card",
			}.Apply(ctx)
		},
	})
}
