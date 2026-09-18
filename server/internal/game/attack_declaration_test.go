package game

import (
	"testing"

	"github.com/google/uuid"
)

// attack_declaration_test.go covers #859 — the attack declaration is
// announced at its LOCK-IN, not at each click.
//
// CR 508.1 declares attackers as one turn-based action and declares
// each creature as an attacker once; the sandbox lets the active
// player re-point an attacker from one defender to another before the
// declaration is complete. Announcing per click meant the single
// EventAttack the creature is owed kept naming the defender it was
// FIRST pointed at, so every "attacks <player>" reader — and every
// trigger condition gated on the defender — saw a player the attacker
// had already left.

// lockInAttackDeclaration is the lock-in, called directly so a test
// can stay in the declare-attackers step across several of them. The
// two production paths into it (the priority wrap and the cursor
// leaving the step) have their own tests below.
func lockInAttackDeclaration(t *testing.T, g *Game) {
	t.Helper()
	g.WithWriteLock(func() { g.commitAttackDeclarationLocked() })
}

// lockInAttacksByPriorityWrap reaches the lock-in the way play does:
// priority passes around the table until it wraps, which is the
// declaration being complete (CR 508.1). Stops early if the pass
// moved the cursor out of the step.
func lockInAttacksByPriorityWrap(t *testing.T, g *Game) {
	t.Helper()
	step := g.Turn.Step
	for i := 0; i < len(g.Seats)+1; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
		if g.Turn.Step != step {
			return
		}
	}
}

// attackDeclEvents returns the logged EventAttacks naming `card`.
func attackDeclEvents(g *Game, card uuid.UUID) []Event {
	var out []Event
	for _, ev := range g.Events {
		if ev.Kind == EventAttack && ev.CardID == card {
			out = append(out, ev)
		}
	}
	return out
}

// attackSetup is a four-player table parked in the declare-attackers
// step with two ready attackers under the active seat.
func attackSetup(t *testing.T, g *Game) (a, b uuid.UUID) {
	t.Helper()
	a = pushPlainCreature(t, g, g.Seats[0], "Alpha")
	b = pushPlainCreature(t, g, g.Seats[0], "Bravo")
	advanceIntoStep(t, g, StepDeclareAttackers)
	return a, b
}

// TestDeclareAttackerAnnouncesNothingBeforeTheLockIn is the half of
// #859 that makes the rest possible: a click stages the attack and
// emits no event, so nothing has been harvested off it yet.
func TestDeclareAttackerAnnouncesNothingBeforeTheLockIn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, _ := attackSetup(t, g)

	if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if n := len(attackDeclEvents(g, a)); n != 0 {
		t.Fatalf("the click announces nothing: %d attack events", n)
	}
	card := findCard(g, a)
	if card.AttackingTarget != g.Seats[1].ID {
		t.Error("the attack is staged on the card all the same")
	}
	if !card.Tapped {
		t.Error("CR 508.1f: declaring still taps at the click")
	}
}

// TestRepointedAttackerAnnouncesOnlyTheFinalDefender is the #859
// repro at the event level: A → B is one attack, naming B.
func TestRepointedAttackerAnnouncesOnlyTheFinalDefender(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, _ := attackSetup(t, g)

	if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(a, g.Seats[2].ID); err != nil {
		t.Fatalf("re-point: %v", err)
	}
	lockInAttackDeclaration(t, g)

	evs := attackDeclEvents(g, a)
	if len(evs) != 1 {
		t.Fatalf("a re-point is one attack (CR 508.1): %d attack events", len(evs))
	}
	if evs[0].Target != g.Seats[2].ID {
		t.Errorf("Target = %s, want the defender the attacker ended on (%s)", evs[0].Target, g.Seats[2].ID)
	}
}

// TestAttackerRepointedBackAnnouncesOnce — A → B → A is still one
// declaration, and it names A.
func TestAttackerRepointedBackAnnouncesOnce(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, _ := attackSetup(t, g)

	for _, target := range []uuid.UUID{g.Seats[1].ID, g.Seats[2].ID, g.Seats[1].ID} {
		if err := g.DeclareAttacker(a, target); err != nil {
			t.Fatalf("DeclareAttacker %s: %v", target, err)
		}
	}
	lockInAttackDeclaration(t, g)

	evs := attackDeclEvents(g, a)
	if len(evs) != 1 {
		t.Fatalf("three clicks are one declaration: %d attack events", len(evs))
	}
	if evs[0].Target != g.Seats[1].ID {
		t.Errorf("Target = %s, want %s", evs[0].Target, g.Seats[1].ID)
	}
}

