package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// discover_cards_test.go — the discover cards (ADR 0099). Each test
// proves the card discovers the right N: the library is a card with
// mana value exactly N on the bottom and one with mana value N+1 on
// top, so the prompt offers the bottom card only when N is right.

const (
	hiddenCataractOracle    = "927979d7-9b5c-4448-aef0-baf2907a89f1"
	hiddenCourtyardOracle   = "e19d5071-4ea1-4883-b067-a21e553f96e0"
	hiddenNecropolisOracle  = "f780ee53-62b0-4c32-b5b7-047651f48e5f"
	hiddenNurseryOracle     = "1a26e2d6-6bfc-4cdc-9bd6-8b37a9be2961"
	hiddenVolcanoOracle     = "a1c7cd7a-0795-4135-b787-effeb981d95b"
	daringDiscoveryOracle   = "a9564593-b5a6-4a83-be1e-2af8caf647d7"
	walkAncestorsOracle     = "ce657a2e-461b-45fd-ab98-35b5c8512777"
	hitMotherLodeOracle     = "a49900b3-cc34-428f-806d-71861dc8059d"
	hurlIntoHistoryOracle   = "016718df-2eb5-499f-adea-18231c0375b3"
	zoyowasJusticeOracle    = "17e11f0c-4ac7-4060-9e0b-27cd5c151e50"
	contestOfClawsOracle    = "e30d5114-40d7-4dc5-8067-9c5e9625e097"
	trumpetingCarnoOracle   = "f2ef8bda-373d-4387-900d-0f1b6ccf72e9"
	etalisFavorOracle       = "df98b93b-0aa6-4bea-9d8d-798ebc795b0b"
	geologicalApprOracle    = "a0a1cd2b-ed7c-4cce-8497-4873e1e5e3df"
	chimilOracle            = "119d9671-61ba-4629-89b1-f94bdec5cb74"
	franklinRichardsOracle  = "af20f452-5cb9-47cc-9a81-71f23204ddd0"
	monstrousVortexOracle   = "6e27956e-ba2f-41a9-8e39-ec54424591b6"
	curatorOracle           = "0e636c98-67c7-4f02-b909-5d43e276fa41"
	pantlazaOracle          = "b828ba28-e98d-498b-9ab4-d4aa7143e407"
	aloyOracle              = "f0554a8f-32de-4069-9f47-5e06ceb3f09d"
	dinosaurEggOracle       = "ac4d5a97-6177-4eac-b0f2-cf10531ad879"
	caparoctiOracle         = "7f87a770-af47-4345-8dc1-0a1e8bae75c5"
	ellieAndAlanOracle      = "0f4d513e-100d-4cbb-b919-09249cef5fad"
	longRangeSensorOracle   = "358297cd-4a9f-48e2-aaa5-131b0849f8dc"
	buriedTreasureOracle    = "56f5ed8d-ef53-4766-9ef3-1e24a10e267b"
	quintoriusKandOracle    = "91babc5f-ebeb-4894-8c35-b0f60249cb80"
	swashbucklersWhipOracle = "5cdb69e5-9395-4ef6-be18-71e3dd70df66"
	zoeticGlyphOracle       = "1c21efcf-1007-45bb-ba21-ac33c2a8d751"
)

// discoverLibrary replaces p's library with a mana-value-n card under a
// mana-value-(n+1) card and returns the first: the card a discover n
// finds, and the one a discover of anything else does not.
func discoverLibrary(p *game.Player, n int) uuid.UUID {
	exact := discoverLibraryCard(p.ID, "Exact Hit", "Artifact", genericCost(n))
	over := discoverLibraryCard(p.ID, "Over By One", "Artifact", genericCost(n+1))
	p.Library.Cards = []game.Card{exact, over}
	return exact.InstanceID
}

func genericCost(n int) string {
	if n == 0 {
		return "{0}"
	}
	out := ""
	for i := 0; i < n; i++ {
		out += "{1}"
	}
	return out
}

