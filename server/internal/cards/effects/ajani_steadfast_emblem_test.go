package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ajani_steadfast_emblem_test.go — #4192. EmblemSpec.Replacements: the
// replacement gather walks every seat's emblems with the emblem as the
// source and its owner as "you" (CR 114.2, ADR 0109 §5), shown by Ajani
// Steadfast's −7.

// ajaniEmblemFor makes owner's Ajani Steadfast emblem through the real
// −7 (loyalty row 2) and returns the walker.
func ajaniEmblemFor(t *testing.T, g *game.Game, owner *game.Player) uuid.UUID {
	t.Helper()
	ajani := pushWalkerForTest(g, owner.ID, "Ajani Steadfast", guffaAjaniOracle, 7)
	if err := g.ActivateCatalogAbility(owner.ID, ajani, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("−7: %v", err)
	}
	passPriorityAroundTable(t, g)
	if emblems := emblemsOf(g, owner.ID); len(emblems) != 1 {
		t.Fatalf("%d emblems after the −7, want 1", len(emblems))
	}
	return ajani
}

func TestAjaniSteadfastIsFullWithItsEmblem(t *testing.T) {
	spec, ok := Lookup(guffaAjaniOracle)
	if !ok {
		t.Fatal("Ajani Steadfast is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Ajani Steadfast: %v with caveats %v, want full with none", spec.Completeness, spec.Caveats)
	}
	if len(spec.Activated) != 3 || spec.Emblem == nil || len(spec.Emblem.Replacements) != 1 {
		t.Errorf("Ajani Steadfast offers %d abilities (emblem %+v), want the three and a one-replacement emblem", len(spec.Activated), spec.Emblem)
	}
}

// The −7 makes the emblem; damage to its owner is cut to 1, damage of 1
// is not touched, and the opponent's life is untouched.
func TestAjaniSteadfastEmblemReducesDamageToItsOwnerToOne(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	advanceToMainOf(t, g, seat)
	ajaniEmblemFor(t, g, me)
	src := pr7Creature(g, opp.ID, "Pinger", 1, "R")

	life, theirs := me.Life, opp.Life
	pr7Hit(t, g, src, me.ID, 5)
	if me.Life != life-1 {
		t.Errorf("life %d after 5 damage, want %d (all but 1 prevented)", me.Life, life-1)
	}
	pr7Hit(t, g, src, me.ID, 1)
	if me.Life != life-2 {
		t.Errorf("life %d after 1 more damage, want %d (1 damage is untouched)", me.Life, life-2)
	}
	pr7Hit(t, g, src, me.ID, 0)
	if me.Life != life-2 {
		t.Errorf("life %d after 0 damage, want %d", me.Life, life-2)
	}
	// Damage to the emblem's owner's opponent is not the owner's.
	other := pr7Creature(g, me.ID, "Mine", 1, "W")
	pr7Hit(t, g, other, opp.ID, 5)
	if opp.Life != theirs-5 {
		t.Errorf("opponent life %d after 5 damage, want %d (the emblem is not theirs)", opp.Life, theirs-5)
	}
}

// Damage to a planeswalker the owner controls is cut to 1; damage to
// another player's planeswalker, and to the owner's creatures, is not.
func TestAjaniSteadfastEmblemReducesDamageToTheOwnersPlaneswalkers(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	advanceToMainOf(t, g, seat)
	ajaniEmblemFor(t, g, me)
	mine := pushWalkerForTest(g, me.ID, "My Walker", "", 6)
	theirs := pushWalkerForTest(g, opp.ID, "Their Walker", "", 6)
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	src := pr7Creature(g, opp.ID, "Pinger", 1, "R")

	pr6Damage(t, g, src, mine, 4)
	if got := loyaltyOf(g, mine); got != 5 {
		t.Errorf("my walker loyalty %d after 4 damage, want 5 (all but 1 prevented)", got)
	}
	pr6Damage(t, g, src, theirs, 4)
	if got := loyaltyOf(g, theirs); got != 2 {
		t.Errorf("their walker loyalty %d after 4 damage, want 2 (not the owner's)", got)
	}
	pr6Damage(t, g, src, bear, 3)
	if got := pr6Marked(g, bear); got != 3 {
		t.Errorf("my creature has %d damage, want 3 (the emblem protects players and planeswalkers only)", got)
	}
}

// "That damage can't be prevented" leaves the emblem applied once and
// preventing nothing (CR 615.12).
func TestAjaniSteadfastEmblemDoesNotPreventUnpreventableDamage(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	advanceToMainOf(t, g, seat)
	ajaniEmblemFor(t, g, me)
	src := pr7Creature(g, opp.ID, "Pinger", 1, "R")
	life := me.Life
	g.WithWriteLock(func() {
		if err := g.DealMarkedDamageForEffect(src, nil, me.ID, 4, game.DamageMarks{CantBePrevented: true}); err != nil {
			t.Fatal(err)
		}
	})
	if me.Life != life-4 {
		t.Errorf("life %d, want %d (the damage can't be prevented)", me.Life, life-4)
	}
}

