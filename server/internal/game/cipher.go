package game

import "github.com/google/uuid"

// cipher.go — CR 702.99, cipher (#3867, ADR 0093's 2026-10-16
// amendment).
//
//	702.99a Cipher appears on some instants and sorceries. It represents
//	        two abilities. The first is a spell ability that functions
//	        while the spell with cipher is on the stack. The second is a
//	        static ability that functions while the card with cipher is
//	        in the exile zone. "Cipher" means "If this spell is
//	        represented by a card, you may exile this card encoded on a
//	        creature you control" and "For as long as this card is
//	        encoded on that creature, that creature has 'Whenever this
//	        creature deals combat damage to a player, you may copy the
//	        encoded card and you may cast the copy without paying its
//	        mana cost.'"
//	702.99b The term "encoded" describes the relationship between the
//	        card with cipher while in the exile zone and the creature
//	        chosen when the spell represented by that card resolves.
//	702.99c The card with cipher remains encoded on the chosen creature
//	        as long as the card with cipher remains exiled and the
//	        creature remains on the battlefield. The card remains encoded
//	        on that object even if it changes controller or stops being a
//	        creature, as long as it remains on the battlefield.
//
// THE SPELL ABILITY. It is the last instruction of the spell, so it is
// asked at the end of the resolution, after every prompt the spell's own
// text opened has been answered (resolution_tail.go): the spell's
// controller picks a creature they control, or declines. Not a target,
// so hexproof does not stop it. Encoded, the card is exiled from the
// stack and Card.EncodedOn names the creature OBJECT {instance, epoch};
// declined, the resolution ends as any other (the graveyard, or its
// replacements). A copy of the spell is not "represented by a card", so
// it is never asked; neither is a card copy made to be cast (CR 707.12).
// The keyword is read off the resolving card with HasKeyword, so a
// deck-imported cipher card is encoded with no catalog entry.
//
// THE STATIC ABILITY is a layer-6 grant (CR 613.1f) from the card in
// exile to the object it is encoded on, written by the layer pass's
// cipher step at the moment of encoding (Card.EncodedAt, CR 613.7): a
// later "loses all abilities" takes the granted ability away, an earlier
// one does not (CR 613.6). The grant names the bundle CipherGrantKey, a
// text-only bundle the catalog files, with the encoded card as its
// grantor, so the view lists the ability on the creature like any other
// granted one. The epoch ends the encoding when the creature leaves the
// battlefield (CR 400.7), and MoveCard clears the link when the card
// leaves exile; a change of controller or of card type ends nothing.
//
// THE TRIGGERED ABILITY is derived from the grant, one per encoded card
// (cipherTriggersFor, read by triggersOf beside the keyword triggers):
// two cards encoded on one creature are two abilities, each copying its
// own card. It belongs to the creature, so the creature's controller
// controls it (CR 702.99c's change of controller), and it is keyed
// (CipherCastBody), so a table waiting on it is still a restore point.
// As it resolves, the encoded card — if it is still in exile — is copied
// and the copy is offered free (CopyCardsAndOfferCastsForEffect, the
// engine's one copy-and-cast door, ADR 0066's 2026-10-13 amendment), and
// cast during that resolution (CR 707.12, CR 608.2g; a resolution_cast
// prompt, ADR 0066's 2026-10-23 amendment, #4243): the "you may" is
// casting it or not, and a copy not cast ceases to exist (CR 704.5e).

// KeywordCipher is CR 702.99's canonical token.
const KeywordCipher = "cipher"

// CipherGrantKey names the bundle a card encoded on a creature grants it
// (CR 702.99a's quoted ability). On-disk identity: never renamed.
const CipherGrantKey = "cipher/encoded"

// cipherEncodeKey is the encoding prompt's keyed continuation, and
// cipherCastKey the trigger's body. On-disk identities: never renamed,
// never reused.
const (
	cipherEncodeKey = "cipher/encode"
	cipherCastKey   = "cipher/cast"
)

// cipherEncodeThen names the encoding prompt's continuation, and
// CipherCastBody the trigger's body. Both are built from their keys and
// registered in init, for annihilatorSacrificeBody's reason: the
// resolution that asks reaches the code that registers them.
var (
	cipherEncodeThen = OptionPickThen{key: cipherEncodeKey}
	CipherCastBody   = BodyRef{key: cipherCastKey}
)