// TestRepointFromPlaneswalkerToPlayerAnnouncesThePlayer — S27 widened
// the defender to a planeswalker or a battle, and a re-point can
// cross that line. "Attacked a player" is a trigger CONDITION on
// several cards, so the event has to carry the final classification,
// not the first.
func TestRepointFromPlaneswalkerToPlayerAnnouncesThePlayer(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, _ := attackSetup(t, g)
	walker := pushPlaneswalkerForTest(g, g.Seats[1].ID, "Test Walker", 4)

	if err := g.DeclareAttacker(a, walker); err != nil {
		t.Fatalf("DeclareAttacker at the planeswalker: %v", err)
	}
	if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
		t.Fatalf("re-point to the player: %v", err)
	}
	lockInAttackDeclaration(t, g)

	evs := attackDeclEvents(g, a)
	if len(evs) != 1 {
		t.Fatalf("%d attack events, want 1", len(evs))
	}
	if evs[0].Target != g.Seats[1].ID {
		t.Errorf("Target = %s, want the player %s the attacker ended on", evs[0].Target, g.Seats[1].ID)
	}
	// And the other way: player → planeswalker.
	if err := g.ClearCombat(); err != nil {
		t.Fatalf("ClearCombat: %v", err)
	}
	if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker at the player: %v", err)
	}
	if err := g.DeclareAttacker(a, walker); err != nil {
		t.Fatalf("re-point to the planeswalker: %v", err)
	}
	lockInAttackDeclaration(t, g)
	evs = attackDeclEvents(g, a)
	if len(evs) != 2 {
		t.Fatalf("%d attack events across two combats, want 2", len(evs))
	}
	if evs[1].Target != walker {
		t.Errorf("Target = %s, want the planeswalker %s", evs[1].Target, walker)
	}
}

// TestRepointedAttackerIsStillOneBatch — two attackers, one of them
// re-pointed, are one declaration and therefore one event batch, so a
// OncePerBatch ability collapses them (#854, CR 603.2c). This is the
// property Adeline's one batch of Humans rides on.
func TestRepointedAttackerIsStillOneBatch(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := attackSetup(t, g)

	if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker a: %v", err)
	}
	if err := g.DeclareAttacker(b, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker b: %v", err)
	}
	if err := g.DeclareAttacker(a, g.Seats[2].ID); err != nil {
		t.Fatalf("re-point a: %v", err)
	}
	lockInAttackDeclaration(t, g)

	var batch uint64
	seen := 0
	for _, ev := range g.Events {
		if ev.Kind != EventAttack {
			continue
		}
		seen++
		if batch == 0 {
			batch = ev.Batch
			continue
		}
		if ev.Batch != batch {
			t.Fatalf("batch %d alongside %d: the declaration must be one batch", ev.Batch, batch)
		}
	}
	if seen != 2 {
		t.Fatalf("two attackers, one of them re-pointed: %d attack events, want 2", seen)
	}
}

// TestAttackDeclarationLocksInOnPriorityWrap — the production path
// through PassPriority. Priority passing all the way around is the
// declaration being complete (CR 508.1), and the engine announces it
// there rather than a step later.
func TestAttackDeclarationLocksInOnPriorityWrap(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, _ := attackSetup(t, g)

	if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(a, g.Seats[2].ID); err != nil {
		t.Fatalf("re-point: %v", err)
	}
	lockInAttacksByPriorityWrap(t, g)

	evs := attackDeclEvents(g, a)
	if len(evs) != 1 {
		t.Fatalf("the priority wrap locks the declaration in: %d attack events", len(evs))
	}
	if evs[0].Target != g.Seats[2].ID {
		t.Errorf("Target = %s, want %s", evs[0].Target, g.Seats[2].ID)
	}
}

