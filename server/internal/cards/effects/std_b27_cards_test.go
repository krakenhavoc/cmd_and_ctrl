package effects

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b27_cards_test.go — Standard-set batches 27–30 (umbrella #4452).

const (
	stdBoarOracle         = "6fbaa853-48d4-4a3b-b22d-27415cc266d7"
	stdPathmakerOracle    = "ce36a450-68a6-4519-a9f0-0e3b772054eb"
	stdOverflowOracle     = "75f7d9c7-5d52-4bfe-a730-62fef187b8b8"
	stdPartingGustOracle  = "367bb252-41e1-4f98-9d22-442dd332cccf"
	stdSwanOracle         = "0f68ae67-1671-4281-8b6b-3e52fdd4915f"
	stdSplashPortalOracle = "29b537f5-4d9f-40ed-b38b-e8faa2b5af7d"
	stdCoyoteOracle       = "f76a86f3-568a-40b3-8e2c-f84db4141293"
)

func stdRegisteredFull(t *testing.T, oracle, name string) {
	t.Helper()
	spec, ok := Lookup(oracle)
	if !ok {
		t.Fatalf("%s is not registered", name)
	}
	if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
	}
}

func TestStdQuaketuskBoarHasItsThreeKeywords(t *testing.T) {
	stdRegisteredFull(t, stdBoarOracle, "Quaketusk Boar")
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	id := blbB1Push(g, me.ID, "Quaketusk Boar", "Creature — Boar", stdBoarOracle, 4, 3)
	for _, kw := range []string{"reach", "trample", "haste"} {
		if !p1g2HasKeyword(g, id, kw) {
			t.Errorf("Quaketusk Boar lacks %s", kw)
		}
	}
}

func TestStdAlaniasPathmakerImpulsesOneCardUntilTheEndOfYourNextTurn(t *testing.T) {
	stdRegisteredFull(t, stdPathmakerOracle, "Alania's Pathmaker")
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	top := seedLibrary(me, "Top", "Next")
	if _, err := blbHCastWithMana(t, g, "Alania's Pathmaker", "Creature — Otter Wizard", stdPathmakerOracle, "{2}{U}", 2, 3, "U", "U", "U"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(top[0]) || g.Exile.Contains(top[1]) {
		t.Fatal("exactly the top card should be exiled")
	}
	if !permissionLive(g, exiledPermission(g, top[0]), me.ID) {
		t.Error("the exiled card is not playable")
	}
}

func TestStdHoardersOverflowCountsStashAndWheels(t *testing.T) {
	stdRegisteredFull(t, stdOverflowOracle, "Hoarder's Overflow")
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	seedLibrary(me, "A", "B", "C", "D")
	id := blbB1Push(g, me.ID, "Hoarder's Overflow", "Enchantment", stdOverflowOracle, 0, 0)
	addCountersForTest(g, id, "stash", 2)
	b06AddMana(me, "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("the enchantment was not sacrificed")
	}
	if me.Hand.Size() != 2 {
		t.Errorf("hand %d, want 2 (discard the hand, draw two stash counters' worth)", me.Hand.Size())
	}
}

func TestStdPartingGustPromisedExilesForGood(t *testing.T) {
	stdRegisteredFull(t, stdPartingGustOracle, "Parting Gust")
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bear := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	mustCastWithGift(t, g, "Parting Gust", "Instant", stdPartingGustOracle, cardTarget(bear), opp.ID)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the creature was not exiled")
	}
	if n, tapped := countTokensNamed(g, opp.ID, "Fish"); n != 1 || tapped != 1 {
		t.Errorf("the gift: %d Fish, %d tapped, want one tapped Fish", n, tapped)
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Their Bear" {
			t.Fatal("a promised gift must leave the card exiled")
		}
	}
	_ = me
}

func TestStdPartingGustUnpromisedReturnsWithACounterAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	bear := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	mustCastWithGift(t, g, "Parting Gust", "Instant", stdPartingGustOracle, cardTarget(bear), uuid.Nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the creature was not exiled")
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	c := semestersEndCard(t, g, "Their Bear")
	if c.Controller != opp.ID || c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("returned under %s with counters %v, want its owner and one +1/+1", c.Controller, c.Counters)
	}
}

func TestStdPartingGustRefusesATokenTarget(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	tok := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Fish", TypeLine: "Token Creature — Fish",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	_, err := castWithGift(t, g, "Parting Gust", "Instant", stdPartingGustOracle, cardTarget(tok), uuid.Nil)
	if err == nil {
		t.Fatal("Parting Gust targeted a token")
	}
}