// discoverPromptFor is the open may_cast prompt addressed to chooser,
// or nil.
func discoverPromptFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMayCast && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// expectDiscovered fails unless a discover prompt for chooser offers
// want, then declines it (the card goes to hand).
func expectDiscovered(t *testing.T, g *game.Game, chooser, want uuid.UUID) {
	t.Helper()
	c := discoverPromptFor(g, chooser)
	if c == nil {
		t.Fatalf("no discover prompt for the discoverer")
	}
	if c.MayCastKeyword != game.MayCastKeywordDiscover {
		t.Fatalf("the prompt is a %q may_cast, not a discover", c.MayCastKeyword)
	}
	if c.MayCastCard != want {
		t.Fatalf("discover offered the wrong card — the N is wrong")
	}
	answerMayCastPrompt(t, g, chooser, false)
}

func TestHiddenCavesDiscoverFour(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Hidden Cataract", hiddenCataractOracle},
		{"Hidden Courtyard", hiddenCourtyardOracle},
		{"Hidden Necropolis", hiddenNecropolisOracle},
		{"Hidden Nursery", hiddenNurseryOracle},
		{"Hidden Volcano", hiddenVolcanoOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			advanceToMain(t, g)
			cave := pushCatalogPermanent(g, me.ID, tc.name, "Land — Cave", tc.oracle, false)
			hit := discoverLibrary(me, 4)
			b16Activate(t, g, me.ID, cave, 0, game.ActivateAbilityParams{})
			if g.Battlefield.Contains(cave) {
				t.Errorf("the Cave was not sacrificed")
			}
			expectDiscovered(t, g, me.ID, hit)
		})
	}
}

func TestDaringDiscoveryStopsBlockersAndDiscoversFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hit := discoverLibrary(me, 4)
	castCatalogSpell(t, g, "Daring Discovery", "Sorcery", daringDiscoveryOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	expectDiscovered(t, g, me.ID, hit)
	c, _ := g.LookupCardForEffect(bear)
	if c.Effective().Restrictions&game.CantBlock == 0 {
		t.Errorf("the targeted creature can still block")
	}
}

func TestWalkWithTheAncestorsReturnsAPermanentCardThenDiscoversFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := game.Card{InstanceID: uuid.New(), Name: "Dead Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(dead)
	hit := discoverLibrary(me, 4)
	castCatalogSpell(t, g, "Walk with the Ancestors", "Sorcery", walkAncestorsOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: dead.InstanceID}})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead.InstanceID) {
		t.Errorf("the permanent card did not return to hand")
	}
	expectDiscovered(t, g, me.ID, hit)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	seven := discoverLibraryCard(me.ID, "Seven Drop", "Artifact", genericCost(7))
	me.Library.Cards = []game.Card{seven}
	castCatalogSpell(t, g, "Hit the Mother Lode", "Sorcery", hitMotherLodeOracle, nil)
	passPriorityAroundTable(t, g)
	expectDiscovered(t, g, me.ID, seven.InstanceID)
	if n, tapped := treasures(g, me.ID); n != 3 || tapped != 3 {
		t.Errorf("Treasures = %d (%d tapped), want 3 tapped", n, tapped)
	}

	// An empty walk discovers no card, so it makes no Treasures (owner
	// decision 6).
	g = newCatalogGame(t)
	me = g.Seats[0]
	me.Library.Cards = []game.Card{discoverLibraryCard(me.ID, "Forest", "Basic Land — Forest", "")}
	castCatalogSpell(t, g, "Hit the Mother Lode", "Sorcery", hitMotherLodeOracle, nil)
	passPriorityAroundTable(t, g)
	if n, _ := treasures(g, me.ID); n != 0 {
		t.Errorf("an empty walk made %d Treasures", n)
	}
}

