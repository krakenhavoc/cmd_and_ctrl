package testsds

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dst_batch1_cards_test.go — Darksteel, batch 1 (#5605, S82): Aether Snap
// through Arcbound Worker.

const (
	dstAetherSnapOracle        = "1a1a16e5-ce4f-42b8-a1c6-05dd635d83cc"
	dstAetherVialOracle        = "fc148e1e-dff0-448e-9f16-625341754356"
	dstAgelessEntityOracle     = "2a93236d-f182-4826-8112-4048bb3cbaf9"
	dstAngelsFeatherOracle     = "3d0d4ba4-d5aa-4d9f-89ab-46687637fecf"
	dstArcaneSpyglassOracle    = "c2fbdc80-ad5d-4fe0-a860-3e4f452e6302"
	dstArcboundBruiserOracle   = "51577a8b-28d6-4f9e-8d3f-6a1380270fab"
	dstArcboundCrusherOracle   = "ee496b34-d8e0-482a-a2c2-3122ea376da0"
	dstArcboundFiendOracle     = "72e6cc74-0574-4421-a6b8-97d180484950"
	dstArcboundHybridOracle    = "4bd7b380-ed4b-46a1-8211-1c4b651a52c6"
	dstArcboundLancerOracle    = "fd6ffa00-170b-46d2-810a-f255de62b9d7"
	dstArcboundReclaimerOracle = "c6424a3e-dbc7-403f-867f-a81cccd17ab8"
	dstArcboundSlithOracle     = "4fa113da-44aa-4c84-8f63-b1cad63da926"
	dstArcboundStingerOracle   = "7e6e2c37-3a06-456a-ae0f-37b7dc1a3e2b"
	dstArcboundWorkerOracle    = "6d9af084-cd5d-4242-b0f0-24a36d0a75d4"
)

func TestDstBatch1CardsAreRegisteredFull(t *testing.T) {
	for name, oracle := range map[string]string{
		"Aether Snap": dstAetherSnapOracle, "Aether Vial": dstAetherVialOracle,
		"Ageless Entity": dstAgelessEntityOracle, "Angel's Feather": dstAngelsFeatherOracle,
		"Arcane Spyglass": dstArcaneSpyglassOracle, "Arcbound Bruiser": dstArcboundBruiserOracle,
		"Arcbound Crusher": dstArcboundCrusherOracle, "Arcbound Fiend": dstArcboundFiendOracle,
		"Arcbound Hybrid": dstArcboundHybridOracle, "Arcbound Lancer": dstArcboundLancerOracle,
		"Arcbound Reclaimer": dstArcboundReclaimerOracle, "Arcbound Slith": dstArcboundSlithOracle,
		"Arcbound Stinger": dstArcboundStingerOracle, "Arcbound Worker": dstArcboundWorkerOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered as %q, %v, caveats %v; want its name, Full and none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// dstHand puts a card with a mana cost, and the given colours, in a
// player's hand.
func dstHand(p *game.Player, name, typeLine, manaCost string, colors ...string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine, c.ManaCost, c.Colors = typeLine, manaCost, colors
	p.Hand.PushTop(c)
	return c.InstanceID
}

// dstToken creates one token for controller and returns its ID.
func dstToken(t *testing.T, g *game.Game, controller uuid.UUID, name, typeLine string) uuid.UUID {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(controller, game.Card{Name: name, TypeLine: typeLine, Power: 1, Toughness: 1}, 1); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	})
	return effectstest.BattlefieldIDNamed(g, controller, name)
}

// --- Aether Snap -------------------------------------------------------------

func TestDstAetherSnapRemovesEveryCounterThenExilesEveryToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, opp.ID)
	effectstest.AddCounters(t, g, bear, game.CounterPlusOne, 2)
	effectstest.AddCounters(t, g, bear, game.CounterMinusOne, 1)
	rock := effectstest.MrdPermanent(g, me.ID, "My Rock", "Artifact", "{2}", 0, 0)
	effectstest.AddCounters(t, g, rock, game.CounterCharge, 3)
	token := dstToken(t, g, opp.ID, "Their Soldier", "Token Creature — Soldier")
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.CastCatalogSpell(t, g, "Aether Snap", "Sorcery", dstAetherSnapOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, rock} {
		c, ok := effectstest.BattlefieldCard(g, id)
		if !ok {
			t.Fatalf("%s left the battlefield; only tokens are exiled", id)
		}
		for kind, n := range c.Counters {
			if n != 0 {
				t.Errorf("%s still has %d %s counters", c.Name, n, kind)
			}
		}
	}
	if g.Battlefield.Contains(token) {
		t.Error("the token should be exiled")
	}
}

