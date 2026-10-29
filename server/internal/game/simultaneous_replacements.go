package game

import "github.com/google/uuid"

// simultaneous_replacements.go — #4521: a replacement whose source leaves
// in the same simultaneous event still applies to the rest of it.
//
// A wipe destroys every creature at once (CR 608.2f), and one state-based
// action pass is one event (CR 704.3). A replacement effect has to exist
// before the event it replaces (CR 614.4), and Liesa, Forgotten
// Archangel's "if a creature an opponent controls would die, exile it
// instead" exists when Wrath of God's destruction begins. So every
// opponent's creature in the wipe is exiled, Liesa's own death included
// in the same event.
//
// The engine moves the legs of that event one at a time
// (simultaneous.go), and the CR 614 gather read only the live
// battlefield. A source processed early in the batch had already left
// when its neighbours' legs asked, so they went to the graveyard
// unreplaced. The trigger harvester had the same problem and solved it by
// reading the batch's pre-move copies (harvestSimultaneousExitLocked);
// the gather now reads the same copies.
//
// What is gathered. While a battlefield exit batch is open and the event
// is one of its legs (a move off the battlefield of a card in the batch),
// every batch member that has ALREADY left contributes its catalog
// replacements, read off the copy taken when the batch opened. That copy
// is the object as it was when the event began, so its characteristics,
// its controller and any ability removal on it are the ones that applied
// then. A member still on the battlefield is gathered by the live walk
// and is skipped here; the leg's own card is skipped too, since its own
// effects are the live walk's.
//
// Where the batch is read from. A fire-and-forget batch is published on
// the game for its whole loop (Game.simultaneousExit). A keyed batch's
// destroy legs open their window before the batch is published, and its
// copies ride the leg's route instead (zoneRoute.simultaneousExit), which
// is also what a restore writes. The route's copies win when both exist.
//
// Who chooses. Nothing changes: a departed source's effect joins the
// gathered set like any other, and CR 616.1 gives the order to the
// affected object's controller (affectedPlayerForEvent).
//
// IDs. A departed source's effects are named by its object serial, as
// every catalog effect is (#3482). The copy is the object that was on the
// battlefield, under the epoch it had there, so its serial is the one the
// live walk handed it. The serial registry's compaction keeps the refs of
// open batches (replacementRefInOpenBatchLocked), so a pause does not
// renumber them.
//
// A pause. activeReplacement.departed marks the effect, and its source is
// the copy. liveReplacementSourceLocked returns it as gathered, because
// the source did not leave before the event: it left in it. A keyed frame
// regathers it on restore from the route's copies. A frame whose event
// does not carry the batch (a fire-and-forget leg) is not written to the
// disk while it holds a departed effect, since the restored table could
// not gather it again.
//
// Umbra armor keeps its ordering (hostsBeforeAttachmentsLocked). Its
// effect reads the Aura on the battlefield, and its instruction destroys
// the Aura, so the host's leg must still find the Aura there.

// departedBatchSourcesLocked is the pre-move copies of the batch members
// that have already left the battlefield, when ev is one leg of an open
// simultaneous battlefield exit. Nil otherwise.
//
// Caller must hold g.mu.
func (g *Game) departedBatchSourcesLocked(ev *ReplacementEvent) []Card {
	if ev == nil || ev.Kind != RepEventMove || ev.OldZone != ZoneBattlefield || ev.CardID == uuid.Nil {
		return nil
	}
	batch := g.simultaneousExit
	if ev.zoneRoute != nil && len(ev.zoneRoute.simultaneousExit) > 0 {
		batch = ev.zoneRoute.simultaneousExit
	}
	if len(batch) < 2 {
		return nil
	}
	member := false
	for i := range batch {
		if batch[i].InstanceID == ev.CardID {
			member = true
			break
		}
	}
	if !member {
		return nil
	}
	var out []Card
	for i := range batch {
		c := batch[i]
		if c.InstanceID == ev.CardID || findCardOnBattlefield(g, c.InstanceID) >= 0 {
			continue
		}
		out = append(out, c)
	}
	return out
}

// departedReplacementSourceLocked is the copy a paused frame gathered a
// departed source's effect from, for the prompt's label. Nil when no
// open frame holds one by that ID.
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) departedReplacementSourceLocked(id ReplacementEffectID) *Card {
	for _, c := range g.PendingChoices {
		if c == nil || c.replacementResume == nil {
			continue
		}
		for _, a := range c.replacementResume.applicable {
			if a.id == id && a.departed && a.source != nil {
				return a.source
			}
		}
	}
	return nil
}

// frameDepartedUnregatherable reports whether a paused frame holds a
// departed source's effect that a restored table could not gather again:
// its event does not carry the batch's copies.
func frameDepartedUnregatherable(f *replacementResumeFrame) bool {
	if f == nil {
		return false
	}
	if r := f.ev.zoneRoute; r != nil && len(r.simultaneousExit) > 0 {
		return false
	}
	for _, a := range f.applicable {
		if a.departed {
			return true
		}
	}
	return false
}

// replacementRefInOpenBatchLocked reports whether ref is a member of an
// open battlefield exit batch, published or riding a paused frame's
// route, so the serial registry's compaction keeps it.
//
// Caller must hold g.mu.
func (g *Game) replacementRefInOpenBatchLocked(ref ObjectRef) bool {
	in := func(batch []Card) bool {
		for i := range batch {
			if batch[i].InstanceID == ref.ID && batch[i].ObjectEpoch == ref.Epoch {
				return true
			}
		}
		return false
	}
	if in(g.simultaneousExit) {
		return true
	}
	for _, c := range g.PendingChoices {
		if c == nil || c.replacementResume == nil || c.replacementResume.ev == nil {
			continue
		}
		if r := c.replacementResume.ev.zoneRoute; r != nil && in(r.simultaneousExit) {
			return true
		}
		for _, a := range c.replacementResume.applicable {
			if a.departed && a.sourceRef == ref {
				return true
			}
		}
	}
	return false
}
