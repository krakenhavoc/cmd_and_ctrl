package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eye of Nidhogg — Legendary Enchantment — Aura {2}{B}:
//
//	"Enchant creature
//	 Enchanted creature is a black Dragon with base power and toughness
//	 4/2, has flying and deathtouch, and is goaded. (It attacks each
//	 combat if able and attacks a player other than you if able.)
//	 When Eye of Nidhogg is put into a graveyard from the battlefield,
//	 return it to its owner's hand."
//
// One continuous effect, written phrase by phrase in the layer each
// lands in: the creature type is REPLACED by Dragon (layer 4, the card
// types stay), the colour is black (layer 5), the base power and
// toughness are 4/2 (layer 7b), flying and deathtouch are layer 6, and
// the goad is the static form (GoadAttached) held by the Aura's
// controller. The return trigger is Glistening Oil's: it moves the card
// only while it is still the graveyard object that died.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c288f184-0beb-4cc9-9b0a-5f3c4b9e70d7",
		Name:         "Eye of Nidhogg",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			{
				Label:                 "Enchanted creature is a Dragon.",
				Layer:                 game.Layer4Type,
				ContinuesAfterRemoval: true,
				AppliesTo:             effects.AttachedToSource,
				// "Dragon" and "black" are words of the Aura's text,
				// read through it (CR 612.2, #3984).
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					c.SetSubtypes(source.TextSubtypes("Dragon"))
				},
			},
			effects.SetAttachedColors("B"),
			effects.SetAttachedBasePT(4, 2),
			effects.GrantToAttached("flying", "deathtouch"),
			effects.GoadAttached(),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.CardDied(ev, source)
			}, "Eye of Nidhogg — return it to its owner's hand", effects.ReturnSelfFromGraveyardToHand),
		},
	})
}
