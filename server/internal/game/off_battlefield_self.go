package game

import (
	"sync"

	"github.com/google/uuid"
)

// off_battlefield_self.go — a card's own "As long as this card isn't on
// the battlefield, it's a ..." (#5427).
//
//	Grist, the Hunger Tide: "As long as Grist isn't on the battlefield,
//	it's a 1/1 Insect creature in addition to its other types."
//
//	CR 113.6c  "An ability that states which zones it doesn't function
//	            in functions everywhere except for the specified zones,
//	            even outside the game and before the game begins."
//	CR 613.1d  "Layer 4: Type-changing effects are applied."
//	CR 613.4b  "Layer 7b: Effects that set power and/or toughness to a
//	            specific number or value are applied."
//	CR 208.3   "A noncreature object not on the battlefield has power or
//	            toughness only if it has a power and toughness printed
//	            on it."
//
// So in a hand, a library, a graveyard, exile, the command zone and on
// the stack the card is a creature card (a creature spell) with the
// stated subtype, power and toughness, and on the battlefield it is its
// printed self. It is not a characteristic-defining ability (CR 604.3a
// lists colours, subtypes, power and toughness, and excludes a
// conditional one), but CR 113.6c gives it the same reach.
//
// # The shape
//
// The sentence has no source on the battlefield, so the zone layer step
// (zone_layers.go), which gathers BATTLEFIELD statics over objects
// elsewhere, never sees it. And no layer pass runs for a card off the
// battlefield with no such static around. So it is applied where a
// bestowed spell's CR 702.103b is (bestow.go): on the cold baseline,
// at the three places an object off the battlefield reads its printed
// self (printedShared, and the type and subtype fast paths in
// hasCardType and hasSubtype). Every other reader goes through those:
// effectiveOf (the view, the power and toughness accessors,
// cdaAnywhereLocked), and zoneLayerBaseline, so a Maskwood Nexus'
// "creature cards you own that aren't on the battlefield" finds the
// card already a creature. That is the CR 613.8a order too: the Nexus'
// effect depends on this one, whichever is older.
//
// The catalog declares it as DATA (effects.Spec.OffBattlefield), filed
// here at boot by key. Nothing is persisted and nothing is stamped on
// the card: the answer is a function of the card's catalog key and of
// whether it is on the battlefield.
//
// # "Isn't on the battlefield"
//
// A Card does not carry its zone. offBattlefieldSelfOf reads the two
// marks a permanent has: the layer pass's characteristic
// (Card.effective) and the entry timestamp (EnteredBattlefieldAt, set
// by the layer listener on the move and zeroed when the permanent
// leaves). A permanent has at least one of them from the entry event
// on, including while a transform or a phasing has dropped its layer
// cache; a card anywhere else has neither. A last-known copy of the
// permanent keeps both, so "whenever a creature dies" does not see a
// creature in a Grist that left the battlefield (CR 603.10a looks
// back), while the card now in the graveyard is a creature card.
//
// The entry window is the one place that asks about a card that is not
// yet a permanent as though it were one (CR 614.12): see
// entryDropsOffBattlefieldSelfLocked.
//
// A face-down object has no catalog key (CR 708.2a), so a face-down
// spell or permanent is not reached. The PrintedIs* accessors are not
// changed: they are the card's type line as printed, which is what a
// face-down PERMANENT's "if it's a creature card, turn it face up"
// reads, and on the battlefield this card is not a creature card.

// OffBattlefieldSelf is what a card is, in addition to its printed
// characteristics, while it is not on the battlefield.
type OffBattlefieldSelf struct {
	// Types and Subtypes are added to the printed ones ("in addition
	// to its other types"), in type-line spelling: "Creature",
	// "Insect".
	Types    []string
	Subtypes []string
	// Power and Toughness are set (CR 613.4b).
	Power     int
	Toughness int
}

