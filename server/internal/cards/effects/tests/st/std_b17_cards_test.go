package testsst

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b17_cards_test.go — Standard set leftovers, batches 17 and 18
// (Foundations, #4465, umbrella #4452). Each test casts, activates or
// attacks through the real engine.

const (
	stdResoluteReinforcementsOracle = "eaaed253-9b11-47e2-8624-0340ab2207f6"
	stdAegisTurtleOracle            = "f84c30e7-2ea8-43ff-9356-1f907558cfd9"
	stdBrazenScourgeOracle          = "422ef4ec-863a-450e-9577-a0707bc16532"
	stdBishopsSoldierOracle         = "84487586-4f71-457a-b983-4f1f7a71deef"
	stdLeoninSkyhunterOracle        = "517c7295-8ba2-47a6-a1c6-aee722f8ca88"
	stdWitnessProtectionOracle      = "a5f5e71e-83a7-468d-829a-6392b764c76e"
	stdSkyknightSquireOracle        = "7e497f6b-69e1-4fab-bb04-f687c99a8310"
	stdElendaOracle                 = "bfd614d5-4c7f-416b-9cb2-bb167c0893a0"
	stdLunarInsightOracle           = "56b4a2f7-2d35-4d05-8d26-4f9f16c5f9f5"
	stdHidetsugusSecondRiteOracle   = "749683c9-8cc0-4e8e-8d90-adda5e5f4e93"
	stdGiadaOracle                  = "48e6d3d8-2f27-4017-acdd-40bce8cdbc02"
	stdExtravagantReplicationOracle = "3a646245-b8b7-4f91-a312-d5eea9a9e49a"
	stdNineLivesFamiliarOracle      = "5049da27-f38b-4fe6-a387-51510bd1f44e"
	stdDrakeHatcherOracle           = "4de0ccab-bb5d-4c7c-827e-36aa1d00b182"
	stdFlamewakePhoenixOracle       = "9e43805d-3576-47d4-9331-524869ab99d0"
	stdKellanOracle                 = "e8fd428f-7521-4e6a-86b7-835f20ee3c5c"
	stdZimoneOracle                 = "9dd674a7-becf-4106-b53f-bca88426d92d"
	stdChandraFlameshaperOracle     = "818f908c-9af6-4edc-bb0e-193bc830e2e4"
	stdVivienReidOracle             = "b5f20a6d-8c3e-452f-9d98-4886f0fa052a"
	stdAbyssalHarvesterOracle       = "68cc5065-0514-4adf-885f-da554e1db00a"
)

func TestStdB17KeywordAndVanillaCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, k := range []struct{ name, typeLine, oracle, kw string }{
		{"Brazen Scourge", "Creature — Gremlin", stdBrazenScourgeOracle, "haste"},
		{"Bishop's Soldier", "Creature — Vampire Soldier", stdBishopsSoldierOracle, "lifelink"},
		{"Leonin Skyhunter", "Creature — Cat Knight", stdLeoninSkyhunterOracle, "flying"},
	} {
		effectstest.AssertKeywords(t, g, effectstest.B12Push(g, me.ID, k.name, k.typeLine, k.oracle, 2, 2), k.kw)
	}
	id := effectstest.B12Push(g, me.ID, "Aegis Turtle", "Creature — Turtle", stdAegisTurtleOracle, 0, 5)
	if c := effectstest.LayeredCard(t, g, id); len(c.Keywords) != 0 {
		t.Errorf("Aegis Turtle is vanilla but has %v", c.Keywords)
	}
	for _, o := range []string{stdAegisTurtleOracle, stdBrazenScourgeOracle, stdBishopsSoldierOracle, stdLeoninSkyhunterOracle} {
		if spec, ok := effects.Lookup(o); !ok || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s is not registered Full", o)
		}
	}
}

func TestStdB17ResoluteReinforcementsMakesASoldierAndHasFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.AvB1ResolveCreature(t, g, me, "Resolute Reinforcements", "Creature — Human Soldier", stdResoluteReinforcementsOracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, id, "flash")
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Soldier" && c.IsToken() {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d Soldier tokens, want 1", n)
	}
}

// --- Witness Protection --------------------------------------------------------------

