package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Search for Glory — Snow Sorcery {2}{W} (Kaldheim):
//
//	"Search your library for a snow permanent card, a legendary card, or a
//	 Saga card, reveal it, put it into your hand, then shuffle. You gain 1
//	 life for each {S} spent to cast this spell."
//
// The life is counted from the snow mana that paid for this spell (CR 107.4h,
// ManaSpent.CountFrom), after the search resolves; it is gained even when the
// search finds nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "fab29bc7-1f0a-4df4-b8f1-c008ce5583d2",
		Name:           "Search for Glory",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsPoolManaOnly,
		WantsManaFrom:  game.ManaSourceSnow,
		Purpose:        game.Purpose{Tutors: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.SearchLibrary{
				Player: item.Controller,
				Predicate: func(c game.Card) bool {
					return (c.IsPermanent() && c.HasSupertype("snow")) || c.IsLegendary() || c.HasSubtype("Saga")
				},
				Dest:    game.ZoneHand,
				Limit:   1,
				Reveal:  true,
				Shuffle: true,
				Reason:  "Search for Glory — a snow permanent card, a legendary card, or a Saga card",
			}).Apply(ctx); err != nil {
				return err
			}
			if n := ctx.ManaSpent().CountFrom(game.ManaSourceSnow); n > 0 {
				return effects.GainLife{Player: item.Controller, Amount: n}.Apply(ctx)
			}
			return nil
		},
	})
}
