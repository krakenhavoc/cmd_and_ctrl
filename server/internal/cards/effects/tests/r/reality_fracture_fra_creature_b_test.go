package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestFraCreatureBCardsAreRegistered(t *testing.T) {
	for oracle, want := range map[string]struct {
		name         string
		completeness effects.Completeness
	}{
		effectstest.RfcbDesperateFuturescribe: {"Desperate Futurescribe", effects.CompletenessFull},
		effectstest.RfcbDiviningDuelist:       {"Divining Duelist", effects.CompletenessFull},
		effectstest.RfcbDraconicVisitor:       {"Draconic Visitor", effects.CompletenessFull},
		effectstest.RfcbEardrumRattler:        {"Eardrum Rattler", effects.CompletenessFull},
		effectstest.RfcbEdgar:                 {"Edgar, Moonlit Sovereign", effects.CompletenessFull},
		effectstest.RfcbEmrakul:               {"Emrakul, the Exigent Doom", effects.CompletenessFull},
		effectstest.RfcbFateshaperAspirant:    {"Fateshaper Aspirant", effects.CompletenessFull},
		effectstest.RfcbFblthpImpossiblyLost:  {"Fblthp, Impossibly Lost", effects.CompletenessFull},
		effectstest.RfcbFblthpKnowsTheWay:     {"Fblthp, Knows the Way", effects.CompletenessFull},
		effectstest.RfcbFlickeringHound:       {"Flickering Hound", effects.CompletenessFull},
		effectstest.RfcbFrostbitePyromental:   {"Frostbite Pyromental", effects.CompletenessFull},
		effectstest.RfcbGalliaTragicHost:      {"Gallia, Tragic Host", effects.CompletenessFull},
		effectstest.RfcbGalliaMerrymaker:      {"Gallia, the Merrymaker", effects.CompletenessFull},
		effectstest.RfcbGeistOfSaintThalia:    {"Geist of Saint Thalia", effects.CompletenessFull},
		effectstest.RfcbGhaltaImmovable:       {"Ghalta the Immovable", effects.CompletenessFull},
		effectstest.RfcbGhaltaUnstoppable:     {"Ghalta the Unstoppable", effects.CompletenessFull},
		effectstest.RfcbGinger:                {"Ginger, Queen of Sweets", effects.CompletenessFull},
		effectstest.RfcbGraftSurgeon:          {"Graft Surgeon", effects.CompletenessFull},
		effectstest.RfcbGreenhousePropagator:  {"Greenhouse Propagator", effects.CompletenessFull},
		effectstest.RfcbGrimRepriser:          {"Grim Repriser", effects.CompletenessFull},
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != want.name || spec.Completeness != want.completeness {
			t.Errorf("%s: registered=%v name=%q completeness=%v, want %q %v",
				oracle, ok, spec.Name, spec.Completeness, want.name, want.completeness)
		}
	}
}

func rfcbCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return effectstest.AllCountersOn(t, g, id)[game.CounterPlusOne]
}

// --- Greenhouse Propagator ------------------------------------------

func TestGreenhousePropagatorGainsLifeForYourOtherCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.RfcbPush(g, me.ID, "Greenhouse Propagator", "Creature — Cat Druid", effectstest.RfcbGreenhousePropagator, 2, 3)
	before := me.Life
	effectstest.CastAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life = %d, want %d after my creature entered", me.Life, before+1)
	}
	effectstest.RfcbPush(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	if me.Life != before+1 {
		t.Errorf("an opponent's creature must not gain me life: %d", me.Life)
	}
}

func TestGreenhousePropagatorDoesNotTriggerOnItsOwnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Life
	effectstest.CastAndResolveCreature(t, g, "Greenhouse Propagator", "Creature — Cat Druid", effectstest.RfcbGreenhousePropagator)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before {
		t.Errorf("life = %d, want %d: \"another\" excludes its own entry", me.Life, before)
	}
}

func TestGreenhousePropagatorTapsForGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.RfcbPush(g, me.ID, "Greenhouse Propagator", "Creature — Cat Druid", effectstest.RfcbGreenhousePropagator, 2, 3)
	effectstest.AdvanceToStepInTurn(t, g, game.StepPrecombatMain)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("pool = %d, want 1 green", len(me.ManaPool))
	}
}

// --- Geist of Saint Thalia ------------------------------------------

