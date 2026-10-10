package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_fra_creature_b_test.go — the Reality Fracture
// creatures of slice fra-creature-b.

const (
	rfcbDesperateFuturescribe = "83c12f50-e5a1-479d-afab-edb0b40412f2"
	rfcbDiviningDuelist       = "d0ce03fa-2284-4217-85b2-4db6b8a88c94"
	rfcbDraconicVisitor       = "6745e850-68f6-480b-a1e0-e06180dde169"
	rfcbEardrumRattler        = "25151d18-8e73-4b0d-a8d2-e5fa291e5729"
	rfcbEdgar                 = "1917bda6-c0e1-4c80-8009-74c28cf6b8e9"
	rfcbEmrakul               = "4421ab7d-6d9b-4edd-b5a0-53a8ed84da6f"
	rfcbFateshaperAspirant    = "bb3996b6-b68d-4ea1-b3dd-1600f8da2f1c"
	rfcbFblthpImpossiblyLost  = "17e7212e-69d8-4014-af98-447baed2cae7"
	rfcbFblthpKnowsTheWay     = "48b891b6-3175-4864-bb98-af131098c557"
	rfcbFlickeringHound       = "162421d2-8761-437f-bed9-578b61c96f1f"
	rfcbFrostbitePyromental   = "9a104f73-597b-49e3-8088-13db31f61900"
	rfcbGalliaTragicHost      = "58285b70-0d13-4723-b5a4-91fb6f9ccf03"
	rfcbGalliaMerrymaker      = "87f39199-3e4b-44fa-8406-62019eb43c10"
	rfcbGeistOfSaintThalia    = "ef32a4a9-14e2-4738-b4c2-53ce5e1d2a53"
	rfcbGhaltaImmovable       = "0d496c6e-8f7b-420f-9197-dacd2feb528f"
	rfcbGhaltaUnstoppable     = "8e5ea773-a60b-4502-9db5-ce93fa61cc89"
	rfcbGinger                = "5ea93956-085d-409b-bd95-dd669fa69eeb"
	rfcbGraftSurgeon          = "fe7938ce-6289-4bf8-a8f5-85ecccbb7f86"
	rfcbGreenhousePropagator  = "b6b77cb9-49a5-4d1a-ae2d-02ae176dc1fa"
	rfcbGrimRepriser          = "e853e2fb-f90d-4934-aa70-c8d3cdf12e56"
)

func TestFraCreatureBCardsAreRegistered(t *testing.T) {
	for oracle, want := range map[string]struct {
		name         string
		completeness Completeness
	}{
		rfcbDesperateFuturescribe: {"Desperate Futurescribe", CompletenessFull},
		rfcbDiviningDuelist:       {"Divining Duelist", CompletenessFull},
		rfcbDraconicVisitor:       {"Draconic Visitor", CompletenessFull},
		rfcbEardrumRattler:        {"Eardrum Rattler", CompletenessFull},
		rfcbEdgar:                 {"Edgar, Moonlit Sovereign", CompletenessFull},
		rfcbEmrakul:               {"Emrakul, the Exigent Doom", CompletenessCaveats},
		rfcbFateshaperAspirant:    {"Fateshaper Aspirant", CompletenessFull},
		rfcbFblthpImpossiblyLost:  {"Fblthp, Impossibly Lost", CompletenessFull},
		rfcbFblthpKnowsTheWay:     {"Fblthp, Knows the Way", CompletenessFull},
		rfcbFlickeringHound:       {"Flickering Hound", CompletenessFull},
		rfcbFrostbitePyromental:   {"Frostbite Pyromental", CompletenessFull},
		rfcbGalliaTragicHost:      {"Gallia, Tragic Host", CompletenessFull},
		rfcbGalliaMerrymaker:      {"Gallia, the Merrymaker", CompletenessFull},
		rfcbGeistOfSaintThalia:    {"Geist of Saint Thalia", CompletenessFull},
		rfcbGhaltaImmovable:       {"Ghalta the Immovable", CompletenessCaveats},
		rfcbGhaltaUnstoppable:     {"Ghalta the Unstoppable", CompletenessFull},
		rfcbGinger:                {"Ginger, Queen of Sweets", CompletenessFull},
		rfcbGraftSurgeon:          {"Graft Surgeon", CompletenessFull},
		rfcbGreenhousePropagator:  {"Greenhouse Propagator", CompletenessFull},
		rfcbGrimRepriser:          {"Grim Repriser", CompletenessFull},
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != want.name || spec.Completeness != want.completeness {
			t.Errorf("%s: registered=%v name=%q completeness=%v, want %q %v",
				oracle, ok, spec.Name, spec.Completeness, want.name, want.completeness)
		}
	}
}