func TestHurlIntoHistoryCountersAndDiscoversTheSpellsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	creature := game.Card{InstanceID: uuid.New(), Name: "Big Beast", TypeLine: "Creature — Beast", ManaCost: "{3}{G}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(creature)
	if err := g.CastSpell(me.ID, creature.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the creature: %v", err)
	}
	hit := discoverLibrary(me, 4)
	hurl := game.Card{InstanceID: uuid.New(), Name: "Hurl into History", TypeLine: "Instant", OracleID: hurlIntoHistoryOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(hurl)
	if err := g.CastSpell(me.ID, hurl.InstanceID, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: creature.InstanceID}}}); err != nil {
		t.Fatalf("cast Hurl into History: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(creature.InstanceID) {
		t.Errorf("the creature spell was not countered")
	}
	expectDiscovered(t, g, me.ID, hit)
}

func TestZoyowasJusticeMakesTheOwnerDiscover(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	target := b16Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == target {
				g.Battlefield.Cards[i].ManaCost = "{2}{R}"
			}
		}
	})
	opp.Library.Cards = nil
	castCatalogSpell(t, g, "Zoyowa's Justice", "Instant", zoyowasJusticeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(target) {
		t.Fatalf("the creature is still on the battlefield")
	}
	// The shuffled library holds only the Ogre, which has mana value 3,
	// so the owner's discover 3 offers it.
	c := discoverPromptFor(g, opp.ID)
	if c == nil || c.MayCastCard != target {
		t.Fatalf("the owner was not offered their own discover 3: %+v", c)
	}
}

func TestContestOfClawsDiscoversTheExcess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b16Creature(g, me.ID, "My Beast", "Creature — Beast", 5, 5)
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	hit := discoverLibrary(me, 3)
	castCatalogSpell(t, g, "Contest of Claws", "Sorcery", contestOfClawsOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs},
	})
	passPriorityAroundTable(t, g)
	expectDiscovered(t, g, me.ID, hit)

	// No excess, no discover.
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	mine = b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs = b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	discoverLibrary(me, 0)
	castCatalogSpell(t, g, "Contest of Claws", "Sorcery", contestOfClawsOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs},
	})
	passPriorityAroundTable(t, g)
	if discoverPromptFor(g, me.ID) != nil {
		t.Errorf("exactly lethal damage discovered")
	}
}

