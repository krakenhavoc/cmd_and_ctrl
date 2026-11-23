package testsso

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soulblast_test.go — #2097: "As an additional cost to cast this spell,
// sacrifice all creatures you control" (Soulblast). CR 601.2b and
// 601.2h: the cost is part of casting, the caster chooses nothing, and
// the whole set is fixed and paid as the spell is cast. CR 118.3: with
// no creatures, nothing is needed, so the cost is paid. CR 701.21a:
// sacrifice is not destruction. CR 702.26b: a phased-out permanent is
// treated as though it does not exist.

const soulblastOracle = "18d4c57b-e2bf-47a0-8823-c4a79498a7ff"

// soulblastCast casts Soulblast at `target` with the given sacrifice_ids
// (normally none: the engine fills them) and returns the spell's ID.
func soulblastCast(t *testing.T, g *game.Game, target uuid.UUID, named ...uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	return effectstest.CastWithTapParams(t, g, "Soulblast", "Instant", "{3}{R}{R}{R}", soulblastOracle,
		game.CastSpellParams{Targets: effectstest.SoPlayer(target), SacrificeIDs: named})
}

func soulblastMustCast(t *testing.T, g *game.Game, target uuid.UUID) uuid.UUID {
	t.Helper()
	id, err := soulblastCast(t, g, target)
	if err != nil {
		t.Fatalf("cast Soulblast: %v", err)
	}
	return id
}

// CR 118.3: with no creatures the cost needs nothing, so it is paid,
// the spell is cast, and it deals no damage.
func TestSoulblastWithNoCreaturesIsCastAndDealsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	life := effectstest.LifeOf(g, opp.ID)
	id := soulblastMustCast(t, g, opp.ID)
	item := g.StackMeta[id]
	if item == nil {
		t.Fatal("Soulblast is not on the stack")
	}
	if item.Paid.Sacrificed != 0 || len(item.Paid.SacrificedObjects) != 0 {
		t.Fatalf("Paid = %+v, want nothing sacrificed", item.Paid)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life {
		t.Errorf("opponent's life %d → %d, want no damage", life, got)
	}
}

// Every creature the caster controls goes, as the spell is cast and
// before it resolves (CR 601.2h); nothing else does — not a noncreature
// permanent, not an opponent's creature. The damage is their total
// power.
func TestSoulblastSacrificesEveryCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.SoCreature(g, me.ID, "A", 2, 2)
	b := effectstest.SoCreature(g, me.ID, "B", 3, 3)
	c := effectstest.SoCreature(g, me.ID, "C", 4, 1)
	relic := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Relic", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID,
	})
	theirs := effectstest.SoCreature(g, opp.ID, "Theirs", 5, 5)
	life := effectstest.LifeOf(g, opp.ID)
	id := soulblastMustCast(t, g, opp.ID)
	for _, gone := range []uuid.UUID{a, b, c} {
		if effects.OnBattlefield(g, gone) {
			t.Fatalf("%v is still on the battlefield with Soulblast on the stack", gone)
		}
		if !me.Graveyard.Contains(gone) {
			t.Fatalf("%v is not in its owner's graveyard", gone)
		}
	}
	if !effects.OnBattlefield(g, relic) || !effects.OnBattlefield(g, theirs) {
		t.Fatal("a noncreature permanent or an opponent's creature was sacrificed")
	}
	if item := g.StackMeta[id]; item == nil || item.Paid.Sacrificed != 3 || len(item.Paid.SacrificedObjects) != 3 {
		t.Fatalf("Paid = %+v, want three creatures recorded", g.StackMeta[id])
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-9 {
		t.Errorf("opponent's life %d → %d, want 9 damage (2 + 3 + 4)", life, got)
	}
}

// CR 701.21a: sacrificing is not destroying, so an indestructible
// creature is sacrificed with the rest and its power counts.
func TestSoulblastSacrificesAnIndestructibleCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Unbreakable", TypeLine: "Creature — Golem",
		Power: 4, Toughness: 4, Keywords: []string{"indestructible"}, Owner: me.ID, Controller: me.ID,
	})
	life := effectstest.LifeOf(g, opp.ID)
	soulblastMustCast(t, g, opp.ID)
	if effects.OnBattlefield(g, wall) {
		t.Fatal("the indestructible creature was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-4 {
		t.Errorf("opponent's life %d → %d, want 4 damage", life, got)
	}
}

// CR 702.26b: a phased-out creature is treated as though it does not
// exist, so it is not among "all creatures you control": it stays
// phased out, and its power is not counted.
func TestSoulblastLeavesAPhasedOutCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	here := effectstest.SoCreature(g, me.ID, "Here", 2, 2)
	away := effectstest.SoCreature(g, me.ID, "Away", 5, 5)
	g.WithWriteLock(func() {
		if err := g.PhaseOutForEffect(uuid.Nil, away); err != nil {
			t.Fatal(err)
		}
	})
	life := effectstest.LifeOf(g, opp.ID)
	soulblastMustCast(t, g, opp.ID)
	if effects.OnBattlefield(g, here) {
		t.Fatal("the phased-in creature was not sacrificed")
	}
	if !g.PhasedOut.Contains(away) || me.Graveyard.Contains(away) {
		t.Fatal("the phased-out creature was sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-2 {
		t.Errorf("opponent's life %d → %d, want 2 damage (the phased-out 5/5 not counted)", life, got)
	}
}

