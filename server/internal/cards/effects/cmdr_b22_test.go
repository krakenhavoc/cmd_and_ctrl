package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b22_test.go — Commander deck cards, batch 22 (#5368).

const (
	cmdrB22SalvageOracle   = "a89ebf63-63c9-42cd-98b4-ec852cd9b0a8"
	cmdrB22HarbingerOracle = "3256029f-6558-4fcc-9fa1-74ceba3e5c92"
	cmdrB22KayaOracle      = "e0edb2fc-2533-407c-85e5-4337f4233f05"
	cmdrB22MerenOracle     = "0310dfdb-0498-488d-9f1a-279f8239e024"
	cmdrB22SarkhanOracle   = "bbcae463-ba05-4cfe-9c25-13f2e64698c7"
	cmdrB22SidarOracle     = "331b6834-6bb4-40fa-b80b-22d67b34734f"
	cmdrB22ThalisseOracle  = "77cf4fb1-593f-4240-96d9-6ae9084baeac"
	cmdrB22VelomachusOracl = "43832745-5d7d-4957-8e30-2072ad012564"
	cmdrB22UrnOracle       = "920c4df3-9517-44cd-8304-28b1ef69d60e"
)

func TestCmdrB22CardsAreFull(t *testing.T) {
	for oracle, name := range map[string]string{
		cmdrB22SalvageOracle:   "Grisly Salvage",
		cmdrB22HarbingerOracle: "Harbinger of the Hunt",
		cmdrB22KayaOracle:      "Kaya, Geist Hunter",
		cmdrB22MerenOracle:     "Meren of Clan Nel Toth",
		cmdrB22SarkhanOracle:   "Sarkhan, Soul Aflame",
		cmdrB22SidarOracle:     "Sidar Kondo of Jamuraa",
		cmdrB22ThalisseOracle:  "Thalisse, Reverent Medium",
		cmdrB22VelomachusOracl: "Velomachus Lorehold",
		cmdrB22UrnOracle:       "Colfenor's Urn",
	} {
		cmdrB21AssertFull(t, oracle, name)
	}
}

func TestCmdrB22GrislySalvageTakesACreatureOrLandAndBinsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := trekB48Library(me,
		game.Card{Name: "Spell A", TypeLine: "Sorcery"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Spell B", TypeLine: "Instant"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Spell C", TypeLine: "Instant"},
		game.Card{Name: "Sixth", TypeLine: "Instant"},
	)
	castFromHand(t, g, game.Card{Name: "Grisly Salvage", TypeLine: "Instant", ManaCost: "{B}{G}", OracleID: cmdrB22SalvageOracle}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	ch := latestChooseCards(g, me.ID)
	if ch == nil || ch.ChooseMin != 0 || ch.ChooseMax != 1 {
		t.Fatalf("want an optional pick of one card: %+v", ch)
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{lib[0]}); err == nil {
		t.Error("a sorcery was accepted")
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{lib[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(lib[1]) {
		t.Error("the chosen creature is not in hand")
	}
	for _, i := range []int{0, 2, 3, 4} {
		if !me.Graveyard.Contains(lib[i]) {
			t.Errorf("revealed card %d is not in the graveyard", i)
		}
	}
	if !me.Library.Contains(lib[5]) {
		t.Error("the sixth card left the library")
	}
}

func TestCmdrB22HarbingerPingsByFlying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	h := b12Push(g, me.ID, "Harbinger of the Hunt", "Creature — Dragon", cmdrB22HarbingerOracle, 5, 3)
	ground := b12Creature(g, opp.ID, "Ground", "Creature — Bear", 2, 3)
	flyer := brCreature(g, opp.ID, "Flyer", "Creature — Bird", 1, 3, "flying")
	advanceToMain(t, g)
	apaMana(me, "C", "C", "R")
	b16Activate(t, g, me.ID, h, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, ground); got != 1 {
		t.Errorf("ground creature has %d damage, want 1", got)
	}
	if got := damageOn(t, g, flyer); got != 0 {
		t.Errorf("flyer has %d damage, want 0", got)
	}
	if got := damageOn(t, g, h); got != 0 {
		t.Errorf("the flying Harbinger took %d from its own ability", got)
	}
	apaMana(me, "C", "C", "G")
	b16Activate(t, g, me.ID, h, 1, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, flyer); got != 1 {
		t.Errorf("flyer has %d damage, want 1", got)
	}
	if got := damageOn(t, g, h); got != 0 {
		t.Errorf("the Harbinger damaged itself with 'each other creature': %d", got)
	}
	if got := damageOn(t, g, ground); got != 1 {
		t.Errorf("ground creature took more damage: %d", got)
	}
}

func TestCmdrB22KayaPlusOneGrantsDeathtouchAndAPlusOneCounterOnAToken(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kaya := pushWalkerForTest(g, me.ID, "Kaya, Geist Hunter", cmdrB22KayaOracle, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Spirit", TypeLine: "Token Creature — Spirit",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, kaya, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("a non-token creature was accepted for the counter")
	}
	b16Activate(t, g, me.ID, kaya, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: token}}})
	passPriorityAroundTable(t, g)
	if loyaltyOf(g, kaya) != 4 {
		t.Errorf("loyalty %d, want 4", loyaltyOf(g, kaya))
	}
	if counterOn(g, token, game.CounterPlusOne) != 1 {
		t.Error("the token got no +1/+1 counter")
	}
	if !hasEffectiveKeyword(t, g, bear, "deathtouch") || !hasEffectiveKeyword(t, g, token, "deathtouch") {
		t.Error("creatures you control should have deathtouch")
	}
	if hasEffectiveKeyword(t, g, theirs, "deathtouch") {
		t.Error("an opponent's creature gained deathtouch")
	}
}