func TestStdB17WitnessProtectionMakesALegitimateBusinessperson(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vehicle := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big Rig", TypeLine: "Legendary Artifact Creature — Vehicle Dragon",
		Keywords: []string{"flying"}, Colors: []string{"R"}, Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{game.CounterPlusOne: 2},
	})
	effectstest.AuraCast(t, g, "Witness Protection", stdWitnessProtectionOracle, vehicle)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.EffectiveCardForTest(t, g, vehicle)
	eff := c.Effective()
	if eff.Name != "Legitimate Businessperson" {
		t.Errorf("name %q", eff.Name)
	}
	if p, th := c.CurrentPower(), c.CurrentToughness(); p != 3 || th != 3 {
		t.Errorf("P/T %d/%d, want base 1/1 plus two counters = 3/3", p, th)
	}
	if subs := effectstest.EffectiveSubtypes(t, g, vehicle); len(subs) != 1 || subs[0] != "Citizen" {
		t.Errorf("subtypes %v, want only Citizen", subs)
	}
	if types := effectstest.EffectiveTypes(t, g, vehicle); len(types) != 1 || types[0] != "Creature" {
		t.Errorf("types %v, want only Creature", types)
	}
	if len(eff.Colors) != 2 || !effectstest.ContainsString(eff.Colors, "G") || !effectstest.ContainsString(eff.Colors, "W") {
		t.Errorf("colors %v, want green and white", eff.Colors)
	}
	if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, vehicle), "flying") {
		t.Error("kept flying")
	}
}

// --- Skyknight Squire ------------------------------------------------------------------

func TestStdB17SkyknightSquireGrowsAndFliesAtThreeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	squire := effectstest.B12Push(g, me.ID, "Skyknight Squire", "Creature — Cat Scout", stdSkyknightSquireOracle, 1, 1)
	for i := 1; i <= 3; i++ {
		g.WithWriteLock(func() {
			_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1}, 1)
		})
		effectstest.PassPriorityAroundTable(t, g)
		if i < 3 && effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, squire), "flying") {
			t.Fatalf("flying with %d counters", effectstest.CountersOn(g, squire, game.CounterPlusOne))
		}
	}
	if n := effectstest.CountersOn(g, squire, game.CounterPlusOne); n != 3 {
		t.Fatalf("%d counters, want 3", n)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, squire), "flying") {
		t.Error("no flying at three counters")
	}
	if subs := effectstest.EffectiveSubtypes(t, g, squire); !effectstest.ContainsString(subs, "Knight") || !effectstest.ContainsString(subs, "Cat") || !effectstest.ContainsString(subs, "Scout") {
		t.Errorf("subtypes %v, want Cat Scout Knight", subs)
	}
}

func TestStdB17SkyknightSquireIgnoresOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	squire := effectstest.B12Push(g, me.ID, "Skyknight Squire", "Creature — Cat Scout", stdSkyknightSquireOracle, 1, 1)
	effectstest.AvB1OpponentCreatureEnters(t, g, opp.ID, "Creature — Bear")
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, squire, game.CounterPlusOne); n != 0 {
		t.Errorf("%d counters from an opponent's creature", n)
	}
}

// --- Elenda, Saint of Dusk -------------------------------------------------------------

func TestStdB17ElendaGrowsWithYourLifeAboveStarting(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elenda := effectstest.B12Push(g, me.ID, "Elenda, Saint of Dusk", "Legendary Creature — Vampire Knight", stdElendaOracle, 4, 4)
	start := g.Settings.StartingLife
	check := func(life, power int, menace bool) {
		t.Helper()
		g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, life-me.Life) })
		if got := effectstest.EffectivePower(t, g, elenda); got != power {
			t.Errorf("life %d: power %d, want %d", life, got, power)
		}
		if got := effectstest.EffectiveToughness(t, g, elenda); got != power {
			t.Errorf("life %d: toughness %d, want %d", life, got, power)
		}
		if got := effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, elenda), "menace"); got != menace {
			t.Errorf("life %d: menace %v, want %v", life, got, menace)
		}
	}
	check(start, 4, false)
	check(start+1, 5, true)
	check(start+9, 5, true)
	check(start+10, 10, true)
	check(start-1, 4, false)
	effectstest.AssertKeywords(t, g, elenda, "lifelink", "hexproof from instants")
}

func TestStdB17ElendaIsHexproofFromInstantsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elenda := effectstest.B12Push(g, me.ID, "Elenda, Saint of Dusk", "Legendary Creature — Vampire Knight", stdElendaOracle, 4, 4)
	sorcery := &game.Characteristic{Colors: []string{"R"}, Types: []string{"Sorcery"}}
	if effectstest.CreatureTargetableFrom(g, opp.ID, elenda, effectstest.SrcRed) {
		t.Error("an opponent's instant targeted Elenda")
	}
	if !effectstest.CreatureTargetableFrom(g, opp.ID, elenda, sorcery) {
		t.Error("an opponent's sorcery should target Elenda")
	}
	if !effectstest.CreatureTargetableFrom(g, me.ID, elenda, effectstest.SrcRed) {
		t.Error("its controller's own instant can target it")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, elenda, "hexproof") {
		t.Error("hexproof from instants is not plain hexproof")
	}
}

