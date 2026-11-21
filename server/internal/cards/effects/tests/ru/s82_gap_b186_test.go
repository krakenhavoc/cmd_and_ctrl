package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b186_test.go — Warhammer 40,000 Commander artifacts, S82 GAP
// batch 186 (#5948).

var gapB186IDs = map[string]string{
	"Convergence of Dominion":  "655aece2-f33a-46e0-8f80-112ae8a457fe",
	"Goliath Truck":            "d94144b0-8967-4904-bb6b-44b0022d1f19",
	"Inquisitorial Rosette":    "4a3a25ae-61f5-4bcd-ae3d-4b31feaa83d8",
	"Necron Monolith":          "20932bca-52be-4288-9c7d-4154650c5ec9",
	"Night Scythe":             "dea9741a-3360-4a17-aeea-a854e2174d25",
	"Resurrection Orb":         "b410f294-cbb6-4016-981f-808fd9b7bbae",
	"Sceptre of Eternal Glory": "ce08aff3-efe6-4186-9557-274e00a7ca38",
	"Thunderhawk Gunship":      "d64a4ad2-c80e-4a7f-93f2-b702916dba07",
	"Drach'Nyen":               "aaf4a72b-9606-4d8c-a286-fbce4aa08f97",
	"Bone Sabres":              "7ef45067-74bd-4df8-9f95-621761a0b4a0",
}

func TestGapB186CardsAreFull(t *testing.T) {
	for name, id := range gapB186IDs {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v spec=%q/%s/%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func gapB186Answer(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			effectstest.AnswerPickTarget(t, g, target)
			return
		}
	}
}

// attackWith declares the creatures as attackers and settles the triggers.
func gapB186Attack(t *testing.T, g *game.Game, opp *game.Player, answer uuid.UUID, ids ...uuid.UUID) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range ids {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	if answer != uuid.Nil {
		gapB186Answer(t, g, answer)
	}
	effectstest.PassAndSettleAttackSplit(t, g, g.Seats[g.Turn.ActiveSeat].ID)
}

func gapB186Vehicle(g *game.Game, me *game.Player, name string) uuid.UUID {
	id := effectstest.PushCatalogPermanent(g, me.ID, name, "Artifact Creature — Vehicle", gapB186IDs[name], false)
	effectstest.Unsick(g, id)
	return id
}

func gapB186Has(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	c := effectstest.FlipState(t, g, id)
	return game.HasKeyword(&c, kw)
}

func gapB186Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return effectstest.FlipState(t, g, id).Counters[game.CounterPlusOne]
}

func TestGoliathTruckPutsTwoCountersOnAnotherAttacker(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	truck := gapB186Vehicle(g, me, "Goliath Truck")
	bear := effectstest.B12Creature(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, 2, 2)
	effectstest.Unsick(g, bear)
	gapB186Attack(t, g, opp, bear, truck, bear)
	if n := gapB186Counters(t, g, bear); n != 2 {
		t.Fatalf("the other attacker has %d +1/+1 counters, want 2", n)
	}
	if n := gapB186Counters(t, g, truck); n != 0 {
		t.Fatalf("the Truck put counters on itself (%d)", n)
	}
}

func TestBoneSabresPutsFourCountersOnTheEquippedAttacker(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	sabres := effectstest.PushCatalogPermanent(g, me.ID, "Bone Sabres", "Artifact — Equipment", gapB186IDs["Bone Sabres"], false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", effectstest.TestCreatureTypeLine, 2, 2)
	effectstest.Unsick(g, bear)
	effectstest.Unsick(g, other)
	effectstest.AttachForTest(t, g, sabres, bear)
	gapB186Attack(t, g, opp, uuid.Nil, bear, other)
	if n := gapB186Counters(t, g, bear); n != 4 {
		t.Fatalf("the equipped attacker has %d counters, want 4", n)
	}
	if n := gapB186Counters(t, g, other); n != 0 {
		t.Fatalf("an unequipped attacker has %d counters, want 0", n)
	}
}

func TestResurrectionOrbGivesLifelinkAndReturnsTheCreatureAtEndStep(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	orb := effectstest.PushCatalogPermanent(g, me.ID, "Resurrection Orb", "Artifact — Equipment", gapB186IDs["Resurrection Orb"], false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, 2, 2)
	effectstest.AttachForTest(t, g, orb, bear)
	if !gapB186Has(t, g, bear, "lifelink") {
		t.Fatal("the equipped creature has no lifelink")
	}
	effectstest.B18Kill(t, g, bear)
	if g.Battlefield.Contains(bear) || !me.Graveyard.Contains(bear) {
		t.Fatal("the Bear did not die")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the Bear did not come back at the beginning of the next end step")
	}
}

func TestResurrectionOrbLeavesACardThatLeftTheGraveyard(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	orb := effectstest.PushCatalogPermanent(g, me.ID, "Resurrection Orb", "Artifact — Equipment", gapB186IDs["Resurrection Orb"], false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, 2, 2)
	effectstest.AttachForTest(t, g, orb, bear)
	effectstest.B18Kill(t, g, bear)
	g.WithWriteLock(func() { _ = g.ExileCardsForEffect([]uuid.UUID{bear}) })
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("an exiled card came back")
	}
}

