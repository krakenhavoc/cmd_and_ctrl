package game

import "github.com/google/uuid"

// token_existence.go is CR 704.5d: "If a token is in a zone other
// than the battlefield, it ceases to exist."
//
// Why it is a state-based action and not a check inside the movers
// (#596). A token that is bounced, destroyed, sacrificed, exiled or
// shuffled away DOES reach the new zone — briefly. CR 111.7 is
// explicit that it exists there long enough for everything that
// looks at the move to see it: the "when this dies" and
// leaves-the-battlefield triggers are judged on a token that got as
// far as the graveyard, and "put into a graveyard from anywhere"
// payoffs count it. Refusing the move at the mover, or deleting the
// token inside it, would silently lose all of that. So the token
// lands, the events go out, and the next state-based check removes
// it — which is exactly the order the rules describe.
//
// Doing it once here rather than per-mover is the same argument
// zone_route.go makes for the CR 903.9 window: there are a dozen
// ways off the battlefield and a thirteenth is written every sprint.
// The bug this fixes — a bot's Cyclonic Rift bouncing a Bird token
// and handing its controller a real card in hand, card advantage
// conjured from nothing — was only ever the hand's instance of a gap
// that every destination had.
//
// Token identity itself lives on Card.IsToken (card.go), where #745
// put it. The catalog's effects.IsToken has always delegated to it,
// so the engine and the card catalog cannot disagree about what a
// token is, and the SBA needs no notion of its own.

// tokenCeaseToExistSBALocked performs CR 704.5d over every zone but
// the battlefield. Reports whether it removed anything, which keeps
// the SBA loop going for one more pass (CR 704.3).
//
// The stack is swept with the rest for completeness. No engine path
// puts a token there today — a copy of a spell is a copy, not a
// token (CR 707.10), and copies already cease to exist through
// ceaseToExistLocked in spell_copy.go — but "every zone other than
// the battlefield" is what the rule says, and leaving one zone out
// is how this class of gap started.
//
// Removal is a two-pass walk: collect first, then remove and
// announce. EmitEvent runs the listener chain synchronously and a
// listener may move cards, so mutating a zone's slice while still
// iterating it would be reading a board the previous removal's
// listeners had already changed.
//
// Caller must hold g.mu.
func (g *Game) tokenCeaseToExistSBALocked() bool {
	type departure struct {
		id    uuid.UUID
		from  ZoneKind
		actor uuid.UUID
	}
	var doomed []departure
	collect := func(z *Zone, actor uuid.UUID) {
		if z == nil {
			return
		}
		for _, c := range z.Cards {
			if c.IsToken() {
				doomed = append(doomed, departure{id: c.InstanceID, from: z.Kind, actor: actor})
			}
		}
	}
	collect(g.Stack, uuid.Nil)
	collect(g.Exile, uuid.Nil)
	for _, p := range g.Seats {
		collect(p.Library, p.ID)
		collect(p.Hand, p.ID)
		collect(p.Graveyard, p.ID)
		collect(p.Command, p.ID)
	}
	if len(doomed) == 0 {
		return false
	}

	gone := make([]departure, 0, len(doomed))
	for _, d := range doomed {
		// Re-find rather than trusting the collected zone: nothing
		// has emitted yet, but the collection above is a snapshot and
		// re-finding is what makes it safe to stop being one.
		z := g.findCardZoneLocked(d.id)
		if z == nil || z.Kind == ZoneBattlefield {
			continue
		}
		if _, err := z.Remove(d.id); err != nil {
			continue
		}
		d.from = z.Kind
		if z.Kind == ZoneStack {
			delete(g.StackMeta, d.id)
		}
		gone = append(gone, d)
	}
	if len(gone) == 0 {
		return false
	}
	for _, d := range gone {
		// An EventZoneMove with an empty NewZone is the engine's
		// "moved to nowhere" — the same shape ceaseToExistLocked
		// emits for a spell copy. The client's zone views drop the
		// object; nothing in the engine watches for a move to
		// nowhere, which is the point.
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			Actor:   d.actor,
			CardID:  d.id,
			OldZone: d.from,
		})
	}
	return true
}
