package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// exiled_card_cost.go — #6270, Back from the Brink: "Exile a creature card
// from your graveyard and pay its mana cost: Create a token that's a copy
// of that card."

// ExileFromGraveyardAndPayItsManaCost is "Exile a <kind> card from your
// graveyard and pay its mana cost" as a CR 602 activated ability's cost:
// the activator names one matching card, and the activation's mana is that
// card's mana cost (game.AbilityCost.ManaFromExiledCard), paid with the
// exile (CR 601.2h). The effect reads the card back through Context.Exiled.
// Combine with Plus for a printed mana or tap component beside it.
func ExileFromGraveyardAndPayItsManaCost(label string, match func(game.Card) bool) game.AbilityCost {
	return game.AbilityCost{ExileCards: ExileCardsFromGraveyard(1, label, match), ManaFromExiledCard: true}
}