func TestTrumpetingCarnosaurDiscoversFiveAndFlingsFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hit := discoverLibrary(me, 5)
	castCatalogSpell(t, g, "Trumpeting Carnosaur", "Creature — Dinosaur", trumpetingCarnoOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	expectDiscovered(t, g, me.ID, hit)

	// The hand ability: discard it, 3 damage to a creature.
	bear := b16Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 4)
	card := game.Card{InstanceID: uuid.New(), Name: "Trumpeting Carnosaur", TypeLine: "Creature — Dinosaur", OracleID: trumpetingCarnoOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	if err := g.ActivateCatalogAbility(me.ID, card.InstanceID, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("activate from hand: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(card.InstanceID) {
		t.Errorf("the Carnosaur was not discarded")
	}
	c, _ := g.LookupCardForEffect(bear)
	if c.DamageMarked != 3 {
		t.Errorf("damage marked = %d, want 3", c.DamageMarked)
	}
}

func TestEtalisFavorPumpsAndDiscoversThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	hit := discoverLibrary(me, 3)
	castCatalogSpell(t, g, "Etali's Favor", "Enchantment — Aura", etalisFavorOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	expectDiscovered(t, g, me.ID, hit)
	if p := effectivePower(t, g, mine); p != 3 {
		t.Errorf("enchanted creature power = %d, want 3", p)
	}
	if !hasEffectiveKeyword(t, g, mine, "trample") {
		t.Errorf("enchanted creature lacks trample")
	}
}

func TestGeologicalAppraiserDiscoversOnlyWhenCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hit := discoverLibrary(me, 3)
	castCatalogSpell(t, g, "Geological Appraiser", "Creature — Human Artificer", geologicalApprOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	expectDiscovered(t, g, me.ID, hit)

	// Put onto the battlefield without being cast: no discover.
	g = newCatalogGame(t)
	me = g.Seats[0]
	discoverLibrary(me, 3)
	card := game.Card{InstanceID: uuid.New(), Name: "Geological Appraiser", TypeLine: "Creature — Human Artificer", OracleID: geologicalApprOracle, Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(card.InstanceID, game.HandEntryOptions{}); err != nil {
			t.Fatalf("put onto the battlefield: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if discoverPromptFor(g, me.ID) != nil || len(g.PendingTriggers) != 0 {
		t.Errorf("an Appraiser that was not cast discovered")
	}
}

func TestChimilDiscoversFiveAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Chimil, the Inner Sun", "Legendary Artifact", chimilOracle, false)
	hit := discoverLibrary(me, 5)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	expectDiscovered(t, g, me.ID, hit)
	spec, _ := Lookup(chimilOracle)
	if spec.Completeness != CompletenessCaveats || len(spec.Caveats) != 1 {
		t.Errorf("Chimil must carry its can't-be-countered caveat (owner decision 5)")
	}
}

func TestFranklinRichardsDiscoversSixAfterANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Franklin Richards, Ascendant", "Legendary Creature — Mutant Hero", franklinRichardsOracle, false)
	discoverLibrary(me, 6)
	advanceToMain(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if discoverPromptFor(g, me.ID) != nil {
		t.Fatalf("Franklin discovered with no noncreature spell cast")
	}

	g = newCatalogGame(t)
	me = g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Franklin Richards, Ascendant", "Legendary Creature — Mutant Hero", franklinRichardsOracle, false)
	castCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	hit := discoverLibrary(me, 6)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	expectDiscovered(t, g, me.ID, hit)
}

func TestMonstrousVortexDiscoversTheBigSpellsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Monstrous Vortex", "Enchantment", monstrousVortexOracle, false)
	advanceToMain(t, g)
	small := game.Card{InstanceID: uuid.New(), Name: "Small Beast", TypeLine: "Creature — Beast", ManaCost: "{2}{G}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(small)
	if err := g.CastSpell(me.ID, small.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if discoverPromptFor(g, me.ID) != nil {
		t.Fatalf("a power-4 creature spell discovered")
	}
	hit := discoverLibrary(me, 6)
	big := game.Card{InstanceID: uuid.New(), Name: "Big Beast", TypeLine: "Creature — Beast", ManaCost: "{5}{G}", Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(big)
	if err := g.CastSpell(me.ID, big.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	expectDiscovered(t, g, me.ID, hit)
}

func TestCuratorDiscoversAgainOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	curator := pushCatalogPermanent(g, me.ID, "Curator of Sun's Creation", "Creature — Human Artificer", curatorOracle, false)
	first := discoverLibraryCard(me.ID, "First", "Artifact", "{2}")
	second := discoverLibraryCard(me.ID, "Second", "Artifact", "{2}")
	third := discoverLibraryCard(me.ID, "Third", "Artifact", "{2}")
	me.Library.Cards = []game.Card{third, second, first}
	g.WithWriteLock(func() {
		_ = g.DiscoverThenForEffect(me.ID, curator, 2, nil)
	})
	expectDiscovered(t, g, me.ID, first.InstanceID)
	passPriorityAroundTable(t, g)
	// "Discover again for the same value": 2.
	expectDiscovered(t, g, me.ID, second.InstanceID)
	passPriorityAroundTable(t, g)
	if discoverPromptFor(g, me.ID) != nil {
		t.Errorf("Curator triggered twice in one turn")
	}
	if !me.Library.Contains(third.InstanceID) {
		t.Errorf("a third discover happened")
	}
}

func TestPantlazaDiscoversTheDinosaursToughnessOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Pantlaza, Sun-Favored", "Legendary Creature — Dinosaur", pantlazaOracle, false)
	hit := discoverLibrary(me, 3)
	advanceToMain(t, g)
	dino := game.Card{InstanceID: uuid.New(), Name: "Raptor", TypeLine: "Creature — Dinosaur", ManaCost: "{1}{R}", Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(dino)
	if err := g.CastSpell(me.ID, dino.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	expectDiscovered(t, g, me.ID, hit)

	// A second Dinosaur this turn: the discover was used.
	discoverLibrary(me, 3)
	dino2 := game.Card{InstanceID: uuid.New(), Name: "Raptor Two", TypeLine: "Creature — Dinosaur", ManaCost: "{1}{R}", Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(dino2)
	if err := g.CastSpell(me.ID, dino2.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && (c.Kind == game.PendingChoiceTriggerPrompt || c.Kind == game.PendingChoiceMayCast) {
			t.Fatalf("Pantlaza offered a second discover this turn")
		}
	}
}

func TestAloyDiscoversTheGreatestAttackingArtifactPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Aloy, Savior of Meridian", "Legendary Creature — Human Warrior", aloyOracle, false)
	thopter := b16Creature(g, me.ID, "Big Construct", "Artifact Creature — Construct", 4, 4)
	small := b16Creature(g, me.ID, "Small Construct", "Artifact Creature — Construct", 1, 1)
	hit := discoverLibrary(me, 4)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{thopter, small} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	expectDiscovered(t, g, me.ID, hit)
	passPriorityAroundTable(t, g)
	if discoverPromptFor(g, me.ID) != nil {
		t.Errorf("two attacking artifact creatures triggered Aloy twice")
	}
}

func TestDinosaurEggMayDiscoverItsToughness(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	egg := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dinosaur Egg", TypeLine: "Creature — Dinosaur Egg",
		OracleID: dinosaurEggOracle, Power: 0, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	hit := discoverLibrary(me, 3)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(egg); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	expectDiscovered(t, g, me.ID, hit)
}

func TestCaparoctiTapsTwoToDiscoverThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	capa := pushCatalogPermanent(g, me.ID, "Caparocti Sunborn", "Legendary Creature — Human Soldier", caparoctiOracle, false)
	a := b16Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b16Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	hit := discoverLibrary(me, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(capa, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
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
	if !b16Tapped(t, g, a) || !b16Tapped(t, g, b) {
		t.Errorf("the two creatures were not tapped")
	}
	expectDiscovered(t, g, me.ID, hit)
}

func TestEllieAndAlanDiscoverTheExiledCardsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	ea := pushCatalogPermanent(g, me.ID, "Ellie and Alan, Paleontologists", "Legendary Creature — Human Scientist", ellieAndAlanOracle, false)
	fossil := game.Card{InstanceID: uuid.New(), Name: "Fossil", TypeLine: "Creature — Dinosaur", ManaCost: "{4}{G}", Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(fossil)
	hit := discoverLibrary(me, 5)
	b16Activate(t, g, me.ID, ea, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{fossil.InstanceID}})
	if !g.Exile.Contains(fossil.InstanceID) {
		t.Errorf("the creature card was not exiled as the cost")
	}
	expectDiscovered(t, g, me.ID, hit)
}

func TestLongRangeSensorChargesOnAttackAndDiscoversFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sensor := pushCatalogPermanent(g, me.ID, "Long-Range Sensor", "Artifact", longRangeSensorOracle, false)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, sensor, "charge"); got != 1 {
		t.Fatalf("charge counters after attacking = %d, want 1", got)
	}

	g = newCatalogGame(t)
	me = g.Seats[0]
	advanceToMain(t, g)
	sensor = pushCatalogPermanent(g, me.ID, "Long-Range Sensor", "Artifact", longRangeSensorOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(sensor, "charge", 2) })
	hit := discoverLibrary(me, 4)
	b16Activate(t, g, me.ID, sensor, 0, game.ActivateAbilityParams{})
	if got := counterOn(g, sensor, "charge"); got != 0 {
		t.Errorf("charge counters after activating = %d, want 0", got)
	}
	expectDiscovered(t, g, me.ID, hit)
}

func TestBuriedTreasureDiscoversFiveFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hit := discoverLibrary(me, 5)
	id, _ := activateFromGraveyard(t, g, "Buried Treasure", "Artifact — Treasure", buriedTreasureOracle, 0, 0, "", game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Errorf("Buried Treasure did not exile itself")
	}
	expectDiscovered(t, g, me.ID, hit)
}

func TestQuintoriusKand(t *testing.T) {
	t.Run("-3 discovers four", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		q := pushCatalogPermanent(g, me.ID, "Quintorius Kand", "Legendary Planeswalker — Quintorius", quintoriusKandOracle, false)
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(q, game.CounterLoyalty, 4) })
		hit := discoverLibrary(me, 4)
		b16Activate(t, g, me.ID, q, 1, game.ActivateAbilityParams{})
		expectDiscovered(t, g, me.ID, hit)
	})
	t.Run("+1 makes a 3/2 Spirit", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		q := pushCatalogPermanent(g, me.ID, "Quintorius Kand", "Legendary Planeswalker — Quintorius", quintoriusKandOracle, false)
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(q, game.CounterLoyalty, 4) })
		b16Activate(t, g, me.ID, q, 0, game.ActivateAbilityParams{})
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
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		q := pushCatalogPermanent(g, me.ID, "Quintorius Kand", "Legendary Planeswalker — Quintorius", quintoriusKandOracle, false)
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(q, game.CounterLoyalty, 6) })
		a := game.Card{InstanceID: uuid.New(), Name: "Shock", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID}
		b := game.Card{InstanceID: uuid.New(), Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID}
		me.Graveyard.PushTop(a)
		me.Graveyard.PushTop(b)
		b16Activate(t, g, me.ID, q, 2, game.ActivateAbilityParams{Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: a.InstanceID}, {Kind: game.TargetCard, ID: b.InstanceID},
		}})
		if !g.Exile.Contains(a.InstanceID) || !g.Exile.Contains(b.InstanceID) {
			t.Fatalf("the cards were not exiled")
		}
		red := 0
		for _, tok := range me.ManaPool {
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
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		q := pushCatalogPermanent(g, me.ID, "Quintorius Kand", "Legendary Planeswalker — Quintorius", quintoriusKandOracle, false)
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(q, game.CounterLoyalty, 4) })
		advanceToMain(t, g)
		spell := game.Card{InstanceID: uuid.New(), Name: "Exiled Sorcery", TypeLine: "Sorcery", ManaCost: "{1}", Owner: me.ID, Controller: me.ID}
		g.Exile.PushTop(spell)
		g.WithWriteLock(func() {
			g.GrantCastPermissionToCardsForEffect(game.CastPermission{Player: me.ID, Zone: game.ZoneExile, Cost: "{0}"}, []game.Card{spell})
		})
		life, oppLife := me.Life, opp.Life
		if err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile)}); err != nil {
			t.Fatalf("cast from exile: %v", err)
		}
		passPriorityAroundTable(t, g)
		if opp.Life != oppLife-2 || me.Life != life+2 {
			t.Errorf("life: me %d→%d, opp %d→%d; want +2 / -2", life, me.Life, oppLife, opp.Life)
		}
	})
}