func TestStdSalvationSwanBlinksAWingedlessCreatureBackWithFlying(t *testing.T) {
	stdRegisteredFull(t, stdSwanOracle, "Salvation Swan")
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	bear := blbB1Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	if _, err := blbHCastWithMana(t, g, "Salvation Swan", "Creature — Bird Soldier", stdSwanOracle, "", 3, 3); err != nil {
		t.Fatalf("cast the Swan: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
			t.Fatalf("pick: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(bear) {
		t.Fatal("the Bear was not exiled")
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	c := semestersEndCard(t, g, "My Bear")
	if c.Counters[game.CounterFlying] != 1 {
		t.Errorf("the Bear returned with %v, want a flying counter", c.Counters)
	}
}

func TestStdSplashPortalDrawsForATribeMemberOnly(t *testing.T) {
	stdRegisteredFull(t, stdSplashPortalOracle, "Splash Portal")
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Splash Portal", "Instant", stdSplashPortalOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("a Bear drew a card")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not blinked")
	}
	rat := blbB1Push(g, me.ID, "Rat", "Creature — Rat", "", 1, 1)
	hand = me.Hand.Size()
	castCatalogSpell(t, g, "Splash Portal", "Instant", stdSplashPortalOracle, cardTarget(rat))
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("a Rat: hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
}

func TestStdDriftgloomCoyoteGrowsOnlyOnSmallCreatures(t *testing.T) {
	stdRegisteredFull(t, stdCoyoteOracle, "Driftgloom Coyote")
	for _, tc := range []struct {
		power, counters int
	}{{2, 1}, {3, 0}} {
		g := newCatalogGame(t)
		me, opp, _ := blbB1Seats(g)
		victim := blbB1Push(g, opp.ID, "Victim", "Creature — Bear", "", tc.power, 4)
		if _, err := blbHCastWithMana(t, g, "Driftgloom Coyote", "Creature — Elemental Coyote", stdCoyoteOracle, "", 5, 3); err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
		if err := blbHAnswerPlayerOrCard(t, g, me.ID, victim); err != nil {
			t.Fatalf("pick: %v", err)
		}
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(victim) {
			t.Fatalf("power %d: the creature was not exiled", tc.power)
		}
		coyote := findOnBattlefieldByOracle(g, stdCoyoteOracle)
		c, _ := cardOnBattlefield(g, coyote)
		if c.Counters[game.CounterPlusOne] != tc.counters {
			t.Errorf("power %d: Coyote counters %v, want %d", tc.power, c.Counters, tc.counters)
		}
	}
}

func blbHAnswerPlayerOrCard(t *testing.T, g *game.Game, chooser, card uuid.UUID) error {
	t.Helper()
	p := latestPickTarget(g, chooser)
	if p == nil {
		t.Fatal("no target prompt is open")
	}
	return g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: card})
}

func TestStdJollyGerbilsDrawsWhenAGiftIsGiven(t *testing.T) {
	stdRegisteredFull(t, "8e22bf51-1b1b-459b-af78-91188b6e3f09", "Jolly Gerbils")
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Jolly Gerbils", "Creature — Hamster Citizen", "8e22bf51-1b1b-459b-af78-91188b6e3f09", 2, 1)
	seedLibrary(me, "A", "B", "C")
	bear := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	hand := me.Hand.Size()
	mustCastWithGift(t, g, "Parting Gust", "Instant", stdPartingGustOracle, cardTarget(bear), opp.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn for the gift", hand, got)
	}
	// No promise, no gift, no draw.
	bear2 := blbB1Push(g, opp.ID, "Their Bear 2", "Creature — Bear", "", 2, 2)
	hand = me.Hand.Size()
	mustCastWithGift(t, g, "Parting Gust", "Instant", stdPartingGustOracle, cardTarget(bear2), uuid.Nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("an unpromised gift drew: hand %d -> %d", hand, got)
	}
}

func TestStdMahaSetsOpposingBaseToughnessToOne(t *testing.T) {
	stdRegisteredFull(t, "129476ae-9dab-4e04-8e8e-f249afd56ff9", "Maha, Its Feathers Night")
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 5)
	mine := blbB1Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 5)
	blbB1Push(g, me.ID, "Maha, Its Feathers Night", "Legendary Creature — Elemental Bird", "129476ae-9dab-4e04-8e8e-f249afd56ff9", 6, 5)
	var theirT, myT int
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(theirs)
		theirT = c.CurrentToughness()
		c, _ = g.LookupCardForEffect(mine)
		myT = c.CurrentToughness()
	})
	if theirT != 1 || myT != 5 {
		t.Errorf("toughness: opposing %d (want 1), own %d (want 5)", theirT, myT)
	}
}

