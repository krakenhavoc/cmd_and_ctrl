package legal_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// exile_permanent_cast_cost_test.go — #3402: Soul Exchange's "exile a
// creature you control" additional cost reaches the move list. A cast is
// offered only with a creature the engine accepts named on
// exile_permanent_ids, says so in its label, and is not offered at all
// when the seat has nothing to exile.

const oracleSoulExchange = "366ea3b4-5e7e-4693-9f4b-1e549972609e"

func soulExchangeSetup(t *testing.T, creatures int) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	spell := handCard(active, game.Card{
		Name: "Soul Exchange", TypeLine: "Sorcery", OracleID: oracleSoulExchange, ManaCost: "{B}{B}",
	})
	active.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Bear", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: active.ID, Controller: active.ID})
	lands(g, active, "Swamp", "Swamp", 2)
	for range creatures {
		battlefieldCard(g, active, creature("Thrull", "{B}", 1, 1))
	}
	advanceTo(t, g, game.StepPrecombatMain)
	return g, active, spell
}

func TestSoulExchangeIsOfferedWithTheCreatureItExiles(t *testing.T) {
	g, active, spell := soulExchangeSetup(t, 2)
	moves := castMovesFor(legal.EnumerateFor(g, active.ID), spell)
	if len(moves) == 0 {
		t.Fatal("Soul Exchange is not offered with a creature to exile")
	}
	for _, m := range moves {
		var p struct {
			ExilePermanentIDs []string `json:"exile_permanent_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("decode %q: %v", m.Label, err)
		}
		if len(p.ExilePermanentIDs) != 1 {
			t.Errorf("%q names %d creatures to exile, want exactly 1", m.Label, len(p.ExilePermanentIDs))
		}
		if !strings.Contains(m.Label, "exiling Thrull") {
			t.Errorf("label %q does not show the exiled creature", m.Label)
		}
	}
	dispatchAll(t, g, active.ID, moves)
}

func TestSoulExchangeIsNotOfferedWithNothingToExile(t *testing.T) {
	g, active, spell := soulExchangeSetup(t, 0)
	if moves := castMovesFor(legal.EnumerateFor(g, active.ID), spell); len(moves) != 0 {
		t.Errorf("Soul Exchange offered with no creature to exile: %v", labels(moves))
	}
}
