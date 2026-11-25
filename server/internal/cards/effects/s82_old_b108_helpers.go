package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b108_helpers.go — shared pieces of Core Set 2021 batch 108 (#5511).

// S82OldB108CastFromGraveyardByLifeAndDiscard is "You may cast this card from
// your graveyard by paying <life> life and discarding a card in addition to
// paying its other costs" (Demonic Embrace): the card's printed mana cost, the
// life, and one discarded card, paid when it is cast from the graveyard. The
// card file opens the zone:
//
//	CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
//	AlternativeCosts: []game.AlternativeCost{S82OldB108CastFromGraveyardByLifeAndDiscard("{1}{B}{B}", 3)},
func S82OldB108CastFromGraveyardByLifeAndDiscard(printedCost string, life int) game.AlternativeCost {
	return game.AlternativeCost{
		Key:             "s82_old_b108_life_and_discard_from_graveyard",
		Label:           "Pay " + NumberWord(life) + " life and discard a card to cast this card from your graveyard",
		ManaCost:        printedCost,
		FromZone:        game.ZoneGraveyard,
		Life:            life,
		DiscardFromHand: CardInYourHand("a card").WithCount(1, 1),
		PayLabel:        "a card",
	}
}
