package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sythis, Harvest's Hand — Legendary Enchantment Creature — Nymph,
// {G}{W}, 1/2 (EDHREC rank 982):
//
//	"Whenever you cast an enchantment spell, you gain 1 life and draw
//	 a card."
//
// The enchantress commander: Mesa Enchantress's trigger, mandatory,
// with a life on top. Fires on cast — the spell is on the stack, so
// the card is drawn before the enchantment resolves, as printed —
// and not on Sythis's own cast, since she is not on the battlefield
// to see it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0fc64fd6-f057-4056-9dca-47accb7ff036",
		Name:         "Sythis, Harvest's Hand",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.EnchantmentSpellCastByYou, "Sythis, Harvest's Hand — gain 1 life and draw a card", effects.Do(
				// The draw waits for a prompt the life gain asks (CR 608.2c, #4155).
				effects.GainLife{Amount: 1},
				effects.DrawCards{N: 1},
			)),
		},
	})
}
