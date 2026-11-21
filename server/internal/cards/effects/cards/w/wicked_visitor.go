package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wicked Visitor — Creature — Nightmare {1}{B}, 2/2 (Wilds of Eldraine,
// #4471):
//
//	"Whenever an enchantment you control is put into a graveyard from the
//	 battlefield, each opponent loses 1 life."
//
// Knight of Doves' trigger (socB13EnchantmentYouControlWasPutIntoAGraveyard),
// the enchantment read as it last existed on the battlefield and under
// whose control (CR 603.10a), so an Aura that falls off its creature and a
// Role token both count. The payout is life loss, not damage, written
// with the footprinted EachOpponentLosesLife primitive.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10330a79-a416-40da-bf5a-8b7cbd9a0498",
		Name:         "Wicked Visitor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.SocB13EnchantmentYouControlWasPutIntoAGraveyard,
				"Wicked Visitor — each opponent loses 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1})),
		},
	})
}