func TestCmdrB22KayaMinusTwoDoublesTokensAndMinusSixExilesGraveyards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kaya := pushWalkerForTest(g, me.ID, "Kaya, Geist Hunter", cmdrB22KayaOracle, 8)
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, kaya, 1, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	if loyaltyOf(g, kaya) != 6 {
		t.Fatalf("loyalty %d, want 6", loyaltyOf(g, kaya))
	}
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, TokenCard("1/1 white Spirit with flying"), 2); err != nil {
			t.Fatal(err)
		}
	})
	if got := countNamed(g, "Spirit"); got != 4 {
		t.Errorf("%d Spirits after creating two with Kaya's −2, want 4", got)
	}

	// The −6 on a later turn: every graveyard card is exiled, a Spirit for each.
	b39NextTurnOf(t, g, 0)
	for i := 0; i < 3; i++ {
		pushGraveyardCardForTest(me, "Mine")
	}
	for i := 0; i < 2; i++ {
		pushGraveyardCardForTest(opp, "Theirs")
	}
	before := countNamed(g, "Spirit")
	inYards := 0
	for _, p := range g.Seats {
		inYards += p.Graveyard.Size()
	}
	b16Activate(t, g, me.ID, kaya, 2, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	if me.Graveyard.Size() != 0 || opp.Graveyard.Size() != 0 {
		t.Errorf("graveyards hold %d and %d cards, want none", me.Graveyard.Size(), opp.Graveyard.Size())
	}
	if got := countNamed(g, "Spirit") - before; got != inYards+1 {
		// Kaya paid her whole loyalty, so she is in the graveyard (CR 704.5i)
		// when the ability resolves and is exiled with the rest.
		t.Errorf("%d new Spirits, want %d (one per exiled card, Kaya's own included)", got, inYards+1)
	}
	if _, still := battlefieldCard(g, kaya); still {
		t.Error("Kaya should be in the graveyard with 0 loyalty")
	}
}

func cmdrB22GraveyardCreature(p *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Test", ManaCost: cost,
		Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID})
	return id
}

func TestCmdrB22MerenCountsOtherCreaturesAndReturnsByExperience(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	meren := b12Push(g, me.ID, "Meren of Clan Nel Toth", "Legendary Creature — Human Shaman", cmdrB22MerenOracle, 3, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := experienceCounters(g, me.ID); got != 1 {
		t.Fatalf("%d experience counters after another creature died, want 1", got)
	}
	cheap := cmdrB22GraveyardCreature(me, "Cheap", "{1}")
	pricey := cmdrB22GraveyardCreature(me, "Pricey", "{3}{G}")
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, cheap); !ok && len(battlefieldIDsNamed(g, "Cheap")) == 0 {
		t.Error("a mana value 1 card with 1 experience counter did not return to the battlefield")
	}
	if me.Hand.Contains(cheap) {
		t.Error("the cheap card went to hand")
	}
	_ = pricey
	_ = meren
}

