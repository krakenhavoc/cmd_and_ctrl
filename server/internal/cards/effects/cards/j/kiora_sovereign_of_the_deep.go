package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kiora, Sovereign of the Deep — Legendary Creature — Merfolk Noble
// {3}{G}{U}, 4/5 (March of the Machine: The Aftermath):
//
//	"Vigilance, ward {3}
//	 Whenever you cast a Kraken, Leviathan, Octopus, or Serpent spell from
//	 your hand, look at the top X cards of your library, where X is that
//	 spell's mana value. You may cast a spell with mana value less than X
//	 from among them without paying its mana cost. Put the rest on the
//	 bottom of your library in a random order."
//
// Ward {3} is the shared Ward trigger. The cast trigger is Taigam's
// from-hand read (YouCastFromYourHand), narrowed to the four creature
// types, each a word of Kiora's text (CR 612.2). X is the spell's mana
// value read while it is still on the stack, its announced X included
// (Chrome Host Seedshark's triggeringSpellManaValue). The look-and-cast is
// Aetherworks Marvel's (lookAtTopAndOfferFreeCast) capped at mana value
// X − 1, since "less than X" is "X − 1 or less": the cast is free, a land
// is never offered, and the rest go to the bottom in a random order. With
// X = 0 nothing is looked at.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "95df54e2-0196-4e54-8144-de5982ce07f5",
		Name:            "Kiora, Sovereign of the Deep",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{3}"), "Kiora, Sovereign of the Deep — ward {3}"),
			effects.On(game.EventCast, kioraCastSeaMonsterFromHand,
				"Kiora, Sovereign of the Deep — look at the top X cards of your library, where X is that spell's mana value; you may cast a spell with lesser mana value from among them without paying its mana cost",
				kioraLookAndCast),
		},
	})
}

var kioraFreeCast = effects.RegisterFreeLibraryCastOffer("choose-cards/kiora-sovereign-of-the-deep-cast", "Kiora, Sovereign of the Deep")

func kioraCastSeaMonsterFromHand(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if !effects.YouCastFromYourHand(nil)(ev, source, lki, g) {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	for _, word := range []string{"Kraken", "Leviathan", "Octopus", "Serpent"} {
		if spell.HasSubtype(source.TextSubtype(word)) {
			return true
		}
	}
	return false
}

func kioraLookAndCast(g *game.Game, item *game.StackItem) error {
	x := effects.TriggeringSpellManaValue(g, item)
	if x <= 0 {
		return nil
	}
	return effects.LookAtTopAndOfferFreeCast(g, item.Controller, item.SourceCardID, kioraFreeCast, x,
		effects.SpellWithManaValueAtMost(x-1))
}
