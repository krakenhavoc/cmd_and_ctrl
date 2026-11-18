package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f4aedfd3-4414-493c-bbba-75db301c0737",
		Name:         "Tempered in Solitude",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, mshB1YouAttackAlone,
				"Tempered in Solitude — exile the top card of your library; you may play it this turn",
				func(g *game.Game, item *game.StackItem) error {
					return exileTopCardYouMayPlayThisTurn(item, NewContext(g, item), 0)
				}),
		},
	})
}
