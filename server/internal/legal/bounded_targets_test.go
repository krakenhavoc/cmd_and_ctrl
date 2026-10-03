package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// bounded_targets_test.go — ADR 0109 §9's enumerator half (#1842). A
// target bounded by the counters an activation removes (Simic
// Manipulator), or by the X a variable sacrifice or tap announces
// (Ruthless Technomancer, Aryel), is judged per payment: the enumerator
// offers only (payment, target) pairs the engine accepts (#544), and
// every one of them dispatches.

const (
	oracleSimicManipulator     = "8f06fcc9-9018-4c55-af63-c44350a6cfeb"
	oracleRuthlessTechnomancer = "4e58ad76-37c7-4531-b207-6890b39a2679"
	oracleAryel                = "9956f9b7-0140-484c-b606-3685690b84cc"
)

// boundTargetIDs is the card targets an activate move names.
func boundTargetIDs(t *testing.T, m legal.Move) []string {
	t.Helper()
	var p struct {
		Targets []struct {
			ID string `json:"id"`
		} `json:"targets"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params %s: %v", string(m.Params), err)
	}
	out := make([]string, 0, len(p.Targets))
	for _, tg := range p.Targets {
		out = append(out, tg.ID)
	}
	return out
}

// boundedTargetsOffered collects every target the source's activations
// name, failing on any outside `allowed`.
func boundedTargetsOffered(t *testing.T, moves []legal.Move, source uuid.UUID, allowed map[string]bool) map[string]bool {
	t.Helper()
	seen := map[string]bool{}
	for _, m := range activationsOf(moves, source) {
		for _, id := range boundTargetIDs(t, m) {
			seen[id] = true
			if !allowed[id] {
				t.Errorf("%q targets %s, which its payment's bound excludes", m.Label, id)
			}
		}
	}
	return seen
}

func TestSimicManipulatorIsOfferedOnlyTargetsWithinItsCounters(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	man := battlefieldCard(g, active, game.Card{
		Name: "Simic Manipulator", TypeLine: "Creature — Mutant Wizard", OracleID: oracleSimicManipulator,
		Power: 0, Toughness: 1, Counters: map[string]int{game.CounterPlusOne: 2},
	})
	one := battlefieldCard(g, opp, creature("One", "{G}", 1, 1))
	two := battlefieldCard(g, opp, creature("Two", "{1}{G}", 2, 2))
	three := battlefieldCard(g, opp, creature("Three", "{2}{G}", 3, 3))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	// The Manipulator itself is a power-2 creature while its target is
	// chosen (CR 601.2c comes before the payment, CR 601.2h), so it is
	// a legal target of its own two-counter activation.
	seen := boundedTargetsOffered(t, moves, man, map[string]bool{one.String(): true, two.String(): true, man.String(): true})
	if !seen[two.String()] {
		t.Errorf("the power-2 creature was never offered: %v", labels(moves))
	}
	if seen[three.String()] {
		t.Error("the power-3 creature was offered for two counters")
	}
	dispatchAll(t, g, active.ID, moves)

	// The wire ships the superset with each candidate's power and says
	// the X is the counters removed, so the client narrows by the same
	// number the engine binds.
	v := protocol.ViewOfGameFor(g, active.ID.String())
	var lt *protocol.LegalTargetsView
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID == man.String() && len(c.ActivatedAbilities) > 0 {
			lt = c.ActivatedAbilities[0].LegalTargets
		}
	}
	if lt == nil {
		t.Fatal("the Manipulator's ability ships no legal targets")
	}
	if !lt.PowerAtMostX || !lt.XFromCountersRemoved || lt.ManaValueAtMostX {
		t.Errorf("bound flags = power %v counters %v mv %v", lt.PowerAtMostX, lt.XFromCountersRemoved, lt.ManaValueAtMostX)
	}
	if lt.Powers[three.String()] != 3 || lt.Powers[one.String()] != 1 {
		t.Errorf("powers = %v", lt.Powers)
	}
}

func TestTechnomancerAndAryelAreOfferedOnlyTargetsWithinTheirCounts(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	tech := battlefieldCard(g, active, game.Card{
		Name: "Ruthless Technomancer", TypeLine: "Creature — Human Wizard", OracleID: oracleRuthlessTechnomancer,
		Power: 2, Toughness: 4,
	})
	aryel := battlefieldCard(g, active, game.Card{
		Name: "Aryel, Knight of Windgrace", TypeLine: "Legendary Creature — Human Knight", OracleID: oracleAryel,
		Power: 4, Toughness: 4,
	})
	battlefieldCard(g, active, game.Card{Name: "Trinket", TypeLine: "Artifact"})
	squire := battlefieldCard(g, active, game.Card{Name: "Squire", TypeLine: "Creature — Human Knight", Power: 1, Toughness: 1})
	deadOne := graveyardCard(active, game.Card{Name: "Dead One", TypeLine: "Creature — Test", ManaCost: "{B}", Power: 1, Toughness: 1})
	deadThree := graveyardCard(active, game.Card{Name: "Dead Three", TypeLine: "Creature — Test", ManaCost: "{2}{B}", Power: 3, Toughness: 3})
	oppOne := battlefieldCard(g, opp, creature("Opp One", "{G}", 1, 1))
	oppThree := battlefieldCard(g, opp, creature("Opp Three", "{2}{G}", 3, 3))
	for i := 0; i < 2; i++ {
		battlefieldCard(g, active, basic("Swamp", "Swamp"))
	}
	mana(g, active, 3)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	techSeen := boundedTargetsOffered(t, moves, tech, map[string]bool{deadOne.String(): true})
	if !techSeen[deadOne.String()] {
		t.Errorf("Technomancer never offered the power-1 card for its one artifact: %v", labels(moves))
	}
	if techSeen[deadThree.String()] {
		t.Error("Technomancer offered a power-3 card for one artifact")
	}
	// Aryel's token row names no target; her destroy row's targets are
	// bounded by the one untapped Knight beside her — which may be its
	// own target, a power-1 creature.
	seen := boundedTargetsOffered(t, moves, aryel, map[string]bool{oppOne.String(): true, squire.String(): true})
	if seen[oppThree.String()] {
		t.Error("Aryel offered a power-3 creature for one Knight")
	}
	dispatchAll(t, g, active.ID, moves)
}
