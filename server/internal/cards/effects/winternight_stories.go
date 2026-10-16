package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Winternight Stories — Sorcery {2}{U} (#3931):
//
//	"Draw three cards. Then discard two cards unless you discard a
//	 creature card.
//	 Harmonize {4}{U}"
//
// Thirst for Knowledge's draw-then-discard with a creature card for the
// artifact (discardTwoUnlessOne): the discard opens once all three draws
// have landed, takes one card only when it is a creature card, and
// otherwise two. Harmonize is CR 702.180.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "d64687e4-ec37-4db2-9e60-4626c7a3b011",
		Name:             "Winternight Stories",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Harmonize("{4}{U}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ctx.Game.DrawNThenForEffect(item.Controller, 3, game.DrawThen{
				Ref: winternightStoriesDiscard, Player: item.Controller, Source: item.SourceCardID, N: 2,
			})
		},
	})
}

// winternightStoriesDiscard is "then discard two cards unless you
// discard a creature card". Its key is an on-disk identity.
var winternightStoriesDiscard = discardTwoUnlessOne("discard-two-unless-creature",
	"Winternight Stories — discard two cards, or one creature card",
	func(c game.Card) bool { return c.IsCreature() })