// --- Aether Vial -------------------------------------------------------------

func TestDstAetherVialOffersOnlyCreaturesWithManaValueEqualToItsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	vial := effectstest.Mrd2Catalog(g, me.ID, "Aether Vial", "Artifact", dstAetherVialOracle, "{1}", 0, 0)
	effectstest.AddCounters(t, g, vial, game.CounterCharge, 2)
	two := dstHand(me, "Two Drop", "Creature — Bear", "{1}{G}")
	three := dstHand(me, "Three Drop", "Creature — Ape", "{2}{G}")
	spell := dstHand(me, "Two Spell", "Sorcery", "{1}{U}")

	if err := g.ActivateCatalogAbility(me.ID, vial, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no pick from the hand")
	}
	if !effectstest.HasID(c.ChooseCards, two) || effectstest.HasID(c.ChooseCards, three) || effectstest.HasID(c.ChooseCards, spell) {
		t.Fatalf("offered %v; want only the mana value 2 creature card", c.ChooseCards)
	}
	if c.ChooseMin != 0 {
		t.Error("the put is a \"you may\": declining must be allowed")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, two)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(two) {
		t.Error("the chosen creature card should be on the battlefield")
	}
}

func TestDstAetherVialMayAddACounterAtYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	vial := effectstest.Mrd2Catalog(g, me.ID, "Aether Vial", "Artifact", dstAetherVialOracle, "{1}", 0, 0)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if effectstest.TriggerOnStack(g, vial) == nil {
		t.Fatal("no trigger at its controller's upkeep")
	}
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, true)
	if got := effectstest.CountersOn(g, vial, game.CounterCharge); got != 1 {
		t.Fatalf("accepted: %d charge counters, want 1", got)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, false)
	if got := effectstest.CountersOn(g, vial, game.CounterCharge); got != 1 {
		t.Errorf("declined: %d charge counters, want still 1", got)
	}
}

// --- Ageless Entity ----------------------------------------------------------

func TestDstAgelessEntityGrowsByTheLifeGained(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	entity := effectstest.Mrd2Catalog(g, me.ID, "Ageless Entity", "Creature — Elemental", dstAgelessEntityOracle, "{3}{G}{G}", 4, 4)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3); err != nil {
			t.Fatalf("gain: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, entity, game.CounterPlusOne); got != 3 {
		t.Fatalf("gained 3: %d counters, want 3", got)
	}
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -2); err != nil {
			t.Fatalf("lose: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, entity, game.CounterPlusOne); got != 3 {
		t.Errorf("a life loss added counters: %d, want 3", got)
	}
}

// --- Angel's Feather ---------------------------------------------------------

func TestDstAngelsFeatherGainsOnAnyPlayersWhiteSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.Mrd2Catalog(g, opp.ID, "Angel's Feather", "Artifact", dstAngelsFeatherOracle, "{2}", 0, 0)
	life := opp.Life

	red := dstHand(me, "Red Thing", "Sorcery", "{R}", "R")
	if err := g.CastSpell(me.ID, red, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast red: %v", err)
	}
	if effectstest.OwedMays(g, opp.ID) != 0 || effectstest.TriggerOnStackFrom(g, "Angel's Feather") {
		t.Error("a red spell triggered the Feather")
	}
	effectstest.PassPriorityAroundTable(t, g)

	white := dstHand(me, "White Thing", "Sorcery", "{W}", "W")
	if err := g.CastSpell(me.ID, white, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast white: %v", err)
	}
	effectstest.B06AnswerAllTriggerPrompts(t, g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life+1 {
		t.Errorf("an opponent's white spell: the Feather's controller has %d life, want %d", opp.Life, life+1)
	}
}