// --- Lunar Insight ------------------------------------------------------------------------

func TestStdB17LunarInsightDrawsOnePerDifferentManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for _, c := range []game.Card{
		{Name: "One A", TypeLine: "Creature — Bear", ManaCost: "{G}"},
		{Name: "One B", TypeLine: "Creature — Bear", ManaCost: "{W}"},
		{Name: "Three", TypeLine: "Artifact", ManaCost: "{3}"},
		{Name: "X Spell", TypeLine: "Enchantment", ManaCost: "{X}{G}"},
		{Name: "Land", TypeLine: "Land"},
		{Name: "Token", TypeLine: "Token Creature — Bear"},
	} {
		c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
		effectstest.PushBattlefieldCardWithTimestamp(g, c)
	}
	// Theirs does not count.
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Artifact", ManaCost: "{6}", Owner: opp.ID, Controller: opp.ID})
	for i := 0; i < 8; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	}
	// Values: 1 (two cards), 3, 1 ({X}{G}, X is 0), and 0 (the token): {0, 1, 3} = 3.
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Lunar Insight", "Sorcery", stdLunarInsightOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3 (mana values 0, 1 and 3)", got)
	}
}

// --- Hidetsugu's Second Rite ----------------------------------------------------------

func TestStdB17HidetsugusSecondRiteDealsTenToAPlayerAtExactlyTen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	g.WithWriteLock(func() { opp.Life = 10 })
	effectstest.CastCatalogSpell(t, g, "Hidetsugu's Second Rite", "Instant", stdHidetsugusSecondRiteOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 0 {
		t.Errorf("life %d, want 0", opp.Life)
	}
}

func TestStdB17HidetsugusSecondRiteDoesNothingAtAnyOtherTotal(t *testing.T) {
	for _, life := range []int{9, 11, 20} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		g.WithWriteLock(func() { opp.Life = life })
		effectstest.CastCatalogSpell(t, g, "Hidetsugu's Second Rite", "Instant", stdHidetsugusSecondRiteOracle,
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Life != life {
			t.Errorf("life %d became %d, want unchanged", life, opp.Life)
		}
	}
}

// --- Giada, Font of Hope -------------------------------------------------------------

func TestStdB17GiadaPutsACounterPerAngelAlreadyControlled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Giada, Font of Hope", "Legendary Creature — Angel", stdGiadaOracle, 2, 2)
	// One Angel (Giada herself) controlled: the next Angel gets one counter.
	first := effectstest.AvB1ResolveCreature(t, g, me, "Angel One", "Creature — Angel", "")
	if n := effectstest.CountersOn(g, first, game.CounterPlusOne); n != 1 {
		t.Errorf("first Angel entered with %d counters, want 1 (Giada)", n)
	}
	// Two Angels now plus Giada: three already controlled.
	second := effectstest.AvB1ResolveCreature(t, g, me, "Angel Two", "Creature — Angel", "")
	if n := effectstest.CountersOn(g, second, game.CounterPlusOne); n != 2 {
		t.Errorf("second Angel entered with %d counters, want 2 (Giada and the first)", n)
	}
	bear := effectstest.AvB1ResolveCreature(t, g, me, "Bear", "Creature — Bear", "")
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 0 {
		t.Errorf("a non-Angel entered with %d counters", n)
	}
}

func TestStdB17GiadaIgnoresAnOpponentsAngels(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Giada, Font of Hope", "Legendary Creature — Angel", stdGiadaOracle, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Angel", "Creature — Angel", 2, 2)
	_ = theirs
	first := effectstest.AvB1ResolveCreature(t, g, me, "Angel One", "Creature — Angel", "")
	if n := effectstest.CountersOn(g, first, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters, want 1: only Giada counts", n)
	}
}

func TestStdB17GiadaManaPaysOnlyForAnAngelSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	giada := effectstest.B12Push(g, me.ID, "Giada, Font of Hope", "Legendary Creature — Angel", stdGiadaOracle, 2, 2)
	if _, err := effectstest.CastFromHandErr(g, me, "Bear", "Creature — Bear", "{W}", "", game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("setup: a {W} Bear castable with an empty pool")
	}
	if err := g.ActivateManaAbility(me.ID, giada, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tapping Giada for mana: %v", err)
	}
	if _, err := effectstest.CastFromHandErr(g, me, "Bear", "Creature — Bear", "{W}", "", game.CastSpellParams{Strict: true}); err == nil {
		t.Error("Giada's mana paid for a non-Angel spell")
	}
	if _, err := effectstest.CastFromHandErr(g, me, "Angel", "Creature — Angel", "{W}", "", game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("Giada's mana should pay for an Angel spell: %v", err)
	}
}

