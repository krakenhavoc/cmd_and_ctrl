package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_ring_pool_b_test.go — ADR 0114 PR 5, pool B.

const (
	dunedainRangersOracle        = "7de2a364-4383-437a-b6d6-55e4721a0141"
	faramirFieldCommanderOracle  = "adeed5e8-9019-4740-83ea-67dcd754e194"
	galadrielOfLothlorienOracle  = "8d3c9646-2b6b-4315-939d-5a920efc0a60"
	gandalfFriendOfTheShireOracl = "7f7c917b-f940-4f43-bafa-31f50a6d9d6f"
	gloriousGaleOracle           = "858265c2-6415-4938-a6ec-734d53046100"
	gollumPatientPlotterOracle   = "c4545884-35b7-47cf-939e-3d5d3fe555a4"
	nowForWrathOracle            = "8141fb64-7b55-40fb-8fe4-5efd805d8fef"
	oneRingToRuleThemAllOracle   = "c75add5e-4e9d-4c68-a8ad-b878b1dbc0a4"
	ringwraithsOracle            = "f09278dc-1e67-4cd8-977d-4b3b94430aac"
	slipOnTheRingOracle          = "a6b28298-9624-49d7-ad73-fcd1fd3556d2"
	smeagolHelpfulGuideOracle    = "6b974499-bf1c-4e5b-a72c-c72220f4e591"
	warOfTheLastAllianceOracle   = "e376702a-f56e-4e47-aa8c-3f80384d12a9"
	frodoBagginsOracle           = "d4dd6edf-1e9e-46fa-92b5-5df9aa0e4338"
	theRingGoesSouthOracle       = "3d36f0ab-d333-4a08-ad1b-5b7e1ddb647c"
	inTheDarknessBindThemOracle  = "c8689a83-cefa-4ef3-8ec1-58b4c64e332b"
)

// ringLibraryTop puts a card on top of `p`'s library.
func ringLibraryTop(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// Dúnedain Rangers: landfall tempts only while you control no
// Ring-bearer.
func TestDunedainRangersTemptsOnlyWithoutARingBearer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rangers := pushCatalogPermanent(g, me.ID, "Dúnedain Rangers", "Creature — Human Ranger", dunedainRangersOracle, false)
	b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	ringSettle(t, g)
	if ringCount(g, me.ID) != 1 || ringBearerOf(g, me.ID) != rangers {
		t.Fatalf("first land: tempted %d times, Ring-bearer %v", ringCount(g, me.ID), ringBearerOf(g, me.ID))
	}
	b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	ringSettle(t, g)
	if ringCount(g, me.ID) != 1 {
		t.Fatal("a second land tempted while you had a Ring-bearer")
	}
}

// Faramir: the end-step draw after a creature died, and the token when
// another creature is chosen, not Faramir.
func TestFaramirFieldCommander(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	faramir := pushCatalogPermanent(g, me.ID, "Faramir, Field Commander", "Legendary Creature — Human Soldier", faramirFieldCommanderOracle, false)
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	ringTempt(t, g, me.ID)
	answerRing(t, g, me.ID, faramir)
	ringSettle(t, g)
	if onBattlefieldNamed(g, "Human Soldier") != 0 {
		t.Fatal("choosing Faramir made a token")
	}
	ringTempt(t, g, me.ID)
	answerRing(t, g, me.ID, bear)
	ringSettle(t, g)
	if onBattlefieldNamed(g, "Human Soldier") != 1 {
		t.Fatal("choosing another creature made no token")
	}

	// No creature died: no draw at the end step.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	pushCatalogPermanent(g2, me2.ID, "Faramir, Field Commander", "Legendary Creature — Human Soldier", faramirFieldCommanderOracle, false)
	advanceTo(t, g2, game.StepEnd)
	if triggerOnStackFrom(g2, "Faramir, Field Commander") {
		t.Fatal("the end-step draw triggered with no creature dead")
	}

	victim := b12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	ringSettle(t, g)
	advanceTo(t, g, game.StepEnd)
	hand := me.Hand.Size()
	ringSettle(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d → %d at the end step after a death, want +1", hand, me.Hand.Size())
	}
}

