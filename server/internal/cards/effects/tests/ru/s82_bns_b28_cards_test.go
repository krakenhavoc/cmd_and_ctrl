package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b28_cards_test.go — S82 BNS batch 28 (#5929): Foundations
// Jumpstart creatures. Each test plays the real catalog card.

const (
	bnsB28DarkSupplicant   = "dc7929cb-50a9-4662-a106-c415fc03a1a9"
	bnsB28DaybreakChaplain = "ebd33f4c-2cc8-46e3-b367-ce5b37c2719b"
	bnsB28DeepwoodDenizen  = "4a6887e3-7a65-4f24-be71-183b00a0ad56"
	bnsB28DestructiveDig   = "56787f7d-d9e9-42f2-a419-31f15e6e436e"
	bnsB28DevouringSwarm   = "11656b2a-b459-4b1c-b505-8547c3507856"
	bnsB28DireFleetHoarder = "7d48ad95-d428-4e2d-b697-56907c3ce579"
	bnsB28DragonHatchling  = "3f48f6a7-906d-4287-8440-06bfabc93b89"
	bnsB28DraugrThief      = "0952cfaf-8031-4df7-b276-dcb0b8f19ab7"
	bnsB28DreadhordeArc    = "4c05b382-58ab-4a2d-a81c-408ea273b6b6"
	bnsB28DrossHopper      = "2ec8c540-73bf-4412-a53a-499e9b9fd213"
	bnsB28Tyrannodon       = "ee811356-ef6a-4972-b827-32c2d329b3c3"
)

func TestBnsB28EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsB28DarkSupplicant: "Dark Supplicant", bnsB28DaybreakChaplain: "Daybreak Chaplain",
		bnsB28DeepwoodDenizen: "Deepwood Denizen", bnsB28DestructiveDig: "Destructive Digger",
		bnsB28DevouringSwarm: "Devouring Swarm", bnsB28DireFleetHoarder: "Dire Fleet Hoarder",
		bnsB28DragonHatchling: "Dragon Hatchling", bnsB28DraugrThief: "Draugr Thought-Thief",
		bnsB28DreadhordeArc: "Dreadhorde Arcanist", bnsB28DrossHopper: "Dross Hopper",
		bnsB28Tyrannodon: "Drowsing Tyrannodon",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestBnsB28KeywordCreaturesHaveTheirKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, oracle string
		kws          []string
	}{
		{"Daybreak Chaplain", bnsB28DaybreakChaplain, []string{"lifelink"}},
		{"Deepwood Denizen", bnsB28DeepwoodDenizen, []string{"vigilance"}},
		{"Devouring Swarm", bnsB28DevouringSwarm, []string{"flying"}},
		{"Dragon Hatchling", bnsB28DragonHatchling, []string{"flying"}},
		{"Dreadhorde Arcanist", bnsB28DreadhordeArc, []string{"trample"}},
		{"Drowsing Tyrannodon", bnsB28Tyrannodon, []string{"defender"}},
	} {
		id := effectstest.B12Push(g, me.ID, tc.name, "Creature — Test", tc.oracle, 1, 1)
		for _, kw := range tc.kws {
			if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s lacks %s", tc.name, kw)
			}
		}
	}
}

func TestBnsB28DireFleetHoarderLeavesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Dire Fleet Hoarder", "Creature — Human Pirate", bnsB28DireFleetHoarder, 2, 1)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Treasure" {
			found = true
		}
	}
	if !found {
		t.Error("no Treasure after it died")
	}
}

func TestBnsB28DragonHatchlingFirebreathes(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	id := effectstest.B12Push(g, me.ID, "Dragon Hatchling", "Creature — Dragon", bnsB28DragonHatchling, 0, 1)
	effectstest.ApaMana(me, "R", "R")
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Strict: true})
	if p := effectstest.EffectivePower(t, g, id); p != 2 {
		t.Errorf("power %d, want 2", p)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with no red mana left")
	}
}

func TestBnsB28DevouringSwarmEatsACreature(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	id := effectstest.B12Push(g, me.ID, "Devouring Swarm", "Creature — Insect", bnsB28DevouringSwarm, 2, 1)
	food := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{food}})
	if g.Battlefield.Contains(food) {
		t.Error("the creature was not sacrificed")
	}
	if p, tg := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 3 || tg != 2 {
		t.Errorf("Swarm is %d/%d, want 3/2", p, tg)
	}
}

func TestBnsB28DrossHopperGainsFlying(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	id := effectstest.B12Push(g, me.ID, "Dross Hopper", "Creature — Phyrexian Insect Horror", bnsB28DrossHopper, 2, 1)
	food := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	if effectstest.HasKeywordOnBattlefield(t, g, id, "flying") {
		t.Fatal("flying before the ability")
	}
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{food}})
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("no flying after sacrificing a creature")
	}
}

