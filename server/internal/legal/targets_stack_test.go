package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// targets_stack_test.go — #1307. Smart autopass needs to tell a
// counterspell-shaped response (a move that answers something already
// on the stack) apart from any other instant or activated ability
// that merely happens to be available. `Move.TargetsStack` is that
// signal; these scenarios pin it against the engine's own stack, the
// same way target_ability_test.go pins #1211's ability targeting.

// Counterspell targeting a spell on the stack sets targets_stack.
func TestCounterspellFlagsTargetsStack(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(active)
	clearHand(opp)
	bolt := handCard(active, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleLightningBolt})
	cs := handCard(opp, game.Card{Name: "Counterspell", TypeLine: "Instant", ManaCost: "{U}{U}", OracleID: oracleCounterspell})
	battlefieldCard(g, active, basic("Mountain", "Mountain"))
	battlefieldCard(g, opp, basic("Island", "Island"))
	battlefieldCard(g, opp, basic("Island", "Island"))
	advanceTo(t, g, game.StepPrecombatMain)

	// active holds priority coming out of advanceTo; casting Bolt
	// puts it on the stack, and one pass hands priority to opp with
	// it still there.
	if err := g.CastSpell(active.ID, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		Strict:  true, AutoTap: true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}

	moves := legal.EnumerateFor(g, opp.ID)
	dispatchAll(t, g, opp.ID, moves)
	var found *legal.Move
	for i := range moves {
		if moves[i].Source == cs && moves[i].Kind == legal.KindCast {
			found = &moves[i]
		}
	}
	if found == nil {
		t.Fatalf("Counterspell not offered against the Bolt on the stack: %v", labels(moves))
	}
	if !found.TargetsStack {
		t.Errorf("Counterspell targeting a spell on the stack should set targets_stack, got %+v", found)
	}
}

// A burn spell targeting a player is offense, not a response — it
// never sets targets_stack, empty stack or not.
func TestBurnSpellNeverFlagsTargetsStack(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	bolt := handCard(active, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleLightningBolt})
	battlefieldCard(g, active, basic("Mountain", "Mountain"))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	var saw bool
	for _, m := range moves {
		if m.Source != bolt {
			continue
		}
		saw = true
		if m.TargetsStack {
			t.Errorf("Lightning Bolt should never set targets_stack, got %+v", m)
		}
	}
	if !saw {
		t.Fatal("Lightning Bolt was not offered at all")
	}
}

// Stifle targeting an activated or triggered ability on the stack
// (CR 115.4, #1211) sets targets_stack even though its target is not
// a card in any zone — StackMeta's synthetic id is what the flag
// reads, exactly as game.stackEmpty does.
func TestStifleTargetingAnAbilityFlagsTargetsStack(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 1, "Island")
	rock := battlefieldCard(g, other, game.Card{Name: "Their Rock", TypeLine: "Artifact"})
	announceTriggerForTest(t, g, other.ID, rock, "Their Rock — do a thing")

	stifleID := handCard(seat, stifle())
	moves := legal.EnumerateFor(g, seat.ID)
	dispatchAll(t, g, seat.ID, moves)
	var found *legal.Move
	for i := range moves {
		if moves[i].Source == stifleID && moves[i].Kind == legal.KindCast {
			found = &moves[i]
		}
	}
	if found == nil {
		t.Fatalf("Stifle not offered against the ability on the stack: %v", labels(moves))
	}
	if !found.TargetsStack {
		t.Errorf("Stifle targeting an ability on the stack should set targets_stack, got %+v", found)
	}
}

// A modal spell with one counter mode and one non-counter mode
// (Izzet Charm) flags only the announcement that actually chose a
// stack target — the OTHER mode, offered in the very same
// enumeration, must not be flagged.
func TestModalSpellFlagsOnlyTheStackTargetingMode(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(active)
	clearHand(opp)
	charm := handCard(active, game.Card{Name: "Izzet Charm", TypeLine: "Instant", ManaCost: "{U}{R}", OracleID: oracleIzzetCharm})
	bolt := handCard(opp, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleLightningBolt})
	battlefieldCard(g, active, basic("Island", "Island"))
	battlefieldCard(g, active, basic("Mountain", "Mountain"))
	battlefieldCard(g, opp, creature("Bear", "{1}{G}", 2, 2))
	battlefieldCard(g, opp, basic("Mountain", "Mountain"))
	advanceTo(t, g, game.StepPrecombatMain)

	// opp casts Bolt at active without needing priority (as
	// non_hand_cast_test.go's castSpellForTest also relies on); active
	// still holds priority from advanceTo and is enumerated with the
	// Bolt on the stack.
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: active.ID}},
		Strict:  true, AutoTap: true,
	}); err != nil {
		t.Fatal(err)
	}

	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)

	var sawCounterMode, sawDamageMode bool
	for _, m := range moves {
		if m.Source != charm {
			continue
		}
		var p struct {
			Modes []int `json:"modes"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		switch {
		case len(p.Modes) == 1 && p.Modes[0] == 0:
			// "Counter target noncreature spell unless its controller
			// pays {2}" — targets the Bolt on the stack.
			sawCounterMode = true
			if !m.TargetsStack {
				t.Errorf("Izzet Charm's counter mode should set targets_stack, got %+v", m)
			}
		case len(p.Modes) == 1 && p.Modes[0] == 1:
			// "Izzet Charm deals 2 damage to target creature" —
			// targets the Bear, not the stack.
			sawDamageMode = true
			if m.TargetsStack {
				t.Errorf("Izzet Charm's damage mode should not set targets_stack, got %+v", m)
			}
		}
	}
	if !sawCounterMode || !sawDamageMode {
		t.Fatalf("want both Izzet Charm's counter and damage modes offered: counter=%v damage=%v (%v)",
			sawCounterMode, sawDamageMode, labels(moves))
	}
}
