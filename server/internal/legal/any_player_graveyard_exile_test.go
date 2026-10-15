package legal_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// any_player_graveyard_exile_test.go — ADR 0187 §7 (#3570): an
// any-player link paid by exiling from the chooser's graveyard is a
// card pick the enumerator offers only when the graveyard is deep
// enough, cheapest fuel first.

// graveyardExileOffer queues an offer of "exile n cards from their
// graveyard" and returns the asked seat's moves.
func graveyardExileOffer(t *testing.T, n, inGraveyard int, opts legal.Options) (*game.Game, *game.Player, []legal.Move, []uuid.UUID) {
	t.Helper()
	g := newTable(t)
	controller := g.Seats[1]
	source := uuid.New()
	known := map[uuid.UUID]bool{}
	for _, s := range g.Seats {
		known[s.ID] = true
	}
	g.Battlefield.PushTop(game.Card{InstanceID: source, Name: "Thing", TypeLine: "Creature — Rat", Owner: controller.ID, Controller: controller.ID, KnownBy: known})
	g.WithWriteLock(func() {
		if err := g.QueueAnyPlayerPaysForEffect(game.AnyPlayerPays{
			Source: source, Controller: controller.ID, Cost: "exile cards from their graveyard", Question: "Thing?",
			Action: &game.PayAction{Kind: game.PayActionExileFromGraveyard, Count: n},
		}); err != nil {
			t.Fatalf("queue: %v", err)
		}
	})
	chooser := g.PendingChoices[0].Chooser
	var seat *game.Player
	for _, s := range g.Seats {
		if s.ID == chooser {
			seat = s
		}
	}
	var ids []uuid.UUID
	for i := 0; i < inGraveyard; i++ {
		ids = append(ids, graveyardCard(seat, game.Card{Name: "Fodder " + string(rune('A'+i)), TypeLine: "Sorcery"}))
	}
	return g, seat, legal.EnumerateForWithOptions(g, chooser, opts), ids
}

func TestGraveyardExileLinkOffersNoPaymentFromAShortGraveyard(t *testing.T) {
	_, _, moves, _ := graveyardExileOffer(t, 3, 2, legal.Options{})
	for _, m := range moves {
		if strings.Contains(m.Label, "to stop") && !strings.Contains(m.Label, "decline") {
			t.Errorf("a two-card graveyard was offered %q for a three-card payment", m.Label)
		}
	}
	if len(moves) != 1 || !strings.Contains(moves[0].Label, "decline") {
		t.Fatalf("moves = %v, want only the decline", labels(moves))
	}
}

func TestGraveyardExileLinkPaymentsAreTheCheapestFuelFirst(t *testing.T) {
	g, seat, _, ids := graveyardExileOffer(t, 2, 4, legal.Options{})
	// The policy would miss the first two cards most.
	precious := map[uuid.UUID]bool{ids[0]: true, ids[1]: true}
	order := func(c legal.TargetCandidate) float64 {
		if precious[c.ID] {
			return 100
		}
		return 1
	}
	moves := legal.EnumerateForWithOptions(g, seat.ID, legal.Options{OrderCostFuel: order})
	var first *legal.Move
	for i := range moves {
		if strings.Contains(moves[i].Label, "exile") && strings.Contains(moves[i].Label, "—") {
			first = &moves[i]
			break
		}
	}
	if first == nil {
		t.Fatalf("no payment among %v", labels(moves))
	}
	var p struct {
		Apply   *bool    `json:"apply"`
		CardIDs []string `json:"card_ids"`
	}
	if err := json.Unmarshal(first.Params, &p); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{ids[2].String(): true, ids[3].String(): true}
	if len(p.CardIDs) != 2 || !want[p.CardIDs[0]] || !want[p.CardIDs[1]] {
		t.Errorf("first payment %v, want the two cheapest cards %v", p.CardIDs, want)
	}
	dispatchAll(t, g, seat.ID, moves)
}
