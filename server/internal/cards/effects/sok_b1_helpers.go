package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// sok_b1_helpers.go — the shapes Saviors of Kamigawa, batch 1 (#5601)
// prints more than once: the Onna cycle's return to hand, the upkeep
// return of a creature of one colour, and "the number of cards in your
// hand".

// sokB1OnnaReturn is the Onna cycle's "Whenever you cast a Spirit or
// Arcane spell, you may return this creature to its owner's hand." The
// "may" is asked as the trigger resolves (CR 603.5). The return goes
// through BounceToHand's CR 400.7 guard, so an Onna that left and came
// back as a new object stays where it is.
func sokB1OnnaReturn(name string) game.TriggeredAbility {
	return Optional(WheneverYouCast(chkB1SpiritOrArcane(), name+" — return it to its owner's hand",
		func(g *game.Game, item *game.StackItem) error {
			return BounceToHand{Target: item.SourceCardID}.Apply(NewContext(g, item))
		}), name+" — return this creature to its owner's hand?")
}

// sokB1ReturnACreatureOfColour is "At the beginning of your upkeep,
// return a <colour> creature you control to its owner's hand" (Stampeding
// Serow and its cycle): Trusted Advisor's shape. It targets nothing, so
// the creature is picked as the trigger resolves, the source itself
// included; with none there is nothing to return. The colour is read off
// the effective colours, as a word of the card's text.
func sokB1ReturnACreatureOfColour(name, colour, word string) game.TriggeredAbility {
	return AtYourUpkeep(name+" — return a "+word+" creature you control", Do(ReturnOneYouControl{
		Match:    func(c game.Card) bool { return c.IsCreature() && c.HasColorWord(colour) },
		Question: name + " — return a " + word + " creature you control to its owner's hand",
	}))
}

// sokB1CardsInYourHand is "the number of cards in your hand", counted as
// the spell or ability resolves (CR 608.2h): a spell that counts its
// caster's hand is on the stack by then, so it does not count itself.
func sokB1CardsInYourHand(ctx *Context) int {
	return cardsInHand(ctx.Game, ctx.Controller())
}
