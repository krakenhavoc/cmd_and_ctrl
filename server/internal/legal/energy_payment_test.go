package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// energy_payment_test.go — ADR 0129 §3 and §7: the enumerator's answers
// to energy paid while an ability resolves.

// A fixed energy payment. "Pay" is a move only when the seat has the
// energy, and it names the energy it spends; the decline is always
// there. Every move offered is accepted.
func TestEnergyPaymentPromptMoves(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() {
		_ = g.QueueEnergyPaymentForEffect(game.EnergyPayment{Chooser: active.ID, N: 2, Question: "pay {E}{E}?"})
	})
	giveEnergy(t, g, active, 1)
	moves := legal.EnumerateFor(g, active.ID)
	if len(moves) != 1 || moves[0].Label != "pay {E}{E}?: decline" {
		t.Fatalf("one energy: want only the decline, got %v", labels(moves))
	}
	giveEnergy(t, g, active, 1)
	moves = legal.EnumerateFor(g, active.ID)
	if len(moves) != 2 {
		t.Fatalf("two energy: want pay and decline, got %v", labels(moves))
	}
	if c := moves[0].Cost; c == nil || c.Energy != 2 {
		t.Errorf("pay cost = %+v, want energy 2", c)
	}
	dispatchAll(t, g, active.ID, moves)
}

// Owner decision 3: a pay_amount prompt is offered as nothing, the
// smallest payment, the card's threshold and the ceiling — four moves
// for a 9-energy prompt, not ten — each accepted, paying nothing marked
// always legal.
func TestPayAmountPromptMoves(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	giveEnergy(t, g, active, 9)
	g.WithWriteLock(func() {
		_ = g.QueuePayEnergyAmountForEffect(game.PayEnergyAmount{
			Chooser: active.ID, Min: 1, Goal: 4, Unit: game.PayAmountDamage, Question: "pay any amount",
		})
	})
	moves := legal.EnumerateFor(g, active.ID)
	var amounts []int
	for _, m := range moves {
		var p struct {
			Amount *int `json:"amount"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil || p.Amount == nil {
			t.Fatalf("move %q carries no amount: %s", m.Label, m.Params)
		}
		amounts = append(amounts, *p.Amount)
		if (*p.Amount == 0) != m.AlwaysLegal {
			t.Errorf("amount %d: AlwaysLegal = %v", *p.Amount, m.AlwaysLegal)
		}
		if *p.Amount > 0 && (m.Cost == nil || m.Cost.Energy != *p.Amount) {
			t.Errorf("amount %d: cost = %+v", *p.Amount, m.Cost)
		}
	}
	if len(amounts) != 4 || amounts[0] != 0 || amounts[1] != 1 || amounts[2] != 4 || amounts[3] != 9 {
		t.Errorf("amounts = %v, want [0 1 4 9]", amounts)
	}
	dispatchAll(t, g, active.ID, moves)
}