func TestGeistOfSaintThaliaDiscountsYourNoncreatureSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Geist of Saint Thalia", "Legendary Creature — Spirit Cleric", effectstest.RfcbGeistOfSaintThalia, false)
	if got := effectstest.PriceInHand(t, g, me, "Cultivate", "Sorcery", "{2}{G}"); got != 2 {
		t.Errorf("own sorcery: %d, want 2", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Sol Ring", "Artifact", "{1}"); got != 0 {
		t.Errorf("own artifact: %d, want 0", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Bear", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("own creature: %d, want 2 (undiscounted)", got)
	}
	if got := effectstest.PriceInHand(t, g, them, "Their Cultivate", "Sorcery", "{2}{G}"); got != 3 {
		t.Errorf("opponent's sorcery: %d, want 3 (undiscounted)", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Counterspell", "Instant", "{U}{U}"); got != 2 {
		t.Errorf("{U}{U}: %d, want 2 (colours untouched)", got)
	}
}

// --- Ghalta the Unstoppable and the Immovable -----------------------

func TestGhaltaTheUnstoppableCostsGreatestPowerLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const tl = "Legendary Creature — Elder Dinosaur"
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.RfcbGhaltaUnstoppable, tl, "{8}{G}"); got != 9 {
		t.Errorf("no creatures: %d, want 9", got)
	}
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Big", TypeLine: "Creature — Beast", Power: 5, Toughness: 5})
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Medium", TypeLine: "Creature — Beast", Power: 3, Toughness: 3})
	effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Theirs", TypeLine: "Creature — Beast", Power: 9, Toughness: 9})
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.RfcbGhaltaUnstoppable, tl, "{8}{G}"); got != 4 {
		t.Errorf("greatest power 5 (not the 8 total, not their 9): %d, want 4", got)
	}
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Huge", TypeLine: "Creature — Beast", Power: 20, Toughness: 20})
	got := effectstest.SelfPriced(t, g, me, effectstest.RfcbGhaltaUnstoppable, tl, "{8}{G}", game.ZoneHand, nil)
	if got.Generic != 0 || len(got.Required) != 1 {
		t.Errorf("power 20: %+v, want exactly {G}", got)
	}
}

func TestGhaltaTheUnstoppableGivesOtherCreaturesTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	ghalta := effectstest.RfcbPush(g, me.ID, "Ghalta the Unstoppable", "Legendary Creature — Elder Dinosaur", effectstest.RfcbGhaltaUnstoppable, 8, 8)
	mine := effectstest.SeedCreature(g, "Mine", me.ID)
	theirs := effectstest.SeedCreature(g, "Theirs", opp.ID)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, mine), "trample") {
		t.Error("my other creature should have trample")
	}
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, theirs), "trample") {
		t.Error("an opponent's creature must not gain trample")
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, ghalta), "trample") {
		t.Error("Ghalta has trample as a printed keyword")
	}
}

func TestGhaltaTheImmovableCostsGreatestToughnessLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	const tl = "Legendary Creature — Elder Dinosaur"
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 7})
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 9, Toughness: 2})
	if got := effectstest.SelfPricedMV(t, g, me, effectstest.RfcbGhaltaImmovable, tl, "{8}{W}"); got != 2 {
		t.Errorf("greatest toughness 7: %d, want 2", got)
	}
}

func TestGhaltaTheImmovableLetsYourDefendersAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.RfcbPush(g, me.ID, "Ghalta the Immovable", "Legendary Creature — Elder Dinosaur", effectstest.RfcbGhaltaImmovable, 0, 7)
	wall := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4,
		Owner: me.ID, Controller: me.ID, Keywords: []string{"defender"},
	})
	theirWall := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4,
		Owner: opp.ID, Controller: opp.ID, Keywords: []string{"defender"},
	})
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, wall), "defender") {
		t.Error("my Wall should lose defender")
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, theirWall), "defender") {
		t.Error("an opponent's Wall keeps defender")
	}
}

// --- Frostbite Pyromental -------------------------------------------

func TestFrostbitePyromentalIsSacrificedAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.RfcbPush(g, me.ID, "Frostbite Pyromental", "Creature — Elemental", effectstest.RfcbFrostbitePyromental, 4, 4)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("Frostbite Pyromental should be sacrificed at the end step")
	}
	if !me.Graveyard.Contains(id) {
		t.Error("Frostbite Pyromental should be in the graveyard")
	}
}

func TestFrostbitePyromentalDrawsTwoOnCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	id := effectstest.RfcbPush(g, me.ID, "Frostbite Pyromental", "Creature — Elemental", effectstest.RfcbFrostbitePyromental, 4, 4)
	before := me.Hand.Size()
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: id, Target: opp.ID, Amount: 4, Combat: true, Actor: me.ID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("hand = %d, want %d", got, before+2)
	}
}
