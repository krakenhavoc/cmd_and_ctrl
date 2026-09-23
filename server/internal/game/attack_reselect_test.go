package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// attack_reselect_test.go — CR 508.7, reselecting what an attacking
// creature is attacking (#1329).

// declaredAttackerAt declares `attacker` against `defender` and locks
// the declaration in, leaving the cursor in declare_attackers — the
// window every printed reselect card resolves in.
func declaredAttackerAt(t *testing.T, g *Game, attacker, defender uuid.UUID) {
	t.Helper()
	advanceTo(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if !g.announcedAttacks[attacker] {
		t.Fatalf("setup: the declaration was not locked in")
	}
}

func reselect(g *Game, attacker, target uuid.UUID) error {
	var err error
	g.WithWriteLock(func() { err = g.ReselectAttackTargetForEffect(attacker, target) })
	return err
}

// The creature attacks the new player, is not announced again (no
// second EventAttack, CR 508.7a), and nothing else about it changes.
func TestReselectMovesTheAttackWithoutAttackingAgain(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushReadyAttackerForTest(g, g.Seats[0].ID, "Attacker", 3)
	declaredAttackerAt(t, g, attacker, g.Seats[1].ID)
	seq := lastSeq(g)

	if err := reselect(g, attacker, g.Seats[2].ID); err != nil {
		t.Fatalf("ReselectAttackTargetForEffect: %v", err)
	}
	c := findCard(g, attacker)
	if c.AttackingTarget != g.Seats[2].ID {
		t.Fatalf("the attacker is attacking %v, want seat 2", c.AttackingTarget)
	}
	if !g.announcedAttacks[attacker] {
		t.Errorf("the attacker lost its announcement; CR 508.7a keeps it in combat")
	}
	for _, ev := range g.Events {
		if ev.Seq > seq && ev.Kind == EventAttack {
			t.Errorf("a reselection announced a second attack (CR 508.7a): %+v", ev)
		}
	}
	// A later lock-in has nothing to announce either.
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	for _, ev := range g.Events {
		if ev.Seq > seq && ev.Kind == EventAttack {
			t.Errorf("the lock-in re-announced a reselected attacker: %+v", ev)
		}
	}
}

// CR 508.7c: not the attacker's controller, not a planeswalker they
// control. Checked against the ATTACKING creature's controller.
func TestReselectRefusesTheAttackersOwnSide(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushReadyAttackerForTest(g, g.Seats[0].ID, "Attacker", 3)
	ownWalker := pushPlaneswalkerForTest(g, g.Seats[0].ID, "Own Walker", 3)
	declaredAttackerAt(t, g, attacker, g.Seats[1].ID)

	for name, target := range map[string]uuid.UUID{
		"its controller":                 g.Seats[0].ID,
		"a planeswalker its side runs":   ownWalker,
		"something that is not a target": uuid.New(),
	} {
		if err := reselect(g, attacker, target); !errors.Is(err, ErrIllegalAttackTarget) {
			t.Errorf("reselecting onto %s: err = %v, want ErrIllegalAttackTarget", name, err)
		}
	}
	if c := findCard(g, attacker); c.AttackingTarget != g.Seats[1].ID {
		t.Errorf("a refused reselection moved the attack")
	}
}

// Only an attacking creature can be reselected — and a staged,
// not-yet-locked-in declaration is still the active player's to change.
func TestReselectRefusesANonAttacker(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	idle := pushReadyAttackerForTest(g, g.Seats[0].ID, "Idle", 2)
	staged := pushReadyAttackerForTest(g, g.Seats[0].ID, "Staged", 2)
	advanceTo(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(staged, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := reselect(g, idle, g.Seats[2].ID); !errors.Is(err, ErrNotAttacking) {
		t.Errorf("a creature not in combat: err = %v, want ErrNotAttacking", err)
	}
	if err := reselect(g, staged, g.Seats[2].ID); !errors.Is(err, ErrNotAttacking) {
		t.Errorf("a staged declaration: err = %v, want ErrNotAttacking", err)
	}
}

// The new defender is the defending player for blocks — the one
// option generator (ADR 0045 Decision 14) offers the attacker to the
// new defender's creatures and not to the old defender's — and takes
// the unblocked damage (CR 510.1b).
func TestReselectedAttackerIsBlockedAndDealsDamageAsTheNewTarget(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushReadyAttackerForTest(g, g.Seats[0].ID, "Attacker", 3)
	declaredAttackerAt(t, g, attacker, g.Seats[1].ID)
	if err := reselect(g, attacker, g.Seats[2].ID); err != nil {
		t.Fatalf("ReselectAttackTargetForEffect: %v", err)
	}
	pushCombatant(t, g, g.Seats[1], "Old Defender's", 1, 1)
	pushCombatant(t, g, g.Seats[2], "New Defender's", 1, 1)
	advanceTo(t, g, StepDeclareBlockers)
	var oldOpts, newOpts int
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		oldOpts = len(g.BlockOptionsLocked(g.Seats[1].ID, 1))
		newOpts = len(g.BlockOptionsLocked(g.Seats[2].ID, 1))
	})
	if oldOpts != 0 {
		t.Errorf("the player it no longer attacks is offered %d blocks", oldOpts)
	}
	if newOpts == 0 {
		t.Errorf("the player it now attacks is offered no block")
	}

	life1, life2 := g.Seats[1].Life, g.Seats[2].Life
	passUntilStep(t, g, StepCombatDamage)
	if g.Seats[2].Life != life2-3 || g.Seats[1].Life != life1 {
		t.Errorf("life after damage: seat1 %d (was %d), seat2 %d (was %d); the 3 should hit seat 2",
			g.Seats[1].Life, life1, g.Seats[2].Life, life2)
	}
}

// A blocked creature stays blocked by the same blockers (CR 508.7a —
// it is not removed from combat).
func TestReselectKeepsABlockedAttackerBlocked(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 3, 3)
	blocker := pushCombatant(t, g, g.Seats[1], "Blocker", 1, 4)
	blockAfterLockIn(t, g, attacker, blocker)

	if err := reselect(g, attacker, g.Seats[2].ID); err != nil {
		t.Fatalf("ReselectAttackTargetForEffect: %v", err)
	}
	if !g.blockedAttackers[attacker] {
		t.Errorf("a reselection unblocked the attacker")
	}
	if c := findCard(g, blocker); c == nil || c.BlockingTarget != attacker {
		t.Errorf("a reselection removed the blocker from combat")
	}
}

