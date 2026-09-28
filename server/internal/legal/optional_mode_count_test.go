package legal_test

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// optional_mode_count_test.go — #1655, ADR 0065's 2026-09-28
// amendment: the enumerator half of a mode count that reads the
// optional costs announced with the modes. Inscription of Ruin's
// "if this spell was kicked, choose any number instead" makes the
// unkicked cast and the kicked cast two different mode ranges, and
// every offer of either must be one the engine accepts (#544).

const oracleInscriptionOfRuin = "760e8561-4ec6-4594-ba5d-f79cb9f25fd0"

// modesByKick groups a card's cast moves by whether they announce the
// kicker, as sorted "0,2" mode keys.
func modesByKick(t *testing.T, moves []legal.Move, source uuid.UUID) (unkicked, kicked map[string]bool) {
	t.Helper()
	unkicked, kicked = map[string]bool{}, map[string]bool{}
	for _, m := range castMovesFor(moves, source) {
		var p struct {
			Modes         []int `json:"modes"`
			OptionalCosts []int `json:"optional_costs"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("params: %v", err)
		}
		sorted := append([]int(nil), p.Modes...)
		sort.Ints(sorted)
		parts := make([]string, len(sorted))
		for i, v := range sorted {
			parts[i] = string(rune('0' + v))
		}
		key := strings.Join(parts, ",")
		if len(p.OptionalCosts) > 0 {
			kicked[key] = true
		} else {
			unkicked[key] = true
		}
	}
	return unkicked, kicked
}

func TestInscriptionOfRuinModeRangesFollowTheKicker(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	spell := handCard(active, game.Card{
		Name: "Inscription of Ruin", TypeLine: "Sorcery",
		OracleID: oracleInscriptionOfRuin, ManaCost: "{2}{B}",
	})
	// A target for every bullet: the opponent, a two-drop in your
	// graveyard, and a small creature on their side.
	active.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Bear", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: active.ID, Controller: active.ID})
	battlefieldCard(g, opp, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	lands(g, active, "Swamp", "Swamp", 7)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	unkicked, kicked := modesByKick(t, moves, spell)
	for _, one := range []string{"0", "1", "2"} {
		if !unkicked[one] {
			t.Errorf("unkicked: bullet %s alone is offered; got %v", one, unkicked)
		}
	}
	for key := range unkicked {
		if strings.Contains(key, ",") || key == "" {
			t.Errorf("unkicked: exactly one bullet, but %q was offered", key)
		}
	}
	// The expansion budget (MaxExpansionPerSource) caps how many of
	// the seven selections are reached; what matters is that the kicked
	// range reaches past one bullet, all the way to three.
	if !kicked["0,1,2"] || !kicked["0,1"] {
		t.Errorf("kicked: any number of bullets is offered; got %v", kicked)
	}
	if kicked[""] {
		t.Error("kicked: choosing no bullet at all must not be offered — the minimum is still one")
	}
	// #544: every offer, both ranges, is one the engine accepts.
	dispatchAll(t, g, active.ID, castMovesFor(moves, spell))
}
