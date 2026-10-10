package game

import (
	"testing"

	"github.com/google/uuid"
)

// pay_unless_test.go covers the S19 sub-PR 6 "unless that player
// pays {N}" prompt: queueing, paying from the pool, paying via
// auto-tap, declining, and the "said yes but can't pay" degrade.

// queuePayUnless queues a {2} prompt for chooser whose decline
// consequence bumps *declined. Returns the choice ID.
func queuePayUnless(t *testing.T, g *Game, chooser uuid.UUID, cost string, declined *int) uuid.UUID {
	t.Helper()
	source := uuid.New()
	g.WithWriteLock(func() {
		if err := g.QueuePayUnlessForEffect(chooser, source, cost, "Pay "+cost+"?", func(_ *Game) error {
			*declined++
			return nil
		}); err != nil {
			t.Fatalf("QueuePayUnlessForEffect: %v", err)
		}
	})
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoicePayUnless && c.Chooser == chooser {
			if c.PayCost != cost {
				t.Errorf("PayCost = %q, want %q", c.PayCost, cost)
			}
			return c.ID
		}
	}
	t.Fatalf("no pay_unless prompt queued for %s", chooser)
	return uuid.Nil
}

func TestPayUnlessPayFromPool(t *testing.T) {
	g := newActiveGame(t)
	payer := g.Seats[1]
	payer.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "G"}, ManaToken{Color: "R"})
	var declined int
	id := queuePayUnless(t, g, payer.ID, "{2}", &declined)

	if err := g.ResolvePayUnless(id, payer.ID, true); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if declined != 0 {
		t.Errorf("decline consequence ran after a successful payment")
	}
	if got := len(payer.ManaPool); got != 1 {
		t.Errorf("pool after paying {2} from 3: %d tokens, want 1", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("prompt not dequeued")
	}
	var spent bool
	for _, ev := range g.Events {
		if ev.Kind == EventManaSpent && ev.Actor == payer.ID {
			spent = true
		}
	}
	if !spent {
		t.Errorf("no EventManaSpent breadcrumb")
	}
}

func TestPayUnlessPayViaAutoTap(t *testing.T) {
	g := newActiveGame(t)
	payer := g.Seats[1]
	m1 := pushBattlefieldForTest(g, payer.ID, "Mountain", "Basic Land — Mountain", "")
	m2 := pushBattlefieldForTest(g, payer.ID, "Mountain", "Basic Land — Mountain", "")
	var declined int
	id := queuePayUnless(t, g, payer.ID, "{2}", &declined)

	if err := g.ResolvePayUnless(id, payer.ID, true); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if declined != 0 {
		t.Errorf("decline consequence ran even though two Mountains could pay")
	}
	tapped := 0
	for _, c := range g.Battlefield.Cards {
		if (c.InstanceID == m1 || c.InstanceID == m2) && c.Tapped {
			tapped++
		}
	}
	if tapped != 2 {
		t.Errorf("auto-tap tapped %d Mountains, want 2", tapped)
	}
	if len(payer.ManaPool) != 0 {
		t.Errorf("pool should be empty after auto-tap + spend, has %d", len(payer.ManaPool))
	}
}

func TestPayUnlessDeclineRunsConsequence(t *testing.T) {
	g := newActiveGame(t)
	payer := g.Seats[1]
	payer.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"})
	var declined int
	id := queuePayUnless(t, g, payer.ID, "{2}", &declined)

	if err := g.ResolvePayUnless(id, payer.ID, false); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if declined != 1 {
		t.Errorf("decline consequence ran %d times, want 1", declined)
	}
	if len(payer.ManaPool) != 2 {
		t.Errorf("declining must not touch the pool: %d tokens, want 2", len(payer.ManaPool))
	}
}

func TestPayUnlessYesButCannotPayDegradesToDecline(t *testing.T) {
	g := newActiveGame(t)
	payer := g.Seats[1]
	payer.ManaPool.AddMana(ManaToken{Color: "C"})
	var declined int
	id := queuePayUnless(t, g, payer.ID, "{2}", &declined)

	if err := g.ResolvePayUnless(id, payer.ID, true); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if declined != 1 {
		t.Errorf("consequence ran %d times, want 1 (yes with an empty wallet is a decline)", declined)
	}
	if len(payer.ManaPool) != 1 {
		t.Errorf("failed payment must leave the pool intact: %d tokens, want 1", len(payer.ManaPool))
	}
}

func TestPayUnlessWrongChooserRejected(t *testing.T) {
	g := newActiveGame(t)
	payer := g.Seats[1]
	other := g.Seats[0]
	var declined int
	id := queuePayUnless(t, g, payer.ID, "{1}", &declined)

	if err := g.ResolvePayUnless(id, other.ID, false); err != ErrNotTheChooser {
		t.Fatalf("ResolvePayUnless by non-chooser: got %v, want ErrNotTheChooser", err)
	}
	if declined != 0 || len(g.PendingChoices) != 1 {
		t.Errorf("non-chooser answer must not resolve the prompt")
	}
	if kind, ok := g.PendingChoiceKindFor(id); !ok || kind != PendingChoicePayUnless {
		t.Errorf("PendingChoiceKindFor = %q,%v", kind, ok)
	}
}

func TestPayUnlessEliminatedChooserRunsConsequenceImmediately(t *testing.T) {
	g := newActiveGame(t)
	payer := g.Seats[1]
	payer.Eliminated = true
	var declined int
	g.WithWriteLock(func() {
		if err := g.QueuePayUnlessForEffect(payer.ID, uuid.New(), "{1}", "Pay?", func(_ *Game) error {
			declined++
			return nil
		}); err != nil {
			t.Fatalf("QueuePayUnlessForEffect: %v", err)
		}
	})
	if declined != 1 {
		t.Errorf("consequence should run at queue time for an eliminated payer, ran %d", declined)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("no prompt should queue for an eliminated payer")
	}
}

// TestSpellsCastThisTurnTallyAndReset: CastSpell bumps the caster's
// per-turn tally (noncreature counted separately) and a new turn
// clears it.
func TestSpellsCastThisTurnTallyAndReset(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	for g.Turn.Step != StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	bolt := uuid.New()
	p.Hand.PushTop(Card{InstanceID: bolt, Name: "Bolt", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	bear := uuid.New()
	p.Hand.PushTop(Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Owner: p.ID, Controller: p.ID})
	// Creature first (sorcery speed wants an empty stack), then the
	// instant on top of it.
	if err := g.CastSpell(p.ID, bear, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell bear: %v", err)
	}
	if err := g.CastSpell(p.ID, bolt, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell bolt: %v", err)
	}
	got := g.CastTallyFor(p.ID)
	if got.Total != 2 || got.Noncreature != 1 {
		t.Errorf("tally after bolt + bear: %+v, want {Total:2 Noncreature:1}", got)
	}
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if got := g.CastTallyFor(p.ID); got != (CastTally{}) {
		t.Errorf("tally not reset on a new turn: %+v", got)
	}
}
