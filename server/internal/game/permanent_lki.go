package game

import "github.com/google/uuid"

// permanent_lki.go — resolution-time last-known information for a
// permanent that has LEFT the battlefield (#1379, CR 608.2h).
//
// # The rule
//
// CR 608.2h: when an effect needs information about an object and that
// object has left the zone it was expected in, the effect uses the
// object's last-known information. Cream of the Crop's "look at the top
// X cards, where X is that creature's power" is read when the ability
// RESOLVES. If the creature was killed in response, X is the power it
// had as it last existed on the battlefield, not zero and not its power
// from when the ability triggered.
//
// # Why the CR 603.10 store could not answer
//
// The engine already snapshots a leaving permanent: snapshotLKILocked
// writes Game.lastKnownBattlefield (and lastKnownCounters) just before
// every battlefield exit. Those entries are for the trigger HARVEST —
// CR 603.10's "look back in time" — and harvestLTB deletes them the
// moment the dispatch for that exit ends. An ability that resolves a
// priority round later finds a hole. They also carry a Characteristic,
// which leaves +1/+1 and -1/-1 counters out, so the power a counter made
// was lost twice over.
//
// # The record
//
// Game.lastKnownPermanents keeps one PermanentInfo per departed
// permanent OBJECT, written at the same moment and from the same live
// card as the CR 603.10 snapshot (battlefieldExitLocked, with the card
// still on the battlefield and its layer cache as it stood). It is keyed
// by instance ID and holds a list, one entry per battlefield object that
// card has been this turn, told apart by Card.ObjectEpoch (CR 400.7). A
// creature blinked twice in response to a trigger is two objects, and
// the trigger names only the first.
//
// It lives for the rest of the turn and is cleared at the turn boundary
// with lastKnownStack, for the reason that record gives: an ability that
// refers to a permanent is a stack object or a pending trigger, and the
// stack is empty before a turn can end. The size is bounded by the
// turn's battlefield exits. Clone and the persisted snapshot carry it, so
// an undo across the removal spell rewinds it and a restore mid-stack
// still has the answer.
//
// # The reader
//
// PermanentForEffect(ref) is the one read. It returns the LIVE permanent
// while the named object is still on the battlefield, so a pump in
// response counts, and the record once it has gone. A card that has come
// back is a new object, so a live permanent with a different epoch is
// not the one the ref names: the reader answers with the record, never
// with the impostor.

// ObjectRef names one OBJECT: a card instance plus the CR 400.7 epoch it
// had (Card.ObjectEpoch). The instance ID survives a zone change; the
// pair does not, which is what lets a resolving ability tell the
// creature it was about from the same card after a blink.
type ObjectRef struct {
	ID    uuid.UUID `json:"id"`
	Epoch int       `json:"epoch"`
}

// cloneObjectRefs copies a list of refs onto its own backing array, so
// an undo snapshot or a restore point never shares one with the live
// game. Nil in, nil out.
func cloneObjectRefs(refs []ObjectRef) []ObjectRef {
	if len(refs) == 0 {
		return nil
	}
	return append([]ObjectRef(nil), refs...)
}

// stamped is the ref as a snapshot field: nil for the zero ref, so an
// unstamped item writes nothing and an old snapshot reads back as
// unstamped (#1418). The ID is the "stamped" bit; epoch zero is real.
func (r ObjectRef) stamped() *ObjectRef {
	if r.ID == uuid.Nil {
		return nil
	}
	return &r
}

// value is the inverse of stamped.
func (r *ObjectRef) value() ObjectRef {
	if r == nil {
		return ObjectRef{}
	}
	return *r
}