func TestCmdrB22MerenPutsACardTooExpensiveIntoHandAndIgnoresItsOwnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	meren := b12Push(g, me.ID, "Meren of Clan Nel Toth", "Legendary Creature — Human Shaman", cmdrB22MerenOracle, 3, 4)
	advanceToMain(t, g)
	pricey := cmdrB22GraveyardCreature(me, "Pricey", "{3}{G}")
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, pricey)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(pricey) {
		t.Error("a mana value 4 card with no experience counters did not go to hand")
	}
	_ = meren
	// Meren dying herself earns nothing ("another creature").
	killCreature(t, g, me.ID, meren)
	passPriorityAroundTable(t, g)
	if got := experienceCounters(g, me.ID); got != 0 {
		t.Errorf("Meren's own death gave %d experience counters", got)
	}
}

func TestCmdrB22SarkhanDiscountsDragonsAndCopiesAnEnteringDragon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sarkhan := b12Push(g, me.ID, "Sarkhan, Soul Aflame", "Legendary Creature — Human Shaman", cmdrB22SarkhanOracle, 2, 4)
	advanceToMain(t, g)
	// {2}{R} Dragon costs {1}{R}; a {2}{R} Bear does not get the discount.
	apaMana(me, "R", "C")
	bear, err := castFromHandErr(g, me, "Bear", "Creature — Bear", "{2}{R}", "", game.CastSpellParams{Strict: true})
	if err == nil {
		t.Fatalf("a non-Dragon got the discount (%s)", bear)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	apaMana(me, "R", "C")
	dragon := castFromHandForTest(t, g, me, "Whelp", "Creature — Dragon", "{2}{R}", "", game.CastSpellParams{Strict: true})
	g.WithWriteLock(func() {
		for i := range g.Stack.Cards {
			if g.Stack.Cards[i].InstanceID == dragon {
				g.Stack.Cards[i].Power, g.Stack.Cards[i].Toughness = 4, 4
			}
		}
	})
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	c := effectiveCardForTest(t, g, sarkhan)
	if c.Name != "Sarkhan, Soul Aflame" {
		t.Errorf("name %q, want Sarkhan, Soul Aflame", c.Name)
	}
	if !c.HasSubtype("Dragon") || !c.IsLegendary() {
		t.Errorf("Sarkhan is not a legendary Dragon copy: %q", c.TypeLine)
	}
	if p := effectivePower(t, g, sarkhan); p != 4 {
		t.Errorf("power %d, want 4 (copy of the Dragon)", p)
	}
	// At end of turn he is himself again.
	b39NextTurnOf(t, g, 0)
	if c := effectiveCardForTest(t, g, sarkhan); c.HasSubtype("Dragon") {
		t.Error("Sarkhan is still a Dragon copy next turn")
	}
}

func TestCmdrB22SidarKondoKeepsGroundCreaturesFromBlockingSmallAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sidar := b12Push(g, me.ID, "Sidar Kondo of Jamuraa", "Legendary Creature — Human Knight", cmdrB22SidarOracle, 2, 5)
	big := b12Creature(g, me.ID, "Big", "Creature — Bear", 3, 3)
	ground := b12Creature(g, opp.ID, "Ground", "Creature — Giant", 3, 3)
	flyer := brCreature(g, opp.ID, "Flyer", "Creature — Bird", 2, 2, "flying")
	spider := brCreature(g, opp.ID, "Spider", "Creature — Spider", 2, 4, "reach")
	brAttack(t, g, sidar, big)
	brRefusal(t, g.DeclareBlocker(ground, sidar), game.BlockReasonCantBlockAttacker)
	if err := g.DeclareBlocker(flyer, sidar); err != nil {
		t.Fatalf("a flyer blocks the power 2 attacker: %v", err)
	}
	if err := g.DeclareBlocker(spider, sidar); err != nil {
		t.Fatalf("a reach creature blocks the power 2 attacker: %v", err)
	}
	if err := g.DeclareBlocker(ground, big); err != nil {
		t.Fatalf("a ground creature blocks the power 3 attacker: %v", err)
	}
}

