package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestHiddenCavesDiscoverFour(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Hidden Cataract", effectstest.HiddenCataractOracle},
		{"Hidden Courtyard", effectstest.HiddenCourtyardOracle},
		{"Hidden Necropolis", effectstest.HiddenNecropolisOracle},
		{"Hidden Nursery", effectstest.HiddenNurseryOracle},
		{"Hidden Volcano", effectstest.HiddenVolcanoOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			cave := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Land — Cave", tc.oracle, false)
			hit := effectstest.DiscoverLibrary(me, 4)
			effectstest.B16Activate(t, g, me.ID, cave, 0, game.ActivateAbilityParams{})
			if g.Battlefield.Contains(cave) {
				t.Errorf("the Cave was not sacrificed")
			}
			effectstest.ExpectDiscovered(t, g, me.ID, hit)
		})
	}
}

func TestDaringDiscoveryStopsBlockersAndDiscoversFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hit := effectstest.DiscoverLibrary(me, 4)
	effectstest.CastCatalogSpell(t, g, "Daring Discovery", "Sorcery", effectstest.DaringDiscoveryOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ExpectDiscovered(t, g, me.ID, hit)
	c, _ := g.LookupCardForEffect(bear)
	if c.Effective().Restrictions&game.CantBlock == 0 {
		t.Errorf("the targeted creature can still block")
	}
}

func TestWalkWithTheAncestorsReturnsAPermanentCardThenDiscoversFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := game.Card{InstanceID: uuid.New(), Name: "Dead Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(dead)
	hit := effectstest.DiscoverLibrary(me, 4)
	effectstest.CastCatalogSpell(t, g, "Walk with the Ancestors", "Sorcery", effectstest.WalkAncestorsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: dead.InstanceID}})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead.InstanceID) {
		t.Errorf("the permanent card did not return to hand")
	}
	effectstest.ExpectDiscovered(t, g, me.ID, hit)
}

func TestHitTheMotherLodeMakesTreasuresForTheDifference(t *testing.T) {
	treasures := func(g *game.Game, owner uuid.UUID) (n, tapped int) {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == owner && c.Name == "Treasure" {
				n++
				if c.Tapped {
					tapped++
				}
			}
		}
		return
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	seven := effectstest.DiscoverLibraryCard(me.ID, "Seven Drop", "Artifact", effectstest.GenericCost(7))
	me.Library.Cards = []game.Card{seven}
	effectstest.CastCatalogSpell(t, g, "Hit the Mother Lode", "Sorcery", effectstest.HitMotherLodeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ExpectDiscovered(t, g, me.ID, seven.InstanceID)
	if n, tapped := treasures(g, me.ID); n != 3 || tapped != 3 {
		t.Errorf("Treasures = %d (%d tapped), want 3 tapped", n, tapped)
	}

	// An empty walk discovers no card, so it makes no Treasures (owner
	// decision 6).
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	me.Library.Cards = []game.Card{effectstest.DiscoverLibraryCard(me.ID, "Forest", "Basic Land — Forest", "")}
	effectstest.CastCatalogSpell(t, g, "Hit the Mother Lode", "Sorcery", effectstest.HitMotherLodeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := treasures(g, me.ID); n != 0 {
		t.Errorf("an empty walk made %d Treasures", n)
	}
}

func TestHurlIntoHistoryCountersAndDiscoversTheSpellsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	creature := game.Card{InstanceID: uuid.New(), Name: "Big Beast", TypeLine: "Creature — Beast", ManaCost: "{3}{G}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(creature)
	if err := g.CastSpell(me.ID, creature.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the creature: %v", err)
	}
	hit := effectstest.DiscoverLibrary(me, 4)
	hurl := game.Card{InstanceID: uuid.New(), Name: "Hurl into History", TypeLine: "Instant", OracleID: effectstest.HurlIntoHistoryOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(hurl)
	if err := g.CastSpell(me.ID, hurl.InstanceID, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: creature.InstanceID}}}); err != nil {
		t.Fatalf("cast Hurl into History: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(creature.InstanceID) {
		t.Errorf("the creature spell was not countered")
	}
	effectstest.ExpectDiscovered(t, g, me.ID, hit)
}

func TestZoyowasJusticeMakesTheOwnerDiscover(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	target := effectstest.B16Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == target {
				g.Battlefield.Cards[i].ManaCost = "{2}{R}"
			}
		}
	})
	opp.Library.Cards = nil
	effectstest.CastCatalogSpell(t, g, "Zoyowa's Justice", "Instant", effectstest.ZoyowasJusticeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(target) {
		t.Fatalf("the creature is still on the battlefield")
	}
	// The shuffled library holds only the Ogre, which has mana value 3,
	// so the owner's discover 3 offers it.
	c := effectstest.DiscoverPromptFor(g, opp.ID)
	if c == nil || effectstest.KeywordCastCard(c) != target {
		t.Fatalf("the owner was not offered their own discover 3: %+v", c)
	}
}

