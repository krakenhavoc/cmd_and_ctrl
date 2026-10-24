package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// ignore_effect_test.go — #1986, CR 116.2d: the enumerator offers the
// ignore a Damping Engine's victim may take, from the engine's own
// accessor, and the move it offers is one the engine accepts.

const oracleDampingEngine1986 = "3584a055-5220-4721-89f8-de33f779122f"

// ignoreMoves lists the seat's ignore_effect special-action moves.
func ignoreMoves(g *game.Game, seat uuid.UUID) []legal.Move {
	var out []legal.Move
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Kind != legal.KindSpecialAction {
			continue
		}
		var p struct {
			Kind string `json:"kind"`
		}
		if json.Unmarshal(m.Params, &p) == nil && p.Kind == string(game.SpecialActionIgnoreEffect) {
			out = append(out, m)
		}
	}
	return out
}

func TestTheRestrictedSeatIsOfferedTheIgnoreAndNoOneElseIs(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[1-g.Turn.ActiveSeat]
	clearHand(active)
	forest := handCard(active, basic("Forest", "Forest"))
	engine := battlefieldCard(g, active, game.Card{
		Name: "Damping Engine", TypeLine: "Artifact", OracleID: oracleDampingEngine1986,
	})
	token := battlefieldCard(g, active, game.Card{Name: "Spirit", TypeLine: "Token Creature — Spirit"})
	advanceTo(t, g, game.StepPrecombatMain)

	// Restricted: no land play, one ignore, paying with the cheapest
	// permanent (the token), which the engine accepts.
	if got := landMovesFor(g, active.ID, forest); len(got) != 0 {
		t.Fatalf("a land play was offered to the restricted seat: %v", labels(got))
	}
	moves := ignoreMoves(g, active.ID)
	if len(moves) != 1 {
		t.Fatalf("%d ignore moves, want 1: %v", len(moves), labels(legal.EnumerateFor(g, active.ID)))
	}
	m := moves[0]
	if m.Source != engine {
		t.Errorf("the move's source is %v, want the Engine", m.Source)
	}
	var params struct {
		CardID      string `json:"card_id"`
		SacrificeID string `json:"sacrifice_id"`
	}
	if err := json.Unmarshal(m.Params, &params); err != nil {
		t.Fatal(err)
	}
	if params.CardID != engine.String() || params.SacrificeID != token.String() {
		t.Errorf("params = %+v, want the Engine and the token as the cost", params)
	}
	if len(ignoreMoves(g, other.ID)) != 0 {
		t.Error("a seat the effect does not bear on was offered an ignore")
	}
	dispatchAll(t, g, active.ID, moves)

	// Taken for real: the land play is offered, and the ignore is not.
	a := actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: active.ID, Params: m.Params}
	if err := actions.Dispatch(g, a); err != nil {
		t.Fatalf("dispatch the ignore: %v", err)
	}
	if got := landMovesFor(g, active.ID, forest); len(got) != 1 {
		t.Errorf("the land play is not offered after the ignore: %v", labels(legal.EnumerateFor(g, active.ID)))
	}
	if got := ignoreMoves(g, active.ID); len(got) != 0 {
		t.Errorf("the ignore is offered again: %v", labels(got))
	}
}

// A special action is neither a cast nor an activation (CR 702.61b), so it
// is offered under a split-second spell while the casts are not.
func TestTheIgnoreIsOfferedUnderSplitSecond(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	battlefieldCard(g, active, game.Card{Name: "Damping Engine", TypeLine: "Artifact", OracleID: oracleDampingEngine1986})
	advanceTo(t, g, game.StepPrecombatMain)
	g.WithWriteLock(func() { g.SplitSecondActive = true })
	if len(ignoreMoves(g, active.ID)) != 1 {
		t.Errorf("the ignore is not offered under split second: %v", labels(legal.EnumerateFor(g, active.ID)))
	}
}
