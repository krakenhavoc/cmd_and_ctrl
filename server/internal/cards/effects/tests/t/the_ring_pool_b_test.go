package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rangers := effectstest.PushCatalogPermanent(g, me.ID, "Dúnedain Rangers", "Creature — Human Ranger", effectstest.DunedainRangersOracle, false)
	effectstest.B12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	effectstest.RingSettle(t, g)
	if effectstest.RingCount(g, me.ID) != 1 || effectstest.RingBearerOf(g, me.ID) != rangers {
		t.Fatalf("first land: tempted %d times, Ring-bearer %v", effectstest.RingCount(g, me.ID), effectstest.RingBearerOf(g, me.ID))
	}
	effectstest.B12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	effectstest.RingSettle(t, g)
	if effectstest.RingCount(g, me.ID) != 1 {
		t.Fatal("a second land tempted while you had a Ring-bearer")
	}
}

// Faramir: the end-step draw after a creature died, and the token when
// another creature is chosen, not Faramir.
func TestFaramirFieldCommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	faramir := effectstest.PushCatalogPermanent(g, me.ID, "Faramir, Field Commander", "Legendary Creature — Human Soldier", effectstest.FaramirFieldCommanderOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	effectstest.RingTempt(t, g, me.ID)
	effectstest.AnswerRing(t, g, me.ID, faramir)
	effectstest.RingSettle(t, g)
	if effectstest.OnBattlefieldNamed(g, "Human Soldier") != 0 {
		t.Fatal("choosing Faramir made a token")
	}
	effectstest.RingTempt(t, g, me.ID)
	effectstest.AnswerRing(t, g, me.ID, bear)
	effectstest.RingSettle(t, g)
	if effectstest.OnBattlefieldNamed(g, "Human Soldier") != 1 {
		t.Fatal("choosing another creature made no token")
	}

	// No creature died: no draw at the end step.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.PushCatalogPermanent(g2, me2.ID, "Faramir, Field Commander", "Legendary Creature — Human Soldier", effectstest.FaramirFieldCommanderOracle, false)
	effectstest.AdvanceTo(t, g2, game.StepEnd)
	if effectstest.TriggerOnStackFrom(g2, "Faramir, Field Commander") {
		t.Fatal("the end-step draw triggered with no creature dead")
	}

	victim := effectstest.B12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	effectstest.RingSettle(t, g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	hand := me.Hand.Size()
	effectstest.RingSettle(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d → %d at the end step after a death, want +1", hand, me.Hand.Size())
	}
}

// Galadriel of Lothlórien: scry 3 on another Ring-bearer, then the
// scry trigger reveals a land and puts it onto the battlefield tapped.
func TestGaladrielOfLothlorienScriesAndFindsALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Galadriel of Lothlórien", "Legendary Creature — Elf Noble", effectstest.GaladrielOfLothlorienOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	land := ringLibraryTop(me, "Forest", "Basic Land — Forest")

	effectstest.RingTempt(t, g, me.ID)
	effectstest.AnswerRing(t, g, me.ID, bear)
	effectstest.RingSettle(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.RingSettle(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.RingSettle(t, g)
	c, ok := effectstest.BattlefieldCard(g, land)
	if !ok || !c.Tapped {
		t.Fatal("the revealed land is not on the battlefield tapped")
	}
}

// Gandalf: sorceries at instant speed, and a draw on another
// Ring-bearer.
func TestGandalfFriendOfTheShire(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	// The draw step: a sorcery can't be cast without Gandalf.
	sorcery := effectstest.B13HandCard(me, "Birthday Escape", "Sorcery", effectstest.BirthdayEscapeOracle)
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{}); err == nil {
		t.Fatal("a sorcery was cast in the draw step without Gandalf")
	}
	gandalf := effectstest.PushCatalogPermanent(g, me.ID, "Gandalf, Friend of the Shire", "Legendary Creature — Avatar Wizard", effectstest.GandalfFriendOfTheShireOracl, false)
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{}); err != nil {
		t.Fatalf("with Gandalf the sorcery is cast as though it had flash: %v", err)
	}
	effectstest.RingSettle(t, g)
	// Birthday Escape tempted with Gandalf the only creature: no draw
	// from Gandalf.
	if effectstest.RingBearerOf(g, me.ID) != gandalf {
		t.Fatal("Gandalf is not the Ring-bearer")
	}
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.RingTempt(t, g, me.ID)
	effectstest.AnswerRing(t, g, me.ID, bear)
	effectstest.RingSettle(t, g)
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
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		spell := effectstest.B13PlayAs(t, g, 1, "Their Creature", tc.typeLine, "")
		if err := g.CastSpell(me.ID, effectstest.B13HandCard(me, "Glorious Gale", "Instant", effectstest.GloriousGaleOracle),
			game.CastSpellParams{Targets: effectstest.B16TargetCard(spell)}); err != nil {
			t.Fatalf("Glorious Gale: %v", err)
		}
		effectstest.RingSettle(t, g)
		if g.Battlefield.Contains(spell) || !opp.Graveyard.Contains(spell) {
			t.Errorf("%s: the spell was not countered", tc.typeLine)
		}
		if effectstest.RingCount(g, me.ID) != tc.tempts {
			t.Errorf("%s: tempted %d times, want %d", tc.typeLine, effectstest.RingCount(g, me.ID), tc.tempts)
		}
	}
}

