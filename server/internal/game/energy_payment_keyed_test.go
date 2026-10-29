package game

import (
	"testing"

	"github.com/google/uuid"
)

// energy_payment_keyed_test.go — #4289: a fixed energy payment whose
// consequence is a registered key (EnergyPayment.Then) is the plain-data
// pay-unless prompt, so a table waiting on it is a restore point.

var (
	energyKeyedPaid, energyKeyedDeclined int
	energyKeyedBranches                  = RegisterPayUnlessBranches(testEffectKeyPrefix+"energy-keyed",
		func(*Game, PayUnlessAnswer) error { energyKeyedPaid++; return nil },
		func(*Game, PayUnlessAnswer) error { energyKeyedDeclined++; return nil })
)

// A keyed energy payment is the same prompt with its consequence a
// registered key: no closure frame, the payment as plain data, and the
// table waiting on it is a restore point. "Yes" from a short payer is the
// decline (CR 118.3).
func TestKeyedEnergyPaymentIsAPlainDataPrompt(t *testing.T) {
	for _, tc := range []struct {
		name              string
		energy            int
		pay               bool
		wantPaid, wantNot int
		wantEnergy        int
	}{
		{"paid", 3, true, 1, 0, 1},
		{"declined", 3, false, 0, 1, 3},
		{"short payer's yes is the decline", 1, true, 0, 1, 1},
	} {
		energyKeyedPaid, energyKeyedDeclined = 0, 0
		g := newActiveGame(t)
		me := g.Seats[0]
		setEnergyForTest(t, g, me, tc.energy)
		id := queueEnergyPayment(t, g, EnergyPayment{
			Chooser: me.ID, Source: uuid.New(), N: 2, Question: "pay {E}{E}?",
			Then: energyKeyedBranches, InThisStep: true,
		})
		c := g.PendingChoices[0]
		if c.PickThen != energyKeyedBranches.Key() {
			t.Fatalf("%s: the prompt has the wrong key %q", tc.name, c.PickThen)
		}
		if a := c.PayAction(); a == nil || a.Kind != PayActionEnergy || a.Count != 2 {
			t.Fatalf("%s: PayAction = %+v, want energy x2", tc.name, a)
		}
		if c.OwedInStep.Turn == 0 {
			t.Errorf("%s: the prompt does not hold the step it was asked in", tc.name)
		}
		if !g.CaptureSnapshot().Restorable() {
			t.Errorf("%s: a table waiting on a keyed energy payment is not a restore point", tc.name)
		}
		if err := g.ResolvePayUnless(id, me.ID, tc.pay); err != nil {
			t.Fatalf("%s: resolve: %v", tc.name, err)
		}
		if energyKeyedPaid != tc.wantPaid || energyKeyedDeclined != tc.wantNot {
			t.Errorf("%s: paid %d declined %d, want %d and %d", tc.name, energyKeyedPaid, energyKeyedDeclined, tc.wantPaid, tc.wantNot)
		}
		if PlayerEnergy(me) != tc.wantEnergy {
			t.Errorf("%s: energy = %d, want %d", tc.name, PlayerEnergy(me), tc.wantEnergy)
		}
	}
}
