package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// flashback_behold_test.go — #4689: Kindle the Inner Flame's
// "Flashback—{1}{R}, Behold three Elementals". The enumerator offers the
// flashback cast with three distinct Elementals on alt_cost_ids, the
// permanents first (choosing one shows the table nothing new), and offers
// none when the seat has fewer than three. Every move is one the engine
// accepts (#544).

const oracleKindleTheInnerFlame = "c7ca6371-73bb-4717-a148-0804b8f4c1e6"

type flashbackMove struct {
	Key  string
	From string
	Pay  []string
}

func flashbackMovesOf(t *testing.T, moves []legal.Move, src uuid.UUID) []flashbackMove {
	t.Helper()
	var out []flashbackMove
	for _, m := range moves {
		if m.Source != src || m.Kind != legal.KindCast {
			continue
		}
		var p struct {
			AlternativeCost string   `json:"alternative_cost"`
			FromZone        string   `json:"from_zone"`
			AltCostIDs      []string `json:"alt_cost_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("decode %s: %v", m.Params, err)
		}
		out = append(out, flashbackMove{Key: p.AlternativeCost, From: p.FromZone, Pay: p.AltCostIDs})
	}
	return out
}

func kindleTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 2, "Mountain")
	kindle := graveyardCard(seat, game.Card{Name: "Kindle the Inner Flame", TypeLine: "Kindred Sorcery — Elemental", ManaCost: "{3}{R}", OracleID: oracleKindleTheInnerFlame})
	return g, seat, kindle
}

func TestFlashbackBeholdOffersThreeDistinctElementals(t *testing.T) {
	g, seat, kindle := kindleTable(t)
	a := battlefieldCard(g, seat, game.Card{Name: "Smoke Elemental", TypeLine: "Creature — Elemental", Power: 2, Toughness: 2})
	b := battlefieldCard(g, seat, game.Card{Name: "Ember Elemental", TypeLine: "Creature — Elemental", Power: 1, Toughness: 1})
	inHand := handCard(seat, game.Card{Name: "Flame Elemental", TypeLine: "Creature — Elemental", ManaCost: "{R}"})
	spare := handCard(seat, game.Card{Name: "Spare Flame", TypeLine: "Creature — Elemental", ManaCost: "{R}"})

	moves := legal.EnumerateFor(g, seat.ID)
	got := flashbackMovesOf(t, moves, kindle)
	if len(got) == 0 {
		t.Fatalf("no flashback cast of Kindle offered; moves = %v", labels(moves))
	}
	for _, m := range got {
		if m.Key != "flashback" || m.From != "graveyard" || len(m.Pay) != 3 {
			t.Fatalf("move = %+v, want a graveyard flashback naming three objects", m)
		}
		seen := map[string]bool{}
		for _, id := range m.Pay {
			if seen[id] {
				t.Fatalf("move = %+v names %s twice", m, id)
			}
			seen[id] = true
		}
	}
	first := got[0].Pay
	if first[0] != a.String() && first[0] != b.String() || first[1] != a.String() && first[1] != b.String() {
		t.Errorf("first payment = %v, want the two Elemental permanents first", first)
	}
	if first[2] != inHand.String() && first[2] != spare.String() {
		t.Errorf("first payment = %v, want a hand Elemental third", first)
	}
	dispatchAll(t, g, seat.ID, moves)
}

func TestFlashbackBeholdNotOfferedWithTwoElementals(t *testing.T) {
	g, seat, kindle := kindleTable(t)
	battlefieldCard(g, seat, game.Card{Name: "Smoke Elemental", TypeLine: "Creature — Elemental", Power: 2, Toughness: 2})
	handCard(seat, game.Card{Name: "Flame Elemental", TypeLine: "Creature — Elemental", ManaCost: "{R}"})
	handCard(seat, game.Card{Name: "Goblin Friend", TypeLine: "Creature — Goblin", ManaCost: "{R}"})

	moves := legal.EnumerateFor(g, seat.ID)
	if got := flashbackMovesOf(t, moves, kindle); len(got) != 0 {
		t.Fatalf("flashback offered with two Elementals: %+v", got)
	}
	dispatchAll(t, g, seat.ID, moves)
}
