package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vicious Shadows — {6}{R} Enchantment (Shards of Alara, S82 OLD b515):
//
//	"Whenever a creature dies, you may have this enchantment deal damage to
//	 target player equal to the number of cards in that player's hand."
//
// Any creature's death, including one of yours and its own controller's. The
// target is chosen when the trigger is put on the stack and the damage is read
// from that player's hand as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9707ac4f-6493-47fe-9c31-4daf469b5c6e",
		Name:         "Vicious Shadows",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Optional(effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, died := effects.DiedCreature(ev, g)
				return died
			}, "Vicious Shadows — it deals damage to target player equal to the number of cards in that player's hand", effects.OldB515DamageEqualToTargetsHand),
				"Vicious Shadows — have it deal damage to target player equal to the cards in their hand?"),
				effects.TargetPlayer("target player")),
		},
	})
}
