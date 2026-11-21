package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b47_cards_test.go — GAP batch 47 (#5948, S82): Modern Horizons 3
// Commander planeswalkers and sorceries. Each test plays the real catalog
// entry.

const (
	gapB47UginTheIneffable  = "18bdc181-9592-4147-81fb-7f83ce137f70"
	gapB47GarrukApex        = "ef9a8bdc-3361-440d-89f6-0bf5ff7a09e8"
	gapB47NissaSteward      = "fdb96b5c-ecac-4db3-8b9a-24f824149f91"
	gapB47SkitteringInvade  = "cf890abd-b282-4dbb-ac86-9e7caec1c1b8"
	gapB47FindFinality      = "212af747-2ac1-4d33-9456-8b3cc8988f5d"
	gapB47AggressiveBiomanc = "9080a7bb-2b8b-42af-9f5b-a9f374ecb801"
	gapB47MarchFromVelisVel = "25ba6f57-3d50-4c3e-ab55-284e91d684a6"
	gapB47UginsInsight      = "ad7b9db0-e061-4360-9b81-0408d18d971d"
)

func TestGapB47CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB47UginTheIneffable:    "Ugin, the Ineffable",
		gapB47GarrukApex:          "Garruk, Apex Predator",
		gapB47NissaSteward:        "Nissa, Steward of Elements",
		gapB47SkitteringInvade:    "Skittering Invasion",
		gapB47FindFinality:        "Find",
		gapB47FindFinality + "#1": "Finality",
		gapB47AggressiveBiomanc:   "Aggressive Biomancy",
		gapB47MarchFromVelisVel:   "March from Velis Vel",
		gapB47UginsInsight:        "Ugin's Insight",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func gapB47Walker(g *game.Game, owner uuid.UUID, name, sub, oracle string, loyalty int) uuid.UUID {
	return pushCatalogWalkerWithLine(g, owner, name, "Legendary Planeswalker — "+sub, oracle, loyalty)
}

func TestGapB47UginTheIneffablePlusOneMakesASpiritThatReturnsTheExiledCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ugin := gapB47Walker(g, me.ID, "Ugin, the Ineffable", "Ugin", gapB47UginTheIneffable, 4)
	toMain(t, g)
	topID := uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: topID, Name: "Secret", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	})
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, ugin, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	spirits := s46Tokens(g, me.ID, "Spirit")
	if len(spirits) != 1 {
		t.Fatalf("%d Spirits, want 1", len(spirits))
	}
	if p, tg := ptOf(t, g, spirits[0].InstanceID); p != 2 || tg != 2 {
		t.Errorf("the Spirit is %d/%d, want 2/2", p, tg)
	}
	if z := g.FindCardZoneForEffect(topID); z == nil || z.Kind != game.ZoneExile {
		t.Fatalf("the top card is not in exile (%v)", z)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(spirits[0].InstanceID) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want the exiled card in hand", hand, me.Hand.Size())
	}
}

func TestGapB47UginTheIneffableMinusThreeNeedsAColoredPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ugin := gapB47Walker(g, me.ID, "Ugin, the Ineffable", "Ugin", gapB47UginTheIneffable, 4)
	toMain(t, g)
	rock := m3cb1Push(g, opp.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	bear := m3cb1Push(g, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Colors: []string{"G"}, Power: 2, Toughness: 2})
	if err := g.ActivateCatalogAbility(me.ID, ugin, 1, game.ActivateAbilityParams{Targets: cardRefs(rock), Strict: true}); err == nil {
		t.Fatal("targeted a colorless permanent")
	}
	if err := g.ActivateCatalogAbility(me.ID, ugin, 1, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("-3: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the colored permanent was not destroyed")
	}
}

func TestGapB47UginTheIneffableDiscountsColorlessSpells(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	gapB47Walker(g, me.ID, "Ugin, the Ineffable", "Ugin", gapB47UginTheIneffable, 4)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Big Rock", TypeLine: "Artifact", ManaCost: "{4}", Owner: me.ID, Controller: me.ID})
	})
	floatMana(t, g, me, "{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the {4} colorless spell with two mana under Ugin: %v", err)
	}
}

func TestGapB47GarrukApexPredatorPlusOneMakesADeathtouchBeast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	garruk := gapB47Walker(g, me.ID, "Garruk, Apex Predator", "Garruk", gapB47GarrukApex, 5)
	toMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, garruk, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("+1 Beast: %v", err)
	}
	passPriorityAroundTable(t, g)
	beasts := s46Tokens(g, me.ID, "Beast")
	if len(beasts) != 1 || !effectiveAbilitiesContain(t, g, beasts[0].InstanceID, "deathtouch") {
		t.Fatalf("Beast tokens %v, want one with deathtouch", beasts)
	}
	if p, tg := ptOf(t, g, beasts[0].InstanceID); p != 3 || tg != 3 || !beasts[0].HasColor("B") {
		t.Errorf("the Beast is %d/%d, want a black 3/3", p, tg)
	}
}

