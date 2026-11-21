package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tempered in Solitude — Enchantment {1}{R} (Kamigawa: Neon Dynasty):
//
//	"Whenever a creature you control attacks alone, exile the top card of
//	 your library. You may play that card this turn."
//
// Agent 13's lone-attack trigger with the impulse exile (Breeches' bullet):
// play, not cast, so a land may be played from exile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f4aedfd3-4414-493c-bbba-75db301c0737",
		Name:         "Tempered in Solitude",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.MshB1YouAttackAlone,
				"Tempered in Solitude — exile the top card of your library; you may play it this turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ExileTopCardYouMayPlayThisTurn(item, effects.NewContext(g, item), 0)
				}),
		},
	})
}
