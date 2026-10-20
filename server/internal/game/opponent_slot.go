package game

import "github.com/google/uuid"

// opponent_slot.go — #4128: "for each opponent, <verb> target permanent
// that player controls" with a REQUIRED target per opponent (Blatant
// Thievery).
//
// opponent_clauses.go (effects) already had the "up to one" form of this
// (Windgrace's Judgment, Lich's Relic): one clause per possible
// opponent, each bound to the n-th seat after the chooser's, and every
// clause optional, so a seat that is not at the table simply offers
// nothing. A REQUIRED clause cannot be written that way. A two-player
// table has one opponent, and a clause bound to a seat that does not
// exist would make the spell uncastable; an "up to one" clause would
// let the caster skip an opponent who has a legal permanent, which is
// stronger than printed (CR 601.2c requires a target for each).
//
// So the clause carries its slot as DATA (TargetSpec.OpponentSlot) and
// the answer to "is this clause required right now" is read from the
// table:
//
//   - the candidate must be controlled by the seat in that slot
//     (opponentSlotAdmits, folded into the one shared walk);
//   - the clause's lower bound is its Min while that seat holds a live
//     opponent, and 0 once the seat is absent or eliminated, because an
//     eliminated player is no longer an opponent (CR 800.4a) and a
//     spell that says "each opponent" does not reach them.
//
// A live opponent who controls nothing targetable leaves Min at 1, so
// the spell is uncastable, which is what CR 601.2c says.

// OpponentSeatAfter is the n-th seat (1-based) after `chooser`'s in turn
// order, or nil when there is no such seat at the table or its player
// has been eliminated. Eliminated seats keep their place, so the
// numbering never shifts mid-game.
//
// Caller must hold g.mu (read or write).
func (g *Game) OpponentSeatAfter(chooser uuid.UUID, n int) *Player {
	seats := len(g.Seats)
	if n < 1 || n >= seats {
		return nil
	}
	at := -1
	for i, p := range g.Seats {
		if p != nil && p.ID == chooser {
			at = i
			break
		}
	}
	if at < 0 {
		return nil
	}
	p := g.Seats[(at+n)%seats]
	if p == nil || p.Eliminated {
		return nil
	}
	return p
}

// opponentSlotAdmits is OpponentSlot's candidate gate: true without the
// slot; with it, a candidate controlled by the live opponent in that
// slot. An absent seat admits nothing.
func (s *TargetSpec) opponentSlotAdmits(g *Game, src TargetSource, c Card) bool {
	if s.OpponentSlot < 1 {
		return true
	}
	p := g.OpponentSeatAfter(src.Controller, s.OpponentSlot)
	return p != nil && c.Controller == p.ID
}

// clauseMinLocked is the lower bound of a clause for `chooser` right
// now: its Min, except that a slot-bound clause whose opponent is not at
// the table is not required. Every reader that judges a clause's
// required-ness (the announce gate, the cast offer, the wire view and
// the bot's enumerator) goes through this one answer.
//
// Caller must hold g.mu.
func (g *Game) clauseMinLocked(chooser uuid.UUID, c *TargetClause) int {
	if c.OpponentSlot > 0 && g.OpponentSeatAfter(chooser, c.OpponentSlot) == nil {
		return 0
	}
	return c.Min
}

// ClauseMinForEffect is clauseMinLocked for a caller holding the lock
// (the wire projection and the enumerator).
func (g *Game) ClauseMinForEffect(chooser uuid.UUID, c *TargetClause) int {
	return g.clauseMinLocked(chooser, c)
}
