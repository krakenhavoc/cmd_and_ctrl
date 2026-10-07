package legal_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// proliferate_test.go — #2525: the enumerator's answers to a proliferate
// prompt. The engine accepts every subset, so what this pins is REACH
// and the always-legal answer: the empty pick is offered first, the
// suggested set is offered whole however wide the board, and a seat is
// labelled by name.

func proliferateTable(t *testing.T, creatures int) (*game.Game, *game.Player, *game.Player, []uuid.UUID) {
	t.Helper()
	g := newTable(t)
	me, opp := g.Seats[0], g.Seats[1]
	var ids []uuid.UUID
	for i := 0; i < creatures; i++ {
		c := game.Card{
			InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		}
		for _, seat := range g.Seats {
			c.AddKnowersAll([]uuid.UUID{seat.ID})
		}
		g.Battlefield.PushTop(c)
		if err := g.AddCounter(c.InstanceID, game.CounterPlusOne, 1); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, c.InstanceID)
	}
	if err := g.AddPlayerCounter(opp.ID, game.CounterPoison, 1); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() {
		if err := g.ProliferateChoosingForEffect(me.ID, uuid.Nil, nil); err != nil {
			t.Fatalf("ProliferateChoosingForEffect: %v", err)
		}
	})
	return g, me, opp, ids
}

func cardIDsOf(t *testing.T, m legal.Move) []string {
	t.Helper()
	var p struct {
		CardIDs []string `json:"card_ids"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params: %v", err)
	}
	return p.CardIDs
}

func TestProliferateEnumerationOffersTheEmptyAnswerAndTheSuggestedSetFirst(t *testing.T) {
	g, me, opp, ids := proliferateTable(t, 2)
	moves := legal.EnumerateFor(g, me.ID)
	if len(moves) == 0 {
		t.Fatal("the proliferating seat was offered nothing: the #544 wedge")
	}

	var sawEmpty, sawAlwaysLegal bool
	for _, m := range moves {
		if len(cardIDsOf(t, m)) == 0 {
			sawEmpty = true
			sawAlwaysLegal = m.AlwaysLegal
		}
	}
	if !sawEmpty {
		t.Error("\"choose nothing\" is not offered, and \"any number\" includes none")
	}
	if !sawAlwaysLegal {
		t.Error("the empty answer is not marked always-legal")
	}

	// The suggested set — both creatures and the poisoned seat — is the
	// first non-empty answer and is labelled with the seat's name.
	var first *legal.Move
	for i := range moves {
		if len(cardIDsOf(t, moves[i])) > 0 {
			first = &moves[i]
			break
		}
	}
	if first == nil {
		t.Fatal("no non-empty answer offered")
	}
	got := map[string]bool{}
	for _, id := range cardIDsOf(t, *first) {
		got[id] = true
	}
	for _, id := range append(ids, opp.ID) {
		if !got[id.String()] {
			t.Errorf("the first answer leaves out %v: %v", id, got)
		}
	}
	if len(got) != 3 {
		t.Errorf("the first answer names %d things, want the 3 suggested", len(got))
	}
	if !strings.Contains(first.Label, opp.Name) {
		t.Errorf("label %q does not name the seat %q", first.Label, opp.Name)
	}
}

func TestProliferateEnumerationReachesTheSuggestedSetPastTheCap(t *testing.T) {
	g, me, _, ids := proliferateTable(t, 16)
	want := len(ids) + 1 // every creature and the poisoned seat
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if len(cardIDsOf(t, m)) == want {
			return
		}
	}
	t.Fatalf("no offered answer names all %d suggested things", want)
}