// --- Arcane Spyglass ---------------------------------------------------------

func TestDstArcaneSpyglassDrawsAndBanksThenCashesInThreeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	glass := effectstest.Mrd2Catalog(g, me.ID, "Arcane Spyglass", "Artifact", dstArcaneSpyglassOracle, "{4}", 0, 0)
	land := effectstest.MrdPermanent(g, me.ID, "Island", "Basic Land — Island", "", 0, 0)
	hand := me.Hand.Size()

	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, glass, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate the first ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land should be sacrificed")
	}
	if me.Hand.Size() != hand+1 || effectstest.CountersOn(g, glass, game.CounterCharge) != 1 {
		t.Fatalf("hand %d → %d, %d charge counters; want one card and one counter",
			hand, me.Hand.Size(), effectstest.CountersOn(g, glass, game.CounterCharge))
	}

	if err := g.ActivateCatalogAbility(me.ID, glass, 1, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("one charge counter paid for \"remove three\"")
	}
	effectstest.AddCounters(t, g, glass, game.CounterCharge, 2)
	if err := g.ActivateCatalogAbility(me.ID, glass, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the second ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 || effectstest.CountersOn(g, glass, game.CounterCharge) != 0 {
		t.Errorf("hand %d → %d, %d charge counters; want a second card and none left",
			hand, me.Hand.Size(), effectstest.CountersOn(g, glass, game.CounterCharge))
	}
}

// --- the keyword Arcbounds ---------------------------------------------------

func TestDstArcboundKeywordCreaturesEnterWithTheirCountersAndKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		counters               int
		keyword                string
	}{
		{"Arcbound Bruiser", "Artifact Creature — Golem", dstArcboundBruiserOracle, 3, ""},
		{"Arcbound Hybrid", "Artifact Creature — Beast", dstArcboundHybridOracle, 2, "haste"},
		{"Arcbound Lancer", "Artifact Creature — Beast", dstArcboundLancerOracle, 4, "first strike"},
		{"Arcbound Stinger", "Artifact Creature — Insect", dstArcboundStingerOracle, 1, "flying"},
		{"Arcbound Worker", "Artifact Creature — Construct", dstArcboundWorkerOracle, 1, ""},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		id := effectstest.MrdEnter(t, g, me, tc.name, tc.typeLine, tc.oracle)
		c := effectstest.LayeredCard(t, g, id)
		if got := c.Counters[game.CounterPlusOne]; got != tc.counters {
			t.Errorf("%s entered with %d +1/+1 counters, want %d", tc.name, got, tc.counters)
		}
		if tc.keyword != "" && !game.HasKeyword(&c, tc.keyword) {
			t.Errorf("%s lacks %s", tc.name, tc.keyword)
		}
	}
}

func TestDstArcboundWorkerPassesItsCountersWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	worker := effectstest.MrdEnter(t, g, me, "Arcbound Worker", "Artifact Creature — Construct", dstArcboundWorkerOracle)
	effectstest.AddCounters(t, g, worker, game.CounterPlusOne, 1)
	thopter := effectstest.MrdPermanent(g, me.ID, "Ornithopter", "Artifact Creature — Thopter", "{0}", 0, 2)

	effectstest.B25Destroy(g, worker)
	effectstest.AnswerPickTarget(t, g, thopter)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, thopter, game.CounterPlusOne); got != 2 {
		t.Errorf("the Thopter got %d counters, want the Worker's 2", got)
	}
}

// --- Arcbound Crusher --------------------------------------------------------

