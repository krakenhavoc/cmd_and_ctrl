package game

// entry_counters_if_type.go — #4193: a putting effect's "each of them
// enters with an additional <counter> on it if it's a <type>" clause
// (Semester's End: "an additional +1/+1 counter on it if it's a
// creature and an additional loyalty counter on it if it's a
// planeswalker").
//
// ZoneEntryOptions.EntersWithCounters gives every card of a batch the
// same counters. This clause gives each card its own, by what the card
// is as it enters, and it has to stay ONE batch: splitting the return
// into a creature batch and a planeswalker batch would make it two
// events, and every permanent it puts onto the battlefield must see the
// others enter (CR 603.6a).
//
// What the card "is" is CR 614.12's question: the characteristics of
// the permanent as it would exist on the battlefield, taking into
// account the replacement effects that have already modified how it
// enters, its own static abilities and the continuous effects that
// already exist. That is the entry look-ahead's dry run
// (entry_lookahead.go), so a Spark Double coming back as a copy of a
// planeswalker gets the loyalty counter and not the +1/+1 counter, and
// a permanent that would be both a creature and a planeswalker gets
// both. The other members of the same batch are not on the battlefield
// for the dry run, which is right: their statics do not exist yet.
//
// The clause is read once per member, when that member's CR 614 window
// has settled on the battlefield (recordEntryBatchMemberLocked), and
// folded into the settled event's EntersWithCounters. Applying it after
// the window's other replacements is an order the affected object's
// controller may choose (CR 616.1e), and it is the only order in which
// the copy a Clone was told to enter as is already known (CR 616.1c
// puts the copy first). A member a replacement sent somewhere else
// (CR 614.6) gets nothing.
//
// The counters then go through the ordinary drain
// (applyEntryCountersLocked), so a counter replacement such as
// Hardened Scales or Doubling Season applies to them (CR 122.6), and an
// additional loyalty counter is put on top of the printed starting
// loyalty rather than instead of it (CR 306.5b).

// EntryCountersIfType is one "an additional N <Counter> counters on it
// if it's a <CardType>" clause of a putting effect. Data, so a batch
// held behind an entry question carries no closure.
type EntryCountersIfType struct {
	// CardType is the card type the entering permanent must have, in
	// lowercase as Card.HasCardType reads it: "creature",
	// "planeswalker".
	CardType string
	// Subtype is the subtype the entering permanent must have, in
	// the casing the card's type line prints ("Hero"; Heroic Return's
	// "If a Hero enters this way"), read from the permanent as it would
	// exist on the battlefield (CR 614.12), so a static that makes it a
	// Hero there counts. When CardType is also set the permanent must
	// have both. Empty asks nothing about subtypes.
	Subtype string `json:"Subtype,omitempty"`
	// Counter is the counter kind (CounterPlusOne, CounterLoyalty).
	Counter string
	// N is how many. Zero or less adds nothing.
	N int
}

// seedEntryCountersIfTypeLocked folds the putting effect's typed
// clauses into one member's settled entry event. A face-down entry
// takes none: no putting effect combines the two, and a face-down
// permanent's characteristics are CR 708.2's, which the look-ahead does
// not compute.
//
// Caller must hold g.mu (write): the look-ahead's dry pass assigns
// layer caches on a copy of the battlefield.
func (g *Game) seedEntryCountersIfTypeLocked(clauses []EntryCountersIfType, ev *ReplacementEvent) {
	if len(clauses) == 0 || !entryTypeClauseApplies(ev) {
		return
	}
	perm, ok := g.entryPermanentForTypeClausesLocked(ev)
	if !ok {
		return
	}
	for _, cl := range clauses {
		if cl.N <= 0 || cl.Counter == "" || (cl.CardType == "" && cl.Subtype == "") {
			continue
		}
		if (cl.CardType == "" || perm.HasCardType(cl.CardType)) &&
			(cl.Subtype == "" || perm.HasSubtype(cl.Subtype)) {
			ev.AddCounterAtETB(cl.Counter, cl.N)
		}
	}
}

// entryTypeClauseApplies reports whether a putting effect's per-card type
// clauses read this settled event: a live move onto the battlefield, face
// up.
func entryTypeClauseApplies(ev *ReplacementEvent) bool {
	return ev != nil && !ev.Canceled && ev.Kind == RepEventMove &&
		ev.NewZone == ZoneBattlefield && ev.FaceDown == FaceDownNone
}

// entryPermanentForTypeClausesLocked is what the entering card would be on
// the battlefield (CR 614.12), for the typed clauses to read.
func (g *Game) entryPermanentForTypeClausesLocked(ev *ReplacementEvent) (Card, bool) {
	perm, ok := g.entryCharacteristicsLocked(ev)
	if ok {
		return perm, true
	}
	// No dry run (the card cannot be found, or the pass computed
	// nothing for it): the card as it is now, with the copy it
	// was told to enter as.
	c, found := g.LookupCardForEffect(ev.CardID)
	if !found {
		return Card{}, false
	}
	if ev.EntersAsCopyOf != nil {
		c.applyCopy(*ev.EntersAsCopyOf, Card{})
	}
	return c, true
}

// seedEntryTappedIfTypeLocked folds the putting effect's "if a <type> enters
// this way, it enters tapped" into one member's settled entry event (Silver
// Surfer, Cosmic Voyager's land clause): a member that would be one of the
// card types on the battlefield (CR 614.12) enters tapped, and the others
// are untouched. It stays one batch for the same reason the counters
// clause does (CR 603.6a).
//
// Caller must hold g.mu (write), as seedEntryCountersIfTypeLocked does.
func (g *Game) seedEntryTappedIfTypeLocked(types []string, ev *ReplacementEvent) {
	if len(types) == 0 || ev == nil || ev.EntersTapped || !entryTypeClauseApplies(ev) {
		return
	}
	perm, ok := g.entryPermanentForTypeClausesLocked(ev)
	if !ok {
		return
	}
	for _, t := range types {
		if t != "" && perm.HasCardType(t) {
			ev.EntersTapped = true
			return
		}
	}
}
