package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// opponent_slot_test.go — #4128 (game.TargetSpec.OpponentSlot): "for
// each opponent, … target permanent that player controls" with a
// REQUIRED pick per opponent. The enumerator must offer a cast naming
// one permanent per LIVE opponent, a bound clause whose opponent is
// out of the game must not be demanded, and a live opponent with
// nothing to target means no cast at all.

const oracleBlatantThievery = "8bb51dd2-cc62-48ea-ae8e-52351073a8c9"

func thieveryCasts(t *testing.T, g *game.Game) (me *game.Player, casts []legal.Move) {
	t.Helper()
	me = g.Seats[g.Turn.ActiveSeat]
	clearHand(me)
	card := handCard(me, game.Card{Name: "Blatant Thievery", TypeLine: "Sorcery", OracleID: oracleBlatantThievery})
	advanceTo(t, g, game.StepPrecombatMain)
	return me, castMovesFor(legal.EnumerateFor(g, me.ID), card)
}

func targetCountOf(t *testing.T, m legal.Move) int {
	t.Helper()
	var p struct {
		Targets []json.RawMessage `json:"targets"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params: %v", err)
	}
	return len(p.Targets)
}

func giveRock(g *game.Game, seat int) uuid.UUID {
	return battlefieldCard(g, g.Seats[seat], game.Card{Name: "Rock", TypeLine: "Artifact"})
}

func TestBlatantThieveryEnumeratesOnePickPerLiveOpponent(t *testing.T) {
	g := newTable(t)
	giveRock(g, 1)
	giveRock(g, 2)
	giveRock(g, 3)
	me, casts := thieveryCasts(t, g)
	if len(casts) == 0 {
		t.Fatal("no cast offered with a target for every opponent")
	}
	for _, m := range casts {
		if n := targetCountOf(t, m); n != 3 {
			t.Fatalf("a cast names %d targets, want one per opponent (3)", n)
		}
	}
	dispatchAll(t, g, me.ID, casts)
}

func TestBlatantThieveryDoesNotDemandAPickForAnEliminatedOpponent(t *testing.T) {
	g := newTable(t)
	giveRock(g, 1)
	giveRock(g, 2)
	g.Seats[3].Eliminated = true
	me, casts := thieveryCasts(t, g)
	if len(casts) == 0 {
		t.Fatal("no cast offered although both live opponents have a permanent")
	}
	for _, m := range casts {
		if n := targetCountOf(t, m); n != 2 {
			t.Fatalf("a cast names %d targets, want one per LIVE opponent (2)", n)
		}
	}
	dispatchAll(t, g, me.ID, casts)
}

func TestBlatantThieveryIsNotOfferedWhenALiveOpponentHasNothingToTarget(t *testing.T) {
	g := newTable(t)
	giveRock(g, 1)
	giveRock(g, 2)
	_, casts := thieveryCasts(t, g)
	if len(casts) != 0 {
		t.Fatalf("%d casts offered although the third opponent controls nothing", len(casts))
	}
}
