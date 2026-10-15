package legal_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// exile_card_cast_cost_test.go — #3536: Abhorrent Oculus's "exile six
// cards from your graveyard" additional cost reaches the move list. A cast
// is offered once, with exactly six cards the engine accepts named on
// exile_ids, and not at all when the graveyard holds fewer.

const oracleAbhorrentOculus = "b1a0e129-27f0-4a09-b00d-497e18777306"

func oculusSetup(t *testing.T, graveyard int) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	spell := handCard(active, game.Card{
		Name: "Abhorrent Oculus", TypeLine: "Creature — Eye", OracleID: oracleAbhorrentOculus,
		ManaCost: "{2}{U}", Power: 5, Toughness: 5,
	})
	for i := range graveyard {
		active.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Fuel " + string(rune('A'+i)),
			TypeLine: "Instant", Owner: active.ID, Controller: active.ID})
	}
	lands(g, active, "Island", "Island", 3)
	advanceTo(t, g, game.StepPrecombatMain)
	return g, active, spell
}

func TestAbhorrentOculusIsOfferedOnceWithSixNamedCards(t *testing.T) {
	g, active, spell := oculusSetup(t, 7)
	moves := castMovesFor(legal.EnumerateFor(g, active.ID), spell)
	if len(moves) != 1 {
		t.Fatalf("Oculus has %d cast moves with a seven-card graveyard, want exactly 1: %v", len(moves), labels(moves))
	}
	m := moves[0]
	var p struct {
		ExileIDs []string `json:"exile_ids"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("decode %q: %v", m.Label, err)
	}
	if len(p.ExileIDs) != 6 {
		t.Errorf("%q names %d cards to exile, want exactly 6", m.Label, len(p.ExileIDs))
	}
	seen := map[string]bool{}
	for _, id := range p.ExileIDs {
		if seen[id] {
			t.Errorf("%q names %s twice", m.Label, id)
		}
		seen[id] = true
	}
	if !strings.Contains(m.Label, "exiling ") {
		t.Errorf("label %q does not show the exiled cards", m.Label)
	}
	dispatchAll(t, g, active.ID, moves)
}

// CR 118.3: five cards cannot pay six, so the cast is not offered.
func TestAbhorrentOculusIsNotOfferedWithFiveCards(t *testing.T) {
	g, active, spell := oculusSetup(t, 5)
	if moves := castMovesFor(legal.EnumerateFor(g, active.ID), spell); len(moves) != 0 {
		t.Errorf("Oculus offered with five graveyard cards: %v", labels(moves))
	}
}
