package game

import "github.com/google/uuid"

// replacement_serials.go — #3482: a paused replacement event names its
// effects by OBJECT, not by position.
//
// A catalog replacement's ReplacementEffectID used to pack its source's
// battlefield index, and a dredge offer's its graveyard index. Both IDs
// outlive the gather that minted them: CR 614.5's once-per-event marks
// (Game.replacementsAppliedThisEvent), a CR 616 prompt's listed effects
// (PendingChoice.ReplacementEffectIDs), an expanded draw's inherited
// history (pendingDraw.applied) and a damage instance's remembered
// redirection order (redirectOrder.order) all hold them across a pause.
// A permanent earlier on the battlefield leaving while one of those was
// open renumbered every later effect, so the marks named the wrong
// effects and an applied Rhox Faithmender doubled the same gain twice.
//
// The IDs now pack a per-game SERIAL for the source object instead of
// its index. The serial is handed out the first time a gather sees the
// object and is keyed by ObjectRef — instance ID plus Card.ObjectEpoch —
// so it follows the object wherever its zone's slice puts it, and a card
// that leaves and comes back is a new object with a new serial
// (CR 400.7). Serials are never reused, so an ID whose object has gone
// names nothing rather than its neighbour.
//
// In the snapshot only beside a keyed CR 616 frame (#4253,
// replacement_resume_keyed.go), whose IDs must name the same objects
// after the restore; a frame that is not keyed is in the continuation
// census, and between prompts no ID outlives the action that minted it,
// so a file with neither mints afresh. Clone and
// RestoreFrom copy the registry, so an undo into an open prompt answers
// against the serials the prompt was built with.

// replacementSerialMinter mints serials for one gather pass.
//
// A gather under a read lock (the auto-tap planner's mana preview) may
// not write the registry. It hands an object the registry has not seen
// a PROVISIONAL serial past the registry's next one instead, in walk
// order, so the IDs are distinct within the pass and the same on every
// pass over an unchanged board, which is all the preview's own applied
// set needs. They are never stored and never reach a prompt.
type replacementSerialMinter struct {
	g           *Game
	readOnly    bool
	provisional int
	seen        map[ObjectRef]int
}

// serial is card's object serial.
func (m *replacementSerialMinter) serial(card *Card) int {
	ref := ObjectRef{ID: card.InstanceID, Epoch: card.ObjectEpoch}
	if s, ok := m.g.replacementObjectSerials[ref]; ok {
		return s
	}
	if !m.readOnly {
		return m.g.mintReplacementObjectSerialLocked(ref)
	}
	if s, ok := m.seen[ref]; ok {
		return s
	}
	if m.seen == nil {
		m.seen = make(map[ObjectRef]int)
	}
	s := m.g.nextReplacementObjectSerial + m.provisional
	m.provisional++
	m.seen[ref] = s
	return s
}

// mintReplacementObjectSerialLocked hands ref the next serial.
//
// The registry is compacted first when it has doubled since the last
// compaction: an object that is in no zone under the epoch it was
// minted for can never be gathered again, so its entry is dead weight.
// Dropping it is safe because the serial is not reused.
//
// Caller must hold g.mu (write).
func (g *Game) mintReplacementObjectSerialLocked(ref ObjectRef) int {
	if g.replacementObjectSerials == nil {
		g.replacementObjectSerials = make(map[ObjectRef]int)
	} else if len(g.replacementObjectSerials) >= g.replacementSerialsCompactAt {
		for r := range g.replacementObjectSerials {
			if c, ok := g.LookupCardForEffect(r.ID); !ok || c.ObjectEpoch != r.Epoch {
				if !g.isEmblemRefLocked(r) {
					delete(g.replacementObjectSerials, r)
				}
			}
		}
		g.replacementSerialsCompactAt = 2*len(g.replacementObjectSerials) + 64
	}
	s := g.nextReplacementObjectSerial
	g.nextReplacementObjectSerial++
	g.replacementObjectSerials[ref] = s
	return s
}

// replacementObjectSerialOfLocked is the serial already handed to card,
// for a reader that resolves an ID back to its source. Read-only.
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) replacementObjectSerialOfLocked(card *Card) (int, bool) {
	s, ok := g.replacementObjectSerials[ObjectRef{ID: card.InstanceID, Epoch: card.ObjectEpoch}]
	return s, ok
}

// cloneReplacementObjectSerials copies the registry onto its own map:
// the live game inserts into it as it mints, and an undo snapshot may
// be restored more than once.
func cloneReplacementObjectSerials(in map[ObjectRef]int) map[ObjectRef]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[ObjectRef]int, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// liveReplacementSourceLocked re-reads a gathered catalog effect's
// source as the object it was gathered from, for a resume that runs
// after a pause.
//
// The battlefield walk hands each catalog effect a pointer into
// Battlefield.Cards, and a permanent leaving shifts that slice, so after
// a pause the pointer can name the permanent that slid into the slot.
// The source must still be on the battlefield as the SAME object: a
// static ability's effect exists only while its source is there
// (CR 400.7), so ok is false once it has left, even if the
// card has since come back as a new object. Every other kind of effect
// is returned as gathered: a self-replacement's and a dredge offer's
// source is already a copy, and the rest have none.
//
// Caller must hold g.mu.
func (g *Game) liveReplacementSourceLocked(a activeReplacement) (activeReplacement, bool) {
	if a.lost {
		// #4253: a restored frame's effect the running binary no longer
		// has. It has gone, as a source that left takes its effect.
		return a, false
	}
	if a.sourceRef.ID == uuid.Nil {
		return a, true
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID == a.sourceRef.ID && c.ObjectEpoch == a.sourceRef.Epoch {
			a.source = c
			return a, true
		}
	}
	// An emblem's effect (#4192): the emblem is in its owner's command
	// zone and nothing removes it, so it is found the same way.
	found := false
	g.forEachEmblemLocked(func(c *Card) bool {
		if c.InstanceID == a.sourceRef.ID && c.ObjectEpoch == a.sourceRef.Epoch {
			a.source, found = c, true
			return false
		}
		return true
	})
	return a, found
}

// emblemWithSerialLocked is the emblem holding the object serial, for a
// reader that resolves a catalog replacement ID back to its source.
//
// Caller must hold g.mu (a read lock is enough).
func (g *Game) emblemWithSerialLocked(serial int) *Card {
	var out *Card
	g.forEachEmblemLocked(func(c *Card) bool {
		if s, ok := g.replacementObjectSerialOfLocked(c); ok && s == serial {
			out = c
			return false
		}
		return true
	})
	return out
}

// isEmblemRefLocked reports whether ref names an emblem that exists, so
// the serial registry's compaction keeps its entry: an emblem is in no
// card index (LookupCardForEffect cannot find one).
//
// Caller must hold g.mu.
func (g *Game) isEmblemRefLocked(ref ObjectRef) bool {
	found := false
	g.forEachEmblemLocked(func(c *Card) bool {
		if c.InstanceID == ref.ID && c.ObjectEpoch == ref.Epoch {
			found = true
			return false
		}
		return true
	})
	return found
}