func TestStdYgraTurnsOtherCreaturesIntoFoodAndGrowsWhenFoodDies(t *testing.T) {
	const oracle = "62511d43-8eaf-405a-9c81-7bd593d3a01e"
	stdRegisteredFull(t, oracle, "Ygra, Eater of All")
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	ygra := blbB1Push(g, me.ID, "Ygra, Eater of All", "Legendary Creature — Cat Elemental", oracle, 6, 6)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	var food, ygraFood bool
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(theirs)
		food = c.HasSubtype("Food") && c.IsArtifact()
		y, _ := g.LookupCardForEffect(ygra)
		ygraFood = y.HasSubtype("Food")
	})
	if !food {
		t.Error("another creature is not a Food artifact")
	}
	if ygraFood {
		t.Error("Ygra herself became a Food")
	}
	// Their creature dying is a Food going to a graveyard: two counters.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	y, _ := cardOnBattlefield(g, ygra)
	if y.Counters[game.CounterPlusOne] != 2 {
		t.Errorf("Ygra counters %v, want two +1/+1", y.Counters)
	}
}

func TestStdEnduringCouragePumpsAndHastesTheNewcomer(t *testing.T) {
	const oracle = "b3de29cb-7d7d-4566-b5bf-37e571bd5d78"
	stdRegisteredFull(t, oracle, "Enduring Courage")
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	blbB1Push(g, me.ID, "Enduring Courage", "Enchantment Creature — Dog Glimmer", oracle, 3, 3)
	if _, err := blbHCastWithMana(t, g, "Bear", "Creature — Bear", "", "", 2, 2); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	id := findOnBattlefieldByName(g, "Bear")
	var p int
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(id)
		p = c.CurrentPower()
	})
	if p != 4 || !p1g2HasKeyword(g, id, "haste") {
		t.Errorf("Bear power %d haste %v, want 4 and haste", p, p1g2HasKeyword(g, id, "haste"))
	}
}

func findOnBattlefieldByName(g *game.Game, name string) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

func TestStdSeasonOfWeavingBouncesNontokenNonlands(t *testing.T) {
	const oracle = "4b473a42-a0c5-43d5-93f9-c599e963a06b"
	stdRegisteredFull(t, oracle, "Season of Weaving")
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	mine := blbB1Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	theirs := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	tok := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Fish", TypeLine: "Token Creature — Fish",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	castCatalogSpellWithModes(t, g, "Season of Weaving", "Sorcery", oracle, []int{2})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("the nontoken creatures were not returned")
	}
	if !g.Battlefield.Contains(tok) {
		t.Error("the token was returned")
	}
}

func TestStdOtterballAnticsCountersOnlyWhenNotCastFromHand(t *testing.T) {
	const oracle = "43a6de3e-152e-4af1-acec-53405742425c"
	stdRegisteredFull(t, oracle, "Otterball Antics")
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Otterball Antics", "Instant", oracle, nil)
	passPriorityAroundTable(t, g)
	for _, c := range tokensNamed(g, "Otter") {
		if c.Counters[game.CounterPlusOne] != 0 {
			t.Errorf("a hand cast put counters on the Otter: %v", c.Counters)
		}
	}
	if len(tokensNamed(g, "Otter")) != 1 {
		t.Errorf("want one Otter token, have %d", len(tokensNamed(g, "Otter")))
	}
}

func TestStdCoilingRebirthCopiesANonlegendaryAsAOneOne(t *testing.T) {
	const oracle = "27a91ebd-39f4-432d-b6cc-cd86b4a3979a"
	stdRegisteredFull(t, oracle, "Coiling Rebirth")
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	seedLibrary(opp, "A", "B")
	gy := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: gy, Name: "Big Bear", TypeLine: "Creature — Bear", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	mustCastWithGift(t, g, "Coiling Rebirth", "Sorcery", oracle, cardTarget(gy), opp.ID)
	passPriorityAroundTable(t, g)
	bears := 0
	var tokenOneOne bool
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Big Bear" {
			bears++
			if c.IsToken() {
				tokenOneOne = c.Power == 1 && c.Toughness == 1
			}
		}
	}
	if bears != 2 || !tokenOneOne {
		t.Errorf("battlefield has %d Big Bears (want the card and a 1/1 token copy), token 1/1: %v", bears, tokenOneOne)
	}
}

