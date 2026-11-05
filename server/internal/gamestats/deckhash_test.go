package gamestats

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// deckhash_test.go is ADR 0217 §5's deck version.

func TestDeckHash(t *testing.T) {
	sol, bolt := "4cbc6901-6a4a-4d0a-83ea-7eefa3b35021", "4457ed35-7c10-48c8-9776-456485fdf070"
	a := DeckHash(map[string]int{sol: 1, bolt: 2})
	// Pinned: the hash is stored in seat_stats and compared with one
	// computed from a saved deck later, so it may never drift.
	if a != "ed05af03e6000a4c" {
		t.Errorf("DeckHash = %q; the definition changed (sha256 of sorted \"<oracle_id> <count>\" lines)", a)
	}
	if len(a) != 16 {
		t.Errorf("DeckHash is %d characters, want 16", len(a))
	}
	if b := DeckHash(map[string]int{bolt: 2, sol: 1}); b != a {
		t.Error("the order of the entries changed the hash")
	}
	if b := DeckHash(map[string]int{"4CBC6901-6A4A-4D0A-83EA-7EEFA3B35021": 1, bolt: 2}); b != a {
		t.Error("the case of an oracle id changed the hash")
	}
	if b := DeckHash(map[string]int{sol: 1, bolt: 3}); b == a {
		t.Error("one more copy did not change the hash")
	}
	if b := DeckHash(map[string]int{sol: 1, bolt: 2, "": 4, "x": 0}); b != a {
		t.Error("an entry with no id or no copies changed the hash")
	}
	if DeckHash(nil) != "" {
		t.Error("an empty deck has a hash")
	}
}

func TestSeatDeckHashIsByOracleNotPrinting(t *testing.T) {
	oracle := uuid.New()
	p1, p2, tok := uuid.New(), uuid.New(), uuid.New()
	idx := fakeIndex{
		p1:  {ID: p1, OracleID: oracle},
		p2:  {ID: p2, OracleID: oracle},
		tok: {ID: tok, OracleID: uuid.New()},
	}
	in := func(ids ...protocol.CardView) map[string]protocol.CardView {
		m := map[string]protocol.CardView{}
		for _, c := range ids {
			m[c.InstanceID] = c
		}
		return m
	}
	mine := func(id string, sf uuid.UUID, opts ...func(*protocol.CardView)) protocol.CardView {
		c := protocol.CardView{InstanceID: id, Owner: ana, ScryfallID: sf.String()}
		for _, o := range opts {
			o(&c)
		}
		return c
	}
	one := seatDeckHash(ana, in(mine("a", p1), mine("b", p1)), idx)
	other := seatDeckHash(ana, in(mine("a", p1), mine("b", p2),
		mine("t", tok, func(c *protocol.CardView) { c.IsToken = true }),
		mine("c", tok, func(c *protocol.CardView) { c.IsCompanion = true }),
		protocol.CardView{InstanceID: "theirs", Owner: bo, ScryfallID: tok.String()}), idx)
	if one == "" || one != other {
		t.Errorf("two printings of one card, plus a token, a companion and another seat's card: %q vs %q", one, other)
	}
	if one != DeckHash(map[string]int{oracle.String(): 2}) {
		t.Error("the seat's hash is not DeckHash of its oracle counts")
	}
	if seatDeckHash(ana, in(mine("a", p1)), nil) != "" {
		t.Error("with no card index there is no hash")
	}
	var _ CardIndex = (*cards.Index)(nil)
}