// --- Extravagant Replication -----------------------------------------------------------

func TestStdB17ExtravagantReplicationCopiesAnotherNonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	repl := effectstest.PushCatalogPermanent(g, me.ID, "Extravagant Replication", "Enchantment", stdExtravagantReplicationOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bear", "Creature — Bear", 2, 2)
	effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		for _, bad := range []uuid.UUID{repl} {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
				t.Error("the enchantment copied itself")
			}
		}
		effectstest.AvB34Pick(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B30TokensNamed(g, me.ID, "Grizzly Bear"); n != 1 {
		t.Errorf("%d Grizzly Bear tokens, want 1", n)
	}
}

// --- Nine-Lives Familiar ----------------------------------------------------------------

func TestStdB17NineLivesFamiliarEntersWithEightRevivalCountersIfCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.AvB1ResolveCreature(t, g, me, "Nine-Lives Familiar", "Creature — Cat", stdNineLivesFamiliarOracle)
	if n := effectstest.CountersOn(g, id, "revival"); n != 8 {
		t.Errorf("%d revival counters, want 8", n)
	}
}

func TestStdB17NineLivesFamiliarNotCastHasNoCountersAndStaysDead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Nine-Lives Familiar", "Creature — Cat", stdNineLivesFamiliarOracle, false)
	if n := effectstest.CountersOn(g, id, "revival"); n != 0 {
		t.Fatalf("a Familiar that was not cast has %d counters", n)
	}
	effectstest.KillCreature(t, g, me.ID, id)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("a Familiar with no revival counter came back")
	}
}

func TestStdB17NineLivesFamiliarReturnsAtTheNextEndStepWithOneFewer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.AvB1ResolveCreature(t, g, me, "Nine-Lives Familiar", "Creature — Cat", stdNineLivesFamiliarOracle)
	effectstest.KillCreature(t, g, me.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Fatal("it returned at once, not at the beginning of the next end step")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("it did not return at the beginning of the next end step")
	}
	if n := effectstest.CountersOn(g, id, "revival"); n != 7 {
		t.Errorf("%d revival counters, want 7", n)
	}
}

func TestStdB17NineLivesFamiliarsLastLifeIsTheLast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Nine-Lives Familiar", TypeLine: "Creature — Cat", OracleID: stdNineLivesFamiliarOracle,
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID, Counters: map[string]int{"revival": 1},
	})
	effectstest.KillCreature(t, g, me.ID, id)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("it did not use its last revival counter")
	}
	if n := effectstest.CountersOn(g, id, "revival"); n != 0 {
		t.Errorf("%d revival counters, want 0", n)
	}
	effectstest.KillCreature(t, g, me.ID, id)
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("a Familiar with no counter scheduled %d delayed triggers", len(g.DelayedTriggers))
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("it came back with no counter")
	}
}

func TestStdB17NineLivesFamiliarDoesNotReturnFromExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.AvB1ResolveCreature(t, g, me, "Nine-Lives Familiar", "Creature — Cat", stdNineLivesFamiliarOracle)
	effectstest.KillCreature(t, g, me.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	// The card leaves the graveyard before the end step: it is a new object.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneExile}, id); err != nil {
		t.Fatalf("exile from the graveyard: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("an exiled Familiar came back")
	}
}

// --- Drake Hatcher ---------------------------------------------------------------------------

func TestStdB17DrakeHatcherCountsDamageAndMakesADrake(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hatcher := effectstest.PushCatalogPermanent(g, me.ID, "Drake Hatcher", "Creature — Human Wizard", stdDrakeHatcherOracle, false)
	effectstest.AssertKeywords(t, g, hatcher, "vigilance", "prowess")

	if err := g.ActivateCatalogAbility(me.ID, hatcher, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no incubation counters")
	}
	effectstest.DealCombatDamageToPlayer(g, hatcher, opp.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, hatcher, "incubation"); n != 4 {
		t.Fatalf("%d incubation counters, want 4 (the damage dealt)", n)
	}
	if err := g.ActivateCatalogAbility(me.ID, hatcher, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, hatcher, "incubation"); n != 1 {
		t.Errorf("%d incubation counters left, want 1", n)
	}
	drakes := effectstest.AvB34Tokens(g, me.ID, "Drake")
	if len(drakes) != 1 {
		t.Fatalf("%d Drake tokens, want 1", len(drakes))
	}
	if !slices.Contains(effectstest.EffectiveAbilities(t, g, drakes[0].InstanceID), "flying") {
		t.Error("the Drake does not fly")
	}
	if p, th := effectstest.EffectivePower(t, g, drakes[0].InstanceID), effectstest.EffectiveToughness(t, g, drakes[0].InstanceID); p != 2 || th != 2 {
		t.Errorf("the Drake is %d/%d, want 2/2", p, th)
	}
	if err := g.ActivateCatalogAbility(me.ID, hatcher, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with one counter left")
	}
}