// Galadriel of Lothlórien: scry 3 on another Ring-bearer, then the
// scry trigger reveals a land and puts it onto the battlefield tapped.
func TestGaladrielOfLothlorienScriesAndFindsALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Galadriel of Lothlórien", "Legendary Creature — Elf Noble", galadrielOfLothlorienOracle, false)
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	land := ringLibraryTop(me, "Forest", "Basic Land — Forest")

	ringTempt(t, g, me.ID)
	answerRing(t, g, me.ID, bear)
	ringSettle(t, g)
	answerScryKeepAll(t, g, me.ID)
	ringSettle(t, g)
	answerMayChoice(t, g, me.ID, true)
	ringSettle(t, g)
	c, ok := battlefieldCard(g, land)
	if !ok || !c.Tapped {
		t.Fatal("the revealed land is not on the battlefield tapped")
	}
}

// Gandalf: sorceries at instant speed, and a draw on another
// Ring-bearer.
func TestGandalfFriendOfTheShire(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	// The draw step: a sorcery can't be cast without Gandalf.
	sorcery := b13HandCard(me, "Birthday Escape", "Sorcery", birthdayEscapeOracle)
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{}); err == nil {
		t.Fatal("a sorcery was cast in the draw step without Gandalf")
	}
	gandalf := pushCatalogPermanent(g, me.ID, "Gandalf, Friend of the Shire", "Legendary Creature — Avatar Wizard", gandalfFriendOfTheShireOracl, false)
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{}); err != nil {
		t.Fatalf("with Gandalf the sorcery is cast as though it had flash: %v", err)
	}
	ringSettle(t, g)
	// Birthday Escape tempted with Gandalf the only creature: no draw
	// from Gandalf.
	if ringBearerOf(g, me.ID) != gandalf {
		t.Fatal("Gandalf is not the Ring-bearer")
	}
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	ringTempt(t, g, me.ID)
	answerRing(t, g, me.ID, bear)
	ringSettle(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d → %d after choosing another Ring-bearer, want +1", hand, me.Hand.Size())
	}
}

// Glorious Gale: counters a creature spell, and tempts only when it was
// legendary.
func TestGloriousGaleTemptsOnALegendarySpell(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		tempts   int
	}{
		{"Creature — Bear", 0},
		{"Legendary Creature — Human", 1},
	} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		spell := b13PlayAs(t, g, 1, "Their Creature", tc.typeLine, "")
		if err := g.CastSpell(me.ID, b13HandCard(me, "Glorious Gale", "Instant", gloriousGaleOracle),
			game.CastSpellParams{Targets: b16TargetCard(spell)}); err != nil {
			t.Fatalf("Glorious Gale: %v", err)
		}
		ringSettle(t, g)
		if g.Battlefield.Contains(spell) || !opp.Graveyard.Contains(spell) {
			t.Errorf("%s: the spell was not countered", tc.typeLine)
		}
		if ringCount(g, me.ID) != tc.tempts {
			t.Errorf("%s: tempted %d times, want %d", tc.typeLine, ringCount(g, me.ID), tc.tempts)
		}
	}
}

