package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// graveyard_order.go — #6194: the card-side spellings of the clauses that
// read where a card sits in a graveyard (CR 404.2). The engine half is
// game/graveyard_order.go.

// ExileTopFromGraveyard is "Exile the top <kind> card of your graveyard"
// as a CR 602 activated ability's cost (Zombie Scavengers, Necratog,
// Alms, Nature's Kiss): the n topmost cards of the activator's graveyard
// that `match` accepts (nil is any card), never ones below them
// (game.ExileCost.TopOnly). A graveyard without n matches can't pay
// (CR 118.3).
func ExileTopFromGraveyard(n int, label string, match func(game.Card) bool) game.AbilityCost {
	return game.AbilityCost{ExileCards: &game.ExileCost{N: n, Label: label, From: game.ZoneGraveyard, Match: match, TopOnly: true}}
}

// ExileTopGraveyardCardPayment is "exile the top <cardType> card of your
// graveyard" as a pay-unless payment (Barrow Ghoul's and Circling
// Vultures' "sacrifice this creature unless you exile the top creature
// card of your graveyard"). The card is not the payer's to pick: it is the
// topmost card of that type in their own graveyard, and a graveyard with
// none can't pay (CR 118.3). Empty cardType is the top card, whatever it
// is.
func ExileTopGraveyardCardPayment(cardType string) UpkeepPayment {
	words := "Exile the top card of your graveyard"
	if cardType != "" {
		words = "Exile the top " + cardType + " card of your graveyard"
	}
	return UpkeepPayment{Action: &game.PayAction{Kind: game.PayActionExileFromGraveyard, Count: 1, CardType: cardType, TopOnly: true}, Words: words}
}
