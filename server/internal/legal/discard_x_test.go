package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// discard_x_test.go — the enumerator half of #6145: "As an additional cost
// to cast this spell, discard X cards" (Channeled Force) announces an X no
// {X} slot prices. The hand is the ceiling, the enumerator offers a small
// X, and the discards number exactly that X. Every move offered is one
// CastSpell accepts (#544), and none is offered at X=0.

const oracleChanneledForce = "e12ed186-01c1-4da5-b90f-cfbb77c4e8d0"

func channeledForceInHand(t *testing.T, g *game.Game, p *game.Player, others int) uuid.UUID {
	t.Helper()
	clearHand(p)
	cid := handCard(p, game.Card{
		Name: "Channeled Force", TypeLine: "Instant", ManaCost: "{2}{U}{R}", OracleID: oracleChanneledForce,
	})
	for range others {
		handCard(p, game.Card{Name: "Filler", TypeLine: "Sorcery", ManaCost: "{5}"})
	}
	lands(g, p, "Island", "Island", 2)
	lands(g, p, "Mountain", "Mountain", 2)
	return cid
}

func TestDiscardXIsAnnouncedWithExactlyXDiscards(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	card := channeledForceInHand(t, g, active, 3)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	casts := castMovesFor(moves, card)
	if len(casts) == 0 {
		t.Fatal("Channeled Force was not offered")
	}
	for _, m := range casts {
		var p struct {
			XValue     int      `json:"x_value"`
			DiscardIDs []string `json:"discard_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("decode %q: %v", m.Label, err)
		}
		if p.XValue < 1 || p.XValue > 2 {
			t.Errorf("%q: X=%d, want 1..2", m.Label, p.XValue)
		}
		if len(p.DiscardIDs) != p.XValue {
			t.Errorf("%q: %d discards at X=%d", m.Label, len(p.DiscardIDs), p.XValue)
		}
	}
	dispatchAll(t, g, active.ID, moves)
}

// With nothing else in hand, X can only be 0, which does nothing: no move.
func TestDiscardXIsNotOfferedWithAnEmptyHand(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	card := channeledForceInHand(t, g, active, 0)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	if casts := castMovesFor(moves, card); len(casts) != 0 {
		t.Errorf("offered %d casts with nothing to discard: %v", len(casts), labels(casts))
	}
	dispatchAll(t, g, active.ID, moves)
}
