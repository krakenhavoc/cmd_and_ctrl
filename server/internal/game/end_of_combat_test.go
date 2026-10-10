package game

import (
	"testing"

	"github.com/google/uuid"
)

// end_of_combat_test.go — #785, CR 511.3: "as soon as the end of
// combat step ends, all creatures, battles and planeswalkers are
// removed from combat."
//
// The engine used to clear combat on ENTRY to end_combat, so every
// attacker and blocker stopped being in combat for the whole of the
// step the rules keep them in. Nothing cast or activated there could
// see an attacking creature: Aetherize, Settle the Wreckage and
// Aetherspouts found nothing, and Desert — "target attacking creature,
// activate only during the end of combat step" — had no legal target
// in the only step it can be activated.

// Attackers and blockers are still in combat for the whole end of
// combat step, and leave it as the step ends.
func TestCreaturesLeaveCombatAsTheEndOfCombatStepEnds(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 2, 2)
	blocker := pushCombatant(t, g, g.Seats[1], "Blocker", 1, 4)
	blockAfterLockIn(t, g, attacker, blocker)

	passUntilStep(t, g, StepEndCombat)
	if c := findCard(g, attacker); c == nil || c.AttackingTarget == uuid.Nil {
		t.Errorf("the attacker is not attacking in the end of combat step (CR 511.3)")
	}
	if c := findCard(g, blocker); c == nil || c.BlockingTarget == uuid.Nil {
		t.Errorf("the blocker is not blocking in the end of combat step (CR 511.3)")
	}
	if !g.blockedAttackers[attacker] {
		t.Errorf("the blocked state was dropped on entry to the end of combat step")
	}

	passUntilStep(t, g, StepPostcombatMain)
	for _, c := range g.Battlefield.Cards {
		if c.AttackingTarget != uuid.Nil || c.BlockingTarget != uuid.Nil {
			t.Errorf("%s is still in combat after the end of combat step ended", c.Name)
		}
	}
	if len(g.blockedAttackers) != 0 || len(g.announcedBlocks) != 0 || len(g.announcedAttacks) != 0 {
		t.Errorf("the combat's bookkeeping survived the end of combat step")
	}
}

// endOfCombatProbe is a test Listener that answers, at the instant the
// end of combat step is announced, whether the attacker is still
// attacking — the question an "at end of combat" trigger harvested
// from that announcement asks.
type endOfCombatProbe struct {
	attacker  uuid.UUID
	announced bool
	attacking bool
}

func (p *endOfCombatProbe) OnEvent(g *Game, ev Event) {
	if ev.Kind != EventStepBegan || ev.Step != StepEndCombat {
		return
	}
	p.announced = true
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID == p.attacker && c.AttackingTarget != uuid.Nil {
			p.attacking = true
		}
	}
}

// "At end of combat" triggers fire as the step BEGINS (CR 511.2), and
// they see the attackers, because the removal is at the other end of
// the step. The engine's step announcement is what the card-side
// AtEndOfYourCombat constructor watches, so this pins the order:
// EventStepBegan for end_combat, with the attacker still attacking.
func TestEndOfCombatStepAnnouncementSeesTheAttackers(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 2, 2)
	declareAttacks(t, g, attacker)

	probe := &endOfCombatProbe{attacker: attacker}
	g.RegisterListener(probe)

	passUntilStep(t, g, StepEndCombat)
	if !probe.announced {
		t.Fatalf("the end of combat step was never announced")
	}
	if !probe.attacking {
		t.Errorf("an at-end-of-combat trigger harvested at the step's announcement saw no attacker")
	}
}

// The turn after the combat is untouched: the next untap step untaps
// the attacker, which is no longer in combat and no longer tapped.
func TestTheNextTurnsUntapIsUnaffected(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 2, 2)
	declareAttacks(t, g, attacker)
	if c := findCard(g, attacker); c == nil || !c.Tapped {
		t.Fatalf("setup: the attacker was not tapped by attacking")
	}

	// Round the table back to seat 0's turn.
	for range len(g.Seats) {
		if err := g.PassTurn(); err != nil {
			t.Fatalf("PassTurn: %v", err)
		}
	}
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("expected to be back on seat 0's turn, at seat %d", g.Turn.ActiveSeat)
	}
	c := findCard(g, attacker)
	if c == nil {
		t.Fatalf("the attacker is gone")
	}
	if c.Tapped {
		t.Errorf("the attacker did not untap on its controller's next untap step")
	}
	if c.AttackingTarget != uuid.Nil {
		t.Errorf("the attacker is still attacking a turn later")
	}
}

// A turn that ends early from inside the end of combat step never
// reaches the step-exit seam, so PassTurn keeps its own clear.
func TestPassTurnFromTheEndOfCombatStepStillClearsCombat(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 2, 2)
	declareAttacks(t, g, attacker)
	passUntilStep(t, g, StepEndCombat)

	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if c := findCard(g, attacker); c == nil || c.AttackingTarget != uuid.Nil {
		t.Errorf("PassTurn out of the end of combat step left the attacker in combat")
	}
}
