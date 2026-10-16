package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// text_words_costs_test.go — #3984: the cost walks that match a card
// with no TargetSpec read the clause's words through the text of the
// permanent whose ability it is (costTextLocked), as the targeting-free
// sacrifice walk does since #3872 (TestCostWalkReadsTheSourcesWords).

// twcChanged puts a source on the battlefield and registers `ch` on its
// text.
func twcChanged(t *testing.T, g *Game, controller uuid.UUID, ch TextWordChange) uuid.UUID {
	t.Helper()
	src := pushBattlefieldForTest(g, controller, "Altar", "Artifact", "")
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if !g.ChangeTextWordForEffect(uuid.Nil, src, ch, IndefiniteDuration(), "test") {
			t.Fatal("setup: the change was registered")
		}
	})
	return src
}

// A sacrifice clause's set rule ("Sacrifice a Swamp and a Forest",
// TargetSpec.EachOf) reads its kinds through the source's text: under
// "Swamp → Island" an Island and a Forest fill it and a Swamp and a
// Forest do not, in the validator's check, the enumerator's payment and
// the view's groups alike. With no source the printed kinds stand.
func TestCostSetRuleReadsTheSourcesWords(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := twcChanged(t, g, me.ID, twChange(TextWordBasicLandType, "Swamp", "Island"))
	swamp := pushBattlefieldForTest(g, me.ID, "Swamp", "Basic Land — Swamp", "")
	island := pushBattlefieldForTest(g, me.ID, "Island", "Basic Land — Island", "")
	forest := pushBattlefieldForTest(g, me.ID, "Forest", "Basic Land — Forest", "")
	spec := &TargetSpec{
		Mode: "permanent", Label: "a Swamp and a Forest", Zones: []ZoneKind{ZoneBattlefield},
		Min: 2, Max: 2,
		EachOf: []SacrificeKind{
			{Label: "a Swamp", Subtypes: []string{"Swamp"}},
			{Label: "a Forest", Subtypes: []string{"Forest"}},
		},
	}
	var changedIsland, changedSwamp, printedSwamp, printedIsland bool
	var pay []uuid.UUID
	var groups []SacrificeSetGroup
	g.ReadSnapshot(func() {
		changedIsland = g.SacrificeSetSatisfiedForEffect(src, spec, []uuid.UUID{island, forest})
		changedSwamp = g.SacrificeSetSatisfiedForEffect(src, spec, []uuid.UUID{swamp, forest})
		printedSwamp = g.SacrificeSetSatisfiedForEffect(uuid.Nil, spec, []uuid.UUID{swamp, forest})
		printedIsland = g.SacrificeSetSatisfiedForEffect(uuid.Nil, spec, []uuid.UUID{island, forest})
		pay = g.SacrificeSetPaymentForEffect(src, spec, []uuid.UUID{swamp, island, forest})
		groups = g.SacrificeSetGroupsForEffect(src, spec, []uuid.UUID{swamp, island, forest})
	})
	if !changedIsland || changedSwamp {
		t.Errorf("under Swamp → Island: Island+Forest %v (want true), Swamp+Forest %v (want false)", changedIsland, changedSwamp)
	}
	if !printedSwamp || printedIsland {
		t.Errorf("printed: Swamp+Forest %v (want true), Island+Forest %v (want false)", printedSwamp, printedIsland)
	}
	if len(pay) != 2 || pay[0] != island || pay[1] != forest {
		t.Errorf("payment %v: want the Island then the Forest", pay)
	}
	if len(groups) != 2 || !slices.Equal(groups[0].Candidates, []uuid.UUID{island}) {
		t.Errorf("the first group is the Island alone: %+v", groups)
	}
}

