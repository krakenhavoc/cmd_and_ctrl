package game

import "github.com/google/uuid"

// exile_cost.go — #1221: exiling the SOURCE CARD as part of an
// activated ability's cost, the component `AbilityCost` was missing
// once abilities started functioning from a graveyard.
//
// Three keywords print it, all of them from the graveyard and all of
// them in the same words:
//
//	scavenge     "Exile this card from your graveyard: Put a number
//	             of +1/+1 counters equal to this card's power on
//	             target creature."            CR 702.96a
//	embalm       "Exile this card from your graveyard: Create a
//	             token that's a copy of it, except …"  CR 702.128a
//	eternalize   the same, 4/4 black Zombie.  CR 702.129a
//
// It is DiscardSelf one zone over (discard_cost.go), and it is a
// second bit rather than a zone parameter on one "the source pays
// itself" component because the two are different rules. Discarding
// is a CR 701.8 keyword action with its own event, its own cause and
// its own CR 614 window over the hand→graveyard move; exiling as a
// cost is a plain CR 406 zone change with none of that. The shared
// part — "the payment is the source, so nothing is announced and
// nothing goes on the wire" — is the part both get for free by being
// a bool.
//
// The payment goes through routeCardToZoneLocked like every other
// exit, so a commander scavenged out of its owner's graveyard still
// gets its CR 903.9 answer, and MustSettleNow is set for the reason
// the discard cost sets it: CR 602.2b activates an ability in one
// indivisible step, so a cost may not stop to ask a question.

// validateExileSelfCostLocked checks the ExileSelf component without
// moving anything — the validate-all-then-pay discipline the rest of
// ActivateCatalogAbility keeps, so a refused activation never leaves
// a half-paid cost behind.
//
// What it enforces is one rule: the source has to be in a GRAVEYARD.
// Every card that prints the component prints "from your graveyard",
// and the activation path has already checked that a non-battlefield
// source is the activator's own card (CR 108.4), so "your" needs no
// second test here.
//
// srcZone is the zone the activation was validated against, so this
// is a free check rather than a second lookup.
//
// Caller must hold g.mu.
func (g *Game) validateExileSelfCostLocked(srcZone ZoneKind, cost AbilityCost) error {
	if !cost.ExileSelf {
		return nil
	}
	if srcZone != ZoneGraveyard {
		return ErrActivationZoneNotAllowed
	}
	return nil
}

// payAbilityExileSelfLocked pays the ExileSelf component: the source
// card leaves the graveyard for exile.
//
// Paid beside the discards, at the very end of the cost block, for
// the same reason they are: it moves the source and so invalidates
// every pointer the payment block was holding. The ability's Effect
// reads the card back by instance ID afterwards — scavenge wants its
// power, embalm wants the whole card to copy — and finds it in exile,
// which is where this put it.
//
// Caller must hold g.mu.
func (g *Game) payAbilityExileSelfLocked(playerID, sourceID uuid.UUID, ab ActivatedAbilityShape) error {
	if !ab.Cost.ExileSelf {
		return nil
	}
	_, err := g.routeCardToZoneLocked(zoneRoute{
		CardID: sourceID,
		Dst:    ZoneExile,
		Actor:  playerID,
		Source: sourceID,
		// CR 601.2h / 602.2b: paying a cost is part of one
		// indivisible step, so the window over this move settles
		// itself rather than pausing on a player prompt — the bit
		// DiscardCauseCost sets for the discard half.
		MustSettleNow: true,
	})
	return err
}