// Gollum, Patient Plotter: leaving tempts; from the graveyard, {B} and
// a creature return it to hand.
func TestGollumPatientPlotter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gollum := effectstest.PushCatalogPermanent(g, me.ID, "Gollum, Patient Plotter", "Legendary Creature — Halfling Horror", effectstest.GollumPatientPlotterOracle, false)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(gollum) })
	effectstest.RingSettle(t, g)
	if effectstest.RingCount(g, me.ID) != 1 {
		t.Fatal("Gollum left the battlefield and the Ring did not tempt")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	fodder := effectstest.B12Creature(g2, me2.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	id, _ := effectstest.ActivateFromGraveyard(t, g2, "Gollum, Patient Plotter", "Legendary Creature — Halfling Horror", effectstest.GollumPatientPlotterOracle, 3, 1, "{B}",
		game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	effectstest.RingSettle(t, g2)
	if !me2.Hand.Contains(id) || g2.Battlefield.Contains(fodder) {
		t.Fatal("Gollum did not return to hand for {B} and a sacrificed creature")
	}
}

// Now for Wrath, Now for Ruin!: a counter and vigilance on each of your
// creatures, then the tempt.
func TestNowForWrathNowForRuin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Now for Wrath, Now for Ruin!", "Sorcery", effectstest.NowForWrathOracle, nil)
	effectstest.RingSettle(t, g)
	effectstest.AnswerRing(t, g, me.ID, a)
	effectstest.RingSettle(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if effectstest.RingBFCard(t, g, id).Counters["+1/+1"] != 1 || !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, id), "vigilance") {
			t.Errorf("%s has no counter or no vigilance", effectstest.RingBFCard(t, g, id).Name)
		}
	}
	if effectstest.RingBFCard(t, g, theirs).Counters["+1/+1"] != 0 {
		t.Error("an opponent's creature got a counter")
	}
	if effectstest.RingCount(g, me.ID) != 1 || effectstest.RingBearerOf(g, me.ID) != a {
		t.Fatal("the Ring did not tempt")
	}
}

// One Ring to Rule Them All: I mills each player by the Ring-bearer's
// power; II destroys every nonlegendary creature, sparing the
// Ring-bearer, which the Ring makes legendary; III drains for creature
// cards.
func TestOneRingToRuleThemAll(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	graves := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		graves[p.ID] = p.Graveyard.Size()
	}
	saga := effectstest.CastCatalogSpell(t, g, "One Ring to Rule Them All", "Enchantment — Saga", effectstest.OneRingToRuleThemAllOracle, nil)
	effectstest.RingSettle(t, g)
	if effectstest.RingBearerOf(g, me.ID) != bear {
		t.Fatal("chapter I: the bear is not the Ring-bearer")
	}
	for _, p := range g.Seats {
		if p.Graveyard.Size() != graves[p.ID]+2 {
			t.Errorf("chapter I: a player milled %d, want 2", p.Graveyard.Size()-graves[p.ID])
		}
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.RingSettle(t, g)
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
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.RingSettle(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	legend := effectstest.TypedCreature(g, opp.ID, "Their Legend", "Legendary Creature — Human", "", 4, 4)
	life := opp.Life
	wraiths := effectstest.CastRingCreature(t, g, "Ringwraiths", "Creature — Wraith Knight", effectstest.RingwraithsOracle, 5, 5)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			effectstest.AnswerPickTarget(t, g, legend)
			break
		}
	}
	effectstest.RingSettle(t, g)
	if p, tt := effectstest.PtOf(t, g, legend); p != 1 || tt != 1 {
		t.Fatalf("the legend is %d/%d, want 1/1", p, tt)
	}
	if opp.Life != life-3 {
		t.Fatalf("the legend's controller is at %d, want %d", opp.Life, life-3)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(wraiths) })
	effectstest.RingSettle(t, g)
	effectstest.RingTempt(t, g, me.ID)
	effectstest.RingSettle(t, g)
	if !me.Hand.Contains(wraiths) {
		t.Fatal("Ringwraiths did not return to hand when the Ring tempted")
	}
}

