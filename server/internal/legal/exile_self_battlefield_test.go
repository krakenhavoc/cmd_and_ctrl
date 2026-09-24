package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// exile_self_battlefield_test.go — the enumerator half of #1404: an
// "Exile this <permanent>:" cost paid from the battlefield. The bot
// sees the activation when its mana is payable, does not when it is
// not, and every activation it is offered is one the engine accepts
// (#544) — and accepting it leaves the source in exile.

const (
	oraclePerpetualTimepiece = "17b4778b-82b1-4845-ad08-00f3ff66877b"
	oracleHangedExecutioner  = "6ff4ff67-ad08-447f-a112-1a071c1474a4"
)

func exileThisMoves(moves []legal.Move, source uuid.UUID, clause string) []legal.Move {
	var out []legal.Move
	for _, m := range movesFrom(moves, source, legal.KindActivate) {
		if strings.Contains(m.Label, clause) {
			out = append(out, m)
		}
	}
	return out
}

// Perpetual Timepiece with two Plains untapped: the exile-to-shuffle
// activation is offered, and dispatching it exiles the Timepiece.
// With one Plains it is not offered at all.
func TestPerpetualTimepieceExileIsEnumeratedWhenPayable(t *testing.T) {
	for _, lands := range []int{1, 2} {
		g := newTable(t)
		seat := g.Seats[g.Turn.ActiveSeat]
		clearHand(seat)
		advanceTo(t, g, game.StepPrecombatMain)
		for i := 0; i < lands; i++ {
			battlefieldCard(g, seat, basic("Plains", "Plains"))
		}
		piece := battlefieldCard(g, seat, game.Card{
			Name: "Perpetual Timepiece", TypeLine: "Artifact", OracleID: oraclePerpetualTimepiece, ManaCost: "{2}",
		})
		graveyardCard(seat, game.Card{Name: "Fuel", TypeLine: "Sorcery", ManaCost: "{1}"})

		moves := legal.EnumerateFor(g, seat.ID)
		got := exileThisMoves(moves, piece, "Exile this artifact")
		if lands < 2 {
			if len(got) != 0 {
				t.Fatalf("one Plains offered %v — the cost is {2}", labels(got))
			}
			continue
		}
		if len(got) == 0 {
			t.Fatalf("two Plains and a Timepiece, but no exile activation: %v", labels(moves))
		}
		dispatchAll(t, g, seat.ID, got)

		clone := g.Clone()
		m := got[0]
		if err := actions.Dispatch(clone, actions.Action{
			Type: actions.Type(m.Type), Player: m.Player, Caller: seat.ID, Params: m.Params,
		}); err != nil {
			t.Fatalf("dispatch: %v", err)
		}
		if clone.Battlefield.Contains(piece) || !clone.Exile.Contains(piece) {
			t.Error("the dispatched activation did not exile the Timepiece")
		}
	}
}

// A creature whose exile ability targets: Hanged Executioner's
// activation is offered against the opponent's creature, and every
// activation offered dispatches.
func TestHangedExecutionerExileIsEnumeratedPerTarget(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	for i := 0; i < 4; i++ {
		battlefieldCard(g, seat, basic("Plains", "Plains"))
	}
	exec := battlefieldCard(g, seat, game.Card{
		Name: "Hanged Executioner", TypeLine: "Creature — Spirit", OracleID: oracleHangedExecutioner,
		ManaCost: "{2}{W}", Power: 1, Toughness: 1,
	})
	battlefieldCard(g, opp, creature("Target Bear", "{1}{G}", 2, 2))

	moves := legal.EnumerateFor(g, seat.ID)
	got := exileThisMoves(moves, exec, "Exile this creature")
	if len(got) == 0 {
		t.Fatalf("four Plains and a target, but no exile activation: %v", labels(moves))
	}
	dispatchAll(t, g, seat.ID, got)
}
