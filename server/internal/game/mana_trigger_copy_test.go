package game

import (
	"testing"

	"github.com/google/uuid"
)

// mana_trigger_copy_test.go — where #763's triggered mana abilities and
// #779's one-colour planning meet #967's copy effects (CR 707.2).
//
// Both rules below are correct by CONSTRUCTION today: PrintedValues
// carries the OracleID, and every ability hook — ManaAbilitiesForCard,
// ManaTriggersForCard — keys on it through CatalogAbilityKey. Neither
// needed a line of code. They are pinned here because a change to what
// is copiable would break either one silently, and both failures are
// stronger-than-printed.

// A copy of a permanent that adds "N mana of any one color" keeps ONE
// pick minting N, and the planner reads the copy the same way (#779).
// If a copy ever took the produced string apart, a copied Gilded Lotus
// would become three independent colour picks — three colours out of
// one activation, and a plan the activation could never honour.
func TestCopyOfAOneColourSourceKeepsOnePick(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	const lotusOracle = "test-gilded-lotus"
	withCatalogHook(t, func(oracleID string) []ManaAbilityShape {
		if oracleID == lotusOracle {
			return oneColorShape("{W3|U3|B3|R3|G3}")
		}
		return nil
	})
	lotus := pushBattlefieldForTest(g, me.ID, "Gilded Lotus", "Artifact", lotusOracle)
	blank := pushBattlefieldForTest(g, me.ID, "Copy Target", "Artifact", "test-blank-artifact")
	copyOntoForTest(g, lotus, blank)

	// One pick, not three: the copy activates exactly as the original.
	if err := g.ActivateManaAbility(me.ID, blank, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility on the copy: %v", err)
	}
	picks := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceMana {
			picks++
		}
	}
	if picks != 1 {
		t.Fatalf("the copy queued %d mana picks, want exactly one", picks)
	}
	pick := choiceByKind(g, PendingChoiceMana)
	if pick.ManaAmounts["U"] != 3 {
		t.Errorf("copy amounts = %v, want three of each colour", pick.ManaAmounts)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := poolColors(me); got["U"] != 3 || len(got) != 1 {
		t.Errorf("pool = %v, want three blue from the copy", got)
	}

	// And the planner reads the copy as a one-colour source too: the
	// still-untapped ORIGINAL covers {G}{G}{G} off its single pick.
	cost, _ := ParseCost("{G}{G}{G}")
	plan, ok := g.AutoTapForCost(me.ID, cost, 0)
	if !ok || len(plan) != 1 || plan[0] != lotus {
		t.Errorf("plan = %v (ok=%v), want the untapped original to cover {G}{G}{G}", plan, ok)
	}
}

// A copy of the LAND a Wild Growth enchants does NOT get the Aura's
// trigger, and this would catch either of the two independent reasons
// going wrong: the trigger is an ability of the AURA, keyed off the
// Aura's oracle ID and never part of the land's copiable values; and
// Wild Growth's own condition asks whether it is attached to the
// permanent that produced, which the copy is not.
//
// A copy of the AURA is the other half, and does carry the trigger.
func TestCopyOfAnEnchantedLandDoesNotCopyTheAurasManaTrigger(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	forest := pushForest(g, me)
	aura := pushAuraOn(g, me, "Wild Growth", wildGrowthTestOracle, forest)
	blank := pushBattlefieldForTest(g, me.ID, "Copy Target", "Land", "test-blank-land")
	withCatalogManaTriggers(t, func(key string) []ManaTrigger {
		if key == wildGrowthTestOracle {
			return []ManaTrigger{wildGrowthTrigger("{G}")}
		}
		return nil
	})
	copyOntoForTest(g, forest, blank)

	g.ReadSnapshot(func() {
		if n := len(ManaTriggersForCard(*g.findCardByIDLocked(blank))); n != 0 {
			t.Errorf("the copied land carries %d mana triggers: the Aura's ability leaked into the land's copiable values", n)
		}
	})

	// Tapping the COPY adds its own {G} and nothing else.
	if err := g.ActivateManaAbility(me.ID, blank, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility on the copy: %v", err)
	}
	if got := poolColors(me); got["G"] != 1 {
		t.Fatalf("pool = %v after tapping the copy, want one {G}", got)
	}

	// Tapping the ORIGINAL still fires it — the copy changed nothing.
	if err := g.ActivateManaAbility(me.ID, forest, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility on the original: %v", err)
	}
	if got := poolColors(me); got["G"] != 3 {
		t.Errorf("pool = %v, want the copy's {G} plus the Forest's and Wild Growth's", got)
	}

	// The other half: the Aura's own copiable values name the Aura, so
	// a copy of Wild Growth is a Wild Growth and has the trigger.
	g.ReadSnapshot(func() {
		if n := len(ManaTriggersForCard(*g.findCardByIDLocked(aura))); n != 1 {
			t.Fatalf("the Aura has %d mana triggers, want its one", n)
		}
		if got := CopiableValuesOf(*g.findCardByIDLocked(aura)).OracleID; got != wildGrowthTestOracle {
			t.Errorf("copiable OracleID = %q, want the Aura's", got)
		}
	})
}

// copyOntoForTest makes `dst` a copy of `src` through the same
// CopiableValuesOf + applyCopy pair every printed copy effect uses
// (copy_choice.go), without needing a Clone that can legally target
// an artifact or a land.
func copyOntoForTest(g *Game, src, dst uuid.UUID) {
	g.WithWriteLock(func() {
		var values PrintedValues
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == src {
				values = CopiableValuesOf(g.Battlefield.Cards[i])
			}
		}
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == dst {
				self := g.Battlefield.Cards[i]
				g.Battlefield.Cards[i].applyCopy(values, self)
			}
		}
	})
}