// PermanentInfo is one battlefield permanent's state, live or as it last
// existed there.
type PermanentInfo struct {
	// Epoch is the object's Card.ObjectEpoch while it was on the
	// battlefield.
	Epoch int `json:"epoch"`

	// Left reports that the object has left the battlefield and the
	// rest of the struct is its last-known information. False for a
	// live read. Always true on a stored record.
	Left bool `json:"left,omitempty"`

	// Controller is who controlled the permanent.
	Controller uuid.UUID `json:"controller"`

	// Characteristic is the permanent's post-layer characteristics
	// (Effective()), the same value the CR 603.10 snapshot holds.
	Characteristic Characteristic `json:"characteristic"`

	// Power and Toughness include +1/+1 and -1/-1 counters and are NOT
	// clamped (Card.PowerForComparison, Card.CurrentToughness). A
	// reader that deals damage clamps at zero itself.
	Power     int `json:"power"`
	Toughness int `json:"toughness"`

	// Counters is a copy of the counters it had. Nil when it had none.
	Counters map[string]int `json:"counters,omitempty"`

	// AttachedTo is what it was attached to (CR 301.5c / 303.4). An
	// Aura's "enchanted creature" reads it after the Aura has gone.
	AttachedTo TargetRef `json:"attachedTo,omitempty"`

	// Tapped is its tapped status (CR 110.5). Not a characteristic,
	// but last-known information all the same: Mana Vault's "if this
	// artifact is tapped" is re-checked when its draw-step trigger
	// resolves (CR 603.4), and a Vault that has left by then is judged
	// as it last existed (#1396). Read live while it is there;
	// MoveCard clears the flag on the way out, a line after the
	// record is written.
	Tapped bool `json:"tapped,omitempty"`

	// Delved is the permanent's CR 607.2q link to the cards delve
	// exiled to pay for the spell that became it (CastProvenance.Delved,
	// ADR 0100 sub-PR 2). Last-known information like the rest of the
	// record: Ethereal Forager's attack trigger "can still find the
	// cards exiled with Ethereal Forager's delve ability" after the
	// Forager has left the battlefield (its 2020-04-17 ruling), and
	// this is where it finds them. Resolve it with
	// Game.DelvedCardsForEffect. Nil for a permanent that delved
	// nothing.
	Delved []ObjectRef `json:"delved,omitempty"`

	// CraftedWith is the permanent's CR 702.167c link to the materials
	// its craft ability exiled (Card.CraftedWith, ADR 0137), as it last
	// existed: a crafted permanent's own trigger that resolves after it
	// has left still finds "the exiled cards used to craft it". Resolve
	// it with Game.CraftMaterialsForEffect. Nil for a permanent no
	// craft ability put onto the battlefield.
	CraftedWith []ObjectRef `json:"craftedWith,omitempty"`

	// ChosenColor and NamedTribe are the colour and creature type chosen
	// for it as it entered (Card.ChosenColor, Card.NamedTribe), as it last
	// existed: Story Circle's and Circle of Solace's "of the chosen
	// color/type" still name the choice when the ability resolves after
	// the enchantment has gone (ADR 0107 §6, CR 608.2h). Empty when none
	// was chosen.
	ChosenColor string `json:"chosenColor,omitempty"`
	NamedTribe  string `json:"namedTribe,omitempty"`

	// ChosenOption is the option chosen for it as it entered
	// (Card.ChosenOption, CR 614.12), as it last existed: Apex
	// Observatory's tap ability still names its card type when it
	// resolves after the Observatory has gone (CR 608.2h, #2709).
	ChosenOption string `json:"chosenOption,omitempty"`

	// ChosenNumber is the number chosen for it as it entered
	// (Card.ChosenNumber, #1941), as it last existed: Phyrexian
	// Processor's token is still the size of the life paid when the
	// Processor is sacrificed in response to its own ability (CR 608.2h).
	ChosenNumber int `json:"chosenNumber,omitempty"`

	// ManaValue is the permanent's mana value as it last existed on the
	// battlefield (#1600): Food Chain's "1 plus the exiled creature's
	// mana value". Read off the card's mana cost while it was a
	// permanent, so a Clone copying a six-drop is six (a copy effect
	// rewrites the cost, CR 707.2, and leaving the battlefield undoes
	// it), and a face-down permanent, which has no mana cost, is zero.
	// A token that is no copy has no mana cost either, and is zero.
	ManaValue int `json:"manaValue,omitempty"`

	// RingBearer is whether it was its controller's Ring-bearer (CR
	// 701.54e) as it last existed on the battlefield. Read by
	// "protection from Ring-bearers" when the source has since left
	// (#2145); the designation is cleared on leaving (CR 400.7), so
	// only this record remembers it.
	RingBearer bool `json:"ringBearer,omitempty"`

	// Suspected is whether it was suspected (CR 701.60) as it last
	// existed on the battlefield. Agency Coroner's "if the sacrificed
	// creature was suspected" reads it after the cost has moved the
	// creature to a graveyard, where the designation is gone (CR
	// 400.7); only this record remembers it.
	Suspected bool `json:"suspected,omitempty"`

	// Renowned is whether it was renowned (CR 702.112b) as it last
	// existed on the battlefield (#2049). "Whenever this creature
	// attacks, if it's renowned" (Consul's Lieutenant) and "if this
	// creature is renowned" (Scab-Clan Berserker) re-check the
	// condition as they resolve (CR 603.4), and a source that has left
	// by then is read as it last existed (CR 608.2h).
	Renowned bool `json:"renowned,omitempty"`

	// Attacking is whether it was an attacking creature (CR 508.1k) as
	// it last existed on the battlefield, and Enchanted whether an Aura
	// was attached to it (#2026). A damage shield against "attacking
	// creatures" prevents a departed creature's noncombat damage only
	// "if it was an attacking creature at the time it left" (Heavy Fog's
	// ruling, CR 608.2h); leaving the battlefield removes it from combat
	// (CR 506.4), so only this record remembers it.
	Attacking bool `json:"attacking,omitempty"`
	Enchanted bool `json:"enchanted,omitempty"`

	// Blocking is every attacker it was blocking, in declaration order
	// (Card.BlockedAttackers), as it last existed on the battlefield.
	// Leaving the battlefield removes it from combat (CR 506.4), so
	// only this record remembers it: Goblin Snowman's "target creature
	// it's blocking" is still judged against these when the Snowman has
	// died in response (#1863, TargetSpec.CombatWithSource). Nil when
	// it was not blocking.
	Blocking []uuid.UUID `json:"blocking,omitempty"`
}

