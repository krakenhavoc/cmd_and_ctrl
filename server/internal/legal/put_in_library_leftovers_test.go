package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// put_in_library_leftovers_test.go — #1298, ADR 0088's 2026-09-23
// amendment. The prompt kind is unchanged, so the table obligations are
// the same ones: the seat owing the answer is offered answers, every
// answer offered dispatches, and nobody else is offered anything. What
// is new is which answers are legal (an exact top count) and who the
// chooser is (not the cards' owner).

// An exact top count is not something "leave it alone" answers: every
// offered answer holds exactly the count on top — one per card that
// could be the one — and every one dispatches.
func TestPutInLibraryTopCountOffersOnlyExactAnswers(t *testing.T) {
	g := newTable(t)
	me := g.Seats[0]
	ids := make([]uuid.UUID, 0, 3)
	for _, name := range []string{"A", "B", "C"} {
		id := uuid.New()
		ids = append(ids, id)
		me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	}
	g.WithWriteLock(func() {
		if err := g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
			Chooser: me.ID, Cards: ids, From: game.ZoneLibrary,
			Placement: game.LibraryPlaceTopOrBottom, TopCount: 1,
		}); err != nil {
			t.Fatalf("queue: %v", err)
		}
	})
	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) != 3 {
		t.Fatalf("%d answers, want one per card that could stay on top: %v", len(moves), labels(moves))
	}
	for _, m := range moves {
		var p struct {
			TopOrder []string `json:"top_order"`
			Bottom   []string `json:"bottom"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("params: %v", err)
		}
		if len(p.TopOrder) != 1 || len(p.Bottom) != 2 {
			t.Errorf("%q holds %d on top and %d under — want exactly 1 and 2", m.Label, len(p.TopOrder), len(p.Bottom))
		}
	}
	dispatchAll(t, g, me.ID, moves)
}

// The looker orders ANOTHER player's library, and a counter's chooser
// is not the spell's owner. Either way only the chooser is offered
// anything, and everything offered dispatches.
func TestPutInLibraryChooserIsNotTheOwner(t *testing.T) {
	t.Run("a look at another player's library", func(t *testing.T) {
		g := newTable(t)
		me, them := g.Seats[0], g.Seats[1]
		g.WithWriteLock(func() {
			looked := g.LookAtTopOfPlayersLibraryForEffect(me.ID, them.ID, 2)
			if err := g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
				Chooser: me.ID, Cards: looked, From: game.ZoneLibrary, Placement: game.LibraryPlaceTop,
			}); err != nil {
				t.Fatalf("queue: %v", err)
			}
		})
		if others := legal.EnumerateFor(g, them.ID); len(others) != 0 {
			t.Errorf("the library's owner was offered %v", labels(others))
		}
		moves := legal.EnumerateFor(g, me.ID)
		if len(moves) != 2 {
			t.Errorf("%d answers, want leave-alone and the second to the front: %v", len(moves), labels(moves))
		}
		dispatchAll(t, g, me.ID, moves)
	})
	t.Run("a counter to top or bottom", func(t *testing.T) {
		g := newTable(t)
		caster, counterer := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		spell := handCard(caster, game.Card{Name: "Their Spell", TypeLine: "Sorcery"})
		advanceTo(t, g, game.StepPrecombatMain)
		if err := g.CastSpell(caster.ID, spell, game.CastSpellParams{}); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		g.WithWriteLock(func() {
			if err := g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
				Chooser: counterer.ID, Cards: []uuid.UUID{spell}, Placement: game.LibraryPlaceTopOrBottom, Counter: true,
			}); err != nil {
				t.Fatalf("queue: %v", err)
			}
		})
		if others := legal.EnumerateFor(g, caster.ID); len(others) != 0 {
			t.Errorf("the spell's owner was offered %v", labels(others))
		}
		moves := legal.EnumerateFor(g, counterer.ID)
		if len(moves) != 2 {
			t.Errorf("%d answers, want top and bottom: %v", len(moves), labels(moves))
		}
		dispatchAll(t, g, counterer.ID, moves)
	})
}
