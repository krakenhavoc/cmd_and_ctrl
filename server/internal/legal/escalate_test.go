package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// escalate_test.go — CR 702.120a, #2126: the enumerator half. An
// escalate spell is one cast per mode COUNT, and a count whose extra
// modes the seat cannot pay for — cards to discard, creatures to tap,
// mana — is not offered. Every offered move is dispatched against a
// clone, so "offered" and "accepted" cannot drift apart (#544).

const (
	oracleCollectiveBrutality  = "22a78443-db21-4656-b38a-e3e3186fd94b"
	oracleCollectiveEffort     = "a97e760a-99c8-47ea-a875-451aca913ee7"
	oracleCollectiveResistance = "bffdfe7b-f17a-41b6-a460-80280fe497d4"
)

type escalateOffer struct {
	modes, discards, taps int
}

func escalateOffers(t *testing.T, moves []legal.Move, source uuid.UUID) map[int]escalateOffer {
	t.Helper()
	out := make(map[int]escalateOffer)
	for _, m := range castMovesFor(moves, source) {
		var p struct {
			Modes       []int    `json:"modes"`
			DiscardIDs  []string `json:"discard_ids"`
			TeamworkIDs []string `json:"teamwork_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("bad params %s: %v", m.Params, err)
		}
		out[len(p.Modes)] = escalateOffer{len(p.Modes), len(p.DiscardIDs), len(p.TeamworkIDs)}
	}
	return out
}

func escalateTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(active)
	return g, active, opp
}

// Discard-shaped: Collective Brutality offers as many modes as the
// seat has cards to pay for, each priced at one discard per extra mode.
func TestEscalateDiscardOffersOnlyPayableModeCounts(t *testing.T) {
	for _, tc := range []struct {
		fodder int
		want   []int
	}{
		{0, []int{1}},
		{1, []int{1, 2}},
		{2, []int{1, 2, 3}},
		{5, []int{1, 2, 3}},
	} {
		g, active, opp := escalateTable(t)
		spell := handCard(active, game.Card{
			Name: "Collective Brutality", TypeLine: "Sorcery",
			OracleID: oracleCollectiveBrutality, ManaCost: "{1}{B}",
		})
		for range tc.fodder {
			handCard(active, game.Card{Name: "Fodder", TypeLine: "Sorcery"})
		}
		battlefieldCard(g, opp, creature("Their Bear", "{1}{G}", 2, 2))
		lands(g, active, "Swamp", "Swamp", 2)
		advanceTo(t, g, game.StepPrecombatMain)

		moves := legal.EnumerateFor(g, active.ID)
		offers := escalateOffers(t, moves, spell)
		for _, n := range tc.want {
			o, ok := offers[n]
			if !ok {
				t.Errorf("%d fodder: %d modes should be offered; got %v", tc.fodder, n, offers)
				continue
			}
			if o.discards != n-1 {
				t.Errorf("%d fodder, %d modes: %d discards named, want %d", tc.fodder, n, o.discards, n-1)
			}
		}
		if len(offers) != len(tc.want) {
			t.Errorf("%d fodder: offered mode counts %v, want exactly %v", tc.fodder, offers, tc.want)
		}
		dispatchAll(t, g, active.ID, castMovesFor(moves, spell))
	}
}

// Tap-shaped: Collective Effort needs one untapped creature per extra
// mode, and a creature named to the tap cannot also have made the mana.
func TestEscalateTapOffersOnlyPayableModeCounts(t *testing.T) {
	for _, tc := range []struct {
		creatures int
		want      []int
	}{
		{0, []int{1}},
		{1, []int{1, 2}},
		{2, []int{1, 2, 3}},
	} {
		g, active, opp := escalateTable(t)
		spell := handCard(active, game.Card{
			Name: "Collective Effort", TypeLine: "Sorcery",
			OracleID: oracleCollectiveEffort, ManaCost: "{1}{W}{W}",
		})
		battlefieldCard(g, opp, creature("Their Giant", "{4}{G}", 5, 5))
		battlefieldCard(g, opp, game.Card{Name: "Their Aura", TypeLine: "Enchantment"})
		for range tc.creatures {
			battlefieldCard(g, active, creature("Mine", "{G}", 1, 1))
		}
		lands(g, active, "Plains", "Plains", 3)
		advanceTo(t, g, game.StepPrecombatMain)

		moves := legal.EnumerateFor(g, active.ID)
		offers := escalateOffers(t, moves, spell)
		for _, n := range tc.want {
			o, ok := offers[n]
			if !ok {
				t.Errorf("%d creatures: %d modes should be offered; got %v", tc.creatures, n, offers)
				continue
			}
			if o.taps != n-1 {
				t.Errorf("%d creatures, %d modes: %d taps named, want %d", tc.creatures, n, o.taps, n-1)
			}
		}
		if len(offers) != len(tc.want) {
			t.Errorf("%d creatures: offered mode counts %v, want exactly %v", tc.creatures, offers, tc.want)
		}
		dispatchAll(t, g, active.ID, castMovesFor(moves, spell))
	}
}

// Mana-shaped: each extra mode adds {G} to the price, so the offered
// counts follow the Forests the seat has.
func TestEscalateManaOffersOnlyAffordableModeCounts(t *testing.T) {
	for _, tc := range []struct {
		forests int
		want    []int
	}{
		{1, nil},
		{2, []int{1}},
		{3, []int{1, 2}},
		{4, []int{1, 2, 3}},
	} {
		g, active, opp := escalateTable(t)
		spell := handCard(active, game.Card{
			Name: "Collective Resistance", TypeLine: "Instant",
			OracleID: oracleCollectiveResistance, ManaCost: "{1}{G}",
		})
		battlefieldCard(g, opp, game.Card{Name: "Their Rock", TypeLine: "Artifact"})
		battlefieldCard(g, opp, game.Card{Name: "Their Aura", TypeLine: "Enchantment"})
		battlefieldCard(g, opp, creature("Their Bear", "{1}{G}", 2, 2))
		lands(g, active, "Forest", "Forest", tc.forests)
		advanceTo(t, g, game.StepPrecombatMain)

		moves := legal.EnumerateFor(g, active.ID)
		offers := escalateOffers(t, moves, spell)
		for _, n := range tc.want {
			if _, ok := offers[n]; !ok {
				t.Errorf("%d Forests: %d modes should be offered; got %v", tc.forests, n, offers)
			}
		}
		if len(offers) != len(tc.want) {
			t.Errorf("%d Forests: offered mode counts %v, want exactly %v", tc.forests, offers, tc.want)
		}
		dispatchAll(t, g, active.ID, castMovesFor(moves, spell))
	}
}