// permanentManaValue is PermanentInfo.ManaValue's reading of a
// battlefield card.
func permanentManaValue(c *Card) int {
	if c.FaceDown {
		return 0
	}
	return c.ManaValue()
}

// permanentInfoOf reads a battlefield card into a PermanentInfo. The
// caller decides whether the layer cache is fresh enough.
func permanentInfoOf(c *Card) PermanentInfo {
	return PermanentInfo{
		Epoch:          c.ObjectEpoch,
		Controller:     c.Controller,
		Characteristic: c.Effective(),
		Power:          c.PowerForComparison(),
		Toughness:      c.CurrentToughness(),
		Counters:       copyStringIntMap(c.Counters),
		AttachedTo:     c.AttachedTo,
		Tapped:         c.Tapped,
		Delved:         append([]ObjectRef(nil), c.Provenance.Delved...),
		CraftedWith:    cloneObjectRefs(c.CraftedWith),
		ChosenColor:    c.ChosenColor,
		ChosenOption:   c.ChosenOption,
		NamedTribe:     c.NamedTribe,
		ChosenNumber:   c.ChosenNumber,
		ManaValue:      permanentManaValue(c),
		RingBearer:     IsRingBearerOf(*c, c.Controller),
		Suspected:      c.Suspected,
		Renowned:       c.Renowned,
		Attacking:      c.AttackingTarget != uuid.Nil,
		Blocking:       c.BlockedAttackers(),
	}
}

// rememberDepartingPermanentLocked records cardID as it stands on the
// battlefield, just before it leaves. Called from battlefieldExitLocked
// beside the CR 603.10 snapshot, and deliberately WITHOUT a layer
// recompute: in a board wipe the permanents leave one at a time, and a
// recompute part-way through would read a creature after its lord had
// already gone, which is not how it last existed. The CR 603.10
// snapshot and the paid-tap freeze read the same cache for the same
// reason. A card that is not on the battlefield records nothing.
//
// Caller must hold g.mu in write mode.
func (g *Game) rememberDepartingPermanentLocked(cardID uuid.UUID) {
	c := findBattlefieldCard(g, cardID)
	if c == nil {
		return
	}
	info := permanentInfoOf(c)
	info.Left = true
	info.Enchanted = g.isEnchantedLocked(cardID)
	if g.lastKnownPermanents == nil {
		g.lastKnownPermanents = make(map[uuid.UUID][]PermanentInfo)
	}
	recs := g.lastKnownPermanents[cardID]
	for i := range recs {
		if recs[i].Epoch == info.Epoch {
			// The same object cannot leave twice, so a second write for
			// one epoch means the first exit did not complete (a
			// replacement kept it where it was). The later reading is
			// the one closer to its real departure.
			recs[i] = info
			return
		}
	}
	g.lastKnownPermanents[cardID] = append(recs, info)
}

// lastKnownPermanentLocked returns the record for one departed object,
// or false when the turn has none.
func (g *Game) lastKnownPermanentLocked(ref ObjectRef) (PermanentInfo, bool) {
	for _, rec := range g.lastKnownPermanents[ref.ID] {
		if rec.Epoch == ref.Epoch {
			return rec, true
		}
	}
	return PermanentInfo{}, false
}

