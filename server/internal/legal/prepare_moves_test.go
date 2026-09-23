package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// prepare_moves_test.go — the enumerator half of CR 722.3c (ADR 0090,
// #1328). The permission over a prepare copy is DERIVED from the
// prepared permanent (game/prepare.go), and the enumerator reads it
// through the same CastPermissionForLocked the cast path does, so the
// bot is offered exactly the cast the engine accepts — the copy's
// prepare-spell face, by the prepared permanent's controller, and by
// nobody else. dispatchAll proves every offered move is accepted.

// preparedFixture seats a preparation card on p's battlefield and
// prepares it, returning the copy's instance ID. {0} on the prepare
// spell so affordability is not the thing under test.
func preparedFixture(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
	t.Helper()
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   "prepare-enum-oracle",
		Layout:     game.LayoutPrepare,
		Owner:      p.ID,
		Controller: p.ID,
		Faces: []game.Face{
			{Name: "Enum Conductor", TypeLine: "Creature — Bird Pilot", ManaCost: "{0}", Power: 2, Toughness: 3},
			{Name: "Enum Idea", TypeLine: "Sorcery", ManaCost: "{0}"},
		},
	}
	c.SetFace(0)
	g.Battlefield.PushTop(c)
	g.WithWriteLock(func() {
		if ok, err := g.BecomePreparedForEffect(c.InstanceID); !ok || err != nil {
			t.Fatalf("BecomePreparedForEffect = %v, %v", ok, err)
		}
	})
	for _, e := range g.Exile.Cards {
		if e.PrepareCopy {
			return e.InstanceID
		}
	}
	t.Fatal("no prepare copy in exile")
	return uuid.Nil
}

func TestEnumeratorOffersThePrepareCopyToItsController(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	copyID := preparedFixture(t, g, seat)

	moves := legal.EnumerateFor(g, seat.ID)
	got := castMovesFor(moves, copyID)
	if len(got) != 1 {
		t.Fatalf("prepare copy offered %d casts, want exactly one: %v", len(got), labels(moves))
	}
	if face := faceOf(t, got[0]); face != 1 {
		t.Errorf("offered face %d, want the prepare spell (1)", face)
	}
	dispatchAll(t, g, seat.ID, moves)
}

func TestEnumeratorOffersNoOpponentThePrepareCopy(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	copyID := preparedFixture(t, g, active)

	for _, p := range g.Seats {
		if p.ID == active.ID {
			continue
		}
		if got := castMovesFor(legal.EnumerateFor(g, p.ID), copyID); len(got) != 0 {
			t.Errorf("seat %s was offered %d casts of another player's prepare copy", p.Name, len(got))
		}
	}
}
