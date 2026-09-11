package actions

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// declare_attackers_test.go covers the bulk declare_attackers verb
// (#318) at the dispatch layer: shape validation, the batch cap, and
// the split between "ineligible, so skipped" and "not yours, so the
// whole batch is rejected".

type attackerEntry struct {
	Attacker string `json:"attacker"`
	Target   string `json:"target"`
}

func bulkParams(t *testing.T, entries ...attackerEntry) json.RawMessage {
	t.Helper()
	return params(t, map[string]any{"attackers": entries})
}

func TestDispatchDeclareAttackersBulk(t *testing.T) {
	g := newGame(t)
	a := pushCreature(t, g, g.Seats[0])
	b := pushCreature(t, g, g.Seats[0])
	advanceTo(t, g, game.StepDeclareAttackers)
	target := g.Seats[1].ID.String()

	act := mustAction(t, TypeDeclareAttackers, bulkParams(t,
		attackerEntry{Attacker: a, Target: target},
		attackerEntry{Attacker: b, Target: target},
	))
	act.Caller = g.Seats[0].ID
	if err := Dispatch(g, act); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	declared := 0
	for _, c := range g.Battlefield.Cards {
		if c.AttackingTarget.String() == target {
			declared++
		}
	}
	if declared != 2 {
		t.Errorf("declared %d attackers, want 2", declared)
	}
}

func TestDispatchDeclareAttackersEmptySet(t *testing.T) {
	g := newGame(t)
	advanceTo(t, g, game.StepDeclareAttackers)
	act := mustAction(t, TypeDeclareAttackers, bulkParams(t))
	if err := Dispatch(g, act); !errors.Is(err, ErrEmptyAttackerSet) {
		t.Fatalf("Dispatch err = %v, want ErrEmptyAttackerSet", err)
	}
}

func TestDispatchDeclareAttackersMissingParams(t *testing.T) {
	g := newGame(t)
	advanceTo(t, g, game.StepDeclareAttackers)
	act := mustAction(t, TypeDeclareAttackers, nil)
	if err := Dispatch(g, act); !errors.Is(err, ErrMissingParams) {
		t.Fatalf("Dispatch err = %v, want ErrMissingParams", err)
	}
}

func TestDispatchDeclareAttackersBatchCap(t *testing.T) {
	g := newGame(t)
	advanceTo(t, g, game.StepDeclareAttackers)
	entries := make([]attackerEntry, 0, MaxBulkAttackers+1)
	for range MaxBulkAttackers + 1 {
		entries = append(entries, attackerEntry{
			Attacker: uuid.NewString(),
			Target:   g.Seats[1].ID.String(),
		})
	}
	act := mustAction(t, TypeDeclareAttackers, bulkParams(t, entries...))
	if err := Dispatch(g, act); !errors.Is(err, ErrTooManyAttackers) {
		t.Fatalf("Dispatch err = %v, want ErrTooManyAttackers", err)
	}
}

// A creature the caller doesn't control is an authorization failure,
// not an eligibility one: it rejects the whole batch rather than being
// quietly dropped, so a client can never half-declare by accident.
func TestDispatchDeclareAttackersRejectsForeignCreature(t *testing.T) {
	g := newGame(t)
	mine := pushCreature(t, g, g.Seats[0])
	theirs := pushCreature(t, g, g.Seats[1])
	advanceTo(t, g, game.StepDeclareAttackers)
	target := g.Seats[1].ID.String()

	act := mustAction(t, TypeDeclareAttackers, bulkParams(t,
		attackerEntry{Attacker: mine, Target: target},
		attackerEntry{Attacker: theirs, Target: target},
	))
	act.Caller = g.Seats[0].ID
	if err := Dispatch(g, act); !errors.Is(err, game.ErrCardCallerMismatch) {
		t.Fatalf("Dispatch err = %v, want ErrCardCallerMismatch", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.AttackingTarget != uuid.Nil {
			t.Errorf("a rejected batch must not declare anything")
		}
	}
}

// One ineligible creature must not sink the strike — the eligible
// ones still attack.
func TestDispatchDeclareAttackersSkipsIneligible(t *testing.T) {
	g := newGame(t)
	fine := pushCreature(t, g, g.Seats[0])
	advanceTo(t, g, game.StepDeclareAttackers)
	target := g.Seats[1].ID.String()

	act := mustAction(t, TypeDeclareAttackers, bulkParams(t,
		attackerEntry{Attacker: fine, Target: target},
		attackerEntry{Attacker: uuid.NewString(), Target: target},
	))
	act.Caller = g.Seats[0].ID
	if err := Dispatch(g, act); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	declared := 0
	for _, c := range g.Battlefield.Cards {
		if c.AttackingTarget != uuid.Nil {
			declared++
		}
	}
	if declared != 1 {
		t.Errorf("declared %d attackers, want 1", declared)
	}
}

func TestDispatchDeclareAttackersAllIneligible(t *testing.T) {
	g := newGame(t)
	advanceTo(t, g, game.StepDeclareAttackers)
	act := mustAction(t, TypeDeclareAttackers, bulkParams(t,
		attackerEntry{Attacker: uuid.NewString(), Target: g.Seats[1].ID.String()},
	))
	if err := Dispatch(g, act); !errors.Is(err, game.ErrNoLegalAttackers) {
		t.Fatalf("Dispatch err = %v, want ErrNoLegalAttackers", err)
	}
}

func TestDispatchDeclareAttackersBadUUID(t *testing.T) {
	g := newGame(t)
	advanceTo(t, g, game.StepDeclareAttackers)
	act := mustAction(t, TypeDeclareAttackers, bulkParams(t,
		attackerEntry{Attacker: "not-a-uuid", Target: g.Seats[1].ID.String()},
	))
	if err := Dispatch(g, act); err == nil {
		t.Fatalf("Dispatch accepted a malformed attacker UUID")
	}
}
