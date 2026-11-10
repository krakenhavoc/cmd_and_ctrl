package game

import "github.com/google/uuid"

// graveyard_effect_grants.go — an ability a RESOLVED effect gives a card in
// a graveyard until end of turn (Cursecloth Wrappings, #4464).
//
//	CR 113.10  "Effects can add or remove abilities of objects."
//	CR 611.2a  a continuous effect from a resolved ability lasts as long as
//	           it says: "until end of turn"
//	CR 400.7   a card that moves is a new object with no memory of its past
//	CR 514.2   "until end of turn" effects end in the cleanup step
//
// Cursecloth Wrappings: "Target creature card in your graveyard gains embalm
// until end of turn. The embalm cost is equal to its mana cost." It is
// exile_effect_grants.go's grant (ADR 0219) given to a card in a graveyard:
// the same Card.ExileGrants rows (plain data, no closure), read by the one
// ability-row accessor, so the activation path, the wire's zone_abilities
// row and the legal-move enumerator see the granted embalm with no change of
// their own. The two differences are the zone, whose bundle is an ordinary
// graveyard AbilityGrant (ActivatedAbility.Zones), and the end: the row
// carries UntilEndOfTurn and the cleanup sweep takes it, while MoveCard takes
// it the moment the card leaves the graveyard (CR 400.7), embalm's own exile
// cost included.

// GrantAbilityToGraveyardCardUntilEndOfTurnForEffect gives the card
// `cardID`, which must be in a graveyard, the bundle named `key` until end
// of turn. `source` is the granting object, for the wire label; `activator`
// is the "you" of the granting effect. A card that is no longer in a
// graveyard is left alone (CR 608.2b), and an empty key gives nothing.
// Reports whether the grant landed.
//
// Caller must hold g.mu (write).
func (g *Game) GrantAbilityToGraveyardCardUntilEndOfTurnForEffect(cardID, source, activator uuid.UUID, key string) bool {
	gk := GrantKey(key)
	if gk == "" {
		return false
	}
	c := graveyardCardByIDLocked(g, cardID)
	if c == nil {
		return false
	}
	// Always a fresh slice, never written in place: a clone or a view may
	// share the old one.
	out := make([]GrantedAbility, 0, len(c.ExileGrants)+1)
	out = append(out, c.ExileGrants...)
	c.ExileGrants = append(out, GrantedAbility{Key: gk, Source: source, Activator: activator, UntilEndOfTurn: true})
	return true
}

// graveyardCardByIDLocked finds a card in any seat's graveyard.
func graveyardCardByIDLocked(g *Game, id uuid.UUID) *Card {
	for _, p := range g.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for i := range p.Graveyard.Cards {
			if p.Graveyard.Cards[i].InstanceID == id {
				return &p.Graveyard.Cards[i]
			}
		}
	}
	return nil
}

// clearTurnGraveyardGrantsLocked is the CR 514.2 sweep: every graveyard
// card loses the grants that were given "until end of turn". Always a fresh
// slice (a clone may share the old one); a card with none is not touched.
//
// Caller must hold g.mu (write).
func (g *Game) clearTurnGraveyardGrantsLocked() {
	for _, p := range g.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for i := range p.Graveyard.Cards {
			c := &p.Graveyard.Cards[i]
			keep := false
			drop := false
			for _, gr := range c.ExileGrants {
				if gr.UntilEndOfTurn {
					drop = true
				} else {
					keep = true
				}
			}
			if !drop {
				continue
			}
			if !keep {
				c.ExileGrants = nil
				continue
			}
			out := make([]GrantedAbility, 0, len(c.ExileGrants))
			for _, gr := range c.ExileGrants {
				if !gr.UntilEndOfTurn {
					out = append(out, gr)
				}
			}
			c.ExileGrants = out
		}
	}
}