func TestSwashbucklersWhipGrantsReachAndItsTwoAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	whip := pushCatalogPermanent(g, me.ID, "Swashbuckler's Whip", "Artifact — Equipment", swashbucklersWhipOracle, false)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b16Activate(t, g, me.ID, whip, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if !hasEffectiveKeyword(t, g, bear, "reach") {
		t.Fatalf("the equipped creature lacks reach")
	}
	// Granted ability 0: tap target artifact or creature.
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b16Activate(t, g, me.ID, bear, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}})
	if !b16Tapped(t, g, theirs) {
		t.Errorf("the tap ability did not tap its target")
	}
	// Granted ability 1: discover 10. Untap the bear first.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(bear) })
	hit := discoverLibrary(me, 10)
	b16Activate(t, g, me.ID, bear, 1, game.ActivateAbilityParams{})
	expectDiscovered(t, g, me.ID, hit)
}

func TestZoeticGlyphAnimatesAndDiscoversWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	relic := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Old Relic", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Zoetic Glyph", "Enchantment — Aura", zoeticGlyphOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: relic}})
	passPriorityAroundTable(t, g)
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
	hit := discoverLibrary(me, 3)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(glyph); err != nil {
			t.Fatalf("destroy the Aura: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	expectDiscovered(t, g, me.ID, hit)
}