// queueReselect queues the prompt for `attacker`, chosen by `chooser`,
// and returns it.
func queueReselect(t *testing.T, g *Game, chooser, attacker uuid.UUID, then func(*Game) error) *PendingChoice {
	t.Helper()
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueReselectAttackForEffect(ReselectAttackPrompt{
			Chooser: chooser, Attacker: attacker, Question: "reselect?", Then: then,
		})
	})
	if id == uuid.Nil {
		t.Fatalf("no prompt was queued")
	}
	for _, c := range g.PendingChoices {
		if c.ID == id {
			return c
		}
	}
	t.Fatalf("the queued prompt is not in the queue")
	return nil
}

// The prompt: "keep" first (the always-legal branch), then every legal
// target but the current one — seats as seats, a planeswalker as a
// card. Picking a planeswalker moves the attack onto it; "keep" moves
// nothing; Then runs either way.
func TestReselectPromptOffersKeepThenEveryOtherTarget(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushReadyAttackerForTest(g, g.Seats[0].ID, "Attacker", 3)
	walker := pushPlaneswalkerForTest(g, g.Seats[3].ID, "Their Walker", 4)
	declaredAttackerAt(t, g, attacker, g.Seats[1].ID)
	chooser := g.Seats[1].ID // a defending player flashing in the card

	thenRuns := 0
	c := queueReselect(t, g, chooser, attacker, func(*Game) error { thenRuns++; return nil })
	if c.Kind != PendingChoiceOptionPick {
		t.Fatalf("kind %s, want option_pick", c.Kind)
	}
	if got := c.PickOptions[0]; got.Player != uuid.Nil || len(got.Cards) != 0 {
		t.Fatalf("the first option is not the no-change \"keep\" branch: %+v", got)
	}
	var seats, cards int
	walkerIdx := -1
	for i, opt := range c.PickOptions[1:] {
		if opt.Player == g.Seats[1].ID {
			t.Errorf("the current target is offered as a change")
		}
		if opt.Player == g.Seats[0].ID {
			t.Errorf("the attacker's own controller is offered (CR 508.7c)")
		}
		if opt.Player != uuid.Nil {
			seats++
		}
		if len(opt.Cards) == 1 && opt.Cards[0] == walker {
			cards++
			walkerIdx = i + 1
		}
	}
	if seats != 2 || cards != 1 {
		t.Fatalf("offered %d seats and %d planeswalkers, want 2 and 1", seats, cards)
	}
	if err := g.ResolveOptionPick(c.ID, chooser, walkerIdx); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if findCard(g, attacker).AttackingTarget != walker {
		t.Errorf("picking the planeswalker did not move the attack onto it")
	}

	// And "keep" leaves it where it is.
	c = queueReselect(t, g, chooser, attacker, func(*Game) error { thenRuns++; return nil })
	if err := g.ResolveOptionPick(c.ID, chooser, 0); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if findCard(g, attacker).AttackingTarget != walker {
		t.Errorf("\"keep\" moved the attack")
	}
	if thenRuns != 2 {
		t.Errorf("Then ran %d times, want once per answered prompt", thenRuns)
	}
}