func TestStdHeirloomEpicLetsCreaturesPayTheMana(t *testing.T) {
	const oracle = "43037ef5-09e8-4881-beff-ed38cf7574c6"
	stdRegisteredFull(t, oracle, "Heirloom Epic")
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	epic := blbB1Push(g, me.ID, "Heirloom Epic", "Artifact", oracle, 0, 0)
	var crew []uuid.UUID
	for i := 0; i < 4; i++ {
		crew = append(crew, blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2))
	}
	advanceToMain(t, g)
	hand := me.Hand.Size()
	// Three creatures cannot pay {4}.
	err := g.ActivateCatalogAbility(me.ID, epic, 0, game.ActivateAbilityParams{Strict: true, WaterbendIDs: crew[:3]})
	if err == nil {
		t.Fatal("three tapped creatures paid for {4}")
	}
	if err := g.ActivateCatalogAbility(me.ID, epic, 0, game.ActivateAbilityParams{Strict: true, WaterbendIDs: crew}); err != nil {
		t.Fatalf("activate with four creatures: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want a card drawn", hand, me.Hand.Size())
	}
	for _, id := range crew {
		if c, _ := cardOnBattlefield(g, id); !c.Tapped {
			t.Error("a creature that paid is untapped")
		}
	}
}

func TestStdEssenceChanneler(t *testing.T) {
	const oracle = "1473580d-7cb2-48f5-a77d-08f49244c5e8"
	stdRegisteredFull(t, oracle, "Essence Channeler")
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	ch := blbB1Push(g, me.ID, "Essence Channeler", "Creature — Bat Cleric", oracle, 2, 1)
	ally := blbB1Push(g, me.ID, "Ally", "Creature — Bear", "", 2, 2)
	if p1g2HasKeyword(g, ch, "flying") {
		t.Fatal("flying before any life was lost")
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -1) })
	if !p1g2HasKeyword(g, ch, "flying") || !p1g2HasKeyword(g, ch, "vigilance") {
		t.Error("no flying and vigilance after losing life")
	}
	gainLifeNow(t, g, uuid.Nil, me, 2)
	passPriorityAroundTable(t, g)
	if c, _ := cardOnBattlefield(g, ch); c.Counters[game.CounterPlusOne] != 1 {
		t.Fatalf("counters after gaining life: %v", c.Counters)
	}
	killCreature(t, g, me.ID, ch)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: ally}); err != nil {
			t.Fatalf("pick: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if c, _ := cardOnBattlefield(g, ally); c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("the Channeler's counter did not move: %v", c.Counters)
	}
}

const stdFomoOracle = "236ce592-1321-4a7c-b695-fd84496f8924"

func TestStdFearOfMissingOutLoots(t *testing.T) {
	stdRegisteredFull(t, stdFomoOracle, "Fear of Missing Out")
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	seedLibrary(me, "A", "B", "C")
	fodder := handCard(me, "Fodder", "Instant")
	before := me.Hand.Size()
	if _, err := blbHCastWithMana(t, g, "Fear of Missing Out", "Enchantment Creature — Nightmare", stdFomoOracle, "", 2, 3); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, fodder)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(fodder) {
		t.Error("the chosen card was not discarded")
	}
	if me.Hand.Size() != before {
		t.Errorf("hand %d, want %d (the discard, then a draw)", me.Hand.Size(), before)
	}
}

func TestStdFearOfMissingOutDeliriumUntapsAndAddsACombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	for i, tl := range []string{"Creature", "Instant", "Land", "Artifact"} {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: fmt.Sprintf("G%d", i), TypeLine: tl, Owner: me.ID, Controller: me.ID})
	}
	fomo := pushDiesCreatureForTest(g, me.ID, "Fear of Missing Out", stdFomoOracle, "Enchantment Creature — Nightmare", 2, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(fomo, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: fomo}); err != nil {
			t.Fatalf("pick: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if isTapped(g, fomo) {
		t.Error("the target creature was not untapped")
	}
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepBeginCombat {
		t.Fatalf("after the first combat: %+v, want an additional combat", next)
	}
}

func TestStdFearOfMissingOutNeedsFourCardTypes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	for i, tl := range []string{"Creature", "Instant", "Land"} {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: fmt.Sprintf("G%d", i), TypeLine: tl, Owner: me.ID, Controller: me.ID})
	}
	fomo := pushDiesCreatureForTest(g, me.ID, "Fear of Missing Out", stdFomoOracle, "Enchantment Creature — Nightmare", 2, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(fomo, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatal("the trigger went on the stack with three card types")
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step == game.StepBeginCombat {
		t.Error("an additional combat was added without delirium")
	}
}
