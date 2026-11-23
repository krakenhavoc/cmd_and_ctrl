package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keen Sense — Enchantment — Aura for {G} (EDHREC rank 2857):
//
//	"Enchant creature
//	 Whenever enchanted creature deals damage to an opponent, you may
//	 draw a card."
//
// Curiosity in green, word for word, and in the batch for the same
// reason two Signets are: a deck that wants the effect wants all the
// copies of it its colours allow, and a green Voltron deck cannot
// play the blue one.
//
// Everything Curiosity's file records applies here — any damage, an
// opponent only, "you" being the Aura's controller, and the "may"
// being a real prompt rather than an assumed yes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8d47b78b-11cc-4351-91d6-c891eb63dd98",
		Name:         "Keen Sense",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttachedCreatureDealtDamageToOpponent(ev, source, g)
			}, "Keen Sense — draw a card", effects.Do(effects.DrawCards{N: 1})), "Keen Sense — draw a card?", 1),
		},
	})
}