// Slip On the Ring: the creature returns as a new object and may be
// chosen.
func TestSlipOnTheRingFlickersThenTempts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, "+1/+1", 1) })
	effectstest.RingCastSettled(t, g, "Slip On the Ring", "Instant", effectstest.SlipOnTheRingOracle, ringTarget(bear))
	if effectstest.OnBattlefieldNamed(g, "Grizzly Bears") != 1 {
		t.Fatal("the bear did not return")
	}
	rb := effectstest.RingBearerOf(g, me.ID)
	if rb == uuid.Nil || effectstest.RingCount(g, me.ID) != 1 {
		t.Fatal("the returned bear was not chosen as the Ring-bearer")
	}
	if effectstest.RingBFCard(t, g, rb).Counters["+1/+1"] != 0 {
		t.Fatal("the returned bear kept its counter")
	}
}

// Sméagol, Helpful Guide: the opponent reveals until a land, which
// enters tapped under your control; the rest go to their graveyard.
func TestSmeagolHelpfulGuideTakesALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Sméagol, Helpful Guide", "Legendary Creature — Halfling Horror", effectstest.SmeagolHelpfulGuideOracle, false)
	land := ringLibraryTop(opp, "Their Forest", "Basic Land — Forest")
	x := ringLibraryTop(opp, "Their Spell", "Sorcery")
	y := ringLibraryTop(opp, "Their Other Spell", "Instant")
	effectstest.RingTempt(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.RingSettle(t, g)
	c, ok := effectstest.BattlefieldCard(g, land)
	if !ok || c.Controller != me.ID || !c.Tapped {
		t.Fatal("the land did not enter tapped under Sméagol's controller")
	}
	if !opp.Graveyard.Contains(x) || !opp.Graveyard.Contains(y) {
		t.Fatal("the rest did not go to their graveyard")
	}
}

// War of the Last Alliance: I finds a legendary creature card.
func TestWarOfTheLastAllianceFindsALegend(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	legend := ringLibraryTop(me, "Some Legend", "Legendary Creature — Elf")
	effectstest.CastCatalogSpell(t, g, "War of the Last Alliance", "Enchantment — Saga", effectstest.WarOfTheLastAllianceOracle, nil)
	effectstest.RingSettle(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if !me.Hand.Contains(legend) {
		t.Fatal("chapter I did not find the legendary creature card")
	}
}

// Frodo Baggins: tempts as it or another legend of yours enters, and
// must be blocked only while it is your Ring-bearer.
func TestFrodoBaggins(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	frodo := effectstest.CastRingCreature(t, g, "Frodo Baggins", "Legendary Creature — Halfling Scout", effectstest.FrodoBagginsOracle, 1, 3)
	effectstest.RingSettle(t, g)
	if effectstest.RingCount(g, me.ID) != 1 || effectstest.RingBearerOf(g, me.ID) != frodo {
		t.Fatal("Frodo entering did not tempt")
	}
	effectstest.CastRingCreature(t, g, "Other Legend", "Legendary Creature — Elf", "", 2, 2)
	effectstest.RingSettle(t, g)
	effectstest.AnswerRing(t, g, me.ID, frodo)
	effectstest.RingSettle(t, g)
	if effectstest.RingCount(g, me.ID) != 2 {
		t.Fatal("another legendary creature entering did not tempt")
	}
	effectstest.CastRingCreature(t, g, "A Bear", "Creature — Bear", "", 2, 2)
	effectstest.RingSettle(t, g)
	if effectstest.RingCount(g, me.ID) != 2 {
		t.Fatal("a nonlegendary creature entering tempted")
	}
}

func TestFrodoBagginsMustBeBlockedAsRingBearer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Seats1684(g)
	frodo := effectstest.TypedCreature(g, me.ID, "Frodo Baggins", "Legendary Creature — Halfling Scout", effectstest.FrodoBagginsOracle, 1, 3, "haste")
	effectstest.TypedCreature(g, opp.ID, "Wall", "Creature — Wall", "", 0, 4)
	effectstest.RingTempt(t, g, me.ID)
	effectstest.ReqToBlockers(t, g, frodo)
	effectstest.ReqRefusal(t, g.PassPriority())

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := effectstest.Seats1684(g2)
	frodo2 := effectstest.TypedCreature(g2, me2.ID, "Frodo Baggins", "Legendary Creature — Halfling Scout", effectstest.FrodoBagginsOracle, 1, 3, "haste")
	effectstest.TypedCreature(g2, opp2.ID, "Wall", "Creature — Wall", "", 0, 4)
	effectstest.ReqToBlockers(t, g2, frodo2)
	if err := g2.PassPriority(); err != nil {
		t.Fatalf("Frodo is not the Ring-bearer, so no block is owed: %v", err)
	}
}

