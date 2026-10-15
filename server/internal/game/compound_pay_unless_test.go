package game

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func queueCompoundWard(t *testing.T, g *Game, payer *Player, cost string, life int) (uuid.UUID, *PendingChoice) {
	t.Helper()
	spell := guardedSpellOnStack(t, g, payer)
	g.WithWriteLock(func() {
		if err := g.QueueCounterUnlessPaidForEffect(CounterUnlessPaidPrompt{
			StackItem: spell, Chooser: payer.ID, Cost: cost, LifeCost: life,
			Question: "Ward — pay mana and life",
		}); err != nil {
			t.Fatal(err)
		}
	})
	return spell, onlyPendingChoice(t, g)
}

func TestCompoundWardPaymentIsAllOrNothing(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		life, mana, lands     int
		pay, locked, wantPaid bool
	}{
		{"floating mana", 40, 2, 0, true, false, true},
		{"auto tap", 40, 0, 2, true, false, true},
		{"decline", 40, 2, 2, false, false, false},
		{"missing mana", 40, 0, 1, true, false, false},
		{"insufficient life", 1, 0, 2, true, false, false},
		{"locked life", 40, 2, 2, true, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newFourPlayerActiveGame(t)
			payer := g.Seats[1]
			payer.Life = tc.life
			for range tc.mana {
				payer.ManaPool.AddMana(ManaToken{Color: "C"})
			}
			var lands []uuid.UUID
			for range tc.lands {
				lands = append(lands, pushBattlefieldForTest(g, payer.ID, "Swamp", "Basic Land — Swamp", ""))
			}
			if tc.locked {
				lockUntilNextTurn(g, payer)
			}
			spell, choice := queueCompoundWard(t, g, payer, "{2}", 2)
			if choice.LifeCost != 2 || choice.PayCost != "{2}" {
				t.Fatalf("incomplete offer: %+v", choice)
			}
			if err := g.ResolvePayUnless(choice.ID, payer.ID, tc.pay); err != nil {
				t.Fatal(err)
			}
			if tc.wantPaid {
				if payer.Life != tc.life-2 || payer.ManaPool.Total() != 0 || g.StackMeta[spell] == nil {
					t.Fatal("complete payment did not preserve the spell and pay both components")
				}
				for _, id := range lands {
					if !tappedForTest(g, id) {
						t.Fatal("funding land was not tapped")
					}
				}
			} else {
				if payer.Life != tc.life || payer.ManaPool.Total() != tc.mana || g.StackMeta[spell] != nil {
					t.Fatal("failed/declined payment charged a component or did not counter")
				}
				for _, id := range lands {
					if tappedForTest(g, id) {
						t.Fatal("failed payment tapped a land")
					}
				}
			}
		})
	}
}

func TestCompoundPaymentAllowsExactlyEnoughLife(t *testing.T) {
	g := newActiveGame(t)
	payer := g.Seats[0]
	payer.Life = 2
	payer.ManaPool.AddMana(ManaToken{Color: "C"})
	payer.ManaPool.AddMana(ManaToken{Color: "C"})
	cost, err := ParseCost("{2}")
	if err != nil {
		t.Fatal(err)
	}
	var paid bool
	g.WithWriteLock(func() { paid = g.payCostWithLifeLocked(payer, cost, uuid.Nil, nil, 0, 2) })
	if !paid || payer.Life != 0 || payer.ManaPool.Total() != 0 {
		t.Fatal("paying exactly the life total is legal before state checks")
	}
}

func TestCompoundPaymentCombinesFixedAndPhyrexianLife(t *testing.T) {
	for _, life := range []int{40, 3} {
		g := newFourPlayerActiveGame(t)
		payer := g.Seats[1]
		payer.Life = life
		spell, choice := queueCompoundWard(t, g, payer, "{B/P}", 2)
		err := g.ResolvePayUnlessWithLife(choice.ID, payer.ID, true, nil, 1)
		if life == 3 {
			if !errors.Is(err, ErrInvalidParam) || payer.Life != 3 || len(g.PendingChoices) != 1 || g.StackMeta[spell] == nil {
				t.Fatal("unpayable combined life claim changed state or lost its prompt")
			}
		} else if err != nil || payer.Life != life-4 || g.StackMeta[spell] == nil {
			t.Fatalf("fixed and symbol life not paid together: %v, life %d", err, payer.Life)
		}
	}
}

func TestCompoundPaymentDoesNotChargeForDepartedSpell(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	payer := g.Seats[1]
	spell, choice := queueCompoundWard(t, g, payer, "{2}", 2)
	payer.ManaPool.AddMana(ManaToken{Color: "C"})
	payer.ManaPool.AddMana(ManaToken{Color: "C"})
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(spell); err != nil {
			t.Fatal(err)
		}
	})
	if err := g.ResolvePayUnless(choice.ID, payer.ID, true); err != nil && !errors.Is(err, ErrPendingChoiceNotFound) {
		t.Fatal(err)
	}
	if payer.Life != 40 || payer.ManaPool.Total() != 2 {
		t.Fatal("a departed guarded spell charged mana or life")
	}
}

func TestCompoundLifeCostSurvivesCloneAndSnapshotCapture(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	payer := g.Seats[1]
	_, choice := queueCompoundWard(t, g, payer, "{2}", 2)
	clone := g.Clone()
	choice.LifeCost = 7
	if clone.PendingChoices[0].LifeCost != 2 {
		t.Fatal("clone shares the choice life cost")
	}
	g.RestoreFrom(clone)
	if g.PendingChoices[0].LifeCost != 2 {
		t.Fatal("undo lost the fixed life cost")
	}
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var capture GameSnapshot
	if err = json.Unmarshal(raw, &capture); err != nil {
		t.Fatal(err)
	}
	if len(capture.PendingChoices) != 1 || capture.PendingChoices[0].LifeCost != 2 {
		t.Fatal("capture lost the existing serialized life cost")
	}
}

func TestCompoundCanceledLifePaymentChargesNoMana(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	payer := g.Seats[1]
	const oracle = "test-2945-cancel-life-payment"
	old := CatalogReplacements
	CatalogReplacements = func(id string) []ReplacementEffect {
		if id == oracle {
			return []ReplacementEffect{{
				Watches:   []EventKind{EventChangeLife},
				AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool { return ev.Kind == RepEventLife && ev.LifeDelta < 0 },
				Replace:   func(ev *ReplacementEvent, _ *Game, _ *Card) error { ev.Canceled = true; return nil },
			}}
		}
		if old != nil {
			return old(id)
		}
		return nil
	}
	t.Cleanup(func() { CatalogReplacements = old })
	pushBattlefieldForTest(g, payer.ID, "Cancel life cost", "Artifact", oracle)
	first := pushBattlefieldForTest(g, payer.ID, "Swamp", "Basic Land — Swamp", "")
	second := pushBattlefieldForTest(g, payer.ID, "Swamp", "Basic Land — Swamp", "")
	spell, choice := queueCompoundWard(t, g, payer, "{2}", 2)
	if err := g.ResolvePayUnless(choice.ID, payer.ID, true); err != nil {
		t.Fatal(err)
	}
	if payer.Life != 40 || payer.ManaPool.Total() != 0 || tappedForTest(g, first) || tappedForTest(g, second) || g.StackMeta[spell] != nil {
		t.Fatal("canceled life cost charged a partial mana payment")
	}
}
