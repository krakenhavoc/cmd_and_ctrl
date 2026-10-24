package game

import "github.com/google/uuid"

// returned_aura.go — a card that returns to the battlefield as an Aura
// (#2809, ADR 0154's 2026-10-23 amendment).
//
//	Harold and Bob, First Numens: "When Harold and Bob dies, if it was a
//	creature, return it to the battlefield. It's an Aura enchantment
//	with enchant Forest you control and 'Enchanted Forest has "{T}: Add
//	three mana of any one color. You get two rad counters."' Harold and
//	Bob loses all other abilities."
//
// Bronzehide Lion prints the same shape with a different clause and
// grant.
//
// # What it is
//
// Three rules meet, and each already had a seam:
//
//   - CR 303.4f: an Aura entering other than by resolving as an Aura
//     spell chooses what it enchants as it enters, by its enchant ability
//     "and any other applicable effects". aura_host.go asks that question
//     of the Aura's catalogued clause. A creature card in a graveyard has
//     none, so ReplacementEvent.entersAsAura says the entry is an Aura's
//     and CardDef.ReturnedAura is its clause. With no legal host the card
//     stays in the graveyard (CR 303.4g).
//   - CR 205.1a, CR 613.1d: "It's an Aura enchantment" SETS its card
//     types, so the creature type and the creature subtypes are gone and
//     the supertypes stay (CR 205.4b: Harold and Bob is still legendary).
//   - CR 613.1f, CR 613.6: "loses all other abilities" and the granted
//     ability are one effect. The ability it keeps is the one the Aura
//     itself has, "Enchanted Forest has …".
//
// The last two are the card's own statics, built by effects.ReturnedAuraForm
// and read through Card.ReturnedAura, which is the one new state. It is
// set as the entry lands (seedReturnedAuraEntryLocked), after the CR 400.7
// reset and before the attach and EventETB, as bestow's flag is, and it is
// cleared by every zone change, so a second death returns nothing: the
// Aura is no creature and has no dies trigger.
//
// The enchant clause is CardDef.ReturnedAura and not Targets: a creature
// spell reads Targets as its own cast-time target. attachmentLegalLocked
// reads it first when the flag is set, so "Forest you control" is
// re-checked every state-based pass as any enchant clause is (CR 704.5m)
// and an Aura attached to nothing is illegal (the uncatalogued-Aura
// exemption does not reach it).

// ReturnedAuraSpecFor is the enchant clause of a card's Aura form, or
// nil for a card that has none.
func ReturnedAuraSpecFor(key string) *TargetSpec {
	if key == "" {
		return nil
	}
	if d := catalogDef(key); d != nil {
		return d.ReturnedAura
	}
	return nil
}

// auraEnchantSpecOf is the enchant clause an Aura on the battlefield is
// held to: its Aura form's when it returned as one, else its catalogued
// clause.
func auraEnchantSpecOf(c *Card) *TargetSpec {
	if c.ReturnedAura {
		return ReturnedAuraSpecFor(catalogKeyOf(c))
	}
	return TargetSpecFor(catalogKeyOf(c))
}

// ReturnFromGraveyardAsAuraForEffect returns a card from a graveyard to
// the battlefield as an Aura: "return it to the battlefield. It's an
// Aura enchantment with enchant …". The controller chooses what it
// enchants as it enters (CR 303.4f); with no legal host it stays in the
// graveyard (CR 303.4g). `controller` uuid.Nil is "under its owner's
// control". Returns the entering permanent's ID, or uuid.Nil when
// nothing entered or the entry paused for a question.
//
// Refuses a card that is not in a graveyard with ErrCardNotFound, and a
// card whose definition names no Aura form with ErrInvalidParam, moving
// nothing.
//
// Caller must hold g.mu.
func (g *Game) ReturnFromGraveyardAsAuraForEffect(cardID, controller uuid.UUID, tapped bool) (uuid.UUID, error) {
	z := g.findCardZoneLocked(cardID)
	if z == nil || z.Kind != ZoneGraveyard {
		return uuid.Nil, ErrCardNotFound
	}
	var key string
	for i := range z.Cards {
		if z.Cards[i].InstanceID == cardID {
			key = catalogKeyOf(&z.Cards[i])
			break
		}
	}
	if ReturnedAuraSpecFor(key) == nil {
		return uuid.Nil, ErrInvalidParam
	}
	return g.returnFromGraveyardEntryLocked(cardID, ZoneBattlefield, controller, tapped, FaceDownNone, nil, nil, true, nil)
}

// seedReturnedAuraEntryLocked makes the permanent that has just landed
// the Aura form of its card. After the CR 400.7 reset and before the
// Aura attach and EventETB in announceEntryLocked, so the attach finds
// an Aura and an enters trigger finds no creature.
//
// Caller must hold g.mu.
func (g *Game) seedReturnedAuraEntryLocked(entered uuid.UUID, moved *Card) {
	if c := findBattlefieldCard(g, entered); c != nil {
		c.ReturnedAura = true
		c.effective = nil
	}
	if moved != nil {
		moved.ReturnedAura = true
	}
	g.layerVersion.Add(1)
}
