package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// pay_unless_nothing_test.go — #4361 S3: no pay-unless holds a frame.
// One with nothing hanging off its answer (the PayUnless primitive with
// no Else, an upkeep payment or an energy payment queued with no
// branches) used to store a payUnlessFrame whose callbacks were both
// nil, and a prompt with a frame was not a restore point. It is keyed
// by the engine's payUnlessNothing now, so a table waiting on it saves
// and restores strictly, and the restored prompt is answered from its
// data.

// TestAPayUnlessWithNothingOnItIsARestorePoint saves and strictly
// restores a table waiting on each frame-free shape, then answers the
// restored prompt.
func TestAPayUnlessWithNothingOnItIsARestorePoint(t *testing.T) {
	for _, tc := range []struct {
		name  string
		queue func(g *Game, payer uuid.UUID) error
		// pay funds the payer on the restored table for a "yes".
		pay func(t *testing.T, g *Game, payer *Player)
		// paid reports whether the payment came off the payer.
		paid func(g *Game, payer *Player) bool
	}{
		{
			name: "plain pay-unless",
			queue: func(g *Game, payer uuid.UUID) error {
				return g.QueuePayUnlessForEffect(payer, uuid.New(), "{1}", "Pay {1}?")
			},
			pay:  func(_ *testing.T, _ *Game, p *Player) { p.ManaPool.AddMana(ManaToken{Color: "C"}) },
			paid: func(_ *Game, p *Player) bool { return p.ManaPool.Total() == 0 },
		},
		{
			name: "upkeep pay-unless",
			queue: func(g *Game, payer uuid.UUID) error {
				return g.QueueUpkeepPayUnlessForEffect(UpkeepPayUnlessPrompt{Chooser: payer, Source: uuid.New(), Cost: "{1}", Question: "Upkeep {1}?"})
			},
			pay:  func(_ *testing.T, _ *Game, p *Player) { p.ManaPool.AddMana(ManaToken{Color: "C"}) },
			paid: func(_ *Game, p *Player) bool { return p.ManaPool.Total() == 0 },
		},
		{
			name: "energy payment",
			queue: func(g *Game, payer uuid.UUID) error {
				return g.QueueEnergyPaymentForEffect(EnergyPayment{Chooser: payer, Source: uuid.New(), N: 2, Question: "Pay {E}{E}?"})
			},
			pay:  func(t *testing.T, g *Game, p *Player) { setEnergyForTest(t, g, p, 3) },
			paid: func(_ *Game, p *Player) bool { return PlayerEnergy(p) == 1 },
		},
	} {
		for _, answerPay := range []bool{false, true} {
			g := newActiveGame(t)
			payer := g.Seats[1].ID
			var err error
			g.WithWriteLock(func() { err = tc.queue(g, payer) })
			if err != nil {
				t.Fatalf("%s: queue: %v", tc.name, err)
			}
			c := onlyPendingChoice(t, g)
			if c.PickThen != payUnlessNothing.Key() {
				t.Fatalf("%s: PickThen = %q, want %q", tc.name, c.PickThen, payUnlessNothing.Key())
			}
			owed := c.OwedInStep

			restored := restoreStrictForTest(t, g)
			rc := onlyPendingChoice(t, restored)
			if rc.Kind != PendingChoicePayUnless || rc.PickThen != payUnlessNothing.Key() || rc.OwedInStep != owed {
				t.Fatalf("%s: restored prompt = %+v", tc.name, rc)
			}
			p := restored.playerByIDLocked(payer)
			if answerPay {
				tc.pay(t, restored, p)
			}
			if err := restored.ResolvePayUnless(rc.ID, payer, answerPay); err != nil {
				t.Fatalf("%s pay=%v: ResolvePayUnless on the restored table: %v", tc.name, answerPay, err)
			}
			if len(restored.PendingChoices) != 0 {
				t.Errorf("%s pay=%v: %d prompts left after the answer", tc.name, answerPay, len(restored.PendingChoices))
			}
			if answerPay && !tc.paid(restored, p) {
				t.Errorf("%s: the restored prompt's payment did not come off the payer", tc.name)
			}
		}
	}
}

// A waterbend counter-unless-pays carries its tap clause as data and is
// a restore point; after a strict restore the clause is rebuilt by its
// registered builder and a decline still counters the spell.
func TestAWaterbendCounterUnlessPaidIsARestorePoint(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster := g.Seats[1]
	spell := guardedSpellOnStack(t, g, caster)
	queueWaterbendWardTax(t, g, spell, caster.ID, "{4}")
	c := onlyPendingChoice(t, g)
	if c.PayTap == nil || c.PayTap.Kind != WaterbendKey || c.PickThen != counterGuardedDecline.Key() {
		t.Fatalf("prompt = %+v, want a keyed waterbend payment", c)
	}

	restored := restoreStrictForTest(t, g)
	rc := onlyPendingChoice(t, restored)
	if tap := rc.PayTapCost(); tap == nil || tap.Key != WaterbendKey || tap.Extra != "{4}" {
		t.Fatalf("restored tap clause = %+v, want waterbend {4}", tap)
	}
	if err := restored.ResolvePayUnless(rc.ID, caster.ID, false); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if restored.StackItemForEffect(spell) != nil {
		t.Error("a declined waterbend ward did not counter the spell after the restore")
	}
}

// With no frame to hold it, a tap clause whose kind this binary cannot
// rebuild is refused rather than queued as a closure.
func TestACounterUnlessPaidWithAnUnregisteredTapClauseIsRefused(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster := g.Seats[1]
	spell := guardedSpellOnStack(t, g, caster)
	var err error
	g.WithWriteLock(func() {
		err = g.QueueCounterUnlessPaidForEffect(CounterUnlessPaidPrompt{
			StackItem: spell, Source: uuid.New(), Cost: "{2}", Question: "Ward",
			Waterbend: &TapPermanentsCost{Key: "test/4361-unregistered-tap", Extra: "{2}", Spec: &TargetSpec{}},
		})
	})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("err = %v, want ErrInvalidParam", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d prompts queued, want none", len(g.PendingChoices))
	}
}
