package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thunderherd Migration — Sorcery {1}{G}:
//
//	"As an additional cost to cast this spell, reveal a Dinosaur card from your hand or pay {1}.
//	 Search your library for a basic land card, put it onto the battlefield tapped, then shuffle."
//
// The additional cost is the either/or branch cost of ADR 0100 §2
// with the reveal branch added by its 2026-10-07 amendment
// (RevealOrPay): the caster announces the branch, names the
// card on reveal_ids, and either shows it to the table or pays {1}
// more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "006bbd8b-2007-419d-8b84-9ff73f2f95b7",
		Name:           "Thunderherd Migration",
		Completeness:   CompletenessFull,
		AdditionalCost: RevealOrPay("a", "Dinosaur", "{1}"),
		Purpose:        game.Purpose{Lands: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Reveal:        true,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Thunderherd Migration — a basic land",
			}.Apply(ctx)
		},
	})
}