// TestAttackDeclarationLocksInBeforeTheCursorLeavesTheStep — the
// other production path. advance_step out of declare_attackers
// announces the declaration BEFORE the cursor moves, so the events
// belong to the step that produced them and their triggers are
// harvested there, ahead of blockers.
func TestAttackDeclarationLocksInBeforeTheCursorLeavesTheStep(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, _ := attackSetup(t, g)

	if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(a, g.Seats[2].ID); err != nil {
		t.Fatalf("re-point: %v", err)
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	evs := attackDeclEvents(g, a)
	if len(evs) != 1 {
		t.Fatalf("leaving the step locks the declaration in: %d attack events", len(evs))
	}
	if evs[0].Target != g.Seats[2].ID {
		t.Errorf("Target = %s, want %s", evs[0].Target, g.Seats[2].ID)
	}
	// The next step's announcement is later in the log, which is what
	// "before the cursor moved" means.
	for _, ev := range g.Events {
		if ev.Kind == EventStepBegan && ev.Step == StepDeclareBlockers && ev.Seq < evs[0].Seq {
			t.Error("the declaration was announced after the cursor had already entered declare blockers")
		}
	}
}

// TestAttackAnnouncementsRewindWithUndo — both halves of an undo
// across a re-point. A clone taken while the declaration is staged
// comes back staged, so the lock-in announces the defender the undo
// restored; a clone taken after the lock-in comes back with the
// announcement, so the creature is not announced as attacking twice.
func TestAttackAnnouncementsRewindWithUndo(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, _ := attackSetup(t, g)

	if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	// Undo of a STAGED re-point: the attacker goes back to seat 1,
	// and the lock-in then announces seat 1.
	staged := g.Clone()
	if err := g.DeclareAttacker(a, g.Seats[2].ID); err != nil {
		t.Fatalf("re-point: %v", err)
	}
	g.WithWriteLock(func() { g.RestoreFrom(staged) })
	if findCard(g, a).AttackingTarget != g.Seats[1].ID {
		t.Fatal("the undo takes the re-point back")
	}
	lockInAttackDeclaration(t, g)
	evs := attackDeclEvents(g, a)
	if len(evs) != 1 || evs[0].Target != g.Seats[1].ID {
		t.Fatalf("after the undo: %d attack events naming %v, want 1 naming seat 1", len(evs), evs)
	}

	// Undo of a re-point made AFTER the lock-in. The announcement
	// rewinds with the log, so re-locking announces nothing new —
	// and does not re-announce the attacker either.
	announced := g.Clone()
	if err := g.DeclareAttacker(a, g.Seats[2].ID); err != nil {
		t.Fatalf("post-lock-in re-point: %v", err)
	}
	lockInAttackDeclaration(t, g)
	if n := len(attackDeclEvents(g, a)); n != 1 {
		t.Fatalf("a re-point after the lock-in is not a second attack: %d attack events", n)
	}
	g.WithWriteLock(func() { g.RestoreFrom(announced) })
	lockInAttackDeclaration(t, g)
	evs = attackDeclEvents(g, a)
	if len(evs) != 1 {
		t.Errorf("the undone declaration is not replayed: %d attack events", len(evs))
	}
	if evs[0].Target != g.Seats[1].ID {
		t.Errorf("Target = %s after the undo, want %s", evs[0].Target, g.Seats[1].ID)
	}
}

// TestClearCombatForgetsAttackAnnouncements — the announcements are
// one combat's bookkeeping. The same creature next combat is a new
// declaration and announces again.
func TestClearCombatForgetsAttackAnnouncements(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, _ := attackSetup(t, g)

	if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttackDeclaration(t, g)
	if err := g.ClearCombat(); err != nil {
		t.Fatalf("ClearCombat: %v", err)
	}
	if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker again: %v", err)
	}
	lockInAttackDeclaration(t, g)
	if n := len(attackDeclEvents(g, a)); n != 2 {
		t.Errorf("a fresh combat announces afresh: %d attack events, want 2", n)
	}
}

// TestPutOntoTheBattlefieldAttackingIsNeverAnnounced — CR 506.3c. A
// permanent put onto the battlefield attacking was never DECLARED as
// an attacker, so the lock-in must not mistake its AttackingTarget
// for a staged declaration and hand it an attack trigger.
func TestPutOntoTheBattlefieldAttackingIsNeverAnnounced(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, _ := attackSetup(t, g)
	if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}

	before := len(attackEvents(g))
	g.WithWriteLock(func() {
		if err := g.CreateTokensAttackingForEffect(g.Seats[0].ID, Card{
			Name: "Angel", TypeLine: "Token Creature — Angel", Power: 4, Toughness: 4,
		}, 2, g.Seats[1].ID); err != nil {
			t.Fatalf("CreateTokensAttackingForEffect: %v", err)
		}
	})
	lockInAttackDeclaration(t, g)
	if n := len(attackEvents(g)) - before; n != 1 {
		t.Errorf("the declared attacker announces and the two tokens do not: %d attack events, want 1", n)
	}
}

// TestBulkDeclareAttackersAnnouncesOncePerAttacker — the bulk verb
// goes through the same lock-in and its contract is unchanged: one
// event per declared creature, all in one batch, drained together.
func TestBulkDeclareAttackersAnnouncesOncePerAttacker(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := attackSetup(t, g)

	declared, err := g.DeclareAttackers(declsAt(g.Seats[1].ID, a, b))
	if err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	if len(declared) != 2 {
		t.Fatalf("declared %d, want 2", len(declared))
	}
	var batch uint64
	seen := 0
	for _, ev := range g.Events {
		if ev.Kind != EventAttack {
			continue
		}
		seen++
		if batch == 0 {
			batch = ev.Batch
		} else if ev.Batch != batch {
			t.Fatalf("batch %d alongside %d: the bulk declaration must be one batch", ev.Batch, batch)
		}
		if ev.Target != g.Seats[1].ID {
			t.Errorf("Target = %s, want %s", ev.Target, g.Seats[1].ID)
		}
	}
	if seen != 2 {
		t.Fatalf("%d attack events for a two-creature bulk declaration, want 2", seen)
	}
	// And the bulk verb announces WITHOUT a further boundary, because
	// it runs the state checks itself.
	if g.Turn.Step != StepDeclareAttackers {
		t.Errorf("the bulk declaration moved the cursor to %s", g.Turn.Step)
	}
}