func TestDstArcboundCrusherGrowsWhenAnotherArtifactEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	crusher := effectstest.MrdEnter(t, g, me, "Arcbound Crusher", "Artifact Creature — Juggernaut", dstArcboundCrusherOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, crusher, game.CounterPlusOne); got != 1 {
		t.Fatalf("the Crusher entered with %d counters, want only modular's 1", got)
	}
	if c := effectstest.LayeredCard(t, g, crusher); !game.HasKeyword(&c, "trample") {
		t.Error("the Crusher lacks trample")
	}

	dstToken(t, g, opp.ID, "Their Thopter", "Token Artifact Creature — Thopter")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, crusher, game.CounterPlusOne); got != 2 {
		t.Fatalf("an opponent's artifact entered: %d counters, want 2", got)
	}
	dstToken(t, g, me.ID, "My Soldier", "Token Creature — Soldier")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, crusher, game.CounterPlusOne); got != 2 {
		t.Errorf("a nonartifact entered: %d counters, want still 2", got)
	}
}

// --- Arcbound Fiend ----------------------------------------------------------

func TestDstArcboundFiendMovesACounterFromTheTargetAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	fiend := effectstest.MrdEnter(t, g, me, "Arcbound Fiend", "Artifact Creature — Horror", dstArcboundFiendOracle)
	bear := effectstest.SeedBear(g, opp.ID)
	effectstest.AddCounters(t, g, bear, game.CounterPlusOne, 2)
	if c := effectstest.LayeredCard(t, g, fiend); !game.HasKeyword(&c, "fear") || c.Counters[game.CounterPlusOne] != 3 {
		t.Fatalf("the Fiend has fear %v and %d counters; want fear and 3",
			game.HasKeyword(&c, "fear"), c.Counters[game.CounterPlusOne])
	}

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, true)
	if got := effectstest.CountersOn(g, fiend, game.CounterPlusOne); got != 4 {
		t.Errorf("the Fiend has %d counters, want 4", got)
	}
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the target has %d counters, want 1", got)
	}
}

func TestDstArcboundFiendMovesNothingFromACreatureWithoutCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	fiend := effectstest.MrdEnter(t, g, me, "Arcbound Fiend", "Artifact Creature — Horror", dstArcboundFiendOracle)
	bear := effectstest.SeedBear(g, opp.ID)

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, true)
	if got := effectstest.CountersOn(g, fiend, game.CounterPlusOne); got != 3 {
		t.Errorf("the Fiend has %d counters, want 3: nothing to move (CR 122.5)", got)
	}
}

// --- Arcbound Reclaimer ------------------------------------------------------

func TestDstArcboundReclaimerTucksAnArtifactCardForACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	reclaimer := effectstest.MrdEnter(t, g, me, "Arcbound Reclaimer", "Artifact Creature — Golem", dstArcboundReclaimerOracle)
	creature := effectstest.PushGraveyardCardForTest(me, "Dead Bear")
	art := game.NewCard("Dead Rock", me.ID)
	art.TypeLine = "Artifact"
	me.Graveyard.PushTop(art)

	if err := g.ActivateCatalogAbility(me.ID, reclaimer, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(creature)}); err == nil {
		t.Fatal("a nonartifact card was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, reclaimer, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(art.InstanceID)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if top, _ := me.Library.Top(); top.InstanceID != art.InstanceID {
		t.Error("the artifact card should be on top of the library")
	}
	if got := effectstest.CountersOn(g, reclaimer, game.CounterPlusOne); got != 1 {
		t.Errorf("the Reclaimer has %d counters, want 1 after paying one", got)
	}
}

// --- Arcbound Slith ----------------------------------------------------------

func TestDstArcboundSlithGrowsOnCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	slith := effectstest.MrdEnter(t, g, me, "Arcbound Slith", "Artifact Creature — Slith", dstArcboundSlithOracle)
	if got := effectstest.CountersOn(g, slith, game.CounterPlusOne); got != 1 {
		t.Fatalf("the Slith entered with %d counters, want 1", got)
	}
	effectstest.DealCombatDamageToPlayer(g, slith, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, slith, game.CounterPlusOne); got != 2 {
		t.Errorf("after combat damage the Slith has %d counters, want 2", got)
	}
}