// The Ring Goes South: X counts the legendary creatures after the
// tempt, so a plain bear chosen as the Ring-bearer makes X = 1.
func TestTheRingGoesSouthRevealsUntilXLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	land2 := ringLibraryTop(me, "Second Forest", "Basic Land — Forest")
	land := ringLibraryTop(me, "Forest", "Basic Land — Forest")
	spell := ringLibraryTop(me, "A Spell", "Sorcery")
	effectstest.RingCastSettled(t, g, "The Ring Goes South", "Sorcery", effectstest.TheRingGoesSouthOracle)
	c, ok := effectstest.BattlefieldCard(g, land)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "In the Darkness Bind Them", "Enchantment — Saga", effectstest.InTheDarknessBindThemOracle, nil)
	effectstest.RingSettle(t, g)
	if effectstest.OnBattlefieldNamed(g, "Wraith") != 1 {
		t.Fatal("no Wraith token")
	}
	rb := effectstest.RingBearerOf(g, me.ID)
	if rb == uuid.Nil || effectstest.RingBFCard(t, g, rb).Name != "Wraith" {
		t.Fatal("the Wraith was not chosen as the only creature")
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, rb), "menace") {
		t.Fatal("the Wraith has no menace")
	}
}

// In the Darkness Bind Them: chapter IV takes up to one creature from
// each opponent until end of turn, then tempts.
func TestInTheDarknessBindThemChapterFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	theirs := effectstest.TypedCreature(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	saga := effectstest.CastCatalogSpell(t, g, "In the Darkness Bind Them", "Enchantment — Saga", effectstest.InTheDarknessBindThemOracle, nil)
	effectstest.RingSettle(t, g)
	for chapter := 2; chapter <= 3; chapter++ {
		effectstest.AdvanceToPrecombatMainOf(t, g, seat)
		effectstest.RingSettle(t, g)
		if c := effectstest.RingPrompt(g, me.ID); c != nil {
			effectstest.AnswerRing(t, g, me.ID, c.ChooseCards[0])
			effectstest.RingSettle(t, g)
		}
	}
	if effectstest.OnBattlefieldNamed(g, "Wraith") != 3 || effectstest.RingCount(g, me.ID) != 3 {
		t.Fatalf("after chapter III: %d Wraiths, tempted %d times; want 3 and 3", effectstest.OnBattlefieldNamed(g, "Wraith"), effectstest.RingCount(g, me.ID))
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.RingSettle(t, g)
	effectstest.AnswerRing(t, g, me.ID, theirs)
	effectstest.RingSettle(t, g)
	if c, ok := effectstest.BattlefieldCard(g, theirs); !ok || c.Controller != me.ID || !effectstest.ContainsString(c.Effective().Abilities, "haste") {
		t.Fatal("chapter IV did not take their creature with haste")
	}
	if effectstest.RingCount(g, me.ID) != 4 || effectstest.RingBearerOf(g, me.ID) != theirs {
		t.Fatal("chapter IV did not tempt, or the stolen creature could not be chosen")
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter IV")
	}
}
