package game

import "github.com/google/uuid"

// goad.go is the goad marker (CR 701.15): who has goaded a creature,
// and until when. #1598, ADR 0045 amendment of 2026-09-28, Decision 53.
//
// CR 701.15a: goading a creature makes it goaded by that player until
// that player's next turn. CR 701.15b: a goaded creature attacks each
// combat if able and attacks a player other than the goading player if
// able. CR 701.15c: a creature can be goaded by several players, and it
// then carries every goader's requirements, so it attacks a player who
// goaded it NONE of those times if it can. So the marker is a set with
// one entry per goader, and each entry has its own end.
//
// The requirements themselves are judged in attack_requirements.go,
// which adds two per entry. Nothing here decides an attack.

// Goad is one player's goad on one creature.
type Goad struct {
	// By is the goading player.
	By uuid.UUID
	// ExpiresAtTurnsBegun is the value of By's Player.TurnsBegun at
	// which the goad is over: that player's next turn (CR 701.15a),
	// stamped as TurnsBegun+1 when the goad is made — the same counter
	// and the same stamp as an UntilYourNextTurn Duration, so a goad
	// made on the goader's own turn and one made on somebody else's
	// both end as the goader's next turn begins, and a goader who has
	// left ends it when that turn would have begun (CR 800.4m).
	ExpiresAtTurnsBegun int
}

// IsGoaded reports whether any player's goad is on the creature.
func (c Card) IsGoaded() bool { return len(c.Goads) > 0 }

// IsGoadedBy reports whether `player` has goaded the creature and that
// goad has not ended.
func (c Card) IsGoadedBy(player uuid.UUID) bool {
	for _, gd := range c.Goads {
		if gd.By == player {
			return true
		}
	}
	return false
}

// Goaders lists every player whose goad is on the creature, oldest goad
// first. Nil when it is not goaded.
func (c Card) Goaders() []uuid.UUID {
	if len(c.Goads) == 0 {
		return nil
	}
	out := make([]uuid.UUID, len(c.Goads))
	for i, gd := range c.Goads {
		out[i] = gd.By
	}
	return out
}

// LatestGoader is the player whose goad is the most recent (a refresh
// counts as a new goad), or uuid.Nil. It is what the pre-#1598 single
// marker held, and it is what the legacy snapshot key and the legacy
// wire field still carry.
func (c Card) LatestGoader() uuid.UUID {
	if len(c.Goads) == 0 {
		return uuid.Nil
	}
	return c.Goads[len(c.Goads)-1].By
}

// cloneGoads copies the slice so a clone never aliases the live card's
// backing array: cloneCard's `out := c` would otherwise share it with
// every undo snapshot, and a goad appended after the snapshot would
// appear in it.
func cloneGoads(in []Goad) []Goad {
	if len(in) == 0 {
		return nil
	}
	return append([]Goad(nil), in...)
}

// goadLocked records `by`'s goad on `c`, stamped to end as `by`'s next
// turn begins. A player who has already goaded the creature refreshes
// that goad — its end moves to the new stamp and it becomes the latest
// — rather than adding a second entry: two goads by one player are
// still one set of CR 701.15b requirements. Caller must hold g.mu.
func (g *Game) goadLocked(c *Card, by uuid.UUID) {
	if c == nil || by == uuid.Nil {
		return
	}
	entry := Goad{By: by, ExpiresAtTurnsBegun: g.turnsBegunForLocked(by) + 1}
	kept := c.Goads[:0:0]
	for _, gd := range c.Goads {
		if gd.By != by {
			kept = append(kept, gd)
		}
	}
	c.Goads = append(kept, entry)
}

// GoadForEffect goads the battlefield creature `cardID` on behalf of
// `by` (see goadLocked), for a card effect that already holds the lock.
// It reports false when there is no such permanent, which is not an
// error — the target left in response.
//
// Caller must hold g.mu.
func (g *Game) GoadForEffect(cardID, by uuid.UUID) bool {
	c, ok := g.battlefieldCardLocked(cardID)
	if !ok || by == uuid.Nil {
		return false
	}
	g.goadLocked(c, by)
	return true
}

