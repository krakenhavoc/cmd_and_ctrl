package game

import "github.com/google/uuid"

// exile_cost.go — #1221: exiling the SOURCE CARD as part of an
// activated ability's cost, the component `AbilityCost` was missing
// once abilities started functioning from a graveyard.
//
// #1228 gave it a second owner one ability kind over —
// ManaAbilityShape.ExileSelf, the Spirit Guides' "Exile this card
// from your hand: Add {R}" — so the validator and the payer below
// take the zone and the bit rather than an AbilityCost. One clause,
// two owners, the way SacrificeOther, RemoveCounters, TapOthers and
// DiscardCards each are: a component declared twice is a component
// that can be paid two ways, and the CR 903.9 answer, the CR 601.2h
// indivisible step and the "the payment IS the source, so nothing
// goes on the wire" property are each written once.
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

// validateExileSelfFromLocked checks an ExileSelf component without
// moving anything — the validate-all-then-pay discipline both
// activation paths keep, so a refused activation never leaves a
// half-paid cost behind.
//
// What it enforces is one rule: the source has to be in the zone the
// ability says the card is exiled FROM. Every card that prints the
// component names a zone in the same breath ("from your graveyard"
// for scavenge, embalm and eternalize; "from your hand" for the
// Spirit Guides), and both activation paths have already checked that
// a non-battlefield source is the activator's own card (CR 108.4), so
// "your" needs no second test here.
//
// srcZone is the zone the activation was validated against and `want`
// the zone the component is printed against, so this is a free check
// rather than a second lookup.
//
// Caller must hold g.mu.
func (g *Game) validateExileSelfFromLocked(srcZone, want ZoneKind, exileSelf bool) error {
	if !exileSelf {
		return nil
	}
	if srcZone != want {
		return ErrActivationZoneNotAllowed
	}
	return nil
}

// validateExileSelfCostLocked is validateExileSelfFromLocked for a
// CR 602 activated ability's cost: scavenge, embalm and eternalize
// all print "Exile this card from your GRAVEYARD".
//
// Caller must hold g.mu.
func (g *Game) validateExileSelfCostLocked(srcZone ZoneKind, cost AbilityCost) error {
	return g.validateExileSelfFromLocked(srcZone, ZoneGraveyard, cost.ExileSelf)
}

// validateManaExileSelfCostLocked is the same check for a CR 605 mana
// ability (#1228): the Spirit Guides print "Exile this card from your
// HAND". The zone is read off the ability's own declaration rather
// than hard-coded, because the component names "this card" and CR
// 113.6 is what says which pile it is in — an ability that declared
// two zones would be exiled from whichever one holds it, and one that
// declared none never reaches here (effects.Register refuses
// ExileSelf on a battlefield-only mana ability at boot).
//
// Caller must hold g.mu.
func (g *Game) validateManaExileSelfCostLocked(srcZone ZoneKind, ab ManaAbilityShape) error {
	if !ab.ExileSelf {
		return nil
	}
	if !ManaAbilityFunctionsFromZone(ab, srcZone) || srcZone == ZoneBattlefield {
		return ErrActivationZoneNotAllowed
	}
	return nil
}

// payExileSelfCostLocked pays an ExileSelf component: the source card
// leaves the zone it was activated from for exile.
//
// Paid beside the discards, at the very end of the cost block, for
// the same reason they are: it moves the source and so invalidates
// every pointer the payment block was holding. A CR 602 ability's
// Effect reads the card back by instance ID afterwards — scavenge
// wants its power, embalm wants the whole card to copy — and finds it
// in exile, which is where this put it. A mana ability has no effect
// to run, but its mint site has the same problem and solves it the
// same way: it snapshots the card before this runs (#1212's
// ManaSourceKinds) rather than reading a pointer this invalidated.
//
// Caller must hold g.mu.
func (g *Game) payExileSelfCostLocked(playerID, sourceID uuid.UUID, exileSelf bool) error {
	if !exileSelf {
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

// payAbilityExileSelfLocked is payExileSelfCostLocked for a CR 602
// activated ability. Caller must hold g.mu.
func (g *Game) payAbilityExileSelfLocked(playerID, sourceID uuid.UUID, ab ActivatedAbilityShape) error {
	return g.payExileSelfCostLocked(playerID, sourceID, ab.Cost.ExileSelf)
}