// --- Flamewake Phoenix ---------------------------------------------------------------------

// phoenixInGraveyard seeds the Phoenix in p's graveyard as a card that was
// put there this turn.
func phoenixInGraveyard(g *game.Game, p *game.Player) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Graveyard.PushTop(game.Card{
			InstanceID: id, Name: "Flamewake Phoenix", TypeLine: "Creature — Phoenix", OracleID: stdFlamewakePhoenixOracle,
			Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID,
		})
	})
	return id
}

func TestStdB17FlamewakePhoenixHasFlyingHasteAndMustAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	phoenix := effectstest.B12Push(g, me.ID, "Flamewake Phoenix", "Creature — Phoenix", stdFlamewakePhoenixOracle, 2, 2)
	effectstest.AssertKeywords(t, g, phoenix, "flying", "haste")
	var reqs []game.AttackRequirement
	g.ReadSnapshot(func() {
		if c, ok := effectstest.BattlefieldCard(g, phoenix); ok {
			reqs = c.Effective().AttackRequirements
		}
	})
	if len(reqs) == 0 {
		t.Error("the Phoenix carries no attack requirement")
	}
}

func TestStdB17FlamewakePhoenixReturnsForRWithAPowerFourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Creature(g, me.ID, "Big Bear", "Creature — Bear", 4, 4)
	phoenix := phoenixInGraveyard(g, me)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FillPoolColored(me, "R", 1) // the pool empties as each step ends
	effectstest.PassPriorityAroundTable(t, g)
	pay := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
	if pay == nil {
		t.Fatal("the Phoenix did not offer the {R}")
	}
	if err := g.ResolvePayUnless(pay.ID, me.ID, true); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if !g.Battlefield.Contains(phoenix) {
		t.Error("the Phoenix did not return")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("the {R} was not spent: %d left", me.ManaPool.Total())
	}
}

func TestStdB17FlamewakePhoenixStaysIfYouDeclineOrCannotPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Creature(g, me.ID, "Big Bear", "Creature — Bear", 4, 4)
	phoenix := phoenixInGraveyard(g, me)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FillPoolColored(me, "R", 1) // the pool empties as each step ends
	effectstest.PassPriorityAroundTable(t, g)
	pay := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
	if pay == nil {
		t.Fatal("the Phoenix did not offer the {R}")
	}
	if err := g.ResolvePayUnless(pay.ID, me.ID, false); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if g.Battlefield.Contains(phoenix) {
		t.Error("the Phoenix returned without payment")
	}
}

func TestStdB17FlamewakePhoenixNeedsAPowerFourCreatureWhenCombatBegins(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Creature(g, me.ID, "Small Bear", "Creature — Bear", 3, 3)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Creature(g, opp.ID, "Their Big Bear", "Creature — Bear", 6, 6)
	phoenix := phoenixInGraveyard(g, me)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FillPoolColored(me, "R", 1) // the pool empties as each step ends
	effectstest.PassPriorityAroundTable(t, g)
	if pay := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless); pay != nil {
		t.Fatal("the Phoenix triggered with no power 4 creature of yours")
	}
	if g.Battlefield.Contains(phoenix) {
		t.Error("the Phoenix returned")
	}
}

func TestStdB17FlamewakePhoenixTheCreatureLeavingInResponseRemovesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	big := effectstest.B12Creature(g, me.ID, "Big Bear", "Creature — Bear", 4, 4)
	phoenix := phoenixInGraveyard(g, me)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FillPoolColored(me, "R", 1) // the pool empties as each step ends
	// The trigger is on the stack; the only power 4 creature dies in response.
	effectstest.KillCreature(t, g, me.ID, big)
	effectstest.PassPriorityAroundTable(t, g)
	if pay := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless); pay != nil {
		t.Error("the {R} was offered although the intervening if no longer held")
	}
	if g.Battlefield.Contains(phoenix) {
		t.Error("the Phoenix returned")
	}
}

// --- Kellan, Planar Trailblazer ------------------------------------------------------------

