package game

import "github.com/google/uuid"

// entry_ordinal.go — ADR 0109 §8 (#1862): the order permanents entered
// the battlefield in, and the world rule that reads it.
//
// # Why an ordinal
//
// CR 704.5k: "If two or more permanents have the supertype world, all
// except the one that has had the world supertype for the shortest
// amount of time are put into their owners' graveyards. In the event
// of a tie for the shortest amount of time, all are put into their
// owners' graveyards." Nothing grants the world supertype, so "had it
// for the shortest time" is "entered the battlefield most recently".
//
// Card.EnteredBattlefieldAt cannot answer that. It is a wall-clock
// stamp taken per event, so two permanents one Replenish puts onto the
// battlefield together get two different stamps (and the later one
// would survive, where the rule says both go), while two separate
// entries in one nanosecond would tie. So the entry gets an ordinal: a
// game counter taken once per MOVE that puts permanents onto the
// battlefield. Every card of one simultaneous entry shares it, and
// every later entry has a strictly larger one.
//
// The stamp rides the one place every entry is already stamped
// (stampBattlefieldEntryLocked, the layer listener's arm for a zone
// move onto the battlefield and a token's creation). A simultaneous
// entry (landEntryBatchLocked) opens a shared ordinal around its
// announcements, so the cards it announces all take that one.

// entryOrdinalForStampLocked is the ordinal the entry being stamped
// takes: the shared one while a simultaneous entry is announcing its
// cards, and otherwise the next one.
//
// Caller must hold g.mu.
func (g *Game) entryOrdinalForStampLocked() int64 {
	if g.entryOrdinalShared != 0 {
		return g.entryOrdinalShared
	}
	g.entryOrdinalSeq++
	return g.entryOrdinalSeq
}

// beginSimultaneousEntryLocked opens one ordinal for every card a
// simultaneous entry is about to announce, and returns the call that
// closes it. A nested call — an entry inside an entry's announcement —
// shares the outer one and closes nothing.
//
// Caller must hold g.mu.
func (g *Game) beginSimultaneousEntryLocked() (release func()) {
	if g.entryOrdinalShared != 0 {
		return func() {}
	}
	g.entryOrdinalSeq++
	g.entryOrdinalShared = g.entryOrdinalSeq
	return func() { g.entryOrdinalShared = 0 }
}

// restoreEntryOrdinalsLocked is the restore half: the counter resumes
// past every ordinal a restored permanent carries — phased-out ones
// included, since phasing is not an entry (CR 702.26d) — so the next
// entry is newer than all of them. A file written before the field
// existed carries none; the world rule orders those permanents by
// their entry stamps instead (worldRuleDoomedLocked), so restore
// rewrites nothing and a file round-trips exactly.
//
// Caller must hold g.mu.
func (g *Game) restoreEntryOrdinalsLocked() {
	var max int64
	for _, z := range []*Zone{g.Battlefield, g.PhasedOut} {
		if z == nil {
			continue
		}
		for i := range z.Cards {
			if z.Cards[i].EntryOrdinal > max {
				max = z.Cards[i].EntryOrdinal
			}
		}
	}
	g.entryOrdinalSeq = max
	g.entryOrdinalShared = 0
}

// worldRuleDoomedLocked is CR 704.5k's collection: when two or more
// permanents on the battlefield have the world supertype, every one
// but the single most recent entrant, and all of them when the most
// recent entry is shared. Read on effective supertypes (HasSupertype),
// so a face-down permanent has none (CR 708.2). A permanent whose exit
// is already paused on a prompt is skipped, as the other state-based
// actions skip it.
//
// Caller must hold g.mu.
func (g *Game) worldRuleDoomedLocked() []uuid.UUID {
	if g.Battlefield == nil {
		return nil
	}
	var worlds []*Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.HasSupertype("world") {
			worlds = append(worlds, c)
		}
	}
	if len(worlds) < 2 {
		return nil
	}
	// The most recent entry: the largest ordinal. A permanent no ordinal
	// was ever stamped on entered before every one that has one — it
	// was restored from a file written before the field existed — so
	// only when NONE of them has one does the order fall back to their
	// entry stamps, the order that binary kept.
	key := func(c *Card) int64 { return c.EntryOrdinal }
	var newest int64
	for _, c := range worlds {
		if c.EntryOrdinal > newest {
			newest = c.EntryOrdinal
		}
	}
	if newest == 0 {
		key = func(c *Card) int64 { return c.EnteredBattlefieldAt }
		for _, c := range worlds {
			if c.EnteredBattlefieldAt > newest {
				newest = c.EnteredBattlefieldAt
			}
		}
	}
	atNewest := 0
	for _, c := range worlds {
		if key(c) == newest {
			atNewest++
		}
	}
	var out []uuid.UUID
	for _, c := range worlds {
		if key(c) == newest && atNewest == 1 {
			continue
		}
		if g.zoneChangePausedLocked(c.InstanceID) {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return out
}
