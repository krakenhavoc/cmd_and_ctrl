package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Psychic Corrosion — Enchantment {2}{U} (EDHREC rank 1782):
//
//	"Whenever you draw a card, each opponent mills two cards."
//
// The mill deck's engine. EventDrawCard fires once per card with the
// drawer in Actor, so a draw-three mills each opponent six — as
// printed, since the printed ability triggers per card too. The
// controller's own draw-step draw counts. The mill body is Altar of
// the Brood's (b12EachOpponentMills).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "328b42f1-d679-4f9c-80e3-38fe3b965d10",
		Name:         "Psychic Corrosion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouDraw("Psychic Corrosion — each opponent mills two cards", func(g *game.Game, item *game.StackItem) error {
				return effects.B12EachOpponentMills(g, item, 2)
			}),
		},
	})
}
