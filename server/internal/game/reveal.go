package game

import "github.com/google/uuid"

// reveal.go is the public half of the S13.5 knowledge model.
//
// "Look at" and "reveal" are two different instructions and the
// engine has only ever had the first. Scry, surveil and Sensei's
// Divining Top tell ONE player something — lookAtTopForEffect marks
// the chooser and only the chooser a knower, and protocol's
// per-viewer filter redacts the cards for everyone else. A reveal
// tells the TABLE, and, the part that makes it a primitive rather
// than a louder version of the same thing, every player is entitled
// to remember it afterwards (CR 701.16).
//
// A reveal therefore has two halves, and both live here:
//
//   - The knowledge. Every seated player joins the card's KnownBy
//     set — the same set a move into a public zone writes and the
//     same set protocol.FilterViewFor reads, not a parallel model.
//     It is sticky, which is what "entitled to remember" means: a
//     card revealed off the top of a library and put into its
//     owner's hand stays readable to every seat while it sits in
//     that hand. It is cleared by a shuffle, which is also right —
//     once the library has been randomised nobody knows where the
//     card went, and clearKnownInZoneLocked already does that.
//
//   - The announcement. One EventRevealCards per card, all stamped
//     with the same RevealSeq, which protocol.publicRevealsOf groups
//     back into one broadcast frame.
//
// The second half is not redundant. Knowledge alone is silent: it
// changes what a client is ALLOWED to see without telling anyone
// that anything happened, and for the common case — reveal the top
// card of your library, then shuffle — the knowledge is gone again
// before the next frame, so the reveal would have been invisible to
// every seat but the revealer's.
//
// What the frame does NOT do is hand out a handle. It carries
// printed identity and no instance ID at all; see protocol/
// reveal_frame.go for why that is the whole design and not an
// oversight.

// RevealSpec describes one reveal: who is showing, what card caused
// it, and which cards the table gets to see.
type RevealSpec struct {
	// Player is whose cards are being shown. Used for the "N
	// revealed ..." line on the client, and it is a seat, never a
	// card handle.
	Player uuid.UUID

	// Source is the card whose effect is doing the revealing (Dark
	// Confidant on the battlefield, Fact or Fiction on the stack).
	// Its NAME reaches the wire, not its ID.
	Source uuid.UUID

	// Reason is the one-line label the client banner shows. Written
	// the way the card is written — "Dark Confidant — reveal the top
	// card of your library".
	Reason string

	// Cards are the instance IDs being revealed, in the order the
	// table should see them. A card that is no longer findable is
	// skipped rather than erroring: a reveal is a look, and a look
	// at something that has left is nothing.
	Cards []uuid.UUID
}

// RevealForEffect performs a reveal: it marks every seated player a
// knower of each named card and announces the whole thing as one
// grouped run of EventRevealCards.
//
// Nothing moves. A reveal is not a zone change (CR 701.16a) and this
// function deliberately has no destination parameter — "reveal the
// top card of your library and put it into your hand" is a reveal
// followed by an ordinary move, in that order, so that the table
// sees the card in the zone it was revealed FROM.
//
// Returns the RevealSeq the run was stamped with, or 0 when nothing
// was revealed (an empty Cards list, or every card already gone).
// Callers do not normally need it; the tests do.
//
// Caller must hold g.mu.
func (g *Game) RevealForEffect(spec RevealSpec) uint64 {
	if len(spec.Cards) == 0 {
		return 0
	}
	seats := make([]uuid.UUID, 0, len(g.Seats))
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		seats = append(seats, p.ID)
	}
	var revealSeq uint64
	for _, id := range spec.Cards {
		zone := g.findCardZoneLocked(id)
		if zone == nil {
			continue
		}
		for i := range zone.Cards {
			if zone.Cards[i].InstanceID != id {
				continue
			}
			zone.Cards[i].AddKnowersAll(seats)
			break
		}
		if revealSeq == 0 {
			// The group key is the Seq the FIRST event of the run
			// will be stamped with. EmitEvent increments before it
			// stamps, so the next Seq is eventSeq+1. Reading it here
			// keeps the key deterministic — a replayed game produces
			// the same event stream byte for byte, which a minted
			// UUID would not.
			revealSeq = g.eventSeq + 1
		}
		g.EmitEvent(Event{
			Kind:      EventRevealCards,
			Actor:     spec.Player,
			Source:    spec.Source,
			CardID:    id,
			Label:     spec.Reason,
			OldZone:   zone.Kind,
			RevealSeq: revealSeq,
		})
	}
	return revealSeq
}

// RevealTopOfLibraryForEffect reveals the top n cards of playerID's
// library and returns their instance IDs, top card first.
//
// The cards stay on the library — this is the reveal and nothing
// else. "Reveal the top card of your library and put it into your
// hand" (Dark Confidant) calls this and then moves the ID it gets
// back; "reveal the top five cards" (Fact or Fiction) calls this and
// then asks somebody a question about them.
//
// A short library reveals what it has; an empty one reveals nothing
// and is not an error. Revealing is not drawing, so running the
// library out this way does NOT set up the CR 704.5b loss.
//
// Caller must hold g.mu.
func (g *Game) RevealTopOfLibraryForEffect(playerID, source uuid.UUID, n int, reason string) []uuid.UUID {
	if n <= 0 {
		return nil
	}
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Library == nil {
		return nil
	}
	size := p.Library.Size()
	if size == 0 {
		return nil
	}
	if n > size {
		n = size
	}
	// The library's top is the LAST element — the same read PopTop
	// and lookAtTopForEffect use. Collected top-first so the table
	// sees them in draw order.
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		ids = append(ids, p.Library.Cards[size-1-i].InstanceID)
	}
	g.RevealForEffect(RevealSpec{
		Player: playerID,
		Source: source,
		Reason: reason,
		Cards:  ids,
	})
	return ids
}
