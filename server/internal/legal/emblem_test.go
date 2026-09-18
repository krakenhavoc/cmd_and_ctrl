package legal_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// emblem_test.go is the bot half of CR 114 (#623, ADR 0064 Decision
// 9): an emblem must never appear in a legal move.
//
// The enumerator walks `p.Hand` and `p.Command` for casts and the
// battlefield for everything else. An emblem lives in `p.Emblems`,
// which none of those reach, so the guard here is a regression test
// rather than a check on new filtering code: it fails the day
// somebody folds the emblem slice into a zone the enumerator reads,
// which is the change that would start offering a bot a cast the
// engine refuses — the hung-table shape #544 is about.

// TestEmblemIsNeverOfferedAsAMove gives the active seat an emblem and
// asserts nothing in its move list names it.
func TestEmblemIsNeverOfferedAsAMove(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	advanceTo(t, g, game.StepPrecombatMain)

	before := len(legal.EnumerateFor(g, active.ID))

	// Built by hand rather than through CreateEmblemForEffect,
	// because the enumerator's job is to ignore the OBJECT however it
	// got there, and this way the test needs no catalog stub.
	emblem := game.Card{
		InstanceID: uuid.New(),
		Name:       "Test Walker emblem",
		OracleID:   game.EmblemKey("test-walker"),
		Owner:      active.ID,
		Controller: active.ID,
	}
	active.Emblems.PushTop(emblem)

	moves := legal.EnumerateFor(g, active.ID)
	if len(moves) != before {
		t.Errorf("the move list went %d → %d after an emblem appeared", before, len(moves))
	}
	id := emblem.InstanceID.String()
	for _, m := range moves {
		if strings.Contains(m.Label, "emblem") {
			t.Errorf("a move names the emblem: %q", m.Label)
		}
		if raw, err := json.Marshal(m); err == nil && strings.Contains(string(raw), id) {
			t.Errorf("a move carries the emblem's instance id: %s", raw)
		}
	}
}
