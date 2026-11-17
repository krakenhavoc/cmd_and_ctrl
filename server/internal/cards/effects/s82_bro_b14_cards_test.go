package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bro_b14_cards_test.go — The Brothers' War, batch 14 (Soldiers II).

const (
	broB14SkystrikeOracle   = "862798c6-1c38-4bb9-b4a1-3ef1fcf8f7f7"
	broB14ArchimandriteOr   = "e5e8368d-43af-4f83-8565-25d03da64360"
	broB14HarbinOracle      = "d827575f-e364-4a20-887e-3091e5501a4f"
	broB14MyrelOracle       = "503094aa-6f8f-42c1-ad5a-9305934e5f0c"
	broB14RecruitmentOracle = "97dfc025-afc2-42d8-a20a-f29d60754a2e"
	broB14SurvivorOracle    = "ca7a25f4-c280-48b8-9e04-b15f4c13e569"
	broB14HajarOracle       = "63584001-7f58-42ee-b072-914baa66c78a"
	broB14VanguardOracle    = "e60b23ba-2258-4da1-b873-452ee215846b"
	broB14StalwartOracle    = "ce291948-125d-493d-8770-8b3aac515f9b"
)

func TestBroB14EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		broB14SkystrikeOracle:   "Skystrike Officer",
		broB14ArchimandriteOr:   "The Archimandrite",
		broB14HarbinOracle:      "Harbin, Vanguard Aviator",
		broB14MyrelOracle:       "Myrel, Shield of Argive",
		broB14RecruitmentOracle: "Recruitment Officer",
		broB14SurvivorOracle:    "Survivor of Korlis",
		broB14HajarOracle:       "Hajar, Loyal Bodyguard",
		broB14VanguardOracle:    "Fallaji Vanguard",
		broB14StalwartOracle:    "Citanul Stalwart",
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

func broB14Soldiers(g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		ids = append(ids, b12Creature(g, owner, "Soldier", "Creature — Human Soldier", 1, 1))
	}
	return ids
}

// --- Skystrike Officer ------------------------------------------------------

func TestBroB14SkystrikeOfficerAttackMakesASoldierToken(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	officer := b12Push(g, me.ID, "Skystrike Officer", "Creature — Human Soldier", broB14SkystrikeOracle, 2, 3)
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, officer)
	passPriorityAroundTable(t, g)
	n, _ := countTokensNamed(g, me.ID, "Soldier")
	if n != 1 {
		t.Fatalf("%d Soldier tokens, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if IsToken(c) && c.Controller == me.ID && (!c.IsArtifact() || !c.IsCreature() || len(c.Colors) != 0) {
			t.Errorf("token %q is not a colorless artifact creature", c.TypeLine)
		}
	}
}

func TestBroB14SkystrikeOfficerTapsThreeSoldiersToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	officer := b12Push(g, me.ID, "Skystrike Officer", "Creature — Human Soldier", broB14SkystrikeOracle, 2, 3)
	soldiers := broB14Soldiers(g, me.ID, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	toMainForCost(t, g)
	hand := handCount(g, me)
	if err := g.ActivateCatalogAbility(me.ID, officer, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{soldiers[0], soldiers[1], bear}}); err == nil {
		t.Fatal("a non-Soldier paid the cost")
	}
	assertTapped(t, g, false, officer, soldiers[0], soldiers[1], bear)
	if err := g.ActivateCatalogAbility(me.ID, officer, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{officer, soldiers[0], soldiers[1]}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	assertTapped(t, g, true, officer, soldiers[0], soldiers[1])
	if got := handCount(g, me); got != hand+1 {
		t.Errorf("hand %d → %d, want one card", hand, got)
	}
}

// --- The Archimandrite ------------------------------------------------------

func TestBroB14ArchimandriteGainsLifeAtUpkeepAndPumpsItsKin(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	arch := b12Push(g, me.ID, "The Archimandrite", "Legendary Creature — Human Advisor", broB14ArchimandriteOr, 0, 5)
	monk := b12Creature(g, me.ID, "Monk", "Creature — Human Monk", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToUpkeepOf(t, g, 0)
	d2dSetHand(me, 6)
	start := me.Life
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if me.Life != start+2 {
		t.Fatalf("life %d → %d, want +2", start, me.Life)
	}
	if p := effectivePower(t, g, arch); p != 2 {
		t.Errorf("Archimandrite power %d, want 2", p)
	}
	if p := effectivePower(t, g, monk); p != 3 {
		t.Errorf("Monk power %d, want 3", p)
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("Bear power %d, want 2", p)
	}
	if !hasEffectiveKeyword(t, g, monk, "vigilance") || hasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("vigilance should go to the Monk only")
	}
}

func TestBroB14ArchimandriteGainsNothingWithASmallHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	arch := b12Push(g, me.ID, "The Archimandrite", "Legendary Creature — Human Advisor", broB14ArchimandriteOr, 0, 5)
	advanceToUpkeepOf(t, g, 0)
	d2dSetHand(me, 3)
	start := me.Life
	passPriorityAroundTable(t, g)
	if me.Life != start || effectivePower(t, g, arch) != 0 {
		t.Errorf("life %d → %d, power %d; want no change", start, me.Life, effectivePower(t, g, arch))
	}
}

func TestBroB14ArchimandriteTapsThreeKinToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	arch := b12Push(g, me.ID, "The Archimandrite", "Legendary Creature — Human Advisor", broB14ArchimandriteOr, 0, 5)
	monk := b12Creature(g, me.ID, "Monk", "Creature — Human Monk", 1, 1)
	smith := b12Creature(g, me.ID, "Smith", "Creature — Human Artificer", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	toMainForCost(t, g)
	hand := handCount(g, me)
	if err := g.ActivateCatalogAbility(me.ID, arch, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{arch, monk, bear}}); err == nil {
		t.Fatal("a Bear paid the cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, arch, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{arch, monk, smith}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := handCount(g, me); got != hand+1 {
		t.Errorf("hand %d → %d, want one card", hand, got)
	}
}

// --- Harbin, Vanguard Aviator -----------------------------------------------

func TestBroB14HarbinTriggersOnFiveSoldiersOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	harbin := b12Push(g, me.ID, "Harbin, Vanguard Aviator", "Legendary Creature — Human Soldier", broB14HarbinOracle, 3, 2)
	soldiers := broB14Soldiers(g, me.ID, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, append([]uuid.UUID{harbin}, soldiers...)...)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 3 {
		t.Errorf("Bear power %d, want 3 (+1/+1)", p)
	}
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the Bear did not gain flying")
	}
}

func TestBroB14HarbinIgnoresFourSoldiersAndANonSoldier(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	harbin := b12Push(g, me.ID, "Harbin, Vanguard Aviator", "Legendary Creature — Human Soldier", broB14HarbinOracle, 3, 2)
	soldiers := broB14Soldiers(g, me.ID, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, append([]uuid.UUID{harbin, bear}, soldiers...)...)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("Bear power %d, want 2: four Soldiers and a Bear attacked", p)
	}
}

// --- Myrel, Shield of Argive ------------------------------------------------

func TestBroB14MyrelLocksOutOpponentsDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedPermanentWithOracle(g, me.ID, "Myrel, Shield of Argive", "Legendary Creature — Human Soldier", broB14MyrelOracle)
	signet := pushCatalogPermanent(g, opp.ID, "Boros Signet", "Artifact", "41c84665-1f99-40ab-aaca-1188649eb263", false)
	toMainForCost(t, g)
	bolt := handCardFull(opp, "Their Bolt", "Instant", "{R}", "", []string{"R"})
	if err := g.AddManaForEffect(opp.ID, uuid.Nil, "{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("an opponent cast a spell during my turn")
	}
	if err := g.AddManaForEffect(opp.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateManaAbility(opp.ID, signet, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("an opponent activated an artifact during my turn")
	}
}

func TestBroB14MyrelAttackMakesATokenPerSoldier(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myrel := b12Push(g, me.ID, "Myrel, Shield of Argive", "Legendary Creature — Human Soldier", broB14MyrelOracle, 3, 4)
	broB14Soldiers(g, me.ID, 2)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, myrel)
	passPriorityAroundTable(t, g)
	// Myrel and two Soldiers: X = 3.
	if n, _ := countTokensNamed(g, me.ID, "Soldier"); n != 3 {
		t.Errorf("%d Soldier tokens, want 3", n)
	}
}

// --- Recruitment Officer ----------------------------------------------------

func TestBroB14RecruitmentOfficerTakesACheapCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	officer := b12Push(g, me.ID, "Recruitment Officer", "Creature — Human Soldier", broB14RecruitmentOracle, 2, 1)
	toMainForCost(t, g)
	me.Library.Cards = nil
	mk := func(name, typeLine, cost string) uuid.UUID {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: me.ID, Controller: me.ID})
		return id
	}
	deep := mk("Deep Cheap", "Creature — Bear", "{1}")
	land := mk("Forest", "Basic Land — Forest", "")
	big := mk("Big", "Creature — Giant", "{4}{G}")
	cheap := mk("Cheap", "Creature — Bear", "{1}{G}")
	spell := mk("Bolt", "Instant", "{R}")
	floatMana(t, g, me, "{C}{C}{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, officer, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != cheap {
		t.Fatalf("offered %v, want only the cheap creature", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(cheap) {
		t.Error("the creature is not in hand")
	}
	for _, id := range []uuid.UUID{land, big, spell} {
		if !me.Library.Contains(id) {
			t.Errorf("card %s should be on the library's bottom", id)
		}
	}
	if n := len(me.Library.Cards); n != 4 || me.Library.Cards[n-1].InstanceID != deep {
		t.Errorf("the fifth card should stay on top of the bottomed three; %d cards", n)
	}
}

// --- Survivor of Korlis -----------------------------------------------------

func TestBroB14SurvivorOfKorlisScriesFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	card := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: card, Name: "Survivor of Korlis", TypeLine: "Creature — Human Soldier",
		OracleID: broB14SurvivorOracle, Owner: me.ID, Controller: me.ID})
	seedLibrary(me, "A", "B", "C")
	floatMana(t, g, me, "{C}{W}")
	b16Activate(t, g, me.ID, card, 0, game.ActivateAbilityParams{})
	if !g.Exile.Contains(card) {
		t.Fatal("the card was not exiled to pay the cost")
	}
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("scry prompt %+v, want two cards", c)
	}
}

// --- Hajar, Loyal Bodyguard -------------------------------------------------

func TestBroB14HajarSacrificeProtectsLegendaryCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hajar := b12Push(g, me.ID, "Hajar, Loyal Bodyguard", "Legendary Creature — Human Soldier", broB14HajarOracle, 3, 3)
	legend := b12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	toMainForCost(t, g)
	if err := g.ActivateCatalogAbility(me.ID, hajar, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(hajar) {
		t.Error("Hajar was not sacrificed")
	}
	if p := effectivePower(t, g, legend); p != 3 {
		t.Errorf("Legend power %d, want 3", p)
	}
	if !hasEffectiveKeyword(t, g, legend, "indestructible") {
		t.Error("the legendary creature is not indestructible")
	}
	if effectivePower(t, g, bear) != 2 || hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("a nonlegendary creature was affected")
	}
}

// --- Fallaji Vanguard -------------------------------------------------------

func TestBroB14FallajiVanguardPumpsOnEachCreatureEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vanguard := b12Push(g, me.ID, "Fallaji Vanguard", "Creature — Human Soldier", broB14VanguardOracle, 2, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	toMainForCost(t, g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 colorless Soldier artifact"), 1) })
	pickTriggerTarget(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 4 {
		t.Errorf("Bear power %d, want 4", p)
	}
	if p := effectivePower(t, g, vanguard); p != 2 {
		t.Errorf("Vanguard power %d, want 2", p)
	}
}

func TestBroB14FallajiVanguardTriggersOnItsOwnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Fallaji Vanguard", "Creature — Human Soldier", broB14VanguardOracle, 2, 3)
	pickTriggerTarget(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 4 {
		t.Errorf("Bear power %d, want 4", p)
	}
}

// --- Citanul Stalwart -------------------------------------------------------

func TestBroB14CitanulStalwartTapsAnotherArtifactOrCreatureForMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	stalwart := b12Push(g, me.ID, "Citanul Stalwart", "Creature — Elf Druid Soldier", broB14StalwartOracle, 1, 1)
	rock := pushCatalogPermanent(g, me.ID, "My Rock", "Artifact", "", false)
	land := pushCatalogPermanent(g, me.ID, "My Land", "Land", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	for name, ids := range map[string][]uuid.UUID{
		"nothing else":           nil,
		"itself":                 {stalwart},
		"a land":                 {land},
		"an opponent's artifact": {theirs},
	} {
		if err := g.ActivateManaAbility(me.ID, stalwart, 0, game.ManaAbilityParams{TapIDs: ids}); err == nil {
			t.Fatalf("%s paid the cost", name)
		}
	}
	activateManaFor(t, g, me.ID, stalwart, 0, game.ManaAbilityParams{TapIDs: []uuid.UUID{rock}})
	resolveManaPickAnyColor(t, g, me.ID)
	assertTapped(t, g, true, stalwart, rock)
	if me.ManaPool.Total() != 1 {
		t.Errorf("mana pool = %v, want one mana", me.ManaPool)
	}
}
