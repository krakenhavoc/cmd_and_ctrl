package game

import "github.com/google/uuid"

// exiled_card_cost.go — #6270, Back from the Brink: "Exile a creature
// card from your graveyard and pay its mana cost: Create a token that's a
// copy of that card." CR 602.2b: activating follows CR 601.2b–i, so the
// card is named as the cost is announced and the total cost (601.2f) is
// then known: that card's mana cost, which is paid with the exile (601.2h,
// all or nothing).

// AbilityWithExiledCardCostForEffect is `ab` with its mana component set
// from the card named to its ExileCards cost (AbilityCost.ManaFromExiledCard):
// the printed Mana, if any, plus that card's mana cost as it sits in the
// activator's graveyard or hand (its printed cost; a card with none adds
// nothing, CR 202.1b). Any other ability, or no card named, is returned
// as it is. Caller must hold g.mu.
func (g *Game) AbilityWithExiledCardCostForEffect(activator uuid.UUID, ab ActivatedAbilityShape, exileIDs []uuid.UUID) ActivatedAbilityShape {
	if !ab.Cost.ManaFromExiledCard || ab.Cost.ExileCards == nil || len(exileIDs) != 1 {
		return ab
	}
	pile := g.exileCostPileLocked(g.playerByIDLocked(activator), ab.Cost.ExileCards)
	c := findZoneCard(pile, exileIDs[0])
	if c == nil {
		return ab
	}
	ab.Cost.Mana += c.ManaCost
	return ab
}
