package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b01_cards_test.go — March of the Machine, batch 1 (#5477):
// twelve vanilla, keyword and one-line creatures.

const (
	s82m1Sanctifier   = "ad172d4e-febf-4844-9544-861d03e27ab5"
	s82m1Crusher      = "6971eac1-afd0-46c3-a37f-4a26d50ef656"
	s82m1Billyrider   = "142d2063-d692-42a6-b828-42d998914589"
	s82m1Blademaster  = "5e1fd76b-46c8-41b5-87b3-e078f7a47535"
	s82m1Recluse      = "d99efecd-1419-434b-9de6-5fbb04f57dda"
	s82m1Analyst      = "e5c6bbee-5e6c-4a59-bc5d-285c5f5cf228"
	s82m1Scrapchomper = "baea7cc5-15ec-4229-ad8b-d4b34c2e21e8"
	s82m1Recycler     = "556c4dad-002e-41a8-8ba3-8b4eda70232f"
	s82m1Familiar     = "e2bc064e-4739-4eb1-ba44-1fd92b2829d7"
	s82m1Informant    = "a70ffefb-840c-48c3-a59d-e12306086bb6"
	s82m1Historian    = "399770b9-eab7-4512-aa1d-65c5bd81094f"
	s82m1Archivist    = "cdef800d-033e-4095-964d-9b6066f9fd36"
)

func TestS82M1AreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82m1Sanctifier: "Alabaster Host Sanctifier", s82m1Crusher: "Copper Host Crusher",
		s82m1Billyrider: "Kithkin Billyrider", s82m1Blademaster: "Iridescent Blademaster",
		s82m1Recluse: "Ruins Recluse", s82m1Analyst: "Furtive Analyst",
		s82m1Scrapchomper: "Akki Scrapchomper", s82m1Recycler: "Dreg Recycler",
		s82m1Familiar: "Etched Familiar", s82m1Informant: "Nezumi Informant",
		s82m1Historian: "Trailblazing Historian", s82m1Archivist: "Phyrexian Archivist",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82M1KeywordCreaturesHaveTheirKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, oracle string
		kws          []string
	}{
		{"Alabaster Host Sanctifier", s82m1Sanctifier, []string{"lifelink"}},
		{"Copper Host Crusher", s82m1Crusher, []string{"trample", "hexproof"}},
		{"Kithkin Billyrider", s82m1Billyrider, []string{"double strike"}},
		{"Ruins Recluse", s82m1Recluse, []string{"reach", "deathtouch"}},
		{"Furtive Analyst", s82m1Analyst, []string{"vigilance"}},
		{"Trailblazing Historian", s82m1Historian, []string{"haste"}},
		{"Phyrexian Archivist", s82m1Archivist, []string{"reach"}},
	} {
		id := effectstest.B12Push(g, me.ID, tc.name, "Creature — Test", tc.oracle, 1, 1)
		for _, kw := range tc.kws {
			if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
				t.Errorf("%s has no %s", tc.name, kw)
			}
		}
	}
}

func TestS82M1IridescentBlademasterPumps(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	id := effectstest.B12Push(g, me.ID, "Iridescent Blademaster", "Creature — Elf Warrior", s82m1Blademaster, 2, 2)
	effectstest.ApaMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without {3}{G}")
	}
	effectstest.ApaMana(me, "G")
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Strict: true})
	if p, tough := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 4 || tough != 4 {
		t.Errorf("P/T %d/%d, want 4/4", p, tough)
	}
}

func TestS82M1RuinsReclusePutsACounter(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	id := effectstest.B12Push(g, me.ID, "Ruins Recluse", "Creature — Spider", s82m1Recluse, 1, 1)
	effectstest.ApaMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without {3}{G}")
	}
	effectstest.ApaMana(me, "G")
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Strict: true})
	if n := effectstest.PlusOneCounters(g, id); n != 1 {
		t.Errorf("+1/+1 counters %d, want 1", n)
	}
}