func TestCmdrB22ThalisseMakesSpiritsEqualToTokensCreatedThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Thalisse, Reverent Medium", "Legendary Creature — Human Cleric", cmdrB22ThalisseOracle, 3, 4)
	advanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, TokenCard("2/2 black Zombie"), 3); err != nil {
			t.Fatal(err)
		}
	})
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := countNamed(g, "Spirit"); got != 3 {
		t.Errorf("%d Spirits, want 3", got)
	}
	// A turn with no tokens made makes none (the opponent's end step too).
	b39NextTurnOf(t, g, 0)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := countNamed(g, "Spirit"); got != 3 {
		t.Errorf("%d Spirits after a turn with no tokens, want still 3", got)
	}
}

func TestCmdrB22VelomachusOffersALesserInstantOrSorceryFromTheTopSeven(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	v := b12Push(g, me.ID, "Velomachus Lorehold", "Legendary Creature — Elder Dragon", cmdrB22VelomachusOracl, 5, 5)
	g.WithWriteLock(func() { me.Library.Cards = nil })
	lib := trekB48Library(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}"},
		game.Card{Name: "Small Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: boltOracleCombat},
		game.Card{Name: "Huge Sorcery", TypeLine: "Sorcery", ManaCost: "{4}{R}{R}"},
		game.Card{Name: "Edge Sorcery", TypeLine: "Sorcery", ManaCost: "{3}{R}{R}"},
		game.Card{Name: "Filler 5", TypeLine: "Land"},
		game.Card{Name: "Filler 6", TypeLine: "Land"},
		game.Card{Name: "Filler 7", TypeLine: "Land"},
		game.Card{Name: "Eighth", TypeLine: "Instant", ManaCost: "{R}"},
	)
	declareAttack(t, g, opp.ID, v)
	passPriorityAroundTable(t, g)
	offer := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, lib[1])
	if !hasID(offer.DiscardOptions, lib[3]) {
		t.Error("a mana value 5 sorcery should be offered to a 5-power Velomachus")
	}
	for _, no := range []uuid.UUID{lib[0], lib[2], lib[7]} {
		if hasID(offer.DiscardOptions, no) {
			t.Errorf("card %s was offered though it is not a castable instant or sorcery from the top seven", no)
		}
	}
	declineResolutionCast(t, g, me.ID)
	passPriorityAroundTable(t, g)
	// The seven looked at went to the bottom; the eighth is now on top.
	if n := len(me.Library.Cards); n != 8 || me.Library.Cards[n-1].InstanceID != lib[7] {
		t.Errorf("library not reordered: %d cards, want the eighth on top", n)
	}
}

func TestCmdrB22ColfenorsUrnExilesBigCreaturesAndReturnsThreeAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	urn := pushCatalogPermanent(g, me.ID, "Colfenor's Urn", "Artifact", cmdrB22UrnOracle, false)
	advanceToMain(t, g)
	small := b12Creature(g, me.ID, "Small", "Creature — Bear", 2, 3)
	killCreature(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a toughness 3 creature triggered the Urn")
	}
	var bigs []uuid.UUID
	for i := 0; i < 3; i++ {
		id := b12Creature(g, me.ID, "Big", "Creature — Wall", 0, 4)
		bigs = append(bigs, id)
		killCreature(t, g, me.ID, id)
		passPriorityAroundTable(t, g)
		answerLatestTriggerPrompt(t, g, me.ID, true)
		passPriorityAroundTable(t, g)
		if !g.Exile.Contains(id) {
			t.Fatalf("big creature %d is not in exile", i)
		}
		if i < 2 {
			if _, still := battlefieldCard(g, urn); !still {
				t.Fatal("the Urn left early")
			}
		}
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if _, still := battlefieldCard(g, urn); still {
		t.Error("the Urn was not sacrificed")
	}
	if got := len(battlefieldIDsNamed(g, "Big")); got != 3 {
		t.Errorf("%d big creatures returned, want 3", got)
	}
}