// rfcbPush puts a catalog creature on the active seat's side of the
// battlefield, not summoning-sick.
func rfcbPush(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tgh int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tgh, Owner: owner, Controller: owner,
	})
}

func rfcbCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return allCountersOn(t, g, id)[game.CounterPlusOne]
}

// --- Greenhouse Propagator ------------------------------------------

func TestGreenhousePropagatorGainsLifeForYourOtherCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	rfcbPush(g, me.ID, "Greenhouse Propagator", "Creature — Cat Druid", rfcbGreenhousePropagator, 2, 3)
	before := me.Life
	castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life = %d, want %d after my creature entered", me.Life, before+1)
	}
	rfcbPush(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	if me.Life != before+1 {
		t.Errorf("an opponent's creature must not gain me life: %d", me.Life)
	}
}

func TestGreenhousePropagatorDoesNotTriggerOnItsOwnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Life
	castAndResolveCreature(t, g, "Greenhouse Propagator", "Creature — Cat Druid", rfcbGreenhousePropagator)
	passPriorityAroundTable(t, g)
	if me.Life != before {
		t.Errorf("life = %d, want %d: \"another\" excludes its own entry", me.Life, before)
	}
}

func TestGreenhousePropagatorTapsForGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := rfcbPush(g, me.ID, "Greenhouse Propagator", "Creature — Cat Druid", rfcbGreenhousePropagator, 2, 3)
	advanceToStepInTurn(t, g, game.StepPrecombatMain)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("pool = %d, want 1 green", len(me.ManaPool))
	}
}

// --- Geist of Saint Thalia ------------------------------------------

func TestGeistOfSaintThaliaDiscountsYourNoncreatureSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Geist of Saint Thalia", "Legendary Creature — Spirit Cleric", rfcbGeistOfSaintThalia, false)
	if got := priceInHand(t, g, me, "Cultivate", "Sorcery", "{2}{G}"); got != 2 {
		t.Errorf("own sorcery: %d, want 2", got)
	}
	if got := priceInHand(t, g, me, "Sol Ring", "Artifact", "{1}"); got != 0 {
		t.Errorf("own artifact: %d, want 0", got)
	}
	if got := priceInHand(t, g, me, "Bear", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("own creature: %d, want 2 (undiscounted)", got)
	}
	if got := priceInHand(t, g, them, "Their Cultivate", "Sorcery", "{2}{G}"); got != 3 {
		t.Errorf("opponent's sorcery: %d, want 3 (undiscounted)", got)
	}
	if got := priceInHand(t, g, me, "Counterspell", "Instant", "{U}{U}"); got != 2 {
		t.Errorf("{U}{U}: %d, want 2 (colours untouched)", got)
	}
}

// --- Ghalta the Unstoppable and the Immovable -----------------------

