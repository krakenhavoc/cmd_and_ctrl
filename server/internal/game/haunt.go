package game

import "github.com/google/uuid"

// haunt.go — haunt (CR 702.55, #6083): a card exiled "haunting" a creature,
// and the abilities that trigger from exile when that creature dies.
//
//	702.55a Haunt is a triggered ability. “Haunt” on a permanent means
//	        “When this permanent is put into a graveyard from the
//	        battlefield, exile it haunting target creature.”
//	702.55b Cards that are in the exile zone as the result of a haunt
//	        ability “haunt” the creature targeted by that ability. The
//	        phrase “creature it haunts” refers to the object targeted by
//	        the haunt ability, regardless of whether or not that object is
//	        still a creature.
//	702.55c Triggered abilities of cards with haunt that refer to the
//	        haunted creature can trigger in the exile zone.
//
// The link is plain data on the exiled card (Card.HauntedID, HauntedEpoch,
// HauntingEpoch): the haunted OBJECT, named by its instance ID and its
// CR 400.7 epoch at the moment it was haunted, and the exiled card's own
// epoch then. So the link ends by itself when either object ends: the
// haunted creature blinked is a new object the card never haunted, and the
// card that leaves exile and comes back haunts nothing. No func is stored,
// and no registry outside the card holds it.
//
// "When … the creature it haunts dies" is an ordinary triggered ability
// whose Zones names the exile zone (trigger_zones.go): the declared-zone
// harvest walks exile for EventLTB, and CreatureItHauntsDiedForEffect is the
// condition. The card's owner controls the trigger, as with every ability
// of a card outside the battlefield (CR 108.4, harvestFromDeclaredZone).

// HauntForEffect is haunt's resolution: "exile it haunting target
// creature". `cardID` is the card with haunt, which the caller has already
// checked is still the object that died (in its owner's graveyard);
// `haunted` is the target, which must be on the battlefield. The card is
// exiled through the ordinary route, and the link is written only if it
// arrived in exile. Reports whether the card now haunts the creature.
//
// Caller must hold g.mu (write).
func (g *Game) HauntForEffect(cardID, haunted uuid.UUID) bool {
	host, ok := g.battlefieldCardLocked(haunted)
	if !ok || host == nil {
		return false
	}
	hauntedEpoch := host.ObjectEpoch
	if err := g.ExileCardForEffect(cardID); err != nil {
		return false
	}
	z := g.findCardZoneLocked(cardID)
	if z == nil || z.Kind != ZoneExile {
		return false
	}
	for i := range z.Cards {
		c := &z.Cards[i]
		if c.InstanceID != cardID {
			continue
		}
		c.HauntedID = haunted
		c.HauntedEpoch = hauntedEpoch
		c.HauntingEpoch = c.ObjectEpoch
		return true
	}
	return false
}

// HauntsForEffect reports whether the card `source` (as the harvest hands
// it) is in exile haunting an object: it carries a link written while it
// was this object.
//
// Caller must hold g.mu.
func (g *Game) HauntsForEffect(source *Card) bool {
	if source == nil || source.HauntedID == uuid.Nil || source.ObjectEpoch != source.HauntingEpoch {
		return false
	}
	z := g.findCardZoneLocked(source.InstanceID)
	return z != nil && z.Kind == ZoneExile
}

// CreatureItHauntsDiedForEffect is the trigger condition "the creature it
// haunts dies" for an exiled card `source`: `ev` is a creature going from
// the battlefield to a graveyard (CR 700.4), and that creature is the
// object `source` haunts. The haunted object died at the epoch it was
// haunted at, so the card in the graveyard is one move on from it; a token
// that has already ceased to exist (CR 704.5d) is not found at all, and is
// that object too, since a token can't come back.
//
// Caller must hold g.mu.
func (g *Game) CreatureItHauntsDiedForEffect(ev Event, source *Card) bool {
	if ev.Kind != EventLTB || ev.NewZone != ZoneGraveyard || !g.HauntsForEffect(source) || ev.CardID != source.HauntedID {
		return false
	}
	if epoch := g.cardObjectEpochLocked(ev.CardID); epoch >= 0 && epoch != source.HauntedEpoch+1 {
		return false
	}
	was, known := ev.WasType("creature")
	if !known {
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && c.IsCreature()
	}
	return was
}