// PermanentRefForEffect names the permanent OBJECT a card is, or most
// recently was: the live object while the card is on the battlefield,
// and otherwise the last battlefield object it was this turn. A trigger
// takes this as it is built so that its resolution reads the object it
// was about, even if that card has come back as a new one since.
//
// False when the card is not on the battlefield and has not left it
// this turn.
//
// Caller must hold g.mu.
func (g *Game) PermanentRefForEffect(cardID uuid.UUID) (ObjectRef, bool) {
	if c := findBattlefieldCard(g, cardID); c != nil {
		return ObjectRef{ID: cardID, Epoch: c.ObjectEpoch}, true
	}
	recs := g.lastKnownPermanents[cardID]
	if len(recs) == 0 {
		return ObjectRef{}, false
	}
	return ObjectRef{ID: cardID, Epoch: recs[len(recs)-1].Epoch}, true
}

// LastKnownPermanentForEffect is the record of the card's most recent
// DEPARTED battlefield object this turn (CR 603.10 / 608.2h), as it
// stood on the battlefield: post-layer characteristics, counter-aware
// power and toughness, and the counters it had. It answers "what did
// the permanent that just left look like", for a dies trigger's fill-in
// Build or intervening-if and for a sacrifice-cost ability's effect;
// unlike PermanentForEffect it never answers with a live object, so a
// card that has come back is not mistaken for the one that left. An
// ability that was handed the object (Trigger.Object) should prefer
// PermanentForEffect on that ref.
//
// False when the card has not left the battlefield this turn.
//
// Caller must hold g.mu.
func (g *Game) LastKnownPermanentForEffect(cardID uuid.UUID) (PermanentInfo, bool) {
	recs := g.lastKnownPermanents[cardID]
	if len(recs) == 0 {
		return PermanentInfo{}, false
	}
	return recs[len(recs)-1], true
}

// PermanentForEffect is the resolution-time read of a permanent an
// effect refers to (CR 608.2h): the permanent as it is NOW while the
// named object is still on the battlefield, and as it last existed there
// once it has left. Left on the answer says which.
//
// The live half recomputes the layer cache first, so a pump or an anthem
// that arrived while the ability waited counts. A live card whose epoch
// differs from the ref is a different object (CR 400.7) and is never
// the answer.
//
// False when the object is neither on the battlefield nor recorded this
// turn — it left by a route that bypasses the battlefield exit (a player
// leaving the game, CR 800.4a), or the ref never named a permanent.
//
// Caller must hold g.mu in write mode.
func (g *Game) PermanentForEffect(ref ObjectRef) (PermanentInfo, bool) {
	g.RecomputeLayersIfStaleLocked()
	if c := findBattlefieldCard(g, ref.ID); c != nil && c.ObjectEpoch == ref.Epoch {
		info := permanentInfoOf(c)
		info.Enchanted = g.isEnchantedLocked(c.InstanceID)
		return info, true
	}
	return g.lastKnownPermanentLocked(ref)
}

// clearLastKnownPermanentsLocked forgets the turn's records. Called at
// the turn boundary. Caller must hold g.mu.
func (g *Game) clearLastKnownPermanentsLocked() {
	g.lastKnownPermanents = nil
}

// forgetLastKnownPermanentLocked drops every record of one card. Called
// when the card leaves the GAME (CR 800.4a): there is no object left to
// know anything about.
func (g *Game) forgetLastKnownPermanentLocked(cardID uuid.UUID) {
	if g.lastKnownPermanents == nil {
		return
	}
	delete(g.lastKnownPermanents, cardID)
	if len(g.lastKnownPermanents) == 0 {
		g.lastKnownPermanents = nil
	}
}

// cloneLastKnownPermanents copies the record for Clone, RestoreFrom and
// the persisted snapshot. Each card's list gets its own backing array,
// because rememberDepartingPermanentLocked appends to it and may
// overwrite an entry in place. Nothing reaches INTO an entry after it is
// stored (its Counters map and Characteristic slices are only read), so
// copying the entries by value suffices — the argument
// lastKnownBattlefield's clone makes.
func cloneLastKnownPermanents(src map[uuid.UUID][]PermanentInfo) map[uuid.UUID][]PermanentInfo {
	if len(src) == 0 {
		return nil
	}
	out := make(map[uuid.UUID][]PermanentInfo, len(src))
	for k, v := range src {
		out[k] = append([]PermanentInfo(nil), v...)
	}
	return out
}
