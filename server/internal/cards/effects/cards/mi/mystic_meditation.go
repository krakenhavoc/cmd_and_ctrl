package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mystic Meditation — Sorcery {3}{U} (Dragons of Tarkir, S82 old batch 329):
//
//	"Draw three cards. Then discard two cards unless you discard a creature card."
//
// Winternight Stories' draw-then-discard: the discard opens once all three
// draws have landed and takes one card only when it is a creature card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d7c18aa7-4184-4658-9a5a-0b64992348ee",
		Name:         "Mystic Meditation",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.DrawNThenForEffect(item.Controller, 3, game.DrawThen{
				Ref: mysticMeditationDiscard, Player: item.Controller, Source: item.SourceCardID, N: 2,
			})
		},
	})
}

// mysticMeditationDiscard is "then discard two cards unless you discard a
// creature card". Its key is an on-disk identity.
var mysticMeditationDiscard = effects.DiscardTwoUnlessOne("mystic-meditation-discard-two-unless-creature",
	"Mystic Meditation — discard two cards, or one creature card",
	func(c game.Card) bool { return c.IsCreature() })
