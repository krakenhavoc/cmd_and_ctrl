package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pay_unless_payments.go — #6187: three pay-unless payments (CR
// 118.12a) and the entry trigger that charges one.
//
//	118.12a Some spells, activated abilities, and triggered abilities
//	        read, "[Do something] unless [a player does something
//	        else]." This means the same thing as "[A player may do
//	        something else]. If [that player doesn't], [do something]."

// RandomDiscardPayment is "Discard N cards at random" as a payment
// (game.PayActionDiscardRandom, Balduvian Horde): the payer says yes or
// no, and the engine picks the cards (CR 701.9b). With fewer than N
// cards in hand there is no payment to make (CR 118.3).
func RandomDiscardPayment(n int) UpkeepPayment {
	words := "Discard a card at random"
	if n != 1 {
		words = "Discard " + NumberWord(n) + " cards at random"
	}
	return UpkeepPayment{Action: &game.PayAction{Kind: game.PayActionDiscardRandom, Count: n}, Words: words}
}

// HandToLibraryTopPayment is "Put N cards from your hand on top of your
// library" as a payment (game.PayActionHandToLibraryTop, Tainted
// Specter): the payer names the cards. It is not a discard.
func HandToLibraryTopPayment(n int) UpkeepPayment {
	words := "Put a card from your hand on top of your library"
	if n != 1 {
		words = "Put " + NumberWord(n) + " cards from your hand on top of your library"
	}
	return UpkeepPayment{Action: &game.PayAction{Kind: game.PayActionHandToLibraryTop, Count: n}, Words: words}
}

// SacrificeTotalPowerPayment is "Sacrifice any number of <plural> with
// total power N or greater" as a payment
// (game.PayActionSacrificeTotalPower, Phyrexian Dreadnought): `of` is the
// permanents that may pay.
func SacrificeTotalPowerPayment(total int, plural string, of game.PermanentQuery) UpkeepPayment {
	return UpkeepPayment{
		Action: &game.PayAction{Kind: game.PayActionSacrificeTotalPower, TotalPower: total, Of: of},
		Words:  "Sacrifice any number of " + plural + " with total power " + strconv.Itoa(total) + " or greater",
	}
}

// EntersSacrificeUnlessPaying is "When this <permanent> enters,
// sacrifice it unless you <payment>" for a payment that is not mana:
// the pay-unless prompt with the shared "sacrifice this permanent"
// decline, as EntersSacrificeUnlessYouTap's. With nothing to pay with,
// "yes" is no payment and the permanent is sacrificed. A source that left
// in response, or left and came back (CR 400.7), asks nothing.
func EntersSacrificeUnlessPaying(name string, pay UpkeepPayment) game.TriggeredAbility {
	words, action := pay.Prompt()
	if action == nil {
		panic("effects.EntersSacrificeUnlessPaying: " + name + " needs a payment that is not mana")
	}
	clause := lowerFirstWord(words)
	return WhenThisEnters(name+" — sacrifice it unless you "+clause,
		func(g *game.Game, item *game.StackItem) error {
			if !OnBattlefield(g, item.SourceCardID) || SourceIsNewObject(g, item) {
				return nil
			}
			// Each trigger asks with its own copy of the payment.
			a := *action
			return UpkeepPayUnless{
				Chooser:  item.Controller,
				Cost:     words,
				Action:   &a,
				Question: name + " — " + clause + ", or sacrifice " + name + "?",
				Else:     UpkeepSacrificeSource,
			}.Apply(NewContext(g, item))
		})
}

// lowerFirstWord lower-cases the first letter of a payment's words.
func lowerFirstWord(s string) string {
	if s == "" {
		return s
	}
	b := []byte(s)
	if b[0] >= 'A' && b[0] <= 'Z' {
		b[0] += 'a' - 'A'
	}
	return string(b)
}
