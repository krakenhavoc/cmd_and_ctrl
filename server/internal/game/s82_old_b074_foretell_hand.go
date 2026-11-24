package game

import "github.com/google/uuid"

// ForetellCardFromHandForEffect is "exile a card from your hand face down.
// It becomes foretold. Its foretell cost is its mana cost reduced by
// {reduce}" (Ethereal Valkyrie, CR 702.143b): ForetellPermanentForEffect's
// twin for a card in the player's hand. The card goes to its owner's exile
// face down with the foretell kind, one routed move, so replacement effects
// apply, and its owner may cast it from exile for the reduced cost on a
// LATER turn (CR 702.143a) for as long as it remains exiled.
//
// It is not the foretell special action, so nothing is paid and
// Game.ForetoldThisTurn is not bumped. The reduction takes generic mana off
// the printed cost and never goes below zero or into a coloured symbol. A
// card with no mana cost is still exiled face down but has no foretell cost
// to pay, so no permission is stamped for it. A card that is not in the
// player's hand is left alone.
//
// Caller must hold g.mu in write mode.
func (g *Game) ForetellCardFromHandForEffect(player, cardID uuid.UUID, reduce int) error {
	p := g.playerByIDLocked(player)
	if p == nil || p.Hand == nil || !p.Hand.Contains(cardID) {
		return nil
	}
	var cost string
	for i := range p.Hand.Cards {
		if p.Hand.Cards[i].InstanceID != cardID {
			continue
		}
		if mc := p.Hand.Cards[i].ManaCost; mc != "" {
			if parsed, err := ParseCost(mc); err == nil {
				cost = reduceGeneric(parsed, reduce).String()
				if cost == "" {
					cost = "{0}" // fully reduced: a foretell cost of {0}
				}
			}
		}
		break
	}
	return g.routeAllKeyedLocked(zoneRoute{
		Dst:      ZoneExile,
		Actor:    player,
		FaceDown: FaceDownForetold,
		Cause:    MoveCause{Kind: MoveCauseEffect, Controller: player},
	}, []uuid.UUID{cardID}, KeyedTail{Ref: foretoldByEffect, Player: player, N: g.Turn.Seq + 1, Word: cost})
}