func TestContestOfClawsDiscoversTheExcess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B16Creature(g, me.ID, "My Beast", "Creature — Beast", 5, 5)
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hit := effectstest.DiscoverLibrary(me, 3)
	effectstest.CastCatalogSpell(t, g, "Contest of Claws", "Sorcery", effectstest.ContestOfClawsOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs},
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ExpectDiscovered(t, g, me.ID, hit)

	// No excess, no discover.
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	mine = effectstest.B16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs = effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.DiscoverLibrary(me, 0)
	effectstest.CastCatalogSpell(t, g, "Contest of Claws", "Sorcery", effectstest.ContestOfClawsOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscoverPromptFor(g, me.ID) != nil {
		t.Errorf("exactly lethal damage discovered")
	}
}

func TestTrumpetingCarnosaurDiscoversFiveAndFlingsFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hit := effectstest.DiscoverLibrary(me, 5)
	effectstest.CastCatalogSpell(t, g, "Trumpeting Carnosaur", "Creature — Dinosaur", effectstest.TrumpetingCarnoOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ExpectDiscovered(t, g, me.ID, hit)

	// The hand ability: discard it, 3 damage to a creature.
	bear := effectstest.B16Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 4)
	card := game.Card{InstanceID: uuid.New(), Name: "Trumpeting Carnosaur", TypeLine: "Creature — Dinosaur", OracleID: effectstest.TrumpetingCarnoOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	if err := g.ActivateCatalogAbility(me.ID, card.InstanceID, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate from hand: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(card.InstanceID) {
		t.Errorf("the Carnosaur was not discarded")
	}
	c, _ := g.LookupCardForEffect(bear)
	if c.DamageMarked != 3 {
		t.Errorf("damage marked = %d, want 3", c.DamageMarked)
	}
}

func TestEtalisFavorPumpsAndDiscoversThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	hit := effectstest.DiscoverLibrary(me, 3)
	effectstest.CastCatalogSpell(t, g, "Etali's Favor", "Enchantment — Aura", effectstest.EtalisFavorOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ExpectDiscovered(t, g, me.ID, hit)
	if p := effectstest.EffectivePower(t, g, mine); p != 3 {
		t.Errorf("enchanted creature power = %d, want 3", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, mine, "trample") {
		t.Errorf("enchanted creature lacks trample")
	}
}

func TestGeologicalAppraiserDiscoversOnlyWhenCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hit := effectstest.DiscoverLibrary(me, 3)
	effectstest.CastCatalogSpell(t, g, "Geological Appraiser", "Creature — Human Artificer", effectstest.GeologicalApprOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ExpectDiscovered(t, g, me.ID, hit)

	// Put onto the battlefield without being cast: no discover.
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.DiscoverLibrary(me, 3)
	card := game.Card{InstanceID: uuid.New(), Name: "Geological Appraiser", TypeLine: "Creature — Human Artificer", OracleID: effectstest.GeologicalApprOracle, Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(card.InstanceID, game.HandEntryOptions{}); err != nil {
			t.Fatalf("put onto the battlefield: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscoverPromptFor(g, me.ID) != nil || len(g.PendingTriggers) != 0 {
		t.Errorf("an Appraiser that was not cast discovered")
	}
}

func TestChimilDiscoversFiveAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Chimil, the Inner Sun", "Legendary Artifact", effectstest.ChimilOracle, false)
	hit := effectstest.DiscoverLibrary(me, 5)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ExpectDiscovered(t, g, me.ID, hit)
	// ADR 0106 PR 2 (#1806) closed the caveat ADR 0099's owner decision
	// 5 shipped it with: its first line is now counter_shields_test.go's.
	spec, _ := effects.Lookup(effectstest.ChimilOracle)
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Chimil is complete since #1806: %v %v", spec.Completeness, spec.Caveats)
	}
}

func TestFranklinRichardsDiscoversSixAfterANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Franklin Richards, Ascendant", "Legendary Creature — Mutant Hero", effectstest.FranklinRichardsOracle, false)
	effectstest.DiscoverLibrary(me, 6)
	effectstest.AdvanceToMain(t, g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscoverPromptFor(g, me.ID) != nil {
		t.Fatalf("Franklin discovered with no noncreature spell cast")
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Franklin Richards, Ascendant", "Legendary Creature — Mutant Hero", effectstest.FranklinRichardsOracle, false)
	effectstest.CastCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	hit := effectstest.DiscoverLibrary(me, 6)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ExpectDiscovered(t, g, me.ID, hit)
}

func TestMonstrousVortexDiscoversTheBigSpellsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Monstrous Vortex", "Enchantment", effectstest.MonstrousVortexOracle, false)
	effectstest.AdvanceToMain(t, g)
	small := game.Card{InstanceID: uuid.New(), Name: "Small Beast", TypeLine: "Creature — Beast", ManaCost: "{2}{G}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(small)
	if err := g.CastSpell(me.ID, small.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscoverPromptFor(g, me.ID) != nil {
		t.Fatalf("a power-4 creature spell discovered")
	}
	hit := effectstest.DiscoverLibrary(me, 6)
	big := game.Card{InstanceID: uuid.New(), Name: "Big Beast", TypeLine: "Creature — Beast", ManaCost: "{5}{G}", Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(big)
	if err := g.CastSpell(me.ID, big.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ExpectDiscovered(t, g, me.ID, hit)
}

func TestCuratorDiscoversAgainOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Curator of Sun's Creation", "Creature — Human Artificer", effectstest.CuratorOracle, false)
	cave := effectstest.PushCatalogPermanent(g, me.ID, "Hidden Courtyard", "Land — Cave", effectstest.HiddenCourtyardOracle, false)
	first := effectstest.DiscoverLibraryCard(me.ID, "First", "Artifact", "{4}")
	second := effectstest.DiscoverLibraryCard(me.ID, "Second", "Artifact", "{4}")
	third := effectstest.DiscoverLibraryCard(me.ID, "Third", "Artifact", "{4}")
	me.Library.Cards = []game.Card{third, second, first}
	effectstest.B16Activate(t, g, me.ID, cave, 0, game.ActivateAbilityParams{})
	effectstest.ExpectDiscovered(t, g, me.ID, first.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	// "Discover again for the same value": 4.
	effectstest.ExpectDiscovered(t, g, me.ID, second.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscoverPromptFor(g, me.ID) != nil {
		t.Errorf("Curator triggered twice in one turn")
	}
	if !me.Library.Contains(third.InstanceID) {
		t.Errorf("a third discover happened")
	}
}

func TestPantlazaDiscoversTheDinosaursToughnessOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Pantlaza, Sun-Favored", "Legendary Creature — Dinosaur", effectstest.PantlazaOracle, false)
	hit := effectstest.DiscoverLibrary(me, 3)
	effectstest.AdvanceToMain(t, g)
	dino := game.Card{InstanceID: uuid.New(), Name: "Raptor", TypeLine: "Creature — Dinosaur", ManaCost: "{1}{R}", Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(dino)
	if err := g.CastSpell(me.ID, dino.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ExpectDiscovered(t, g, me.ID, hit)

	// A second Dinosaur this turn: the discover was used.
	effectstest.DiscoverLibrary(me, 3)
	dino2 := game.Card{InstanceID: uuid.New(), Name: "Raptor Two", TypeLine: "Creature — Dinosaur", ManaCost: "{1}{R}", Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(dino2)
	if err := g.CastSpell(me.ID, dino2.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && (c.Kind == game.PendingChoiceTriggerPrompt || effectstest.IsKeywordCastPrompt(c)) {
			t.Fatalf("Pantlaza offered a second discover this turn")
		}
	}
}

func TestAloyDiscoversTheGreatestAttackingArtifactPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Aloy, Savior of Meridian", "Legendary Creature — Human Warrior", effectstest.AloyOracle, false)
	thopter := effectstest.B16Creature(g, me.ID, "Big Construct", "Artifact Creature — Construct", 4, 4)
	small := effectstest.B16Creature(g, me.ID, "Small Construct", "Artifact Creature — Construct", 1, 1)
	hit := effectstest.DiscoverLibrary(me, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{thopter, small} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ExpectDiscovered(t, g, me.ID, hit)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscoverPromptFor(g, me.ID) != nil {
		t.Errorf("two attacking artifact creatures triggered Aloy twice")
	}
}

func TestDinosaurEggMayDiscoverItsToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	egg := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dinosaur Egg", TypeLine: "Creature — Dinosaur Egg",
		OracleID: effectstest.DinosaurEggOracle, Power: 0, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	hit := effectstest.DiscoverLibrary(me, 3)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(egg); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ExpectDiscovered(t, g, me.ID, hit)
}

func TestCaparoctiTapsTwoToDiscoverThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	capa := effectstest.PushCatalogPermanent(g, me.ID, "Caparocti Sunborn", "Legendary Creature — Human Soldier", effectstest.CaparoctiOracle, false)
	a := effectstest.B16Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B16Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	hit := effectstest.DiscoverLibrary(me, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(capa, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	var prompt *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == me.ID {
			prompt = c
		}
	}
	if prompt == nil {
		t.Fatalf("no tap-two prompt")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !effectstest.B16Tapped(t, g, a) || !effectstest.B16Tapped(t, g, b) {
		t.Errorf("the two creatures were not tapped")
	}
	effectstest.ExpectDiscovered(t, g, me.ID, hit)
}

func TestEllieAndAlanDiscoverTheExiledCardsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ea := effectstest.PushCatalogPermanent(g, me.ID, "Ellie and Alan, Paleontologists", "Legendary Creature — Human Scientist", effectstest.EllieAndAlanOracle, false)
	fossil := game.Card{InstanceID: uuid.New(), Name: "Fossil", TypeLine: "Creature — Dinosaur", ManaCost: "{4}{G}", Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(fossil)
	hit := effectstest.DiscoverLibrary(me, 5)
	effectstest.B16Activate(t, g, me.ID, ea, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{fossil.InstanceID}})
	if !g.Exile.Contains(fossil.InstanceID) {
		t.Errorf("the creature card was not exiled as the cost")
	}
	effectstest.ExpectDiscovered(t, g, me.ID, hit)
}

func TestLongRangeSensorChargesOnAttackAndDiscoversFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sensor := effectstest.PushCatalogPermanent(g, me.ID, "Long-Range Sensor", "Artifact", effectstest.LongRangeSensorOracle, false)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, sensor, "charge"); got != 1 {
		t.Fatalf("charge counters after attacking = %d, want 1", got)
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sensor = effectstest.PushCatalogPermanent(g, me.ID, "Long-Range Sensor", "Artifact", effectstest.LongRangeSensorOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(sensor, "charge", 2) })
	hit := effectstest.DiscoverLibrary(me, 4)
	effectstest.B16Activate(t, g, me.ID, sensor, 0, game.ActivateAbilityParams{})
	if got := effectstest.CounterOn(g, sensor, "charge"); got != 0 {
		t.Errorf("charge counters after activating = %d, want 0", got)
	}
	effectstest.ExpectDiscovered(t, g, me.ID, hit)
}

func TestBuriedTreasureDiscoversFiveFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hit := effectstest.DiscoverLibrary(me, 5)
	id, _ := effectstest.ActivateFromGraveyard(t, g, "Buried Treasure", "Artifact — Treasure", effectstest.BuriedTreasureOracle, 0, 0, "", game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Errorf("Buried Treasure did not exile itself")
	}
	effectstest.ExpectDiscovered(t, g, me.ID, hit)
}

func TestQuintoriusKand(t *testing.T) {
	t.Run("-3 discovers four", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		q := effectstest.PushCatalogPermanent(g, me.ID, "Quintorius Kand", "Legendary Planeswalker — Quintorius", effectstest.QuintoriusKandOracle, false)
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(q, game.CounterLoyalty, 4) })
		hit := effectstest.DiscoverLibrary(me, 4)
		effectstest.B16Activate(t, g, me.ID, q, 1, game.ActivateAbilityParams{})
		effectstest.ExpectDiscovered(t, g, me.ID, hit)
	})
	t.Run("+1 makes a 3/2 Spirit", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		q := effectstest.PushCatalogPermanent(g, me.ID, "Quintorius Kand", "Legendary Planeswalker — Quintorius", effectstest.QuintoriusKandOracle, false)
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(q, game.CounterLoyalty, 4) })
		effectstest.B16Activate(t, g, me.ID, q, 0, game.ActivateAbilityParams{})
		found := false
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Spirit" && c.Power == 3 && c.Toughness == 2 && c.Controller == me.ID {
				found = true
			}
		}
		if !found {
			t.Errorf("no 3/2 Spirit token")
		}
	})
	t.Run("-6 exiles, adds red and lets you play them", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		q := effectstest.PushCatalogPermanent(g, me.ID, "Quintorius Kand", "Legendary Planeswalker — Quintorius", effectstest.QuintoriusKandOracle, false)
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(q, game.CounterLoyalty, 6) })
		a := game.Card{InstanceID: uuid.New(), Name: "Shock", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID}
		b := game.Card{InstanceID: uuid.New(), Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID}
		me.Graveyard.PushTop(a)
		me.Graveyard.PushTop(b)
		effectstest.B16Activate(t, g, me.ID, q, 2, game.ActivateAbilityParams{Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: a.InstanceID}, {Kind: game.TargetCard, ID: b.InstanceID},
		}})
		if !g.Exile.Contains(a.InstanceID) || !g.Exile.Contains(b.InstanceID) {
			t.Fatalf("the cards were not exiled")
		}
		red := 0
		for _, tok := range me.ManaPool.Units() {
			if tok.Color == "R" {
				red++
			}
		}
		if red != 2 {
			t.Errorf("red mana = %d, want 2", red)
		}
		for _, id := range []uuid.UUID{a.InstanceID, b.InstanceID} {
			perm := g.CastPermissionOnCardByIDForEffect(id)
			if perm == nil || perm.CastOnly || !g.CastPermissionActiveForEffect(perm, me.ID) {
				t.Errorf("no play permission on an exiled card: %+v", perm)
			}
		}
	})
	t.Run("a spell cast from exile drains each opponent", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		q := effectstest.PushCatalogPermanent(g, me.ID, "Quintorius Kand", "Legendary Planeswalker — Quintorius", effectstest.QuintoriusKandOracle, false)
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(q, game.CounterLoyalty, 4) })
		effectstest.AdvanceToMain(t, g)
		spell := game.Card{InstanceID: uuid.New(), Name: "Exiled Sorcery", TypeLine: "Sorcery", ManaCost: "{1}", Owner: me.ID, Controller: me.ID}
		g.Exile.PushTop(spell)
		g.WithWriteLock(func() {
			g.GrantCastPermissionToCardsForEffect(game.CastPermission{Player: me.ID, Zone: game.ZoneExile, Cost: "{0}"}, []game.Card{spell})
		})
		life, oppLife := me.Life, opp.Life
		if err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile)}); err != nil {
			t.Fatalf("cast from exile: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Life != oppLife-2 || me.Life != life+2 {
			t.Errorf("life: me %d→%d, opp %d→%d; want +2 / -2", life, me.Life, oppLife, opp.Life)
		}
	})
}