func TestStdB17KellanClimbsHisLadder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kellan := effectstest.B12Push(g, me.ID, "Kellan, Planar Trailblazer", "Legendary Creature — Human Faerie Scout", stdKellanOracle, 2, 1)
	activate := func(idx int, r int) {
		t.Helper()
		effectstest.FillPoolColored(me, "R", r)
		if err := g.ActivateCatalogAbility(me.ID, kellan, idx, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activate %d: %v", idx, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	subs := func() []string { return effectstest.EffectiveSubtypes(t, g, kellan) }
	want := func(label string, types ...string) {
		t.Helper()
		got := subs()
		if len(got) != len(types) {
			t.Fatalf("%s: subtypes %v, want %v", label, got, types)
		}
		for _, ty := range types {
			if !slices.Contains(got, ty) {
				t.Fatalf("%s: subtypes %v, want %v", label, got, types)
			}
		}
	}

	// A Scout: the second ability does nothing.
	activate(1, 3)
	want("{2}{R} on a Scout", "Human", "Faerie", "Scout")
	if p := effectstest.EffectivePower(t, g, kellan); p != 2 {
		t.Errorf("power %d after a no-op, want 2", p)
	}
	// The first ability makes a Detective, and only once is anything left to do.
	activate(0, 2)
	want("{1}{R}", "Human", "Faerie", "Detective")
	activate(0, 2)
	want("{1}{R} on a Detective", "Human", "Faerie", "Detective")

	// The granted trigger: combat damage exiles the top card, playable this turn.
	top := uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: top, Name: "Top Card", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	})
	effectstest.DealCombatDamageToPlayer(g, kellan, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Error("the top card was not exiled")
	}
	if len(me.CastPermissions) == 0 {
		t.Error("no permission to play the exiled card")
	}

	// The second ability: a 3/2 Human Faerie Rogue with double strike.
	activate(1, 3)
	want("{2}{R}", "Human", "Faerie", "Rogue")
	if p, th := effectstest.EffectivePower(t, g, kellan), effectstest.EffectiveToughness(t, g, kellan); p != 3 || th != 2 {
		t.Errorf("%d/%d, want 3/2", p, th)
	}
	effectstest.AssertKeywords(t, g, kellan, "double strike")
	// It keeps the granted trigger (no duration): a Rogue still exiles on damage.
	if got := g.Exile.Size(); got < 1 {
		t.Errorf("exile holds %d", got)
	}
}

// --- Zimone, Paradox Sculptor ---------------------------------------------------------------

func TestStdB17ZimoneGrowsUpToTwoCreaturesAtBeginningOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Zimone, Paradox Sculptor", "Legendary Creature — Human Wizard", stdZimoneOracle, 1, 4)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt at the beginning of combat")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err == nil {
		t.Error("an opponent's creature was accepted")
	}
	effectstest.AvB34Pick(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != 1 {
			t.Errorf("%s has %d counters, want 1", id, n)
		}
	}
	if effectstest.CountersOn(g, theirs, game.CounterPlusOne) != 0 {
		t.Error("an opponent's creature got a counter")
	}
}

func TestStdB17ZimoneCanChooseNoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	zimone := effectstest.B12Push(g, me.ID, "Zimone, Paradox Sculptor", "Legendary Creature — Human Wizard", stdZimoneOracle, 1, 4)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("choosing no targets: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, zimone, game.CounterPlusOne); n != 0 {
		t.Errorf("Zimone got %d counters with no target chosen", n)
	}
}

func TestStdB17ZimoneDoublesEachKindOfCounterOnUpToTwoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	zimone := effectstest.B12Push(g, me.ID, "Zimone, Paradox Sculptor", "Legendary Creature — Human Wizard", stdZimoneOracle, 1, 4)
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 2, "shield": 1},
	})
	relic := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Relic", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{"charge": 3},
	})
	land := effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	_ = land
	effectstest.ToMain(t, g)
	effectstest.FillPoolColored(me, "G", 1)
	effectstest.FillPoolColored(me, "U", 1)
	if err := g.ActivateCatalogAbility(me.ID, zimone, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}); err == nil {
		t.Fatal("a land was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, zimone, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: relic}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 4 {
		t.Errorf("+1/+1 counters %d, want 4", n)
	}
	if n := effectstest.CountersOn(g, bear, "shield"); n != 2 {
		t.Errorf("shield counters %d, want 2", n)
	}
	if n := effectstest.CountersOn(g, relic, "charge"); n != 6 {
		t.Errorf("charge counters %d, want 6", n)
	}
	if !effectstest.BattlefieldCardPtr(t, g, zimone).Tapped {
		t.Error("Zimone did not tap")
	}
}

// --- Chandra, Flameshaper --------------------------------------------------------------------