func TestS82M1FurtiveAnalystLoots(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	id := effectstest.B12Push(g, me.ID, "Furtive Analyst", "Creature — Human Wizard", s82m1Analyst, 1, 4)
	effectstest.PutOnTopForTest(me, "Top Card", "Sorcery")
	effectstest.ApaMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with one mana for {2}")
	}
	effectstest.ApaMana(me, "C")
	hand := me.Hand.Size()
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Strict: true})
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d, want %d after the draw", me.Hand.Size(), hand+1)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d after the discard", me.Hand.Size(), hand)
	}
	if !effectstest.TappedNow(g, id) {
		t.Error("the Analyst did not tap")
	}
}

func TestS82M1AkkiScrapchomperSacrificesAnArtifactOrLandToDraw(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	id := effectstest.B12Push(g, me.ID, "Akki Scrapchomper", "Creature — Goblin", s82m1Scrapchomper, 1, 1)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	land := effectstest.B12Permanent(g, me.ID, "My Land", "Basic Land — Mountain")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.PutOnTopForTest(me, "Top Card", "Sorcery")
	effectstest.ApaMana(me, "C", "R")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a plain creature paid the cost")
	}
	_ = land
	hand := me.Hand.Size()
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{rock}})
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestS82M1DregRecyclerCanSacrificeItself(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	id := effectstest.B12Push(g, me.ID, "Dreg Recycler", "Creature — Phyrexian Beast", s82m1Recycler, 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "My Land", "Basic Land — Swamp")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("a land paid the cost")
	}
	mine, theirs := me.Life, opp.Life
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{id}})
	if g.Battlefield.Contains(id) {
		t.Error("the Recycler was not sacrificed")
	}
	if me.Life != mine+1 || opp.Life != theirs-1 {
		t.Errorf("life %d/%d, want %d/%d", me.Life, opp.Life, mine+1, theirs-1)
	}
}

func TestS82M1EtchedFamiliarDrainsWhenItDies(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	id := effectstest.B12Push(g, me.ID, "Etched Familiar", "Artifact Creature — Phyrexian Fox", s82m1Familiar, 3, 2)
	mine, theirs := me.Life, opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id, game.DestroyOptions{}) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine+2 || opp.Life != theirs-2 {
		t.Errorf("life %d/%d, want %d/%d", me.Life, opp.Life, mine+2, theirs-2)
	}
}

func TestS82M1NezumiInformantMakesEachOpponentDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastAndResolveCreature(t, g, "Nezumi Informant", "Creature — Rat Rogue", s82m1Informant)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		owed := effectstest.DiscardOwed(g, p.ID)
		if i == 0 && owed != 0 {
			t.Errorf("you owe nothing, got %d", owed)
		}
		if i != 0 && owed != 1 {
			t.Errorf("seat %d owes one discard, got %d", i, owed)
		}
	}
}

func TestS82M1TrailblazingHistorianGivesAnotherCreatureHaste(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	h := effectstest.B12Push(g, me.ID, "Trailblazing Historian", "Creature — Human Shaman", s82m1Historian, 1, 3)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, h, 0, game.ActivateAbilityParams{Targets: effectstest.RfCardTarget(h), Strict: true}); err == nil {
		t.Fatal("the Historian targeted itself")
	}
	effectstest.P7Activate(t, g, me, h, 0, game.ActivateAbilityParams{Targets: effectstest.RfCardTarget(bear), Strict: true})
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the target did not gain haste")
	}
}

func TestS82M1PhyrexianArchivistTucksAGraveyardCard(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	a := effectstest.B12Push(g, me.ID, "Phyrexian Archivist", "Artifact Creature — Phyrexian Construct", s82m1Archivist, 4, 5)
	dead := effectstest.PushGraveyardCardTyped(opp, "Their Dead Bear", "Creature — Bear")
	effectstest.ApaMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, a, 0, game.ActivateAbilityParams{Targets: effectstest.RfCardTarget(dead), Strict: true}); err == nil {
		t.Fatal("activated with one mana for {2}")
	}
	effectstest.ApaMana(me, "C")
	effectstest.P7Activate(t, g, me, a, 0, game.ActivateAbilityParams{Targets: effectstest.RfCardTarget(dead), Strict: true})
	if opp.Graveyard.Contains(dead) {
		t.Fatal("the card is still in the graveyard")
	}
	if cards := opp.Library.Cards; len(cards) == 0 || cards[0].InstanceID != dead {
		t.Error("the card is not on the bottom of its owner's library")
	}
}