// Gollum, Patient Plotter: leaving tempts; from the graveyard, {B} and
// a creature return it to hand.
func TestGollumPatientPlotter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gollum := pushCatalogPermanent(g, me.ID, "Gollum, Patient Plotter", "Legendary Creature — Halfling Horror", gollumPatientPlotterOracle, false)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(gollum) })
	ringSettle(t, g)
	if ringCount(g, me.ID) != 1 {
		t.Fatal("Gollum left the battlefield and the Ring did not tempt")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	fodder := b12Creature(g2, me2.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	id, _ := activateFromGraveyard(t, g2, "Gollum, Patient Plotter", "Legendary Creature — Halfling Horror", gollumPatientPlotterOracle, 3, 1, "{B}",
		game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	ringSettle(t, g2)
	if !me2.Hand.Contains(id) || g2.Battlefield.Contains(fodder) {
		t.Fatal("Gollum did not return to hand for {B} and a sacrificed creature")
	}
}

// Now for Wrath, Now for Ruin!: a counter and vigilance on each of your
// creatures, then the tempt.
func TestNowForWrathNowForRuin(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Now for Wrath, Now for Ruin!", "Sorcery", nowForWrathOracle, nil)
	ringSettle(t, g)
	answerRing(t, g, me.ID, a)
	ringSettle(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if ringBFCard(t, g, id).Counters["+1/+1"] != 1 || !containsString(effectiveAbilities(t, g, id), "vigilance") {
			t.Errorf("%s has no counter or no vigilance", ringBFCard(t, g, id).Name)
		}
	}
	if ringBFCard(t, g, theirs).Counters["+1/+1"] != 0 {
		t.Error("an opponent's creature got a counter")
	}
	if ringCount(g, me.ID) != 1 || ringBearerOf(g, me.ID) != a {
		t.Fatal("the Ring did not tempt")
	}
}

// One Ring to Rule Them All: I mills each player by the Ring-bearer's
// power; II destroys every nonlegendary creature, sparing the
// Ring-bearer, which the Ring makes legendary; III drains for creature
// cards.
func TestOneRingToRuleThemAll(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	graves := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		graves[p.ID] = p.Graveyard.Size()
	}
	saga := castCatalogSpell(t, g, "One Ring to Rule Them All", "Enchantment — Saga", oneRingToRuleThemAllOracle, nil)
	ringSettle(t, g)
	if ringBearerOf(g, me.ID) != bear {
		t.Fatal("chapter I: the bear is not the Ring-bearer")
	}
	for _, p := range g.Seats {
		if p.Graveyard.Size() != graves[p.ID]+2 {
			t.Errorf("chapter I: a player milled %d, want 2", p.Graveyard.Size()-graves[p.ID])
		}
	}

	advanceToPrecombatMainOf(t, g, seat)
	ringSettle(t, g)
	if g.Battlefield.Contains(theirs) || !g.Battlefield.Contains(bear) {
		t.Fatal("chapter II: want their bear destroyed and the legendary Ring-bearer spared")
	}

	creatures := 0
	for _, c := range opp.Graveyard.Cards {
		if c.IsCreature() {
			creatures++
		}
	}
	life := opp.Life
	advanceToPrecombatMainOf(t, g, seat)
	ringSettle(t, g)
	if opp.Life != life-creatures {
		t.Fatalf("chapter III: opponent %d → %d, want -%d", life, opp.Life, creatures)
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

// Ringwraiths: -3/-3 and 3 life from a legendary target's controller;
// back to hand from the graveyard when the Ring tempts you.
func TestRingwraiths(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	legend := typedCreature(g, opp.ID, "Their Legend", "Legendary Creature — Human", "", 4, 4)
	life := opp.Life
	wraiths := castRingCreature(t, g, "Ringwraiths", "Creature — Wraith Knight", ringwraithsOracle, 5, 5)
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			answerPickTarget(t, g, legend)
			break
		}
	}
	ringSettle(t, g)
	if p, tt := ptOf(t, g, legend); p != 1 || tt != 1 {
		t.Fatalf("the legend is %d/%d, want 1/1", p, tt)
	}
	if opp.Life != life-3 {
		t.Fatalf("the legend's controller is at %d, want %d", opp.Life, life-3)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(wraiths) })
	ringSettle(t, g)
	ringTempt(t, g, me.ID)
	ringSettle(t, g)
	if !me.Hand.Contains(wraiths) {
		t.Fatal("Ringwraiths did not return to hand when the Ring tempted")
	}
}