func init() {
	// Carry is the spell card followed by the creatures offered, in
	// option order; the option after them is "don't encode it".
	RegisterOptionPickThen(cipherEncodeKey, func(g *Game, r OptionPicked) error {
		if len(r.Carry) < 2 || r.Option == nil {
			return nil
		}
		creatures := r.Carry[1:]
		if r.Index < 0 || r.Index >= len(creatures) {
			return nil
		}
		g.encodeCipherCardLocked(r.Carry[0], creatures[r.Index], r.Chooser)
		return nil
	})
	DelayedBody(cipherCastKey, func(g *Game, item *StackItem, p EffectParams) error {
		g.castCipherCopyLocked(item, p.Object.ID)
		return nil
	})
}

// snapshotEncodedOn and restoreEncodedOn carry Card.EncodedOn through
// the snapshot's pointer (nil for a card that is not encoded).
func snapshotEncodedOn(r PermissionCardRef) *PermissionCardRef {
	if r.ID == uuid.Nil {
		return nil
	}
	return &r
}

func restoreEncodedOn(r *PermissionCardRef) PermissionCardRef {
	if r == nil {
		return PermissionCardRef{}
	}
	return *r
}

// askCipherLocked asks the resolving spell's controller whether to exile
// it encoded on a creature they control, and reports whether it asked.
// When it did, the rest of the resolution is parked behind the answer
// with cipherAsked set, so the tail that runs after it routes the card
// without asking again.
//
// Caller must hold g.mu in write mode, inside a resolution function.
func (g *Game) askCipherLocked(top Card, item *StackItem) bool {
	if item == nil || item.IsCopy || top.CardCopy || top.IsPermanent() || !HasKeyword(&top, KeywordCipher) {
		return false
	}
	chooser := item.Controller
	if chooser == uuid.Nil {
		chooser = top.Owner
	}
	g.RecomputeLayersIfStaleLocked()
	var creatures []uuid.UUID
	var options []ChoiceOption
	if g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.Controller != chooser || !c.IsCreature() {
				continue
			}
			creatures = append(creatures, c.InstanceID)
			options = append(options, ChoiceOption{Label: "Encode it on " + c.Name, Cards: []uuid.UUID{c.InstanceID}})
		}
	}
	if len(creatures) == 0 {
		return false
	}
	options = append(options, ChoiceOption{Label: "Don't encode it"})
	id := g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:  chooser,
		Source:   top.InstanceID,
		Question: top.Name + " — cipher: exile it encoded on a creature you control?",
		Options:  options,
		ThenKey:  cipherEncodeThen,
		Carry:    append([]uuid.UUID{top.InstanceID}, creatures...),
	})
	if id == uuid.Nil {
		return false
	}
	g.pausedResolutionTail = &resolutionTail{item: item, cipherAsked: true}
	return true
}

// encodeCipherCardLocked exiles the spell card `cardID` from the stack
// encoded on `creatureID`, which must still be a creature `chooser`
// controls on the battlefield; otherwise nothing happens and the
// resolution's tail routes the card as usual. The link is written only
// once the card has really reached exile.
//
// Caller must hold g.mu in write mode.
func (g *Game) encodeCipherCardLocked(cardID, creatureID, chooser uuid.UUID) {
	if g.Stack == nil || !g.Stack.Contains(cardID) {
		return
	}
	g.RecomputeLayersIfStaleLocked()
	host := findBattlefieldCard(g, creatureID)
	if host == nil || !host.IsCreature() || host.Controller != chooser {
		return
	}
	ref := PermissionCardRef{ID: host.InstanceID, Epoch: host.ObjectEpoch}
	if err := g.ExileCardForEffect(cardID); err != nil {
		return
	}
	if c := exiledCardByIDLocked(g, cardID); c != nil {
		c.EncodedOn = ref
		c.EncodedAt = timeNowUnixNano()
		g.layerVersion.Add(1)
	}
}

// EncodedOnForEffect is the creature object the card `cardID` is encoded
// on, and whether it is encoded at all: in exile, linked, and the linked
// object still on the battlefield (CR 702.99c).
//
// Caller must hold g.mu.
func (g *Game) EncodedOnForEffect(cardID uuid.UUID) (PermissionCardRef, bool) {
	c := exiledCardByIDLocked(g, cardID)
	if c == nil || c.EncodedOn.ID == uuid.Nil {
		return PermissionCardRef{}, false
	}
	host := findBattlefieldCard(g, c.EncodedOn.ID)
	if host == nil || host.ObjectEpoch != c.EncodedOn.Epoch {
		return PermissionCardRef{}, false
	}
	return c.EncodedOn, true
}