func TestGhaltaTheUnstoppableCostsGreatestPowerLess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const tl = "Legendary Creature — Elder Dinosaur"
	if got := selfPricedMV(t, g, me, rfcbGhaltaUnstoppable, tl, "{8}{G}"); got != 9 {
		t.Errorf("no creatures: %d, want 9", got)
	}
	pushPermanent(g, me.ID, game.Card{Name: "Big", TypeLine: "Creature — Beast", Power: 5, Toughness: 5})
	pushPermanent(g, me.ID, game.Card{Name: "Medium", TypeLine: "Creature — Beast", Power: 3, Toughness: 3})
	pushPermanent(g, opp.ID, game.Card{Name: "Theirs", TypeLine: "Creature — Beast", Power: 9, Toughness: 9})
	if got := selfPricedMV(t, g, me, rfcbGhaltaUnstoppable, tl, "{8}{G}"); got != 4 {
		t.Errorf("greatest power 5 (not the 8 total, not their 9): %d, want 4", got)
	}
	pushPermanent(g, me.ID, game.Card{Name: "Huge", TypeLine: "Creature — Beast", Power: 20, Toughness: 20})
	got := selfPriced(t, g, me, rfcbGhaltaUnstoppable, tl, "{8}{G}", game.ZoneHand, nil)
	if got.Generic != 0 || len(got.Required) != 1 {
		t.Errorf("power 20: %+v, want exactly {G}", got)
	}
}

func TestGhaltaTheUnstoppableGivesOtherCreaturesTrample(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	ghalta := rfcbPush(g, me.ID, "Ghalta the Unstoppable", "Legendary Creature — Elder Dinosaur", rfcbGhaltaUnstoppable, 8, 8)
	mine := seedCreature(g, "Mine", me.ID)
	theirs := seedCreature(g, "Theirs", opp.ID)
	if !hasAbility(effectiveAbilities(t, g, mine), "trample") {
		t.Error("my other creature should have trample")
	}
	if hasAbility(effectiveAbilities(t, g, theirs), "trample") {
		t.Error("an opponent's creature must not gain trample")
	}
	if !hasAbility(effectiveAbilities(t, g, ghalta), "trample") {
		t.Error("Ghalta has trample as a printed keyword")
	}
}

func TestGhaltaTheImmovableCostsGreatestToughnessLess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	const tl = "Legendary Creature — Elder Dinosaur"
	pushPermanent(g, me.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 7})
	pushPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 9, Toughness: 2})
	if got := selfPricedMV(t, g, me, rfcbGhaltaImmovable, tl, "{8}{W}"); got != 2 {
		t.Errorf("greatest toughness 7: %d, want 2", got)
	}
}

func TestGhaltaTheImmovableLetsYourDefendersAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	rfcbPush(g, me.ID, "Ghalta the Immovable", "Legendary Creature — Elder Dinosaur", rfcbGhaltaImmovable, 0, 7)
	wall := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4,
		Owner: me.ID, Controller: me.ID, Keywords: []string{"defender"},
	})
	theirWall := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4,
		Owner: opp.ID, Controller: opp.ID, Keywords: []string{"defender"},
	})
	if hasAbility(effectiveAbilities(t, g, wall), "defender") {
		t.Error("my Wall should lose defender")
	}
	if !hasAbility(effectiveAbilities(t, g, theirWall), "defender") {
		t.Error("an opponent's Wall keeps defender")
	}
}

// --- Frostbite Pyromental -------------------------------------------

func TestFrostbitePyromentalIsSacrificedAtTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := rfcbPush(g, me.ID, "Frostbite Pyromental", "Creature — Elemental", rfcbFrostbitePyromental, 4, 4)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("Frostbite Pyromental should be sacrificed at the end step")
	}
	if !me.Graveyard.Contains(id) {
		t.Error("Frostbite Pyromental should be in the graveyard")
	}
}

func TestFrostbitePyromentalDrawsTwoOnCombatDamageToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	id := rfcbPush(g, me.ID, "Frostbite Pyromental", "Creature — Elemental", rfcbFrostbitePyromental, 4, 4)
	before := me.Hand.Size()
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: id, Target: opp.ID, Amount: 4, Combat: true, Actor: me.ID})
	})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("hand = %d, want %d", got, before+2)
	}
}
