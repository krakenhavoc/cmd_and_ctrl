package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Practical Research — Instant {3}{U}{R} (Strixhaven: School of Mages, #5502):
//
//	"Draw four cards. Then discard two cards unless you discard an instant or
//	 sorcery card."
//
// Thirst for Discovery's shape: all four cards are drawn first, then one
// prompt over the whole hand, so a card just drawn can be the one pitched. The
// answer is one instant or sorcery card or any two cards (the registered
// discard-two-unless rule); a hand of one card is discarded whole and an empty
// hand asks nothing (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9110480b-a5d6-4cf8-a72c-dc1a50f11fac",
		Name:         "Practical Research",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 4, Discards: 2},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.DrawNThenForEffect(item.Controller, 4, game.DrawThen{
				Ref: practicalResearchDiscard, Player: item.Controller, Source: item.SourceCardID, N: 2,
			})
		},
	})
}

// practicalResearchDiscard is "then discard two cards unless you discard an
// instant or sorcery card". An on-disk key.
var practicalResearchDiscard = effects.DiscardTwoUnlessOne("old-b043-discard-two-unless-instant-or-sorcery",
	"Practical Research — discard two cards, or one instant or sorcery card",
	func(c game.Card) bool { return c.IsInstant() || c.IsSorcery() })