// cipherGrantEffect is one encoded card's static ability (CR 702.99a's
// second), as a layer-6 continuous effect on the object it is encoded on.
type cipherGrantEffect struct {
	card, owner uuid.UUID
	host        PermissionCardRef
	timestamp   int64
}

func (e cipherGrantEffect) Layer() (Layer, SubLayer) { return Layer6Ability, 0 }
func (e cipherGrantEffect) Timestamp() int64         { return e.timestamp }
func (e cipherGrantEffect) AppliesTo(target *Card, _ *Game) bool {
	return target != nil && target.InstanceID == e.host.ID && target.ObjectEpoch == e.host.Epoch
}
func (e cipherGrantEffect) Apply(c *Characteristic, _ *Card, _ *Game) {
	c.GrantAbility(CipherGrantKey, e.card, e.owner)
}
func (e cipherGrantEffect) RemovesAbilities() bool      { return false }
func (e cipherGrantEffect) ContinuesAfterRemoval() bool { return false }

// cipherContinuousEffectsLocked is the layer pass's source list for
// cipher: one grant per card in exile that is encoded on something. A
// link to an object no longer on the battlefield applies to nothing.
//
// Caller must hold g.mu in write mode (the layer pass does).
func (g *Game) cipherContinuousEffectsLocked() []ContinuousEffect {
	if g.Exile == nil {
		return nil
	}
	var out []ContinuousEffect
	for i := range g.Exile.Cards {
		c := &g.Exile.Cards[i]
		if c.EncodedOn.ID == uuid.Nil {
			continue
		}
		out = append(out, cipherGrantEffect{card: c.InstanceID, owner: c.Owner, host: c.EncodedOn, timestamp: c.EncodedAt})
	}
	return out
}

// cipherTriggersFor is the triggered ability each cipher grant on `c`
// is: one per encoded card, each about its own card.
func cipherTriggersFor(c *Card) []TriggeredAbility {
	var out []TriggeredAbility
	key := GrantKey(CipherGrantKey)
	for _, gr := range layeredGrants(c) {
		if gr.Key == key {
			out = append(out, cipherTriggerFor(gr.Source))
		}
	}
	return out
}

// cipherTriggerFor is "Whenever this creature deals combat damage to a
// player, you may copy the encoded card and you may cast the copy
// without paying its mana cost", about the card `encoded`.
func cipherTriggerFor(encoded uuid.UUID) TriggeredAbility {
	return TriggeredAbility{
		Keyword: KeywordCipher,
		Watches: []EventKind{EventDealDamage},
		AppliesTo: func(ev Event, source *Card, _ Characteristic, g *Game) bool {
			return ev.Combat && ev.Amount > 0 && ev.Source == source.InstanceID && g.playerByIDLocked(ev.Target) != nil
		},
		Build: func(_ Event, source *Card, _ Characteristic, g *Game) *StackItem {
			name := "the encoded card"
			if c := exiledCardByIDLocked(g, encoded); c != nil && c.Name != "" {
				name = c.Name
			}
			return NewKeyedTriggeredItem(source, "Cipher — you may cast a copy of "+name+" without paying its mana cost",
				CipherCastBody, EffectParams{Object: ObjectRef{ID: encoded}})
		},
	}
}

// castCipherCopyLocked resolves a cipher trigger: the encoded card, if
// it is still in exile, is copied and the copy offered to the trigger's
// controller free. A card that has left exile since the trigger is no
// longer there to copy (CR 608.2b's reasoning), so nothing is offered.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) castCipherCopyLocked(item *StackItem, cardID uuid.UUID) {
	if item == nil || exiledCardByIDLocked(g, cardID) == nil {
		return
	}
	offer := CardCopyOffer{
		Player: item.Controller,
		Source: item.SourceCardID,
		Cost:   "{0}",
		Timing: TimingFlash,
		Lapse:  LapseStaysInExile,
		// CR 702.99a, CR 707.12, CR 608.2g: the copy is cast while the
		// trigger resolves (#4243).
		DuringResolution: true,
	}
	if host := findBattlefieldCard(g, item.SourceCardID); host != nil {
		offer.SourceName = host.Name
	}
	g.CopyCardsAndOfferCastsForEffect(offer, []uuid.UUID{cardID})
}
