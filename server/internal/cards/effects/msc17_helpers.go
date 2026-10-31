package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// msc17CastFromGraveyardByDiscarding is "You may cast this card from your
// graveyard by discarding a card in addition to paying its other costs"
// (Alien Symbiosis, Dragon Man). It is Retrace's offer (ADR 0066,
// 2026-10-07 amendment) with any card as the discard instead of a land:
// the printed mana cost, spelled exactly as the card prints it, plus one
// card from the caster's hand discarded as a cost. The card file still
// opens the zone with CastableZones. The key is an on-disk identity.
func msc17CastFromGraveyardByDiscarding(printedCost string) game.AlternativeCost {
	return game.AlternativeCost{
		Key:             "discard_card_from_graveyard",
		Label:           "Discard a card to cast this card from your graveyard",
		ManaCost:        printedCost,
		FromZone:        game.ZoneGraveyard,
		DiscardFromHand: CardInYourHand("a card"),
		PayLabel:        "a card",
	}
}

// msc17CantBeBlockedTheTarget is "target creature can't be blocked this
// turn" for an activated ability: it resolves against the first target
// that is still legal (CR 608.2b skips one that left or stopped being a
// legal pick) and makes it unblockable until the end of the turn. Merfolk
// Sovereign and Daughter of the Deep share it.
func msc17CantBeBlockedTheTarget(label string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			return RestrictUntilEOT{
				Target:       t.ID,
				Restrictions: game.CantBeBlocked,
				Label:        label,
			}.Apply(ctx)
		}
		return nil
	}
}
