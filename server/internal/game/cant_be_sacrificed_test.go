package game

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// cant_be_sacrificed_test.go pins #3275: "<that permanent> can't be
// sacrificed this turn" (Slicer, Hired Muscle). The record is read by
// the cost validator, the effect's candidate walk and the sacrifice
// route, so each test here is one of those three. The static form
// (Alexios, Deimos of Kosmos) is pinned with the card in
// cards/effects/alexios_deimos_of_kosmos_test.go.

// markCantBeSacrificed is "<id> can't be sacrificed this turn".
func markCantBeSacrificed(t *testing.T, g *Game, id uuid.UUID) {
	t.Helper()
	var ok bool
	g.WithWriteLock(func() { ok = g.CantBeSacrificedThisTurnForEffect(uuid.Nil, id, "Slicer, Hired Muscle") })
	if !ok {
		t.Fatalf("CantBeSacrificedThisTurnForEffect(%s) registered nothing", id)
	}
	if !g.CantBeSacrificedForEffect(id) {
		t.Fatalf("the record is not read back for %s", id)
	}
}

// CR 118.3: a sacrifice cost can't be paid with it, and the refused
// payment moves nothing. Another artifact still pays.
func TestCantBeSacrificedRefusesASacrificeCost(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	var seen int
	src := pushVarSacSource(g, me, openSacrificeCost(), &seen)
	kept := pushVarSacArtifact(g, me, "Kept Rock")
	free := pushVarSacArtifact(g, me, "Free Rock")
	markCantBeSacrificed(t, g, kept)

	err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{SacrificeIDs: []uuid.UUID{kept, free}})
	if !errors.Is(err, ErrCantBeSacrificed) || !errors.Is(err, ErrIllegalTarget) {
		t.Fatalf("activate naming the protected rock: %v, want ErrCantBeSacrificed (an ErrIllegalTarget)", err)
	}
	if !g.Battlefield.Contains(kept) || !g.Battlefield.Contains(free) {
		t.Fatal("a refused payment moved something")
	}
	var cands []uuid.UUID
	g.WithWriteLock(func() {
		cands = g.SacrificeCostCandidatesForEffect(me.ID, uuid.Nil, openSacrificeCost().SacrificeOther)
	})
	if slices.Contains(cands, kept) || !slices.Contains(cands, free) {
		t.Errorf("cost candidates = %v, want the free rock and not the protected one", cands)
	}

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{SacrificeIDs: []uuid.UUID{free}}); err != nil {
		t.Fatalf("activate with the free rock: %v", err)
	}
	if g.Battlefield.Contains(free) || !g.Battlefield.Contains(kept) {
		t.Error("the free rock should be sacrificed and the protected one kept")
	}
}

// "Sacrifice this" can't be paid by a source that can't be sacrificed:
// a Treasure under the record makes no mana and stays.
func TestCantBeSacrificedRefusesASacrificeThisCost(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	id := pushIntrinsicPermanent(g, me, "Treasure", "Token Artifact — Treasure", treasureAbility(), nil)
	markCantBeSacrificed(t, g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, ManaAbilityParams{}); !errors.Is(err, ErrCantBeSacrificed) {
		t.Fatalf("ActivateManaAbility: %v, want ErrCantBeSacrificed", err)
	}
	if !g.Battlefield.Contains(id) || len(me.ManaPool) != 0 || hasEvent(g, EventSacrifice, id) {
		t.Error("the refused Treasure left, made mana or announced a sacrifice")
	}
}