// #994's hazard, for this prompt: a seat that concedes while the
// question is open is pruned off it, which renumbers every option
// after it. The answer is read off the option picked, so the attack
// lands on the seat the chooser clicked, not the one a stale index
// names.
func TestReselectPromptAnswerSurvivesASeatPrunedMidQuestion(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushReadyAttackerForTest(g, g.Seats[0].ID, "Attacker", 3)
	declaredAttackerAt(t, g, attacker, g.Seats[1].ID)
	c := queueReselect(t, g, g.Seats[0].ID, attacker, nil)
	// Options: keep, seat 2, seat 3.
	if len(c.PickOptions) != 3 || c.PickOptions[1].Player != g.Seats[2].ID || c.PickOptions[2].Player != g.Seats[3].ID {
		t.Fatalf("unexpected option list: %+v", c.PickOptions)
	}
	if err := g.Concede(g.Seats[2].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if len(c.PickOptions) != 2 || c.PickOptions[1].Player != g.Seats[3].ID {
		t.Fatalf("the departed seat was not pruned: %+v", c.PickOptions)
	}
	if err := g.ResolveOptionPick(c.ID, g.Seats[0].ID, 1); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if got := findCard(g, attacker).AttackingTarget; got != g.Seats[3].ID {
		t.Errorf("the attack went to %v, want seat 3 — the option that was clicked", got)
	}
}

// Nothing to ask — not attacking, or nowhere else to go — queues
// nothing and still runs Then, so a per-attacker chain moves on.
func TestReselectPromptWithNothingToAskRunsThen(t *testing.T) {
	g := newActiveGameWithSeats(t, 2)
	attacker := pushReadyAttackerForTest(g, g.Seats[0].ID, "Attacker", 3)
	declaredAttackerAt(t, g, attacker, g.Seats[1].ID)
	ran := false
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueReselectAttackForEffect(ReselectAttackPrompt{
			Chooser: g.Seats[0].ID, Attacker: attacker,
			Then: func(*Game) error { ran = true; return nil },
		})
	})
	if id != uuid.Nil {
		t.Errorf("a two-player table has nowhere else to attack, but a prompt was queued")
	}
	if !ran {
		t.Errorf("Then did not run when there was nothing to ask")
	}
}