func TestStdB17ChandraFlameshaperPlusTwoAddsManaAndLetsYouPlayOneExiledCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	chandra := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Chandra, Flameshaper", "Legendary Planeswalker — Chandra", stdChandraFlameshaperOracle, 6)
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		for _, n := range []string{"Card C", "Card B", "Card A"} {
			id := uuid.New()
			ids = append(ids, id)
			me.Library.PushTop(game.Card{InstanceID: id, Name: n, TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, chandra, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate +2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyCount(g, chandra); got != 8 {
		t.Errorf("loyalty %d, want 8", got)
	}
	if me.ManaPool.Total() != 3 {
		t.Errorf("pool %d, want {R}{R}{R}", me.ManaPool.Total())
	}
	for _, id := range ids {
		if !g.Exile.Contains(id) {
			t.Fatalf("%s was not exiled", id)
		}
	}
	pick := effectstest.LatestChoiceOfKind(g, game.PendingChoiceChooseCards)
	if pick == nil {
		t.Fatal("no prompt to choose one of the exiled cards")
	}
	chosen := ids[1]
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{chosen}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	granted := map[uuid.UUID]bool{}
	for _, perm := range me.CastPermissions {
		for _, ref := range perm.Cards {
			granted[ref.ID] = true
		}
	}
	if !granted[chosen] || len(granted) != 1 {
		t.Errorf("play permission over %v, want only the chosen card %s", granted, chosen)
	}
}

func TestStdB17ChandraFlameshaperPlusOneMakesAHastyCopyThatIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	chandra := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Chandra, Flameshaper", "Legendary Planeswalker — Chandra", stdChandraFlameshaperOracle, 6)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, chandra, 1, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate +1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	toks := effectstest.AvB34Tokens(g, me.ID, "Grizzly Bear")
	if len(toks) != 1 {
		t.Fatalf("%d token copies, want 1", len(toks))
	}
	tok := toks[0].InstanceID
	if !slices.Contains(effectstest.EffectiveAbilities(t, g, tok), "haste") {
		t.Error("the copy has no haste")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tok) {
		t.Error("the copy survived the end step")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the original was sacrificed")
	}
}

func TestStdB17ChandraFlameshaperMinusFourDividesEightDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	chandra := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Chandra, Flameshaper", "Legendary Planeswalker — Chandra", stdChandraFlameshaperOracle, 6)
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Ogre B", "Creature — Ogre", 3, 5)
	walker := effectstest.PushCatalogWalkerWithLine(g, opp.ID, "Chandra, Flameshaper", "Legendary Planeswalker — Chandra", stdChandraFlameshaperOracle, 6)
	err := g.ActivateCatalogAbility(me.ID, chandra, 2, game.ActivateAbilityParams{
		Strict:       true,
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: walker}},
		Distribution: map[uuid.UUID]int{a: 2, b: 5, walker: 1},
	})
	if err != nil {
		t.Fatalf("activate -4: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("the Bear or the Ogre survived its share")
	}
	if got := effectstest.LoyaltyCount(g, walker); got != 5 {
		t.Errorf("the planeswalker's loyalty is %d, want 5", got)
	}
}

// --- Vivien Reid -----------------------------------------------------------------------------

func TestStdB17VivienReidPlusOneTakesACreatureOrLandAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vivien := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Vivien Reid", "Legendary Planeswalker — Vivien", stdVivienReidOracle, 5)
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		for _, c := range []game.Card{
			{Name: "Spell D", TypeLine: "Sorcery"},
			{Name: "Spell C", TypeLine: "Instant"},
			{Name: "Bear B", TypeLine: "Creature — Bear"},
			{Name: "Spell A", TypeLine: "Sorcery"},
		} {
			c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
			ids = append(ids, c.InstanceID)
			me.Library.PushTop(c)
		}
	})
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, vivien, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate +1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKind(g, game.PendingChoiceChooseCards)
	if pick == nil {
		t.Fatal("no prompt for the creature or land card")
	}
	// Only the creature card can be taken.
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[0]}); err == nil {
		t.Error("a Sorcery was accepted")
	}
	bear := ids[2]
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	if !me.Hand.Contains(bear) {
		t.Error("the Bear is not in hand")
	}
	for _, id := range []uuid.UUID{ids[0], ids[1], ids[3]} {
		if me.Library.Contains(id) {
			continue
		}
		t.Errorf("%s left the library", id)
	}
	if got := effectstest.LoyaltyCount(g, vivien); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func TestStdB17VivienReidMinusThreeDestroysAnArtifactEnchantmentOrFlyer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vivien := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Vivien Reid", "Legendary Planeswalker — Vivien", stdVivienReidOracle, 12)
	art := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	ench := effectstest.B12Permanent(g, opp.ID, "Aura Thing", "Enchantment")
	flyer := effectstest.B30Flyer(g, opp.ID, "Drake", "Creature — Drake", 2, 2)
	walker := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	activate := func(target uuid.UUID) error {
		return g.ActivateCatalogAbility(me.ID, vivien, 1, game.ActivateAbilityParams{
			Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		})
	}
	if err := activate(walker); err == nil {
		t.Fatal("a Bear without flying was a legal target")
	}
	for i, id := range []uuid.UUID{art, ench, flyer} {
		if i > 0 {
			// Loyalty to spare, and one activation per turn: reset the once-a-turn gate.
			g.WithWriteLock(func() { delete(g.LoyaltyActivatedThisTurn, vivien) })
		}
		if err := activate(id); err != nil {
			t.Fatalf("activate -3 on %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(id) {
			t.Errorf("target %d survived", i)
		}
	}
}

