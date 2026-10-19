package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_dominion_bracelet_test.go — #4000: a granted ability whose cost
// exiles the Equipment that grants it (AbilityCost.ExileGrantor), with
// an own cost clause that reads the equipped creature's power.

const dominionBraceletOracle = "c017f8ae-9103-4965-a07a-f259f44a157d"

// braceletTable is a catalog game in the active seat's main phase with
// a creature of power `power` (before the Bracelet's +1/+1) equipped
// with The Dominion Bracelet. It returns the game, the active player,
// the creature, the Bracelet and the granted row's index and ref.
func braceletTable(t *testing.T, power int) (*game.Game, *game.Player, uuid.UUID, uuid.UUID, int, string) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", power, 3)
	bracelet := gaPut(g, me.ID, "The Dominion Bracelet", "Legendary Artifact — Equipment", dominionBraceletOracle)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	equipTo(t, g, me.ID, bracelet, bear)
	idx, ref := grantedActivatedIndex(t, g, bear)
	if idx < 0 {
		t.Fatal("the equipped creature has no granted ability")
	}
	return g, me, bear, bracelet, idx, ref
}

// The whole card: the equipped creature's ability costs {15} less its
// power, exiles the Bracelet as a cost, and controls the opponent's
// next turn with no extra turn after it.
func TestDominionBraceletExilesItselfAndControlsTheirNextTurn(t *testing.T) {
	g, me, bear, bracelet, idx, ref := braceletTable(t, 3)
	if c := gaLayered(t, g, bear); c.PowerForComparison() != 4 {
		t.Fatalf("equipped power = %d, want 4 (+1/+1)", c.PowerForComparison())
	}
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	price, err := g.PriceActivation(me.ID, bear, idx, nil)
	if err != nil {
		t.Fatalf("PriceActivation: %v", err)
	}
	if price.Total.String() != "{11}" {
		t.Fatalf("price = %s, want {11} ({15} less power 4)", price.Total.String())
	}
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, bear, idx, game.ActivateAbilityParams{
		Ref: ref, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// The exile is a cost: the Bracelet is gone before the ability
	// resolves, and the creature is still there, without the grant.
	if !inExile(g, bracelet) || g.Battlefield.Contains(bracelet) {
		t.Fatal("The Dominion Bracelet was not exiled as a cost")
	}
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the equipped creature left; the cost exiles the Bracelet, not the creature")
	}
	if c := gaLayered(t, g, bear); c.PowerForComparison() != 3 {
		t.Errorf("power after the Bracelet left = %d, want 3", c.PowerForComparison())
	}
	passPriorityAroundTable(t, g)
	w := scheduledControls(g)
	if len(w) != 1 || w[0].Controller != me.ID || w[0].Controlled != opp.ID || w[0].ExtraTurn != nil || w[0].SourceName != "The Dominion Bracelet" {
		t.Fatalf("scheduled %+v, want The Dominion Bracelet: %s controls %s with no extra turn", w, me.Name, opp.Name)
	}
	walkControlledTurnWithNoExtraTurn(t, g, me, opp)
}

// "This creature's power": the discount follows the equipped creature,
// and never goes below {0} of generic or above the creature's power.
func TestDominionBraceletDiscountIsTheCreaturesPower(t *testing.T) {
	for _, tc := range []struct {
		power, want int
	}{
		{0, 14},
		{6, 8},
		{20, 0},
	} {
		g, me, bear, _, idx, _ := braceletTable(t, tc.power)
		price, err := g.PriceActivation(me.ID, bear, idx, nil)
		if err != nil {
			t.Fatalf("PriceActivation: %v", err)
		}
		if got := price.Total.Generic; got != tc.want || len(price.Total.Required) != 0 {
			t.Errorf("power %d (+1): price %+v, want {%d}", tc.power, price.Total, tc.want)
		}
	}
}