func TestGapB47GarrukApexPredatorMinusThreeDestroysAndGainsLifeEqualToToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	garruk := gapB47Walker(g, me.ID, "Garruk, Apex Predator", "Garruk", gapB47GarrukApex, 5)
	toMain(t, g)
	ogre := m3cb1Push(g, opp.ID, game.Card{Name: "Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 4})
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, garruk, 2, game.ActivateAbilityParams{Targets: cardRefs(ogre), Strict: true}); err != nil {
		t.Fatalf("-3: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) || me.Life != life+4 {
		t.Errorf("ogre on battlefield %v, life %d -> %d; want it destroyed and +4", g.Battlefield.Contains(ogre), life, me.Life)
	}
}

func TestGapB47GarrukApexPredatorMinusEightGivesAnOpponentTheEmblem(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	garruk := gapB47Walker(g, me.ID, "Garruk, Apex Predator", "Garruk", gapB47GarrukApex, 8)
	toMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, garruk, 3, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Strict: true}); err != nil {
		t.Fatalf("-8: %v", err)
	}
	passPriorityAroundTable(t, g)
	attacker := m3cb1Push(g, me.ID, game.Card{Name: "Squire", TypeLine: "Creature — Human", Power: 1, Toughness: 1})
	declareAttack(t, g, opp.ID, attacker)
	passPriorityAroundTable(t, g)
	if p, tg := ptOf(t, g, attacker); p != 6 || tg != 6 {
		t.Errorf("the attacker is %d/%d, want 6/6 under the emblem", p, tg)
	}
	if !effectiveAbilitiesContain(t, g, attacker, "trample") {
		t.Error("the attacker lacks trample")
	}
}

func TestGapB47NissaStewardOfElementsMinusSixAnimatesTwoLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	nissa := gapB47Walker(g, me.ID, "Nissa, Steward of Elements", "Nissa", gapB47NissaSteward, 6)
	l1 := m3cb1Push(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	l2 := m3cb1Push(g, me.ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, l1).Tapped = true
		findBattlefieldCardForTest(g, l2).Tapped = true
	})
	toMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, nissa, 2, game.ActivateAbilityParams{Targets: cardRefs(l1, l2), Strict: true}); err != nil {
		t.Fatalf("-6: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{l1, l2} {
		c := layeredCard(t, g, id)
		if c.Tapped || !c.IsLand() || !c.IsCreature() || !c.HasSubtypeWord("Elemental") {
			t.Errorf("land %s: tapped %v, %q, want an untapped Elemental creature that is still a land", c.Name, c.Tapped, c.TypeLine)
		}
		if p, tg := ptOf(t, g, id); p != 5 || tg != 5 {
			t.Errorf("%s is %d/%d, want 5/5", c.Name, p, tg)
		}
		if !effectiveAbilitiesContain(t, g, id, "flying") || !effectiveAbilitiesContain(t, g, id, "haste") {
			t.Errorf("%s lacks flying or haste", c.Name)
		}
	}
}

func TestGapB47NissaStewardOfElementsEntersWithXLoyalty(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Nissa, Steward of Elements", TypeLine: "Legendary Planeswalker — Nissa",
			OracleID: gapB47NissaSteward, ManaCost: "{X}{G}{U}", Owner: me.ID, Controller: me.ID})
	})
	floatMana(t, g, me, "{G}{U}{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := loyaltyCount(g, id); got != 3 {
		t.Errorf("loyalty %d, want 3", got)
	}
}

func TestGapB47SkitteringInvasionMakesFiveSpawn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Skittering Invasion", "Kindred Sorcery — Eldrazi", gapB47SkitteringInvade, nil)
	passPriorityAroundTable(t, g)
	if got := len(s46Tokens(g, me.ID, "Eldrazi Spawn")); got != 5 {
		t.Errorf("%d Eldrazi Spawn, want 5", got)
	}
}

func TestGapB47FindReturnsTwoCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	var dead []uuid.UUID
	for _, n := range []string{"A", "B", "C"} {
		id := uuid.New()
		dead = append(dead, id)
		g.WithWriteLock(func() {
			me.Graveyard.PushTop(game.Card{InstanceID: id, Name: n, TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		})
	}
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Find", "Sorcery", gapB47FindFinality, cardRefs(dead[0], dead[1]))
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d -> %d, want +2", hand, got)
	}
}