func TestStdB17VivienReidMinusEightMakesAnEmblemThatPumpsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vivien := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Vivien Reid", "Legendary Planeswalker — Vivien", stdVivienReidOracle, 8)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, vivien, 2, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate -8: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Emblems.Size() != 1 {
		t.Fatalf("%d emblems, want 1", me.Emblems.Size())
	}
	if p, th := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || th != 4 {
		t.Errorf("your Bear is %d/%d, want 4/4", p, th)
	}
	effectstest.AssertKeywords(t, g, bear, "vigilance", "trample", "indestructible")
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's Bear is %d power, want 2", p)
	}
}

// --- Abyssal Harvester -----------------------------------------------------------------------

func TestStdB17AbyssalHarvesterCopiesACardPutIntoAGraveyardThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	harvester := effectstest.B12Push(g, me.ID, "Abyssal Harvester", "Creature — Demon Warlock", stdAbyssalHarvesterOracle, 3, 2)
	dead := effectstest.B12Creature(g, opp.ID, "Grizzly Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, opp.ID, dead)
	older := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: older, Name: "Old Bones", TypeLine: "Creature — Skeleton", Owner: me.ID, Controller: me.ID})
	})
	// A Nightmare token you control, an opponent's, and a plain token.
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Spook", TypeLine: "Token Creature — Nightmare",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Spook", TypeLine: "Token Creature — Nightmare",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	plain := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})

	if err := g.ActivateCatalogAbility(me.ID, harvester, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: older}},
	}); err == nil {
		t.Fatal("a card put into the graveyard in an earlier turn was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, harvester, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		if z := g.FindCardZoneForEffect(dead); z != nil && z.Kind == game.ZoneGraveyard {
			t.Error("the card is still in the graveyard")
		}
	})
	copies := effectstest.AvB34Tokens(g, me.ID, "Grizzly Bear")
	if len(copies) != 1 {
		t.Fatalf("%d Grizzly Bear tokens, want 1", len(copies))
	}
	subs := effectstest.EffectiveSubtypes(t, g, copies[0].InstanceID)
	if !slices.Contains(subs, "Bear") || !slices.Contains(subs, "Nightmare") {
		t.Errorf("the copy's subtypes are %v, want Bear and Nightmare", subs)
	}
	if g.Battlefield.Contains(mine) {
		t.Error("your other Nightmare token was not exiled")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("an opponent's Nightmare token was exiled")
	}
	if !g.Battlefield.Contains(plain) {
		t.Error("a token that is not a Nightmare was exiled")
	}
	if !effectstest.BattlefieldCardPtr(t, g, harvester).Tapped {
		t.Error("the Harvester did not tap")
	}
}

func TestStdB17AbyssalHarvesterSecondCopyReplacesTheFirstNightmare(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	harvester := effectstest.B12Push(g, me.ID, "Abyssal Harvester", "Creature — Demon Warlock", stdAbyssalHarvesterOracle, 3, 2)
	first := effectstest.B12Creature(g, me.ID, "Grizzly Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, first)
	if err := g.ActivateCatalogAbility(me.ID, harvester, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: first}},
	}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	firstCopy := effectstest.AvB34Tokens(g, me.ID, "Grizzly Bear")
	if len(firstCopy) != 1 {
		t.Fatalf("%d copies after the first activation", len(firstCopy))
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == harvester {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	second := effectstest.B12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	effectstest.KillCreature(t, g, me.ID, second)
	if err := g.ActivateCatalogAbility(me.ID, harvester, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: second}},
	}); err != nil {
		t.Fatalf("second activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.AvB34Tokens(g, me.ID, "Grizzly Bear")) != 0 {
		t.Error("the first Nightmare copy was not exiled")
	}
	if len(effectstest.AvB34Tokens(g, me.ID, "Hill Giant")) != 1 {
		t.Error("the second Nightmare copy is missing")
	}
}