// Short of the discounted price, the activation is refused and nothing
// is paid: the Bracelet stays on the creature.
func TestDominionBraceletShortPriceExilesNothing(t *testing.T) {
	g, me, bear, bracelet, idx, ref := braceletTable(t, 3)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, bear, idx, game.ActivateAbilityParams{
		Ref: ref, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err == nil {
		t.Fatal("activated with {10} for an {11} ability")
	}
	if !g.Battlefield.Contains(bracelet) {
		t.Error("a refused activation exiled the Bracelet")
	}
	if n := len(scheduledControls(g)); n != 0 {
		t.Errorf("%d controls scheduled by a refused activation", n)
	}
}

// "Activate only as a sorcery": refused outside the main phase.
func TestDominionBraceletIsSorcerySpeed(t *testing.T) {
	g, me, bear, bracelet, idx, ref := braceletTable(t, 3)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for g.Turn.Step != game.StepBeginCombat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, bear, idx, game.ActivateAbilityParams{
		Ref: ref, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err == nil {
		t.Fatal("activated in the beginning of combat step")
	}
	if !g.Battlefield.Contains(bracelet) {
		t.Error("a refused activation exiled the Bracelet")
	}
}

func TestDominionBraceletIsFull(t *testing.T) {
	if spec, ok := Lookup(dominionBraceletOracle); !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("The Dominion Bracelet is not Full without caveats: %v %v", spec.Completeness, spec.Caveats)
	}
}

// --- the boot-time refusals (#4000, #3999) ---------------------------

func TestRegisterRefusesExileGrantorOutsideAGrant(t *testing.T) {
	mustPanic(t, "not a granted ability", func() {
		Register(Spec{
			OracleID: "exile-grantor-own-row", Name: "Own Row",
			Activated: []ActivatedAbility{{
				Label: "{1}, Exile it: Draw.", Cost: Plus(ManaCost("{1}"), ExileTheGrantor()),
				Effect: func(*game.Game, *game.StackItem) error { return nil },
			}},
		})
	})
	mustPanic(t, "picks permanents", func() {
		Register(Spec{
			OracleID: "exile-grantor-beside-sacrifice", Name: "Beside Sacrifice",
			Grants: []AbilityGrant{{
				Key: "exile-grantor-test/beside-sacrifice",
				Activated: []ActivatedAbility{{
					Label:  "{1}, Exile it, Sacrifice two creatures: Draw.",
					Cost:   Plus(ManaCost("{1}"), ExileTheGrantor(), SacrificeAnotherN(2, "two creatures", Creature())),
					Effect: func(*game.Game, *game.StackItem) error { return nil },
				}},
				Text: "test",
			}},
		})
	})
}

func TestRegisterRefusesAMalformedOptionalWaterbend(t *testing.T) {
	mustPanic(t, "mandatory AdditionalCost slot", func() {
		ac := OptionalWaterbend("{3}")
		ac.Optional, ac.Key, ac.ManaCost = false, "", ""
		Register(Spec{OracleID: "waterbend-mandatory", Name: "Mandatory", AdditionalCost: &ac})
	})
	mustPanic(t, "malformed waterbend cost", func() {
		oc := OptionalWaterbend("{3}")
		oc.ManaCost = "{4}"
		Register(Spec{OracleID: "waterbend-wrong-mana", Name: "Wrong Mana", OptionalCosts: []game.AdditionalCost{oc}})
	})
	mustPanic(t, "a cast carries one tap cost", func() {
		Register(Spec{
			OracleID: "waterbend-with-convoke", Name: "With Convoke",
			TapCost: Convoke(), OptionalCosts: []game.AdditionalCost{OptionalWaterbend("{3}")},
		})
	})
	mustPanic(t, "no waterbend clause", func() {
		oc := Kicker("{3}")
		oc.Key = game.WaterbendKey
		Register(Spec{OracleID: "waterbend-key-only", Name: "Key Only", OptionalCosts: []game.AdditionalCost{oc}})
	})
}