func TestSwashbucklersWhipGrantsReachAndItsTwoAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	whip := effectstest.PushCatalogPermanent(g, me.ID, "Swashbuckler's Whip", "Artifact — Equipment", effectstest.SwashbucklersWhipOracle, false)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B16Activate(t, g, me.ID, whip, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if !effectstest.HasEffectiveKeyword(t, g, bear, "reach") {
		t.Fatalf("the equipped creature lacks reach")
	}
	// Granted ability 0: tap target artifact or creature.
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B16Activate(t, g, me.ID, bear, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}})
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Errorf("the tap ability did not tap its target")
	}
	// Granted ability 1: discover 10. Untap the bear first.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(bear) })
	hit := effectstest.DiscoverLibrary(me, 10)
	effectstest.B16Activate(t, g, me.ID, bear, 1, game.ActivateAbilityParams{})
	effectstest.ExpectDiscovered(t, g, me.ID, hit)
}

func TestZoeticGlyphAnimatesAndDiscoversWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	relic := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Old Relic", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Zoetic Glyph", "Enchantment — Aura", effectstest.ZoeticGlyphOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: relic}})
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := g.LookupCardForEffect(relic)
	eff := c.Effective()
	if eff.Power != 5 || eff.Toughness != 4 || !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Golem") {
		t.Fatalf("enchanted artifact = %d/%d types %v %v, want a 5/4 Golem artifact creature", eff.Power, eff.Toughness, eff.Types, eff.Subtypes)
	}
	var glyph uuid.UUID
	for _, bc := range g.Battlefield.Cards {
		if bc.Name == "Zoetic Glyph" {
			glyph = bc.InstanceID
		}
	}
	hit := effectstest.DiscoverLibrary(me, 3)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(glyph); err != nil {
			t.Fatalf("destroy the Aura: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ExpectDiscovered(t, g, me.ID, hit)
}
