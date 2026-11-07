package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// collect_evidence_event_test.go — #5223: "whenever you collect evidence"
// (EventCollectEvidence) as the cast path emits it, and the collect
// evidence X gate (CR 701.59a, 107.3f). The card files' tests, with the
// ability cost, the effect forms and the four cards, are
// cards/effects/mkm_collect_evidence_x_cards_test.go.

// collectEvidenceEvents is every EventCollectEvidence in the log.
func collectEvidenceEvents(g *Game) []Event {
	var out []Event
	for _, ev := range g.Events {
		if ev.Kind == EventCollectEvidence {
			out = append(out, ev)
		}
	}
	return out
}

// A cast that takes the collect evidence offer emits one event, after the
// cards are in exile, naming the caster, the spell and the cards with their
// total mana value.
func TestCollectEvidenceCostEmitsOneEventWithTheCards(t *testing.T) {
	withCollectEvidence(t, 4)
	g := newActiveGame(t)
	me := g.Seats[0]
	gy := evidenceInGraveyard(me, "{2}{U}", "{1}", "{5}")
	named := []uuid.UUID{gy[1], gy[0]}
	spell, err := castCollectingEvidence(t, g, me, true, named)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	evs := collectEvidenceEvents(g)
	if len(evs) != 1 {
		t.Fatalf("%d collect evidence events, want 1", len(evs))
	}
	ev := evs[0]
	if ev.Actor != me.ID || ev.Source != spell || ev.Amount != 4 || !slices.Equal(ev.CardIDs, named) {
		t.Errorf("event %+v, want actor %s, source %s, amount 4, cards %v", ev, me.ID, spell, named)
	}
	for _, e := range g.Events[slices.IndexFunc(g.Events, func(e Event) bool { return e.Kind == EventCollectEvidence }):] {
		if e.Kind == EventZoneMove && slices.Contains(named, e.CardID) {
			t.Errorf("card %s moved after the collection was announced", e.CardID)
		}
	}
}

// Declining the offer, and a refused cast, collect nothing and emit nothing.
func TestCollectEvidenceCostDeclinedOrRefusedEmitsNothing(t *testing.T) {
	t.Run("declined", func(t *testing.T) {
		withCollectEvidence(t, 4)
		g := newActiveGame(t)
		me := g.Seats[0]
		evidenceInGraveyard(me, "{5}")
		if _, err := castCollectingEvidence(t, g, me, false, nil); err != nil {
			t.Fatalf("cast: %v", err)
		}
		if n := len(collectEvidenceEvents(g)); n != 0 {
			t.Errorf("%d events for a declined offer", n)
		}
	})
	t.Run("short", func(t *testing.T) {
		withCollectEvidence(t, 4)
		g := newActiveGame(t)
		me := g.Seats[0]
		gy := evidenceInGraveyard(me, "{1}", "{2}")
		if _, err := castCollectingEvidence(t, g, me, true, gy); err == nil {
			t.Fatal("collect evidence 4 with a total of 3 was accepted")
		}
		if n := len(collectEvidenceEvents(g)); n != 0 {
			t.Errorf("%d events for a refused cast", n)
		}
	})
}

// Collect evidence X: every graveyard card may be chosen, X = 0 is always
// met (no card, an empty graveyard, or cards with no mana value), X up to
// the chosen cards' total is met, and an X above it, a negative X, a card
// not in the graveyard and a card named twice are refused.
func TestCollectEvidenceXGate(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if opts := g.CollectEvidenceXOptionsForEffect(me.ID); len(opts) != 0 {
		t.Errorf("an empty graveyard offers %v", opts)
	}
	if !g.CollectEvidenceXMetForEffect(me.ID, nil, 0) {
		t.Error("X = 0 with no card refused on an empty graveyard")
	}
	gy := evidenceInGraveyard(me, "{2}{U}", "", "{4}")
	theirs := evidenceInGraveyard(opp, "{7}")
	if opts := g.CollectEvidenceXOptionsForEffect(me.ID); !slices.Equal(sortedEvidenceIDs(opts), sortedEvidenceIDs(gy)) {
		t.Errorf("options %v, want the whole graveyard %v", opts, gy)
	}
	if ceiling, ok := g.CollectEvidenceXCeilingForEffect(me.ID, gy); !ok || ceiling != 7 {
		t.Errorf("ceiling %d (%v), want 7", ceiling, ok)
	}
	for _, tc := range []struct {
		name string
		ids  []uuid.UUID
		x    int
		want bool
	}{
		{"X = 0, no card", nil, 0, true},
		{"X = 0, a card with no mana value", []uuid.UUID{gy[1]}, 0, true},
		{"X = the total", []uuid.UUID{gy[0], gy[2]}, 7, true},
		{"X below the total", []uuid.UUID{gy[2]}, 2, true},
		{"X above the total", []uuid.UUID{gy[0], gy[2]}, 8, false},
		{"X above the whole graveyard", gy, 8, false},
		{"a negative X", nil, -1, false},
		{"another player's card", []uuid.UUID{theirs[0]}, 0, false},
		{"a card named twice", []uuid.UUID{gy[2], gy[2]}, 4, false},
	} {
		if got := g.CollectEvidenceXMetForEffect(me.ID, tc.ids, tc.x); got != tc.want {
			t.Errorf("%s: met %v, want %v", tc.name, got, tc.want)
		}
	}
}

func sortedEvidenceIDs(ids []uuid.UUID) []uuid.UUID {
	out := slices.Clone(ids)
	slices.SortFunc(out, func(a, b uuid.UUID) int { return slices.Compare(a[:], b[:]) })
	return out
}
