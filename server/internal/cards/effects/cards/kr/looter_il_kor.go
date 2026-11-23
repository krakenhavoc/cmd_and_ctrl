package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Looter il-Kor — "Shadow (This creature can block or be blocked by only
// creatures with shadow.) Whenever this creature deals damage to an opponent,
// draw a card, then discard a card."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c87e3b1d-2d24-4f1c-84e8-1c21a347cdc2",
		Name:            "Looter il-Kor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsDamageToAnOpponent("Looter il-Kor — draw, then discard", func(g *game.Game, item *game.StackItem) error {
				return effects.B16DrawThenDiscard(g, item, 1)
			}),
		},
	})
}