func TestNightScytheMakesANecronWarrior(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	effectstest.CastFromHandForTest(t, g, me, "Night Scythe", "Artifact — Vehicle", "{0}", gapB186IDs["Night Scythe"], game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	w := effectstest.BattlefieldTokenNamed(t, g, "Necron Warrior")
	if w.Power != 2 || w.Toughness != 2 || !w.IsArtifact() {
		t.Fatalf("Necron Warrior = %+v", w)
	}
}

func TestThunderhawkMakesTwoAstartesAndGrantsFlyingToAttackers(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	effectstest.CastFromHandForTest(t, g, me, "Thunderhawk Gunship", "Artifact — Vehicle", "{0}", gapB186IDs["Thunderhawk Gunship"], game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Astartes Warrior" {
				n++
			}
		}
	})
	if n != 2 {
		t.Fatalf("%d Astartes Warriors, want 2", n)
	}
	ship := gapB186Vehicle(g, me, "Thunderhawk Gunship")
	bear := effectstest.B12Creature(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, 2, 2)
	effectstest.Unsick(g, bear)
	gapB186Attack(t, g, opp, uuid.Nil, ship, bear)
	if !gapB186Has(t, g, bear, "flying") {
		t.Fatal("an attacking creature did not gain flying")
	}
}

func TestInquisitorialRosetteMakesAnAttackingTokenAndGivesMenace(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	rosette := effectstest.PushCatalogPermanent(g, me.ID, "Inquisitorial Rosette", "Artifact — Equipment", gapB186IDs["Inquisitorial Rosette"], false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, 2, 2)
	effectstest.Unsick(g, bear)
	effectstest.AttachForTest(t, g, rosette, bear)
	gapB186Attack(t, g, opp, uuid.Nil, bear)
	tok := effectstest.BattlefieldTokenNamed(t, g, "Astartes Warrior")
	if tok.AttackingTarget == uuid.Nil {
		t.Fatal("the token is not attacking")
	}
	if !gapB186Has(t, g, tok.InstanceID, "menace") || !gapB186Has(t, g, bear, "menace") {
		t.Fatal("attacking creatures did not gain menace")
	}
}

func TestDrachNyenExilesACreatureAndPumpsByItsPower(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	victim := effectstest.B12Creature(g, opp.ID, "Big Bear", effectstest.TestCreatureTypeLine, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", effectstest.TestCreatureTypeLine, 2, 2)
	effectstest.CastFromHandForTest(t, g, me, "Drach'Nyen", "Legendary Artifact — Equipment", "{0}", gapB186IDs["Drach'Nyen"], game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	gapB186Answer(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(victim) {
		t.Fatal("the creature was not exiled")
	}
	var dn uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Drach'Nyen" {
				dn = c.InstanceID
			}
		}
	})
	effectstest.AttachForTest(t, g, dn, bear)
	c := effectstest.FlipState(t, g, bear)
	if c.CurrentPower() != 6 || !gapB186Has(t, g, bear, "menace") {
		t.Fatalf("equipped Bear is %d/%d menace=%v, want power 6 with menace", c.CurrentPower(), c.CurrentToughness(), gapB186Has(t, g, bear, "menace"))
	}
}

func TestNecronMonolithMillsThreeAndMakesAWarriorPerCreatureCard(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	mon := gapB186Vehicle(g, me, "Necron Monolith")
	g.WithWriteLock(func() {
		me.Library.Cards = nil
		for _, tl := range []string{"Creature — Bear", "Sorcery", "Artifact Creature — Golem", "Creature — Elf"} {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: tl, TypeLine: tl, Owner: me.ID, Controller: me.ID})
		}
	})
	gapB186Attack(t, g, opp, uuid.Nil, mon)
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Necron Warrior" {
				n++
			}
		}
	})
	if me.Graveyard.Size() != 3 || n != 2 {
		t.Fatalf("graveyard %d, warriors %d: want 3 milled and 2 Warriors (the top three were Elf, Golem, Sorcery... two creature cards)", me.Graveyard.Size(), n)
	}
}

func TestSceptreOfEternalGloryThreeManaNeedsThreeSameNamedLands(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	sceptre := effectstest.PushCatalogPermanent(g, me.ID, "Sceptre of Eternal Glory", "Legendary Artifact", gapB186IDs["Sceptre of Eternal Glory"], false)
	effectstest.Unsick(g, sceptre)
	for range 2 {
		effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	}
	if err := g.ActivateManaAbility(me.ID, sceptre, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("the three-mana ability worked with only two same-named lands")
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	if err := g.ActivateManaAbility(me.ID, sceptre, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("three same-named lands: %v", err)
	}
}

func TestConvergenceOfDominionNeedsYourCommander(t *testing.T) {
	spec, _ := effects.Lookup(gapB186IDs["Convergence of Dominion"])
	if len(spec.CostModifiers) != 1 || !spec.CostModifiers[0].Activations {
		t.Fatalf("cost modifiers = %+v, want one activation reduction", spec.CostModifiers)
	}
	g, me, _ := effectstest.FlipMain(t)
	cod := effectstest.PushCatalogPermanent(g, me.ID, "Convergence of Dominion", "Artifact", gapB186IDs["Convergence of Dominion"], false)
	effectstest.Unsick(g, cod)
	g.WithWriteLock(func() { me.Library.Cards = nil })
	for range 5 {
		g.WithWriteLock(func() {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		})
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, cod, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Translocation Protocols: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Size() != 3 {
		t.Fatalf("milled %d, want 3", me.Graveyard.Size())
	}
}