// Slip On the Ring: the creature returns as a new object and may be
// chosen.
func TestSlipOnTheRingFlickersThenTempts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, "+1/+1", 1) })
	ringCastSettled(t, g, "Slip On the Ring", "Instant", slipOnTheRingOracle, ringTarget(bear))
	if onBattlefieldNamed(g, "Grizzly Bears") != 1 {
		t.Fatal("the bear did not return")
	}
	rb := ringBearerOf(g, me.ID)
	if rb == uuid.Nil || ringCount(g, me.ID) != 1 {
		t.Fatal("the returned bear was not chosen as the Ring-bearer")
	}
	if ringBFCard(t, g, rb).Counters["+1/+1"] != 0 {
		t.Fatal("the returned bear kept its counter")
	}
}

// Sméagol, Helpful Guide: the opponent reveals until a land, which
// enters tapped under your control; the rest go to their graveyard.
func TestSmeagolHelpfulGuideTakesALand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Sméagol, Helpful Guide", "Legendary Creature — Halfling Horror", smeagolHelpfulGuideOracle, false)
	land := ringLibraryTop(opp, "Their Forest", "Basic Land — Forest")
	x := ringLibraryTop(opp, "Their Spell", "Sorcery")
	y := ringLibraryTop(opp, "Their Other Spell", "Instant")
	ringTempt(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPickTargetPlayer(t, g, opp.ID)
	ringSettle(t, g)
	c, ok := battlefieldCard(g, land)
	if !ok || c.Controller != me.ID || !c.Tapped {
		t.Fatal("the land did not enter tapped under Sméagol's controller")
	}
	if !opp.Graveyard.Contains(x) || !opp.Graveyard.Contains(y) {
		t.Fatal("the rest did not go to their graveyard")
	}
}

// War of the Last Alliance: I finds a legendary creature card.
func TestWarOfTheLastAllianceFindsALegend(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	legend := ringLibraryTop(me, "Some Legend", "Legendary Creature — Elf")
	castCatalogSpell(t, g, "War of the Last Alliance", "Enchantment — Saga", warOfTheLastAllianceOracle, nil)
	ringSettle(t, g)
	if !me.Hand.Contains(legend) {
		t.Fatal("chapter I did not find the legendary creature card")
	}
}

// Frodo Baggins: tempts as it or another legend of yours enters, and
// must be blocked only while it is your Ring-bearer.
func TestFrodoBaggins(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	frodo := castRingCreature(t, g, "Frodo Baggins", "Legendary Creature — Halfling Scout", frodoBagginsOracle, 1, 3)
	ringSettle(t, g)
	if ringCount(g, me.ID) != 1 || ringBearerOf(g, me.ID) != frodo {
		t.Fatal("Frodo entering did not tempt")
	}
	castRingCreature(t, g, "Other Legend", "Legendary Creature — Elf", "", 2, 2)
	ringSettle(t, g)
	answerRing(t, g, me.ID, frodo)
	ringSettle(t, g)
	if ringCount(g, me.ID) != 2 {
		t.Fatal("another legendary creature entering did not tempt")
	}
	castRingCreature(t, g, "A Bear", "Creature — Bear", "", 2, 2)
	ringSettle(t, g)
	if ringCount(g, me.ID) != 2 {
		t.Fatal("a nonlegendary creature entering tempted")
	}
}

