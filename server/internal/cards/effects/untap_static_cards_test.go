package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// untap_static_cards_test.go exercises the printed cards which use the
// untap-step restriction seam.  untap_wave_test.go covers registration and
// one-shot effects; these tests keep the card-specific clauses observable.

func untapStaticPendingPay(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		choice := g.PendingChoices[i]
		if choice != nil && choice.Kind == game.PendingChoicePayUnless && choice.Chooser == chooser {
			return choice
		}
	}
	t.Fatalf("no optional-payment prompt for %s", chooser)
	return nil
}

func untapStaticTapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("battlefield card %s is missing", id)
	}
	return c.Tapped
}

func TestManaVaultUpkeepOfferAndDrawInterveningCondition(t *testing.T) {
	t.Run("pay and decline", func(t *testing.T) {
		for _, tc := range []struct {
			name string
			pay  bool
		}{
			{name: "pay", pay: true},
			{name: "decline", pay: false},
		} {
			t.Run(tc.name, func(t *testing.T) {
				g := newCatalogGame(t)
				owner := g.Seats[0]
				vault := pushPermanentForTest(g, owner.ID, "Mana Vault", "736892cb-a34b-4bb9-b56c-e26e3db207a2", "Artifact")
				if err := g.TapCard(vault, true); err != nil {
					t.Fatal(err)
				}

				advanceToUpkeepOf(t, g, 0)
				if triggerOnStack(g, vault) == nil {
					t.Fatal("Mana Vault upkeep trigger was not put on the stack")
				}
				passPriorityAroundTable(t, g)
				choice := untapStaticPendingPay(t, g, owner.ID)
				if tc.pay {
					owner.ManaPool.AddMana(
						game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"},
						game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"},
					)
				}
				if err := g.ResolvePayUnless(choice.ID, owner.ID, tc.pay); err != nil {
					t.Fatalf("answer Mana Vault upkeep offer: %v", err)
				}
				if got := untapStaticTapped(t, g, vault); got == tc.pay {
					t.Errorf("Mana Vault tapped after pay=%v: got %v, want %v", tc.pay, got, !tc.pay)
				}
			})
		}
	})

	t.Run("draw trigger rechecks whether it remains tapped", func(t *testing.T) {
		g := newCatalogGame(t)
		owner := g.Seats[0]
		vault := pushPermanentForTest(g, owner.ID, "Mana Vault", "736892cb-a34b-4bb9-b56c-e26e3db207a2", "Artifact")
		if err := g.TapCard(vault, true); err != nil {
			t.Fatal(err)
		}

		advanceToUpkeepOf(t, g, 0)
		passPriorityAroundTable(t, g)
		choice := untapStaticPendingPay(t, g, owner.ID)
		if err := g.ResolvePayUnless(choice.ID, owner.ID, false); err != nil {
			t.Fatalf("decline Mana Vault upkeep offer: %v", err)
		}

		lifeBefore := owner.Life
		if _, err := g.AdvanceStep(); err != nil || g.Turn.Step != game.StepDraw {
			t.Fatalf("advance to draw: turn=%+v err=%v", g.Turn, err)
		}
		if triggerOnStack(g, vault) == nil {
			t.Fatal("tapped Mana Vault did not trigger at draw step")
		}
		g.WithWriteLock(func() {
			if err := g.UntapTargetForEffect(vault); err != nil {
				t.Errorf("untap Mana Vault before draw trigger resolves: %v", err)
			}
		})
		passPriorityAroundTable(t, g)
		if owner.Life != lifeBefore {
			t.Errorf("Mana Vault dealt damage after becoming untapped before resolution: life %d -> %d", lifeBefore, owner.Life)
		}
	})

	t.Run("draw trigger deals damage while it remains tapped", func(t *testing.T) {
		g := newCatalogGame(t)
		owner := g.Seats[0]
		vault := pushPermanentForTest(g, owner.ID, "Mana Vault", "736892cb-a34b-4bb9-b56c-e26e3db207a2", "Artifact")
		if err := g.TapCard(vault, true); err != nil {
			t.Fatal(err)
		}

		advanceToUpkeepOf(t, g, 0)
		passPriorityAroundTable(t, g)
		choice := untapStaticPendingPay(t, g, owner.ID)
		if err := g.ResolvePayUnless(choice.ID, owner.ID, false); err != nil {
			t.Fatalf("decline Mana Vault upkeep offer: %v", err)
		}

		lifeBefore := owner.Life
		if _, err := g.AdvanceStep(); err != nil || g.Turn.Step != game.StepDraw {
			t.Fatalf("advance to draw: turn=%+v err=%v", g.Turn, err)
		}
		passPriorityAroundTable(t, g)
		if owner.Life != lifeBefore-1 {
			t.Errorf("Mana Vault draw trigger: life %d -> %d, want %d", lifeBefore, owner.Life, lifeBefore-1)
		}
	})
}

