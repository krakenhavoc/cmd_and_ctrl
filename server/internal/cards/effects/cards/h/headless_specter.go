package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Headless Specter — Creature — Specter {1}{B}{B}, 2/2 (Modern Horizons):
//
//	"Flying
//	 Hellbent — Whenever this creature deals combat damage to a player, if you
//	 have no cards in hand, that player discards a card at random."
//
// Hellbent is an ability word (CR 207.2c): the trigger carries an intervening
// "if" (CR 603.4) about the Specter's controller's hand, read as the damage is
// dealt and again as the ability resolves. The player who discards is the one
// the Specter damaged, and the discard is random (CR 701.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9aa9c7ca-e160-4bfa-8ba5-87114222fd86",
		Name:            "Headless Specter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you have no cards in hand",
				func(g *game.Game, you uuid.UUID) bool { return effects.R35YouHaveNoCardsInHand(g, you, uuid.Nil) },
				effects.WheneverThisDealsCombatDamageToAPlayer("Headless Specter — that player discards a card at random",
					func(g *game.Game, item *game.StackItem) error {
						return effects.DiscardCards{Player: item.Trigger.Event.Target, N: 1}.Apply(effects.NewContext(g, item))
					})),
		},
	})
}
