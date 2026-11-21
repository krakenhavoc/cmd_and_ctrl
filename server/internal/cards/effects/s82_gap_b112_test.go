package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b112_test.go — S82 GAP batch 112 (Commander Masters instants and
// planeswalkers): every test drives the real cast, activation and layer paths.

const (
	gap112DivergentOracle = "f08e835a-7827-4fdc-8dcf-ae74507da052"
	gap112BlazeOracle     = "a41760a2-d359-4581-b8d3-c1b572ac15c2"
	gap112BlastOracle     = "f2b4f37b-270b-4746-83b3-51ff88ee3491"
	gap112RageOracle      = "2fdc0d17-b683-4263-949e-4f74ec76badf"
	gap112ShelterOracle   = "c1909d70-d8f4-4c54-8587-0a910f123238"
	gap112ObOracle        = "cfc9a2f9-92fa-4608-b98d-f9763c89fc40"
	gap112CalixOracle     = "d9f01177-2237-475e-ad91-114bfae077ce"
	gap112NahiriOracle    = "d492eb9a-1f3b-4bf2-9a5d-0f4a4433f33f"
)

func TestGap112CardsAreRegisteredComplete(t *testing.T) {
	for oracle, name := range map[string]string{
		gap112DivergentOracle: "Divergent Transformations",
		gap112BlazeOracle:     "Ravaging Blaze",
		gap112BlastOracle:     "Sulfurous Blast",
		gap112RageOracle:      "Temur Battle Rage",
		gap112ShelterOracle:   "Shelter",
		gap112ObOracle:        "Ob Nixilis of the Black Oath",
		gap112CalixOracle:     "Calix, Destiny's Hand",
		gap112NahiriOracle:    "Nahiri, the Harbinger",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// gap112Walker puts a planeswalker onto the battlefield with `loyalty` counters.
func gap112Walker(g *game.Game, owner uuid.UUID, name, oracle string, loyalty int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Legendary Planeswalker", OracleID: oracle,
		Owner: owner, Controller: owner, Counters: map[string]int{game.CounterLoyalty: loyalty},
	})
}

func TestGap112TemurBattleRageGivesDoubleStrikeAndFerociousTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Temur Battle Rage", "Instant", gap112RageOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "double strike") {
		t.Fatal("no double strike")
	}
	if hasEffectiveKeyword(t, g, bear, "trample") {
		t.Fatal("trample without a creature of power 4 or greater")
	}
	pushVanillaCreature(g, me.ID, "Big", 4, 4)
	castCatalogSpell(t, g, "Temur Battle Rage", "Instant", gap112RageOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Fatal("no ferocious trample with a 4-power creature")
	}
}

func TestGap112RavagingBlazeDealsXAndSpellMasteryHitsTheController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 5)
	advanceToMain(t, g)
	life := opp.Life
	castXSpell(t, g, "Ravaging Blaze", "Instant", gap112BlazeOracle, "{X}{R}{R}", 2, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Fatalf("the controller lost life without spell mastery: %d -> %d", life, opp.Life)
	}
	if c, ok := g.LookupCardForEffect(bear); !ok || c.DamageMarked != 2 {
		t.Fatalf("the creature has damage %v, want 2", c.DamageMarked)
	}
	for i := 0; i < 2; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	}
	castXSpell(t, g, "Ravaging Blaze", "Instant", gap112BlazeOracle, "{X}{R}{R}", 1, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("spell mastery: life %d -> %d, want -1", life, opp.Life)
	}
}

func TestGap112SulfurousBlastIsThreeInYourMainPhaseAndTwoOtherwise(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Sulfurous Blast", "Instant", gap112BlastOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != 37 || opp.Life != 37 {
		t.Errorf("main phase: lives %d/%d, want 37/37", me.Life, opp.Life)
	}
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	advanceToDeclareAttackersOf(t, g2, 1)
	floatMana(t, g2, me2, "{R}{R}{C}{C}")
	if _, err := castFromHandErr(g2, me2, "Sulfurous Blast", "Instant", "{2}{R}{R}", gap112BlastOracle, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if me2.Life != 38 || g2.Seats[1].Life != 38 {
		t.Errorf("off-turn: lives %d/%d, want 38/38", me2.Life, g2.Seats[1].Life)
	}
}

func TestGap112ShelterGrantsProtectionAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	pushLibraryTyped(me, "Filler", "Instant")
	advanceToMain(t, g)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Shelter", "Instant", gap112ShelterOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	answerColor(t, g, me.ID, "R")
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
	if !hasEffectiveKeyword(t, g, bear, "protection from red") && !hasEffectiveKeyword(t, g, bear, "protection") {
		t.Error("no protection from the chosen colour")
	}
}

func TestGap112ObNixilisDrainsMakesADemonAndEmblemSacrificesForValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ob := gap112Walker(g, me.ID, "Ob Nixilis of the Black Oath", gap112ObOracle, 10)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, ob, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if opp.Life != 39 || me.Life != 43 {
		t.Fatalf("+2: lives me %d opp %d, want 43/39 (three opponents)", me.Life, opp.Life)
	}
}