// offBattlefieldSelfIndex is the catalog-wide table, by catalog key.
// Written at boot from effects.Register, read-only afterwards — the
// contract every Catalog* hook carries.
var offBattlefieldSelfIndex = struct {
	sync.RWMutex
	byKey map[string]*OffBattlefieldSelf
}{byKey: map[string]*OffBattlefieldSelf{}}

// IndexOffBattlefieldSelf files `self` under `key`. Called once per
// card from effects.Register; a nil `self` is a card without the
// sentence and files nothing.
func IndexOffBattlefieldSelf(key string, self *OffBattlefieldSelf) {
	if key == "" || self == nil {
		return
	}
	own := &OffBattlefieldSelf{
		Types:     append([]string(nil), self.Types...),
		Subtypes:  append([]string(nil), self.Subtypes...),
		Power:     self.Power,
		Toughness: self.Toughness,
	}
	offBattlefieldSelfIndex.Lock()
	offBattlefieldSelfIndex.byKey[key] = own
	offBattlefieldSelfIndex.Unlock()
}

// offBattlefieldSelfOf is the card's own off-battlefield sentence when
// it applies now: the card has one, and it is not on the battlefield
// (see the file comment for how that is read). Nil otherwise, which is
// nearly every card: one map read.
func offBattlefieldSelfOf(c *Card) *OffBattlefieldSelf {
	if c.effective != nil || c.EnteredBattlefieldAt != 0 {
		return nil
	}
	key := catalogKeyOf(c)
	if key == "" {
		return nil
	}
	offBattlefieldSelfIndex.RLock()
	self := offBattlefieldSelfIndex.byKey[key]
	offBattlefieldSelfIndex.RUnlock()
	return self
}

// entryDropsOffBattlefieldSelfLocked reports whether `ev` puts a card
// onto the battlefield face up that has its own off-battlefield sentence
// where it is now. CR 614.12: "To determine which replacement effects
// apply and how they apply, check the characteristics of the permanent
// as it would exist on the battlefield". There the sentence is off, so
// "creatures your opponents control enter tapped" must not read the
// creature spell or the creature card that is entering: the entry
// window answers for it with the dry run's characteristics
// (entryOverwriteViewLocked).
//
// Caller must hold g.mu.
func (g *Game) entryDropsOffBattlefieldSelfLocked(ev *ReplacementEvent) bool {
	if ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield || ev.FaceDown != FaceDownNone || ev.CardID == uuid.Nil {
		return false
	}
	z, pos := g.locateCardLocked(ev.CardID)
	if z == nil || z.Kind == ZoneBattlefield {
		return false
	}
	return offBattlefieldSelfOf(&z.Cards[pos]) != nil
}

// hasType and hasSubtype answer the type and subtype fast paths.
func (o *OffBattlefieldSelf) hasType(lowerType string) bool {
	return typeListHas(o.Types, lowerType)
}

func (o *OffBattlefieldSelf) hasSubtype(subtype string) bool {
	return typeListHas(o.Subtypes, subtype)
}

// applyTo writes the sentence onto a baseline. The baseline's slices
// may be the printed cache's (printed_cache.go's read-only rule), so a
// list that grows is a new slice, never an append in place.
func (o *OffBattlefieldSelf) applyTo(ch *Characteristic) {
	ch.Types = withAddedTypeWords(ch.Types, o.Types)
	ch.Subtypes = withAddedTypeWords(ch.Subtypes, o.Subtypes)
	ch.Power = o.Power
	ch.Toughness = o.Toughness
}

// withAddedTypeWords is `have` with each word of `add` it lacks, as a
// new slice when anything was added.
func withAddedTypeWords(have, add []string) []string {
	out := have
	grown := false
	for _, w := range add {
		if typeListHas(out, w) {
			continue
		}
		if !grown {
			out = append(make([]string, 0, len(have)+len(add)), have...)
			grown = true
		}
		out = append(out, w)
	}
	return out
}