// A craft clause's "one of each" subtypes (ExilePermanentsCost.EachSubtype,
// Throne of the Grim Captain) and its colour read the source's text too.
func TestExilePermanentsCostReadsTheSourcesWords(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := twcChanged(t, g, me.ID, twChange(TextWordCreatureType, "Dinosaur", "Elf"))
	dino := pushBattlefieldForTest(g, me.ID, "Dino", "Creature — Dinosaur", "")
	elf := pushBattlefieldForTest(g, me.ID, "Elf", "Creature — Elf", "")
	merfolk := pushBattlefieldForTest(g, me.ID, "Merfolk", "Creature — Merfolk", "")
	ec := &ExilePermanentsCost{Count: 2, Label: "a Dinosaur and a Merfolk", EachSubtype: []string{"Dinosaur", "Merfolk"}}
	var opts, printed, pay []uuid.UUID
	var validChanged, validPrinted error
	g.WithWriteLock(func() {
		opts = g.ExilePermanentsOptionsForEffect(me.ID, src, ec)
		printed = g.ExilePermanentsOptionsForEffect(me.ID, uuid.Nil, ec)
		pay = g.ExilePermanentsPaymentForEffect(me.ID, src, ec, opts)
		validChanged = g.validateExilePermanentsCostLocked(me.ID, src, ec, []uuid.UUID{dino, merfolk}, nil)
		validPrinted = g.validateExilePermanentsCostLocked(me.ID, src, ec, []uuid.UUID{elf, merfolk}, nil)
	})
	if !slices.Contains(opts, elf) || slices.Contains(opts, dino) {
		t.Errorf("options %v: want the Elf and not the Dinosaur", opts)
	}
	if !slices.Contains(printed, dino) || slices.Contains(printed, elf) {
		t.Errorf("printed options %v: want the Dinosaur and not the Elf", printed)
	}
	if len(pay) != 2 || !slices.Contains(pay, elf) || !slices.Contains(pay, merfolk) {
		t.Errorf("payment %v: want the Elf and the Merfolk", pay)
	}
	if validChanged == nil {
		t.Error("a Dinosaur paid \"an Elf and a Merfolk\"")
	}
	if validPrinted != nil {
		t.Errorf("an Elf and a Merfolk pay it: %v", validPrinted)
	}
}

// A declared mana restriction and a declared spend rider name the
// source's words (restrictionsFor, spendRidersFor); a computed restriction
// names a chosen word and is left alone.
func TestManaRestrictionsReadTheSourcesWords(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := twcChanged(t, g, me.ID, twChange(TextWordCreatureType, "Sliver", "Elf"))
	g.WithWriteLock(func() {
		if ok := g.ChangeTextWordForEffect(uuid.Nil, src, twChange(TextWordColor, "R", "G"), IndefiniteDuration(), "test"); !ok {
			t.Fatal("setup: the colour change was registered")
		}
	})
	declared := &ManaAbilityShape{
		Restrictions: []string{ManaRestrictCast, ManaRestrictSubtype("Sliver"), ManaRestrictType("Creature")},
		SpendRiders:  []ManaSpendRider{{When: []string{ManaRestrictColor("R")}}},
	}
	computed := &ManaAbilityShape{RestrictionsFunc: func(*Game, uuid.UUID, uuid.UUID) []string {
		return []string{ManaRestrictSubtype("Sliver")}
	}}
	var rs, chosen []string
	var riders []ManaSpendRider
	g.ReadSnapshot(func() {
		rs = restrictionsFor(g, declared, me.ID, src)
		riders = spendRidersFor(g, declared, src)
		chosen = restrictionsFor(g, computed, me.ID, src)
	})
	if !slices.Contains(rs, ManaRestrictSubtype("Elf")) || slices.Contains(rs, ManaRestrictSubtype("Sliver")) {
		t.Errorf("declared restrictions %v: want the Elf tag", rs)
	}
	if !slices.Contains(rs, ManaRestrictCast) || !slices.Contains(rs, ManaRestrictType("Creature")) {
		t.Errorf("declared restrictions %v: the other tags are kept as they are", rs)
	}
	if len(riders) != 1 || !slices.Equal(riders[0].When, []string{ManaRestrictColor("G")}) {
		t.Errorf("rider filter %+v: want green", riders)
	}
	if !slices.Equal(chosen, []string{ManaRestrictSubtype("Sliver")}) {
		t.Errorf("a computed (chosen) restriction is not text: %v", chosen)
	}
	if declared.Restrictions[1] != ManaRestrictSubtype("Sliver") || declared.SpendRiders[0].When[0] != ManaRestrictColor("R") {
		t.Error("the catalog's declared lists were written through")
	}
}