func TestGapB47FinalityPutsCountersThenShrinksEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], m3cb1Opponent(g)
	big := m3cb1Push(g, me.ID, game.Card{Name: "Big", TypeLine: "Creature — Giant", Power: 5, Toughness: 5})
	small := m3cb1Push(g, opp.ID, game.Card{Name: "Small", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	castCatalogSpell(t, g, "Finality", "Sorcery", gapB47FindFinality+"#1", nil)
	passPriorityAroundTable(t, g)
	answerPermanentsPick(t, g, me.ID, big)
	passPriorityAroundTable(t, g)
	if p, tg := ptOf(t, g, big); p != 3 || tg != 3 {
		t.Errorf("the Giant is %d/%d, want 3/3 (two counters, then -4/-4)", p, tg)
	}
	if g.Battlefield.Contains(small) {
		t.Error("the 2/2 survived -4/-4")
	}
}

func TestGapB47AggressiveBiomancyMakesXCopiesThatFight(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], m3cb1Opponent(g)
	bear := m3cb1Push(g, me.ID, game.Card{Name: "Grizzly", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	victim := m3cb1Push(g, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 2})
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Aggressive Biomancy", TypeLine: "Sorcery", OracleID: gapB47AggressiveBiomanc,
			ManaCost: "{X}{X}{G}{U}", Owner: me.ID, Controller: me.ID})
	})
	floatMana(t, g, me, "{G}{U}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(bear), XValue: 1}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerPickTargetIfOpen(t, g, victim)
	passPriorityAroundTable(t, g)
	if got := len(s46Tokens(g, me.ID, "Grizzly")); got != 1 {
		t.Fatalf("%d Grizzly copies, want 1", got)
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the copy did not fight the Wall to death (2 damage on a 0/2)")
	}
}

func TestGapB47UginsInsightScriesTheGreatestManaValueThenDrawsThree(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	m3cb1Push(g, me.ID, game.Card{Name: "Expensive", TypeLine: "Creature — Giant", ManaCost: "{4}", Power: 4, Toughness: 4})
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Ugin's Insight", "Sorcery", gapB47UginsInsight, nil)
	passPriorityAroundTable(t, g)
	scry := latestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID)
	if scry == nil {
		t.Fatal("no scry prompt")
	}
	if len(scry.ScryCards) != 4 {
		t.Errorf("scry %d, want 4 (the greatest mana value among your permanents)", len(scry.ScryCards))
	}
	if err := g.ResolveScry(scry.ID, me.ID, nil, scry.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d -> %d, want +3", hand, got)
	}
}

func TestGapB47MarchFromVelisVelTurnsNonbasicLandsIntoCopiesWithHaste(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := m3cb1Push(g, me.ID, game.Card{Name: "Grizzly", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	desert := m3cb1Push(g, me.ID, game.Card{Name: "Test Desert", TypeLine: "Land — Desert"})
	forest := m3cb1Push(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	castCatalogSpell(t, g, "March from Velis Vel", "Sorcery", gapB47MarchFromVelisVel, cardRefs(bear))
	passPriorityAroundTable(t, g)
	pick := latestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no land type prompt")
	}
	idx := -1
	for i, o := range pick.PickOptions {
		if o.Label == "Desert" {
			idx = i
		}
		if o.Label == "Forest" {
			t.Fatal("a basic land type is offered")
		}
	}
	if idx < 0 {
		t.Fatal("Desert is not offered")
	}
	if err := g.ResolveOptionPick(pick.ID, me.ID, idx); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, desert)
	if c.Name != "Grizzly" || !c.IsCreature() {
		t.Errorf("the Desert is %q %q, want a copy of the Grizzly", c.Name, c.TypeLine)
	}
	if !effectiveAbilitiesContain(t, g, desert, "haste") {
		t.Error("the copy lacks haste")
	}
	if f := layeredCard(t, g, forest); f.Name != "Forest" {
		t.Errorf("the Forest became %q", f.Name)
	}
}

func TestGapB47NissaStewardOfElementsZeroOffersALandOrAFittingCreature(t *testing.T) {
	for _, tc := range []struct {
		name  string
		card  game.Card
		offer bool
	}{
		{"land", game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}, true},
		{"creature that fits", game.Card{Name: "Cub", TypeLine: "Creature — Bear", ManaCost: "{2}", Power: 2, Toughness: 2}, true},
		{"creature too expensive", game.Card{Name: "Giant", TypeLine: "Creature — Giant", ManaCost: "{4}", Power: 4, Toughness: 4}, false},
		{"noncreature spell", game.Card{Name: "Bolt", TypeLine: "Instant", ManaCost: "{R}"}, false},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		nissa := gapB47Walker(g, me.ID, "Nissa, Steward of Elements", "Nissa", gapB47NissaSteward, 3)
		toMain(t, g)
		top := tc.card
		top.InstanceID, top.Owner, top.Controller = uuid.New(), me.ID, me.ID
		g.WithWriteLock(func() { me.Library.PushTop(top) })
		if err := g.ActivateCatalogAbility(me.ID, nissa, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("%s: 0: %v", tc.name, err)
		}
		passPriorityAroundTable(t, g)
		pc := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
		if (pc != nil) != tc.offer {
			t.Errorf("%s: prompt open %v, want %v", tc.name, pc != nil, tc.offer)
			continue
		}
		if pc != nil {
			if err := g.ResolveChooseCards(pc.ID, me.ID, []uuid.UUID{top.InstanceID}); err != nil {
				t.Fatalf("%s: ResolveChooseCards: %v", tc.name, err)
			}
			if !g.Battlefield.Contains(top.InstanceID) {
				t.Errorf("%s: the card did not enter the battlefield", tc.name)
			}
		}
	}
}
