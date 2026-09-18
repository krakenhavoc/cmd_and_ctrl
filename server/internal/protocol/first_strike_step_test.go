package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// first_strike_step_test.go — #717's wire surface. The first-strike
// combat damage step is a step like any other: `TurnView.step` names
// it, its phase is `combat`, and the public log gets its own step line
// so the two combat damage steps read as two moments (CR 510.4).
//
// The client mirrors the step list by hand
// (`client/src/lib/turn.ts`), so the string here is the contract
// `docs/protocol.md` records.

func TestTurnViewProjectsTheFirstStrikeDamageStep(t *testing.T) {
	g := buildActiveGame(t)
	atk := g.Seats[0]
	striker := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: striker, Name: "Youthful Knight", TypeLine: "Creature — Human Knight",
			Power: 2, Toughness: 1, Keywords: []string{"first strike"},
			Owner: atk.ID, Controller: atk.ID,
		})
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(striker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepFirstStrikeDamage)

	v := ViewOfGame(g)
	if v.Turn.Step != "first_strike_damage" {
		t.Errorf("TurnView.step = %q, want %q", v.Turn.Step, "first_strike_damage")
	}
	if v.Turn.Phase != "combat" {
		t.Errorf("TurnView.phase = %q, want %q", v.Turn.Phase, "combat")
	}
	if v.Turn.PriorityHolder != v.Turn.ActiveSeat {
		t.Errorf("TurnView.priority_holder = %d, want the active seat %d — the step grants priority",
			v.Turn.PriorityHolder, v.Turn.ActiveSeat)
	}

	// The log's step spine: two damage steps, two lines, the
	// first-strike one first. log.go collapses consecutive step
	// entries that name the same step, so distinct names are what
	// keeps them apart.
	advanceTo(t, g, game.StepCombatDamage)
	var steps []string
	for _, e := range ViewOfGame(g).Log {
		if e.Kind == LogStep {
			steps = append(steps, e.Step)
		}
	}
	first, regular := -1, -1
	for i, s := range steps {
		if s == "first_strike_damage" && first < 0 {
			first = i
		}
		if s == "combat_damage" && regular < 0 {
			regular = i
		}
	}
	if first < 0 {
		t.Fatalf("no first_strike_damage step line in the log: %v", steps)
	}
	if regular < 0 {
		t.Fatalf("no combat_damage step line in the log: %v", steps)
	}
	if first > regular {
		t.Errorf("step lines out of order: %v", steps)
	}
}