// The damage reads what the cost paid, not the board at resolution:
// each sacrificed creature's power as it last existed (its +1/+1
// counters count, CR 608.2h), and a creature that arrives after the
// cast is neither sacrificed nor counted.
func TestSoulblastCountsWhatTheCostPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grown := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grown", TypeLine: "Creature — Beast",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 2},
	})
	life := effectstest.LifeOf(g, opp.ID)
	id := soulblastMustCast(t, g, opp.ID)
	late := effectstest.SoCreature(g, me.ID, "Late", 6, 6)
	if item := g.StackMeta[id]; item == nil || len(item.Paid.SacrificedObjects) != 1 || item.Paid.SacrificedObjects[0].ID != grown {
		t.Fatalf("Paid = %+v, want only the creature on the battlefield at the cast", g.StackMeta[id])
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-3 {
		t.Errorf("opponent's life %d → %d, want 3 damage (1 + two counters)", life, got)
	}
	if !effects.OnBattlefield(g, late) {
		t.Error("the creature that arrived after the cast was sacrificed")
	}
}

// The caster chooses nothing: sacrifice_ids may name exactly the set,
// in any order, and nothing else — not a part of it, and not a creature
// they don't control.
func TestSoulblastAcceptsOnlyTheWholeSetByName(t *testing.T) {
	t.Run("the whole set, reordered", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		a := effectstest.SoCreature(g, me.ID, "A", 1, 1)
		b := effectstest.SoCreature(g, me.ID, "B", 2, 2)
		if _, err := soulblastCast(t, g, opp.ID, b, a); err != nil {
			t.Fatalf("cast naming the whole set: %v", err)
		}
		if effects.OnBattlefield(g, a) || effects.OnBattlefield(g, b) {
			t.Fatal("the named set was not sacrificed")
		}
	})
	t.Run("a part of it", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		a := effectstest.SoCreature(g, me.ID, "A", 1, 1)
		b := effectstest.SoCreature(g, me.ID, "B", 2, 2)
		_, err := soulblastCast(t, g, opp.ID, a)
		if !errors.Is(err, game.ErrSacrificeAllMismatch) {
			t.Fatalf("err = %v, want ErrSacrificeAllMismatch", err)
		}
		if !effects.OnBattlefield(g, a) || !effects.OnBattlefield(g, b) {
			t.Fatal("a refused cast sacrificed something")
		}
	})
	t.Run("an opponent's creature", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		a := effectstest.SoCreature(g, me.ID, "A", 1, 1)
		theirs := effectstest.SoCreature(g, opp.ID, "Theirs", 2, 2)
		if _, err := soulblastCast(t, g, opp.ID, a, theirs); !errors.Is(err, game.ErrSacrificeAllMismatch) {
			t.Fatalf("err = %v, want ErrSacrificeAllMismatch", err)
		}
	})
}

// A copy deals the original's damage (CR 707.10): it carries the
// original's payment record.
func TestACopiedSoulblastUsesTheOriginalsSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SoCreature(g, me.ID, "A", 2, 2)
	effectstest.SoCreature(g, me.ID, "B", 3, 3)
	life := effectstest.LifeOf(g, opp.ID)
	id := soulblastMustCast(t, g, opp.ID)
	if cp := effectstest.SoCopyTheSpell(t, g, id, me.ID); cp.Paid.Sacrificed != 2 {
		t.Fatalf("copy's Paid = %+v, want the original's two creatures", cp.Paid)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-10 {
		t.Errorf("opponent's life %d → %d, want 5 + 5", life, got)
	}
}

// A restore point taken with Soulblast on the stack resolves to the
// same damage: the payment record, and the last-known power it reads,
// survive the round trip.
func TestSoulblastOnTheStackSurvivesARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SoCreature(g, me.ID, "A", 2, 2)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grown", TypeLine: "Creature — Beast",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 3},
	})
	life := effectstest.LifeOf(g, opp.ID)
	id := soulblastMustCast(t, g, opp.ID)
	restored := effectstest.RestoreThroughJSON(t, g)
	item := restored.StackMeta[id]
	if item == nil || item.Paid.Sacrificed != 2 || len(item.Paid.SacrificedObjects) != 2 {
		t.Fatalf("restored Paid = %+v, want two creatures recorded", restored.StackMeta[id])
	}
	effectstest.PassPriorityAroundTable(t, restored)
	if got := effectstest.LifeOf(restored, opp.ID); got != life-6 {
		t.Errorf("opponent's life %d → %d after the restore, want 6 damage (2 + 4)", life, got)
	}
}

// Register keeps "sacrifice all" on the one slot every printed card has
// it in.
func TestRegisterRefusesSacrificeAllOutsideTheMandatoryCost(t *testing.T) {
	all := effects.SacrificeAllCost("creatures you control", effects.Creature())
	effectstest.ExpectRegisterPanic(t, "optional cost", func() {
		oc := *all
		oc.Optional, oc.Key = true, game.KickerKey
		effects.CheckSacrificeAllCost(effects.Spec{Name: "Test", OptionalCosts: []game.AdditionalCost{oc}})
	})
	effectstest.ExpectRegisterPanic(t, "either/or branch", func() {
		effects.CheckSacrificeAllCost(effects.Spec{Name: "Test", AdditionalCost: effects.EitherCost(all.Keyed("all"), effects.DiscardCost(1).Keyed("discard"))})
	})
	effectstest.ExpectRegisterPanic(t, "with a count", func() {
		bad := effects.SacrificeNCost(2, "two creatures", effects.Creature())
		bad.SacrificeAll = true
		effects.CheckSacrificeAllCost(effects.Spec{Name: "Test", AdditionalCost: bad})
	})
	effectstest.ExpectRegisterPanic(t, "no sacrifice clause", func() {
		effects.CheckSacrificeAllCost(effects.Spec{Name: "Test", AdditionalCost: &game.AdditionalCost{SacrificeAll: true, DiscardCards: 1}})
	})
	effects.CheckSacrificeAllCost(effects.Spec{Name: "Test", AdditionalCost: all})
}