func TestBasaltAndGrimMonolithActivatedUntaps(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		cost         int
	}{
		{name: "Basalt Monolith", oracle: "6b8cf2a0-b045-4d91-9d91-c602d40c6237", cost: 3},
		{name: "Grim Monolith", oracle: "229d6627-1292-4ae1-8849-b0f956fa6540", cost: 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			owner := g.Seats[0]
			id := pushPermanentForTest(g, owner.ID, tc.name, tc.oracle, "Artifact")
			if err := g.TapCard(id, true); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < tc.cost; i++ {
				owner.ManaPool.AddMana(game.ManaToken{Color: "C"})
			}
			if err := g.ActivateCatalogAbility(owner.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("activate %s untap ability: %v", tc.name, err)
			}
			passPriorityAroundTable(t, g)
			if untapStaticTapped(t, g, id) {
				t.Errorf("%s remained tapped after its activated untap ability resolved", tc.name)
			}
			if got := len(owner.ManaPool); got != 0 {
				t.Errorf("%s left %d mana after paying its activation", tc.name, got)
			}
		})
	}
}

func TestMeekstoneUsesEffectivePowerAndBackToBasicsFiltersBasics(t *testing.T) {
	t.Run("Meekstone sees a plus-one counter", func(t *testing.T) {
		g := newCatalogGame(t)
		owner := g.Seats[0]
		pushPermanentForTest(g, g.Seats[1].ID, "Meekstone", "5ba73182-30a7-4bad-9cb6-c0feecc2db33", "Artifact")
		high := uuid.New()
		low := uuid.New()
		g.Battlefield.PushTop(game.Card{InstanceID: high, Name: "Boosted Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
			Counters: map[string]int{game.CounterPlusOne: 1}, Owner: owner.ID, Controller: owner.ID, Tapped: true})
		g.Battlefield.PushTop(game.Card{InstanceID: low, Name: "Small Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
			Owner: owner.ID, Controller: owner.ID, Tapped: true})

		advanceToUpkeepOf(t, g, 0)
		if !untapStaticTapped(t, g, high) {
			t.Error("Meekstone allowed a 2/2 with a +1/+1 counter to untap")
		}
		if untapStaticTapped(t, g, low) {
			t.Error("Meekstone kept an effective-power-2 creature tapped")
		}
	})

	t.Run("Back to Basics leaves basic lands alone", func(t *testing.T) {
		g := newCatalogGame(t)
		owner := g.Seats[0]
		pushPermanentForTest(g, g.Seats[1].ID, "Back to Basics", "05c2dec2-d2f7-4036-b91f-4fccba10a8bb", "Enchantment")
		basic := pushPermanentForTest(g, owner.ID, "Island", "", "Basic Land — Island")
		nonbasic := pushPermanentForTest(g, owner.ID, "Steam Vents", "", "Land — Island Mountain")
		if err := g.TapCard(basic, true); err != nil {
			t.Fatal(err)
		}
		if err := g.TapCard(nonbasic, true); err != nil {
			t.Fatal(err)
		}

		advanceToUpkeepOf(t, g, 0)
		if untapStaticTapped(t, g, basic) {
			t.Error("Back to Basics kept a basic land tapped")
		}
		if !untapStaticTapped(t, g, nonbasic) {
			t.Error("Back to Basics allowed a nonbasic land to untap")
		}
	})
}

func TestIntruderAlarmUntapsForCreatureETBOnly(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		wantUntapped   bool
	}{
		{name: "creature", typeLine: "Creature — Bear", wantUntapped: true},
		{name: "noncreature", typeLine: "Artifact", wantUntapped: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			owner := g.Seats[0]
			pushPermanentForTest(g, g.Seats[1].ID, "Intruder Alarm", "1e943e04-e213-4781-b1a7-935aad8790e1", "Enchantment")
			creature := pushPermanentForTest(g, owner.ID, "Tapped Bear", "", "Creature — Bear")
			if err := g.TapCard(creature, true); err != nil {
				t.Fatal(err)
			}

			castCatalogSpell(t, g, "Entrant", tc.typeLine, "", nil)
			passPriorityAroundTable(t, g)
			if got := untapStaticTapped(t, g, creature); got == tc.wantUntapped {
				t.Errorf("tapped creature after %s ETB: tapped=%v, want %v", tc.name, got, !tc.wantUntapped)
			}
		})
	}
}