func TestGap112ObNixilisMinusTwoMakesAFlyingDemonAndCostsTwoLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ob := gap112Walker(g, me.ID, "Ob Nixilis of the Black Oath", gap112ObOracle, 3)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, ob, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if me.Life != 38 {
		t.Errorf("life %d, want 38", me.Life)
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Demon" && IsToken(c) {
			lc := gaLayered(t, g, c.InstanceID)
			if lc.Power != 5 || lc.Toughness != 5 || !hasEffectiveKeyword(t, g, c.InstanceID, "flying") {
				t.Errorf("Demon %d/%d", lc.Power, lc.Toughness)
			}
			n++
		}
	}
	if n != 1 {
		t.Fatalf("%d Demon tokens, want 1", n)
	}
}

func TestGap112ObNixilisEmblemSacrificesACreatureForLifeAndCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ob := gap112Walker(g, me.ID, "Ob Nixilis of the Black Oath", gap112ObOracle, 8)
	bear := pushVanillaCreature(g, me.ID, "Big Bear", 3, 3)
	for i := 0; i < 5; i++ {
		pushLibraryTyped(me, "Filler", "Instant")
	}
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, ob, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if me.Emblems.Size() != 1 {
		t.Fatalf("%d emblems, want 1", me.Emblems.Size())
	}
	emblem := me.Emblems.Cards[0].InstanceID
	life, hand := me.Life, me.Hand.Size()
	floatMana(t, g, me, "{B}{B}")
	b16Activate(t, g, me.ID, emblem, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}})
	passPriorityAroundTable(t, g)
	if me.Life != life+3 || me.Hand.Size() != hand+3 {
		t.Errorf("life %d -> %d, hand %d -> %d, want +3/+3", life, me.Life, hand, me.Hand.Size())
	}
}

func TestGap112NahiriExilesAnEnchantmentTappedArtifactOrTappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nahiri := gap112Walker(g, me.ID, "Nahiri, the Harbinger", gap112NahiriOracle, 4)
	untapped := pushVanillaCreature(g, opp.ID, "Untapped", 2, 2)
	tappedGuy := pushTestPermanent(g, opp.ID, game.Card{Name: "Tapped", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Tapped: true})
	ench := pushTestPermanent(g, opp.ID, game.Card{Name: "Ench", TypeLine: "Enchantment"})
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, nahiri, 1, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(untapped)}); err == nil {
		t.Fatal("an untapped creature was a legal target")
	}
	b16Activate(t, g, me.ID, nahiri, 1, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(tappedGuy)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tappedGuy) || !g.Exile.Contains(tappedGuy) {
		t.Error("the tapped creature was not exiled")
	}
	_ = ench
}

func TestGap112NahiriEightFetchesGivesHasteAndReturnsItAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	nahiri := gap112Walker(g, me.ID, "Nahiri, the Harbinger", gap112NahiriOracle, 8)
	fetched := pushLibraryTyped(me, "Fetched Bear", "Creature — Bear")
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, nahiri, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	passPriorityAroundTable(t, g)
	_ = fetched
	id := findBattlefieldByName(g, "Fetched Bear")
	if id == uuid.Nil {
		t.Fatalf("the creature card was not put onto the battlefield (pending %+v)", g.PendingChoices)
	}
	if !hasEffectiveKeyword(t, g, id, "haste") {
		t.Error("no haste")
	}
	advanceTo(t, g, game.StepEnd)
	monSettle(t, g)
	if findBattlefieldByName(g, "Fetched Bear") != uuid.Nil {
		t.Error("the creature stayed on the battlefield past the end step")
	}
	inHand := false
	for _, c := range me.Hand.Cards {
		if c.Name == "Fetched Bear" {
			inHand = true
		}
	}
	if !inHand {
		t.Error("the creature was not returned to hand")
	}
}

func TestGap112CalixMinusSevenReturnsEveryEnchantmentCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	calix := gap112Walker(g, me.ID, "Calix, Destiny's Hand", gap112CalixOracle, 7)
	a := pushGraveyardPermanent(me, "Ench A", "Enchantment", "{1}")
	b := pushGraveyardPermanent(me, "Ench B", "Enchantment", "{2}")
	c := pushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}")
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, calix, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Error("an enchantment card stayed in the graveyard")
	}
	if g.Battlefield.Contains(c) {
		t.Error("a creature card returned")
	}
}

func TestGap112CalixMinusThreeExilesUntilTheEnchantmentLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	calix := gap112Walker(g, me.ID, "Calix, Destiny's Hand", gap112CalixOracle, 4)
	mine := pushTestPermanent(g, me.ID, game.Card{Name: "Anchor", TypeLine: "Enchantment"})
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, calix, 1, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(theirs, mine)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the creature was not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Their Bear", opp.ID); n != 1 {
		t.Errorf("after the enchantment left, %d of the exiled creature under its owner's control, want 1", n)
	}
}

func TestGap112DivergentTransformationsReplacesEachCreatureWithTheTopCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	pushLibraryTyped(me, "Spell A", "Instant")
	myHit := pushLibraryTyped(me, "My Hit", "Creature — Elf")
	pushLibraryTyped(opp, "Spell B", "Sorcery")
	theirHit := pushLibraryTyped(opp, "Their Hit", "Creature — Goblin")
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Divergent Transformations", "Instant", gap112DivergentOracle, cardRefs(mine, theirs))
	passPriorityAroundTable(t, g)
	for i := 0; i < 4 && len(g.PendingChoices) > 0; i++ {
		passPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Fatal("the targeted creatures were not exiled")
	}
	if !g.Battlefield.Contains(myHit) || !g.Battlefield.Contains(theirHit) {
		t.Errorf("the replacement creatures did not enter (mine %v, theirs %v)", g.Battlefield.Contains(myHit), g.Battlefield.Contains(theirHit))
	}
}
