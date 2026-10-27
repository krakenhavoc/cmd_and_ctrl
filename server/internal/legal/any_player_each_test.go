package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// any_player_each_test.go — ADR 0187 §7 (#3572): a leg of a per-object
// any-player offer (Cleansing) is a choose_cards whose moves are priced
// in life, named for the card, led by "pay for none" and the suggested
// set, and ordered so the expansion budget is spent on the chooser's own
// lands that nobody has chosen yet.

type eachTable struct {
	g     *game.Game
	lands [2][]uuid.UUID
}

func newEachTable(t *testing.T) *eachTable {
	t.Helper()
	g := newTable(t)
	known := map[uuid.UUID]bool{}
	for _, s := range g.Seats {
		known[s.ID] = true
	}
	controller := g.Seats[2]
	source := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: source, Name: "Cleansing", TypeLine: "Sorcery", Owner: controller.ID, Controller: controller.ID, KnownBy: known})
	et := &eachTable{g: g}
	var all []uuid.UUID
	for i := 0; i < 2; i++ {
		for _, name := range []string{"Forest", "Island"} {
			id := uuid.New()
			g.Battlefield.PushTop(game.Card{
				InstanceID: id, Name: name, TypeLine: "Basic Land — " + name,
				Owner: g.Seats[i].ID, Controller: g.Seats[i].ID, KnownBy: known,
			})
			et.lands[i] = append(et.lands[i], id)
			all = append(all, id)
		}
	}
	g.WithWriteLock(func() {
		if err := g.QueueAnyPlayerPaysForEachForEffect(game.AnyPlayerPaysForEach{
			Source: source, Controller: controller.ID, Objects: all, EachLife: 1, Noun: "land",
			// #4361 S7b: the consequence is registered (the closure form
			// is gone); this one does nothing.
			Then: game.RegisterAnyPlayerPaysEach("test/4361-s7b-each-legal-"+uuid.NewString(),
				func(*game.Game, game.AnyPlayerEachAnswer) error { return nil }),
		}); err != nil {
			t.Fatalf("QueueAnyPlayerPaysForEachForEffect: %v", err)
		}
	})
	return et
}

func (et *eachTable) answer(t *testing.T, seat int, picks ...uuid.UUID) {
	t.Helper()
	c := et.g.PendingChoices[0]
	if c.Chooser != et.g.Seats[seat].ID {
		t.Fatalf("open leg is seat %v's, want seat %d's", c.Chooser, seat)
	}
	if err := et.g.ResolveChooseCards(c.ID, c.Chooser, picks); err != nil {
		t.Fatalf("seat %d answer: %v", seat, err)
	}
}

func TestPerObjectOfferLegLeadsWithNoneThenTheSuggestedSetPricedInLife(t *testing.T) {
	et := newEachTable(t)
	g := et.g
	moves := legal.EnumerateFor(g, g.Seats[0].ID)
	if len(moves) < 3 {
		t.Fatalf("only %d moves: %v", len(moves), labels(moves))
	}
	if moves[0].Label != "Cleansing: pay for none" || !moves[0].AlwaysLegal {
		t.Fatalf("first move = %q (always legal %v), want the always-legal \"pay for none\"", moves[0].Label, moves[0].AlwaysLegal)
	}
	if moves[1].Label != "Cleansing: pay 2 life for Forest, Island" {
		t.Fatalf("second move = %q, want the suggested set (the chooser's own two lands)", moves[1].Label)
	}
	if moves[1].Cost == nil || moves[1].Cost.Life != 2 {
		t.Fatalf("the suggested set's cost = %+v, want 2 life", moves[1].Cost)
	}
	for _, m := range moves {
		if strings.Contains(m.Label, "choose") {
			t.Errorf("move %q reads like an ordinary pick, not a payment", m.Label)
		}
	}
	// Only the chooser is offered anything: the leg blocks the table.
	if others := legal.EnumerateFor(g, g.Seats[1].ID); len(others) != 0 && others[0].Kind == legal.KindChoice {
		t.Errorf("seat 1 is offered %v while seat 0's leg is open", labels(others))
	}
}

func TestPerObjectOfferLegNamesALandAnEarlierPlayerAlreadyPaysFor(t *testing.T) {
	et := newEachTable(t)
	g := et.g
	et.answer(t, 0, et.lands[0][0])
	moves := legal.EnumerateFor(g, g.Seats[1].ID)
	want := "Cleansing: pay 1 life for Forest (Seat0 also pays for Forest)"
	found := false
	for _, m := range moves {
		if m.Label == want {
			found = true
		}
	}
	if !found {
		t.Fatalf("no move %q among %v", want, labels(moves))
	}
	// The suggested set leaves out nothing of seat 1's own, and it is
	// offered right after "none".
	if moves[1].Label != "Cleansing: pay 2 life for Forest, Island" {
		t.Fatalf("second move = %q, want seat 1's own two lands", moves[1].Label)
	}
	// Seat 0's chosen Forest is still on offer: paying twice is legal.
	if !hasLabel(moves, "Cleansing: pay 1 life for Forest (Seat0 also pays") {
		t.Errorf("a land an earlier player chose is not offered at all: %v", labels(moves))
	}
}

// The expansion budget is spent on the chooser's own lands that nobody
// has chosen first, so a tight cap never hides them behind everyone
// else's.
func TestPerObjectOfferLegSpendsTheBudgetOnTheChoosersOwnUnchosenLandsFirst(t *testing.T) {
	et := newEachTable(t)
	g := et.g
	et.answer(t, 0)
	moves := legal.EnumerateForWithOptions(g, g.Seats[1].ID, legal.Options{MaxExpansionPerSource: 3})
	// Seat 1 has two lands of its own, so a set of one or two lands that
	// names one of seat 0's has jumped the queue; only the set of three
	// has to reach past seat 1's board.
	for _, m := range moves[1:] {
		for _, other := range et.lands[0] {
			if strings.Contains(string(m.Params), other.String()) && !strings.HasPrefix(m.Label, "Cleansing: pay 3 life") {
				t.Fatalf("move %q names seat 0's land before seat 1's own are exhausted: %v", m.Label, labels(moves))
			}
		}
	}
	if len(moves) < 3 {
		t.Fatalf("only %d moves with a cap of 3: %v", len(moves), labels(moves))
	}
}