// An emblem is gathered with its owner as "you" whoever's turn it is
// and whatever else is on the table: two players with an emblem each
// are each protected, and an emblem outlives its Ajani.
func TestAjaniSteadfastEmblemsAreEachTheirOwners(t *testing.T) {
	g := newCatalogGame(t)
	a, b := g.Seats[0], g.Seats[1]
	for _, p := range []*game.Player{a, b} {
		ajani := pushCatalogPermanent(g, p.ID, "Ajani Steadfast", "Legendary Planeswalker — Ajani", guffaAjaniOracle, false)
		g.WithWriteLock(func() {
			if err := g.CreateEmblemForEffect(p.ID, ajani); err != nil {
				t.Fatal(err)
			}
			// The walker leaves; the emblem stays (CR 114.2).
			_ = g.DestroyPermanentForEffect(ajani)
		})
	}
	src := pr7Creature(g, g.Seats[2].ID, "Pinger", 1, "R")
	la, lb := a.Life, b.Life
	pr7Hit(t, g, src, a.ID, 4)
	pr7Hit(t, g, src, b.ID, 3)
	if a.Life != la-1 || b.Life != lb-1 {
		t.Errorf("lives %d and %d, want %d and %d", a.Life, b.Life, la-1, lb-1)
	}
	// The third player has no emblem.
	c := g.Seats[2]
	lc := c.Life
	pr7Hit(t, g, pr7Creature(g, a.ID, "Mine", 1, "W"), c.ID, 4)
	if c.Life != lc-4 {
		t.Errorf("third player's life %d, want %d", c.Life, lc-4)
	}
}

// The emblem's owner orders it against other effects that would apply
// to the same damage (CR 616.1): against a redirection to a creature, the
// order decides whether 1 or all 3 damage is dealt to it.
func TestAjaniSteadfastEmblemIsOrderedAgainstOtherEffects(t *testing.T) {
	for _, tc := range []struct {
		name        string
		emblemFirst bool
		wantHost    int
	}{
		{"emblem first, then the redirection of what is left", true, 1},
		{"redirection first, so the emblem has nothing left to prevent", false, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			seat := g.Turn.ActiveSeat
			me := g.Seats[seat]
			opp := g.Seats[(seat+1)%len(g.Seats)]
			advanceToMainOf(t, g, seat)
			ajaniEmblemFor(t, g, me)
			host := pr7Creature(g, me.ID, "Host", 1, "W")
			pushAuraOnLand(t, g, me.ID, "Pariah", pr9Pariah, host)
			src := pr7Creature(g, opp.ID, "Pinger", 1, "R")
			life := me.Life
			g.WithWriteLock(func() {
				if err := g.DealDamageToPlayerForEffect(src, me.ID, 3); err != nil {
					t.Fatal(err)
				}
			})
			var order *game.PendingChoice
			for _, c := range g.PendingChoices {
				if c != nil && c.Kind == game.PendingChoiceReplacementOrder {
					order = c
				}
			}
			if order == nil || order.Chooser != me.ID || len(order.ReplacementEffectIDs) != 2 {
				t.Fatalf("want a CR 616 ordering prompt for me between the emblem and Pariah, got %+v", order)
			}
			var emblemID, pariahID game.ReplacementEffectID
			g.ReadSnapshot(func() {
				for _, id := range order.ReplacementEffectIDs {
					label, _ := g.ReplacementOptionMetaForEffect(id)
					if label == "Ajani Steadfast emblem — prevent all but 1 of that damage" {
						emblemID = id
					} else {
						pariahID = id
					}
				}
			})
			if emblemID == 0 || pariahID == 0 {
				t.Fatalf("the prompt did not name the emblem and Pariah: %v", order.ReplacementEffectIDs)
			}
			ids := []game.ReplacementEffectID{pariahID, emblemID}
			if tc.emblemFirst {
				ids = []game.ReplacementEffectID{emblemID, pariahID}
			}
			if err := g.ResolveReplacementOrder(order.ID, me.ID, ids); err != nil {
				t.Fatal(err)
			}
			g.RunStateChecksForTest()
			if me.Life != life {
				t.Errorf("life %d, want %d (Pariah redirected the damage)", me.Life, life)
			}
			if got := pr6Marked(g, host); got != tc.wantHost {
				t.Errorf("host has %d damage, want %d", got, tc.wantHost)
			}
		})
	}
}

// An emblem over the replacement slot budget is refused at registration,
// as a Spec is; an emblem with only replacements is a legal emblem.
func TestEmblemReplacementsAreCheckedAtRegistration(t *testing.T) {
	checkEmblemSpec("Test", &EmblemSpec{
		Label: "Test emblem", Text: "Test.",
		Replacements: []game.ReplacementEffect{{}},
	})
	defer func() {
		if recover() == nil {
			t.Error("an emblem over the replacement slot budget was accepted")
		}
	}()
	checkEmblemSpec("Test", &EmblemSpec{
		Label: "Test emblem", Text: "Test.",
		Replacements: make([]game.ReplacementEffect, game.MaxCatalogReplacementSlots+1),
	})
}