func TestBnsB28DestructiveDiggerSacrificesAnArtifactOrLandToDraw(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	id := effectstest.B12Push(g, me.ID, "Destructive Digger", "Creature — Goblin", bnsB28DestructiveDig, 3, 2)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.PutOnTopForTest(me, "Top Card", "Sorcery")
	effectstest.ApaMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a plain creature paid the cost")
	}
	hand := me.Hand.Size()
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{rock}})
	if g.Battlefield.Contains(rock) || me.Hand.Size() != hand+1 {
		t.Errorf("rock on battlefield=%v, hand %d want %d", g.Battlefield.Contains(rock), me.Hand.Size(), hand+1)
	}
}

func TestBnsB28DeepwoodDenizenCostsLessPerCounter(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	id := effectstest.B12Push(g, me.ID, "Deepwood Denizen", "Creature — Elf Warrior", bnsB28DeepwoodDenizen, 3, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AddCountersForTest(g, id, "+1/+1", 1)
	effectstest.AddCountersForTest(g, other, "+1/+1", 2)
	effectstest.AddCountersForTest(g, theirs, "+1/+1", 5) // not yours: doesn't count
	effectstest.PutOnTopForTest(me, "Top Card", "Sorcery")
	effectstest.ApaMana(me, "G", "C", "C") // {5}{G} less 3 = {2}{G}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activation with three counters for {2}{G}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The same pool is not enough with no counters on board.
	g2, me2, _ := effectstest.SpendTable(t)
	id2 := effectstest.B12Push(g2, me2.ID, "Deepwood Denizen", "Creature — Elf Warrior", bnsB28DeepwoodDenizen, 3, 2)
	effectstest.ApaMana(me2, "G", "C", "C")
	if err := g2.ActivateCatalogAbility(me2.ID, id2, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated for {2}{G} with no counters")
	}
}

func TestBnsB28DrowsingTyrannodonAttacksOnlyWithAPowerFourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.B12Push(g, me.ID, "Drowsing Tyrannodon", "Creature — Dinosaur", bnsB28Tyrannodon, 3, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, opp.ID); err == nil {
		t.Fatal("a defender attacked with no power-4 creature")
	}
	effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 4, 4)
	if err := g.DeclareAttacker(id, opp.ID); err != nil {
		t.Fatalf("could not attack with a power-4 creature: %v", err)
	}
}

func TestBnsB28DarkSupplicantSacrificesThreeClerics(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	id := effectstest.B12Push(g, me.ID, "Dark Supplicant", "Creature — Human Cleric", bnsB28DarkSupplicant, 1, 1)
	c1 := effectstest.B12Creature(g, me.ID, "Cleric A", "Creature — Human Cleric", 1, 1)
	c2 := effectstest.B12Creature(g, me.ID, "Cleric B", "Creature — Human Cleric", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	scion := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: scion, Name: "Scion of Darkness", TypeLine: "Creature — Avatar", Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{c1, c2, bear}}); err == nil {
		t.Fatal("a non-Cleric paid the cost")
	}
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{id, c1, c2}})
	effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneLibrary)
	effectstest.AnswerSearchByID(t, g, me.ID, scion)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(scion) {
		t.Error("Scion of Darkness did not come from the library")
	}
	if g.Battlefield.Contains(c1) || g.Battlefield.Contains(c2) {
		t.Error("the Clerics were not sacrificed")
	}
}

func TestBnsB28DraugrThoughtThiefMillsTheTopCardOnYes(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		top := effectstest.SeedLibraryTop(opp, "Their Top", "Sorcery", "{1}")
		effectstest.CastHoldCreature(t, g, "Draugr Thought-Thief", "Creature — Zombie Rogue", bnsB28DraugrThief, 3, 2)
		effectstest.PickPlayer(t, g, me.ID, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerConfirmFor(t, g, me, yes)
		inGY := opp.Graveyard.Contains(top)
		if inGY != yes {
			t.Errorf("answer %v: card in their graveyard = %v", yes, inGY)
		}
	}
}

func TestBnsB28DreadhordeArcanistCastsASmallSpellFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.B12Push(g, me.ID, "Dreadhorde Arcanist", "Creature — Zombie Wizard", bnsB28DreadhordeArc, 1, 3)
	cheap := effectstest.GyCard(me, "Cheap Sorcery", "Sorcery", "{1}", "")
	dear := effectstest.GyCard(me, "Dear Sorcery", "Sorcery", "{2}", "")
	creature := effectstest.GyCard(me, "Old Bear", "Creature — Bear", "{1}", "")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if !effectstest.HasID(p.PickTargetCards, cheap) {
		t.Error("the mana value 1 sorcery was not offered")
	}
	for _, bad := range []uuid.UUID{dear, creature} {
		if effectstest.HasID(p.PickTargetCards, bad) {
			t.Errorf("illegal card offered: %v", bad)
		}
	}
	effectstest.PickCard(t, g, me.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, cheap)
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, cheap, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("casting for free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(cheap) || !effectstest.InExile(g, cheap) {
		t.Error("the sorcery should be exiled, not back in the graveyard")
	}
}