func TestFrodoBagginsMustBeBlockedAsRingBearer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	frodo := typedCreature(g, me.ID, "Frodo Baggins", "Legendary Creature — Halfling Scout", frodoBagginsOracle, 1, 3, "haste")
	typedCreature(g, opp.ID, "Wall", "Creature — Wall", "", 0, 4)
	ringTempt(t, g, me.ID)
	reqToBlockers(t, g, frodo)
	reqRefusal(t, g.PassPriority())

	g2 := newCatalogGame(t)
	me2, opp2 := seats1684(g2)
	frodo2 := typedCreature(g2, me2.ID, "Frodo Baggins", "Legendary Creature — Halfling Scout", frodoBagginsOracle, 1, 3, "haste")
	typedCreature(g2, opp2.ID, "Wall", "Creature — Wall", "", 0, 4)
	reqToBlockers(t, g2, frodo2)
	if err := g2.PassPriority(); err != nil {
		t.Fatalf("Frodo is not the Ring-bearer, so no block is owed: %v", err)
	}
}

// The Ring Goes South: X counts the legendary creatures after the
// tempt, so a plain bear chosen as the Ring-bearer makes X = 1.
func TestTheRingGoesSouthRevealsUntilXLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	land2 := ringLibraryTop(me, "Second Forest", "Basic Land — Forest")
	land := ringLibraryTop(me, "Forest", "Basic Land — Forest")
	spell := ringLibraryTop(me, "A Spell", "Sorcery")
	ringCastSettled(t, g, "The Ring Goes South", "Sorcery", theRingGoesSouthOracle)
	c, ok := battlefieldCard(g, land)
	if !ok || !c.Tapped {
		t.Fatal("the first land is not on the battlefield tapped")
	}
	if g.Battlefield.Contains(land2) {
		t.Fatal("more lands than X entered")
	}
	if me.Library.Cards[0].InstanceID != spell {
		t.Fatal("the revealed nonland card is not on the bottom")
	}
}

// In the Darkness Bind Them: chapter I's Wraith may be the Ring-bearer.
func TestInTheDarknessBindThemChapterOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "In the Darkness Bind Them", "Enchantment — Saga", inTheDarknessBindThemOracle, nil)
	ringSettle(t, g)
	if onBattlefieldNamed(g, "Wraith") != 1 {
		t.Fatal("no Wraith token")
	}
	rb := ringBearerOf(g, me.ID)
	if rb == uuid.Nil || ringBFCard(t, g, rb).Name != "Wraith" {
		t.Fatal("the Wraith was not chosen as the only creature")
	}
	if !containsString(effectiveAbilities(t, g, rb), "menace") {
		t.Fatal("the Wraith has no menace")
	}
}

// In the Darkness Bind Them: chapter IV takes up to one creature from
// each opponent until end of turn, then tempts.
func TestInTheDarknessBindThemChapterFour(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	theirs := typedCreature(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	saga := castCatalogSpell(t, g, "In the Darkness Bind Them", "Enchantment — Saga", inTheDarknessBindThemOracle, nil)
	ringSettle(t, g)
	for chapter := 2; chapter <= 3; chapter++ {
		advanceToPrecombatMainOf(t, g, seat)
		ringSettle(t, g)
		if c := ringPrompt(g, me.ID); c != nil {
			answerRing(t, g, me.ID, c.ChooseCards[0])
			ringSettle(t, g)
		}
	}
	if onBattlefieldNamed(g, "Wraith") != 3 || ringCount(g, me.ID) != 3 {
		t.Fatalf("after chapter III: %d Wraiths, tempted %d times; want 3 and 3", onBattlefieldNamed(g, "Wraith"), ringCount(g, me.ID))
	}
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, theirs)
	ringSettle(t, g)
	answerRing(t, g, me.ID, theirs)
	ringSettle(t, g)
	if c, ok := battlefieldCard(g, theirs); !ok || c.Controller != me.ID || !containsString(c.Effective().Abilities, "haste") {
		t.Fatal("chapter IV did not take their creature with haste")
	}
	if ringCount(g, me.ID) != 4 || ringBearerOf(g, me.ID) != theirs {
		t.Fatal("chapter IV did not tempt, or the stolen creature could not be chosen")
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter IV")
	}
}