// CR 101.2: an edict passes over it. A player with a protected creature
// and a free one sacrifices the free one without being asked (the only
// legal pick, ADR 0175), and a player whose only creature is protected
// is not asked at all.
func TestCantBeSacrificedIsNotAnEdictChoice(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushBear(g, me.ID)
	spare := pushBear(g, me.ID)
	theirs := pushBear(g, opp.ID)
	markCantBeSacrificed(t, g, mine)
	markCantBeSacrificed(t, g, theirs)

	var candidates []uuid.UUID
	g.WithWriteLock(func() { candidates = g.EffectSacrificeCandidatesForEffect(me.ID, nil) })
	if slices.Contains(candidates, mine) || !slices.Contains(candidates, spare) {
		t.Fatalf("edict candidates = %v, want the spare bear only", candidates)
	}

	run := eachPlayerSacrificesThen(t, g, uuid.Nil, uuid.Nil)
	if n := openSacrificePrompts(g); n != 0 {
		t.Fatalf("%d sacrifice prompts open, want none (forced pick, and nothing to pick)", n)
	}
	if run.ran != 1 {
		t.Fatalf("continuation ran %d times, want 1", run.ran)
	}
	if g.Battlefield.Contains(spare) {
		t.Error("the spare bear should have been sacrificed")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(theirs) {
		t.Error("a bear that can't be sacrificed was sacrificed")
	}
	if run.got.Sacrificed(opp.ID) {
		t.Error("the opponent sacrificed something with only a protected bear")
	}
}

// The move itself: "sacrifice it" from a delayed trigger or a "sacrifice
// all" does nothing to it, announces nothing, and the rest of the batch
// still goes. The player's own sacrifice action is refused.
func TestCantBeSacrificedLeavesItWhereItIs(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	kept := pushBear(g, me.ID)
	gone := pushBear(g, me.ID)
	markCantBeSacrificed(t, g, kept)

	var landed []uuid.UUID
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(kept); err != nil {
			t.Fatalf("SacrificePermanentForEffect: %v, want nil (the instruction does nothing)", err)
		}
		if err := g.SacrificeAllThenForEffect(uuid.Nil, []uuid.UUID{kept, gone}, func(_ *Game, l []uuid.UUID) error {
			landed = l
			return nil
		}); err != nil {
			t.Fatalf("SacrificeAllThenForEffect: %v", err)
		}
	})
	if !g.Battlefield.Contains(kept) || hasEvent(g, EventSacrifice, kept) {
		t.Error("the protected bear left or announced a sacrifice")
	}
	if g.Battlefield.Contains(gone) || !slices.Equal(landed, []uuid.UUID{gone}) {
		t.Errorf("landed = %v, want only the other bear", landed)
	}
	if err := g.SacrificePermanent(me.ID, kept); !errors.Is(err, ErrCantBeSacrificed) {
		t.Errorf("SacrificePermanent: %v, want ErrCantBeSacrificed", err)
	}
}

// CR 514.2: "this turn" ends at cleanup.
func TestCantBeSacrificedEndsAtCleanup(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushBear(g, me.ID)
	markCantBeSacrificed(t, g, bear)
	g.WithWriteLock(func() { g.sweepTurnEndLocked() })
	if g.CantBeSacrificedForEffect(bear) {
		t.Fatal("the record survived the cleanup sweep")
	}
	if err := g.SacrificePermanent(me.ID, bear); err != nil {
		t.Fatalf("SacrificePermanent after cleanup: %v", err)
	}
}

// CR 400.7: the record names the object, so a permanent that left and
// came back may be sacrificed; and nothing is registered for a card that
// isn't on the battlefield.
func TestCantBeSacrificedDoesNotFollowAFlicker(t *testing.T) {
	g := newActiveGame(t)
	bear := pushBear(g, g.Seats[0].ID)
	markCantBeSacrificed(t, g, bear)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].EnteredBattlefieldAt += 1_000_000
			}
		}
	})
	if g.CantBeSacrificedForEffect(bear) {
		t.Error("the new object still carries the old object's record")
	}
	before := len(g.ScopedEffects)
	var ok bool
	g.WithWriteLock(func() { ok = g.CantBeSacrificedThisTurnForEffect(uuid.Nil, uuid.New(), "x") })
	if ok || len(g.ScopedEffects) != before {
		t.Errorf("ok=%v, %d records (was %d); want false and no new record", ok, len(g.ScopedEffects), before)
	}
}
