package game

import (
	"testing"

	"github.com/google/uuid"
)

// counter_unless_paid_onpaid_test.go — #3591. OnPaid is the paid half
// of the same answer OnDecline is the unpaid half of: Assimilate
// Essence's "If they do, you incubate 2".

// A payment that is made runs OnPaid and not the decline; a decline, or
// a "pay" that cannot be funded, runs only the decline.
func TestACounterUnlessPaidRunsOnPaidOnlyForAPaymentThatWasMade(t *testing.T) {
	for _, tc := range []struct {
		name         string
		answerPay    bool
		fund         bool
		wantPaid     int
		wantDeclined int
	}{
		{"paid", true, true, 1, 0},
		{"declined", false, true, 0, 1},
		{"pay it cannot fund", true, false, 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newFourPlayerActiveGame(t)
			caster := g.Seats[1]
			spell := guardedSpellOnStack(t, g, caster)
			if tc.fund {
				caster.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"})
			}
			paid, declined := 0, 0
			g.WithWriteLock(func() {
				if err := g.QueueCounterUnlessPaidForEffect(CounterUnlessPaidPrompt{
					StackItem: spell, Source: uuid.New(), Cost: "{2}", Question: "Assimilate",
					Then: testPayUnlessHook, Carry: payUnlessHookCarry(
						func(*Game) error { paid++; return nil },
						func(*Game) error { declined++; return nil }),
				}); err != nil {
					t.Fatalf("queue: %v", err)
				}
			})
			if err := g.ResolvePayUnless(onlyPendingChoice(t, g).ID, caster.ID, tc.answerPay); err != nil {
				t.Fatalf("ResolvePayUnless: %v", err)
			}
			if paid != tc.wantPaid || declined != tc.wantDeclined {
				t.Errorf("paid ran %d (want %d), decline ran %d (want %d)", paid, tc.wantPaid, declined, tc.wantDeclined)
			}
		})
	}
}

// OnPaid runs under the decline's stack guard: an object that has left
// the stack has nothing for the payment to buy off.
func TestACounterUnlessPaidOnPaidSkipsADepartedObject(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster := g.Seats[1]
	spell := guardedSpellOnStack(t, g, caster)
	caster.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "C"})
	ran := false
	g.WithWriteLock(func() {
		_ = g.QueueCounterUnlessPaidForEffect(CounterUnlessPaidPrompt{
			StackItem: spell, Source: uuid.New(), Cost: "{2}", Question: "Assimilate",
			Then: testPayUnlessHook, Carry: payUnlessHookCarry(func(*Game) error { ran = true; return nil }, nil),
		})
		_ = g.CounterTargetForEffect(spell)
	})
	if len(g.PendingChoices) == 1 {
		if err := g.ResolvePayUnless(g.PendingChoices[0].ID, caster.ID, true); err != nil {
			t.Fatalf("ResolvePayUnless: %v", err)
		}
	}
	if ran {
		t.Error("OnPaid ran for a spell that was no longer on the stack")
	}
}
