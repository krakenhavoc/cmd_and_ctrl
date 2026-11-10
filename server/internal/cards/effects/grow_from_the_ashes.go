package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grow from the Ashes — Sorcery {2}{G} (Foundations, #4465):
//
//	"Kicker {2} (You may pay an additional {2} as you cast this spell.)
//	 Search your library for a basic land card, put it onto the
//	 battlefield, then shuffle. If this spell was kicked, instead search
//	 your library for two basic land cards, put them onto the
//	 battlefield, then shuffle."
//
// One search either way, for one basic land card or, kicked, two (Harrow's
// search). The lands enter untapped, as printed. A search for a stated
// quality may find fewer (CR 701.23b). The purpose declares one land, and
// Purpose.Kicked adds the second on a kicked cast (ADR 0126's amendment of
// 2026-10-15).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "645ab3c7-ee1d-4dd0-811f-2dc7f7c7e792",
		Name:          "Grow from the Ashes",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Lands: 1, LandsUntapped: 1, Kicked: &game.Purpose{Lands: 1, LandsUntapped: 1}},
		OptionalCosts: []game.AdditionalCost{Kicker("{2}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			search := SearchLibrary{
				Player:    item.Controller,
				Predicate: IsBasicLand,
				Dest:      game.ZoneBattlefield,
				Limit:     1,
				Shuffle:   true,
				Reason:    "Grow from the Ashes — a basic land card to put onto the battlefield",
			}
			if ctx.WasKicked() {
				search.Limit = 2
				search.Reason = "Grow from the Ashes — two basic land cards to put onto the battlefield"
			}
			return search.Apply(ctx)
		},
	})
}