// goadExpiredLocked reports whether one goad has ended. Caller must hold
// g.mu.
func (g *Game) goadExpiredLocked(gd Goad) bool {
	return g.turnsBegunForLocked(gd.By) >= gd.ExpiresAtTurnsBegun
}

// sweepExpiredGoadsLocked drops every goad whose goader's next turn has
// begun (CR 701.15a), one entry at a time — a creature goaded by two
// players loses only the goad of the player whose turn it is. It runs
// from onTurnBeganLocked, where every other "until your next turn"
// sweep runs, and covers phased-out permanents too: CR 702.26d keeps a
// phased-out permanent's state, but a duration still runs out while it
// is away.
//
// Caller must hold g.mu.
func (g *Game) sweepExpiredGoadsLocked() {
	for _, z := range []*Zone{g.Battlefield, g.PhasedOut} {
		if z == nil {
			continue
		}
		for i := range z.Cards {
			c := &z.Cards[i]
			if len(c.Goads) == 0 {
				continue
			}
			kept := c.Goads[:0:0]
			for _, gd := range c.Goads {
				if !g.goadExpiredLocked(gd) {
					kept = append(kept, gd)
				}
			}
			if len(kept) == 0 {
				kept = nil
			}
			c.Goads = kept
		}
	}
}

// ExpireGoadsForEffect runs the turn-start goad sweep on behalf of the
// legacy "goad — the goad ends" delayed trigger (effects'
// clearListedGoadsBody). In this binary the sweep has already run as
// the goader's turn began, so the trigger ends nothing; it is still
// scheduled so that a restore point read by a binary from before #1598,
// which knows only the latest goader and ends it with that trigger,
// does not keep the goad forever. It never ends a goad made after the
// goader's turn began (a refresh in their upkeep), because that goad's
// stamp is the NEXT turn.
//
// Caller must hold g.mu.
func (g *Game) ExpireGoadsForEffect() { g.sweepExpiredGoadsLocked() }

// backfillLegacyGoadsLocked stamps the end of every goad a restore
// point from before #1598 carried as the single `goadedBy` ID. Such a
// file has no stamp, and the right one is recomputable: the goad ends
// as that goader's next turn begins, and the goader's TurnsBegun as the
// file was written is the turn the goad was made in or after, so
// TurnsBegun+1 is that next turn — exactly what goadLocked would have
// stamped. Runs after the seats are restored. Caller must hold g.mu (or
// own the game exclusively, as restore does).
func (g *Game) backfillLegacyGoadsLocked() {
	for _, z := range []*Zone{g.Battlefield, g.PhasedOut} {
		if z == nil {
			continue
		}
		for i := range z.Cards {
			for j := range z.Cards[i].Goads {
				gd := &z.Cards[i].Goads[j]
				if gd.ExpiresAtTurnsBegun == 0 {
					gd.ExpiresAtTurnsBegun = g.turnsBegunForLocked(gd.By) + 1
				}
			}
		}
	}
}

// goadSnapshot is one Goad on disk (#1598). The card's legacy
// `goadedBy` key is still written beside the list, holding
// LatestGoader, so a v7 binary from before #1598 — which drops `goads`
// — restores the one goad it can represent, the one it would have kept
// itself when a second goad replaced the first.
type goadSnapshot struct {
	By                  uuid.UUID `json:"by"`
	ExpiresAtTurnsBegun int       `json:"expiresAtTurnsBegun"`
}

func snapshotGoads(in []Goad) []goadSnapshot {
	if len(in) == 0 {
		return nil
	}
	out := make([]goadSnapshot, len(in))
	for i, gd := range in {
		out[i] = goadSnapshot(gd)
	}
	return out
}

// restoreGoads reads a card's goads back. A file written before #1598
// has no `goads` key and at most one goader, in `goadedBy`: that comes
// back as a one-entry set with no stamp yet, which
// backfillLegacyGoadsLocked supplies once the seats are restored.
func restoreGoads(in []goadSnapshot, legacy uuid.UUID) []Goad {
	if len(in) == 0 {
		if legacy == uuid.Nil {
			return nil
		}
		return []Goad{{By: legacy}}
	}
	out := make([]Goad, len(in))
	for i, gd := range in {
		out[i] = Goad(gd)
	}
	return out
}
