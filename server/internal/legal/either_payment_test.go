package legal_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// either_payment_test.go — ADR 0187 §7 (#3575): an either payment (Worms
// of the Earth) is one question with a move per payable branch, each
// naming its branch by index, plus the decline.

// eitherOffer queues Worms' offer (branch 0 sacrifices two lands, branch
// 1 takes 5 damage) and puts `lands` lands on the asked seat's side. It
// returns the table, the asked seat and its moves.
func eitherOffer(t *testing.T, lands int) (*game.Game, *game.Player, []legal.Move, []uuid.UUID) {
	t.Helper()
	g := newTable(t)
	controller := g.Seats[1]
	source := uuid.New()
	known := map[uuid.UUID]bool{}
	for _, s := range g.Seats {
		known[s.ID] = true
	}
	g.Battlefield.PushTop(game.Card{InstanceID: source, Name: "Worms", TypeLine: "Enchantment", Owner: controller.ID, Controller: controller.ID, KnownBy: known})
	g.WithWriteLock(func() {
		if err := g.QueueAnyPlayerPaysForEffect(game.AnyPlayerPays{
			Source: source, Controller: controller.ID, Cost: "sacrifice two lands or take 5 damage", Question: "Worms?",
			Action: &game.PayAction{Kind: game.PayActionEither, Either: []game.PayAction{
				{Kind: game.PayActionSacrifice, Count: 2, Of: game.PermanentQuery{Types: []string{"land"}}, Words: "sacrifice two lands"},
				{Kind: game.PayActionDamage, Count: 5, Words: "take 5 damage"},
			}},
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
	for i := 0; i < lands; i++ {
		id := uuid.New()
		g.Battlefield.PushTop(game.Card{InstanceID: id, Name: "Forest " + string(rune('A'+i)), TypeLine: "Basic Land — Forest", Owner: seat.ID, Controller: seat.ID, KnownBy: known})
		ids = append(ids, id)
	}
	return g, seat, legal.EnumerateFor(g, chooser), ids
}

type eitherMoveParams struct {
	Apply     *bool    `json:"apply"`
	PayOption *int     `json:"pay_option"`
	CardIDs   []string `json:"card_ids"`
}

func eitherParams(t *testing.T, m legal.Move) eitherMoveParams {
	t.Helper()
	var p eitherMoveParams
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatal(err)
	}
	return p
}

// With three lands the move list holds the damage branch, one move for
// each pair of lands (the sacrifice branch), and the decline; every one
// is accepted by the dispatcher.
func TestEitherPaymentOffersAMovePerBranch(t *testing.T) {
	g, seat, moves, lands := eitherOffer(t, 3)
	var damage, sacrifice, decline int
	for _, m := range moves {
		p := eitherParams(t, m)
		switch {
		case p.Apply != nil && !*p.Apply:
			decline++
			if p.PayOption != nil {
				t.Errorf("decline %q names a branch", m.Label)
			}
		case p.PayOption != nil && *p.PayOption == 1:
			damage++
			if len(p.CardIDs) != 0 || !strings.Contains(m.Label, "take 5 damage") {
				t.Errorf("damage move %q carries cards or the wrong words", m.Label)
			}
		case p.PayOption != nil && *p.PayOption == 0:
			sacrifice++
			if len(p.CardIDs) != 2 {
				t.Errorf("sacrifice move %q names %d lands, want 2", m.Label, len(p.CardIDs))
			}
			for _, id := range p.CardIDs {
				found := false
				for _, l := range lands {
					found = found || l.String() == id
				}
				if !found {
					t.Errorf("sacrifice move %q names a card that is not the seat's land", m.Label)
				}
			}
			if !strings.Contains(m.Label, "sacrifice two lands") || !strings.Contains(m.Label, "Forest") {
				t.Errorf("sacrifice move label %q does not name the branch and the lands", m.Label)
			}
		default:
			t.Errorf("unexpected move %q %s", m.Label, string(m.Params))
		}
	}
	if damage != 1 || sacrifice != 3 || decline != 1 {
		t.Errorf("damage %d, sacrifice %d, decline %d; want 1, 3 (the pairs of three lands) and 1: %v", damage, sacrifice, decline, labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}

// The sacrifice branch with fewer than two lands is not offered, and the
// damage branch always is.
func TestEitherPaymentDropsASacrificeBranchTheSeatCannotPay(t *testing.T) {
	for _, lands := range []int{0, 1} {
		g, seat, moves, _ := eitherOffer(t, lands)
		var branches []int
		for _, m := range moves {
			if p := eitherParams(t, m); p.PayOption != nil {
				branches = append(branches, *p.PayOption)
			}
		}
		if len(branches) != 1 || branches[0] != 1 {
			t.Errorf("%d lands: branches offered %v, want only the damage branch (1): %v", lands, branches, labels(moves))
		}
		if !containsLabel(moves, "decline") {
			t.Errorf("%d lands: no decline in %v", lands, labels(moves))
		}
		dispatchAll(t, g, seat.ID, moves)
	}
}

func containsLabel(moves []legal.Move, sub string) bool {
	for _, m := range moves {
		if strings.Contains(m.Label, sub) {
			return true
		}
	}
	return false
}
