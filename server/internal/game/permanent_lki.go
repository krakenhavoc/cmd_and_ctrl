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
		return permanentInfoOf(c), true
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
