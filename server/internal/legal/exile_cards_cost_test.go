package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// exile_cards_cost_test.go — the enumerator half of #1297: a CR 602
// ability whose cost is "Exile N cards from your graveyard" (or hand).
// #544's invariant again: every activation offered is one the engine
// accepts (dispatchAll), and none is offered that it would refuse.

const (
	oracleGrimLavamancer = "37445e06-88a1-4e2e-a432-383736c9b977"
	oracleTomeShredder   = "b145952b-52e3-4a66-b47e-f08a489f9443"
	oracleHolisticWisdom = "7e108285-52da-473c-accd-d48e646a49c0"
)

type exileParams struct {
	ExileIDs   []string `json:"exile_ids"`
	DiscardIDs []string `json:"discard_ids"`
}

func exileParamsOf(t *testing.T, m legal.Move) exileParams {
	t.Helper()
	var p exileParams
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params: %v", err)
	}
	return p
}

// lavamancerBoard is a Grim Lavamancer with a Mountain to pay its {R}
// and `n` cards in its controller's graveyard.
func lavamancerBoard(t *testing.T, n int) (*game.Game, *game.Player, uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	battlefieldCard(g, seat, basic("Mountain", "Mountain"))
	lava := battlefieldCard(g, seat, game.Card{
		Name: "Grim Lavamancer", TypeLine: "Creature — Human Wizard", OracleID: oracleGrimLavamancer,
		ManaCost: "{R}", Power: 1, Toughness: 1,
	})
	var fuel []uuid.UUID
	for i := 0; i < n; i++ {
		fuel = append(fuel, graveyardCard(seat, game.Card{Name: "Fuel", TypeLine: "Sorcery", ManaCost: "{1}"}))
	}
	return g, seat, lava, fuel
}

// One card short of "two cards" is no activation at all; two is an
// activation that names both on exile_ids and nothing on discard_ids.
func TestGrimLavamancerIsEnumeratedOnlyWithTwoGraveyardCards(t *testing.T) {
	g, seat, lava, _ := lavamancerBoard(t, 1)
	if got := movesFrom(legal.EnumerateFor(g, seat.ID), lava, legal.KindActivate); len(got) != 0 {
		t.Fatalf("one graveyard card offered %v — the cost exiles two", labels(got))
	}

	g, seat, lava, fuel := lavamancerBoard(t, 2)
	moves := legal.EnumerateFor(g, seat.ID)
	got := movesFrom(moves, lava, legal.KindActivate)
	if len(got) == 0 {
		t.Fatalf("two graveyard cards and a Mountain, but no activation: %v", labels(moves))
	}
	for _, m := range got {
		p := exileParamsOf(t, m)
		if len(p.ExileIDs) != 2 || len(p.DiscardIDs) != 0 {
			t.Errorf("params = %s, want two exile_ids and no discard_ids", string(m.Params))
		}
		want := map[string]bool{fuel[0].String(): true, fuel[1].String(): true}
		for _, id := range p.ExileIDs {
			if !want[id] {
				t.Errorf("exile_ids %v name a card outside the graveyard", p.ExileIDs)
			}
		}
	}
	dispatchAll(t, g, seat.ID, moves)
}

// The payment the enumerator offers spends what the policy would miss
// least — Options.OrderCostFuel, the price escape's graveyard is paid
// by (#1013). Zone order would have eaten the oldest card first.
func TestGrimLavamancerExilesTheCheapestFuelFirst(t *testing.T) {
	g, seat, lava, fuel := lavamancerBoard(t, 4)
	precious := fuel[0]
	order := func(c legal.TargetCandidate) float64 {
		if c.ID == precious {
			return 100
		}
		return 1
	}
	moves := legal.EnumerateForWithOptions(g, seat.ID, legal.Options{OrderCostFuel: order})
	got := movesFrom(moves, lava, legal.KindActivate)
	if len(got) == 0 {
		t.Fatalf("no activation: %v", labels(moves))
	}
	for _, id := range exileParamsOf(t, got[0]).ExileIDs {
		if id == precious.String() {
			t.Errorf("the payment exiled the card the policy priced highest: %s", string(got[0].Params))
		}
	}
	dispatchAll(t, g, seat.ID, moves)
}

// A filtered clause pays only with what it admits: Tome Shredder's "an
// instant or sorcery card" is not paid with a creature card.
func TestTomeShredderIsEnumeratedOnlyWithAnInstantOrSorcery(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	shredder := battlefieldCard(g, seat, game.Card{
		Name: "Tome Shredder", TypeLine: "Creature — Wolf", OracleID: oracleTomeShredder,
		ManaCost: "{2}{R}", Power: 2, Toughness: 2,
	})
	graveyardCard(seat, creature("Dead Beast", "{2}{G}", 2, 2))
	if got := movesFrom(legal.EnumerateFor(g, seat.ID), shredder, legal.KindActivate); len(got) != 0 {
		t.Fatalf("a creature card paid for \"an instant or sorcery card\": %v", labels(got))
	}
	bolt := graveyardCard(seat, game.Card{Name: "Spent Bolt", TypeLine: "Instant", ManaCost: "{R}"})
	moves := legal.EnumerateFor(g, seat.ID)
	got := movesFrom(moves, shredder, legal.KindActivate)
	if len(got) != 1 {
		t.Fatalf("want one Shredder activation, got %v", labels(got))
	}
	if p := exileParamsOf(t, got[0]); len(p.ExileIDs) != 1 || p.ExileIDs[0] != bolt.String() {
		t.Errorf("params = %s, want exile_ids [%s]", string(got[0].Params), bolt)
	}
	dispatchAll(t, g, seat.ID, moves)
}

// The HAND form (Holistic Wisdom): the payment comes out of the hand,
// on exile_ids, and an empty hand offers nothing.
func TestHolisticWisdomIsEnumeratedWithAHandExile(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	battlefieldCard(g, seat, basic("Forest", "Forest"))
	battlefieldCard(g, seat, basic("Forest", "Forest"))
	wisdom := battlefieldCard(g, seat, game.Card{Name: "Holistic Wisdom", TypeLine: "Enchantment", OracleID: oracleHolisticWisdom})
	graveyardCard(seat, creature("Dead Beast", "{2}{G}", 2, 2))
	if got := movesFrom(legal.EnumerateFor(g, seat.ID), wisdom, legal.KindActivate); len(got) != 0 {
		t.Fatalf("an empty hand offered %v — there is nothing to exile", labels(got))
	}
	pitch := handCard(seat, creature("Spare Beast", "{5}{G}", 5, 5))
	moves := legal.EnumerateFor(g, seat.ID)
	got := movesFrom(moves, wisdom, legal.KindActivate)
	if len(got) == 0 {
		t.Fatalf("no Holistic Wisdom activation: %v", labels(moves))
	}
	for _, m := range got {
		if p := exileParamsOf(t, m); len(p.ExileIDs) != 1 || p.ExileIDs[0] != pitch.String() || len(p.DiscardIDs) != 0 {
			t.Errorf("params = %s, want exile_ids [%s] and no discard_ids", string(m.Params), pitch)
		}
	}
	dispatchAll(t, g, seat.ID, moves)
}
