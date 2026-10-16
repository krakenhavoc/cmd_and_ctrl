package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// exile_x_cast_cost_test.go — #1948: "exile X cards from your graveyard"
// reaches the move list with an X the graveyard can pay and exactly X
// cards on exile_ids, and Soaring Stoneglider's exile branch is offered
// only when the graveyard holds two cards.

const (
	oracleHarvestPyre        = "00e36265-0cd8-4084-adf5-d1ed44f0d689"
	oracleSoaringStoneglider = "62d18d2b-fa68-49cd-b62d-803728ddfa8b"
)

func exileXSetup(t *testing.T, graveyard int, card game.Card, land, sub string, n int) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	spell := handCard(active, card)
	for i := range graveyard {
		active.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Fuel " + string(rune('A'+i)),
			TypeLine: "Sorcery", Owner: active.ID, Controller: active.ID})
	}
	lands(g, active, land, sub, n)
	advanceTo(t, g, game.StepPrecombatMain)
	return g, active, spell
}

func TestHarvestPyreIsOfferedAtTheGraveyardsXWithXNamedCards(t *testing.T) {
	g, active, spell := exileXSetup(t, 3, game.Card{
		Name: "Harvest Pyre", TypeLine: "Instant", OracleID: oracleHarvestPyre, ManaCost: "{1}{R}",
	}, "Mountain", "Mountain", 2)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})

	moves := castMovesFor(legal.EnumerateFor(g, active.ID), spell)
	if len(moves) == 0 {
		t.Fatal("Harvest Pyre is not offered with three graveyard cards")
	}
	for _, m := range moves {
		var p struct {
			XValue   int      `json:"x_value"`
			ExileIDs []string `json:"exile_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("decode %q: %v", m.Label, err)
		}
		if p.XValue < 1 || p.XValue > 3 {
			t.Errorf("%q announces X = %d, want 1..3 (the graveyard holds three)", m.Label, p.XValue)
		}
		if len(p.ExileIDs) != p.XValue {
			t.Errorf("%q names %d cards for X = %d", m.Label, len(p.ExileIDs), p.XValue)
		}
		if m.Value != nil {
			t.Errorf("%q leaves X open (%+v), but the exiled cards fix it", m.Label, m.Value)
		}
	}
	dispatchAll(t, g, active.ID, moves)
}

// An empty graveyard pays only X = 0, which the card is not offered at.
func TestHarvestPyreIsNotOfferedWithAnEmptyGraveyard(t *testing.T) {
	g, active, spell := exileXSetup(t, 0, game.Card{
		Name: "Harvest Pyre", TypeLine: "Instant", OracleID: oracleHarvestPyre, ManaCost: "{1}{R}",
	}, "Mountain", "Mountain", 2)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	if moves := castMovesFor(legal.EnumerateFor(g, active.ID), spell); len(moves) != 0 {
		t.Errorf("Harvest Pyre offered with an empty graveyard: %v", labels(moves))
	}
}

func TestSoaringStonegliderExileBranchNeedsTwoGraveyardCards(t *testing.T) {
	for _, tc := range []struct {
		graveyard int
		wantExile bool
	}{{graveyard: 2, wantExile: true}, {graveyard: 1, wantExile: false}} {
		g, active, spell := exileXSetup(t, tc.graveyard, game.Card{
			Name: "Soaring Stoneglider", TypeLine: "Creature — Elephant Cleric", OracleID: oracleSoaringStoneglider,
			ManaCost: "{2}{W}", Power: 4, Toughness: 3,
		}, "Plains", "Plains", 5)
		moves := castMovesFor(legal.EnumerateFor(g, active.ID), spell)
		sawExile, sawMana := false, false
		for _, m := range moves {
			var p struct {
				CostBranch *int     `json:"cost_branch"`
				ExileIDs   []string `json:"exile_ids"`
			}
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatalf("decode %q: %v", m.Label, err)
			}
			if p.CostBranch == nil {
				t.Fatalf("%q names no branch", m.Label)
			}
			switch *p.CostBranch {
			case 0:
				sawExile = true
				if len(p.ExileIDs) != 2 {
					t.Errorf("%q names %d cards for the exile branch", m.Label, len(p.ExileIDs))
				}
			case 1:
				sawMana = true
				if len(p.ExileIDs) != 0 {
					t.Errorf("%q names cards on the mana branch", m.Label)
				}
			}
		}
		if sawExile != tc.wantExile || !sawMana {
			t.Errorf("graveyard %d: exile branch offered %v (want %v), mana branch %v: %v",
				tc.graveyard, sawExile, tc.wantExile, sawMana, labels(moves))
		}
		dispatchAll(t, g, active.ID, moves)
	}
}
