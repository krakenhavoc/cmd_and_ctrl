package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b32_b_cards_test.go — Bria, Getaway Glamer, Lassoed by the Law,
// One Last Job, Another Round, Fleeting Reflection, Jailbreak Scheme,
// Hollow Marauder and Lively Dirge (Standard leftovers, batches 32 and 33).

const (
	briaOracle               = "d627996b-cbcd-4162-bd13-9ab286de64d1"
	getawayGlamerOracle      = "98eac0f7-3f4d-46ef-9575-680c52e967cf"
	lassoedByTheLawOracle    = "9457f3c1-a087-46cd-a8be-cf665466b77c"
	oneLastJobOracle         = "55cba59c-932c-4295-87dc-9179fbdfbd53"
	anotherRoundOracle       = "8095ca78-db19-4724-a6ff-eacc85fa2274"
	fleetingReflectionOracle = "c915f815-77d7-4e08-b178-82b29a58bb03"
	jailbreakSchemeOracle    = "bc7f39e1-97bf-4725-a84a-2a04121530c6"
	hollowMarauderOracle     = "1a4075f3-6c18-4cbe-99e0-f403dd78a7b0"
	livelyDirgeOracle        = "0b23d9aa-d4ed-4ccf-8810-4a4d11d6fb9c"
)

func TestStdB32BriaGrantsProwessToOthersNotOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bria := b12Push(g, me.ID, "Bria, Riptide Rogue", "Legendary Creature — Otter Rogue", briaOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, bria, "prowess") || !hasEffectiveKeyword(t, g, bear, "prowess") {
		t.Error("Bria and your other creatures have prowess")
	}
	if hasEffectiveKeyword(t, g, theirs, "prowess") {
		t.Error("an opponent's creature must not have prowess")
	}
}

func TestStdB32BriaMakesATargetUnblockableOnANoncreatureCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Bria, Riptide Rogue", "Legendary Creature — Otter Rogue", briaOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opp.ID))
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	answerTriggerOrderInOfferedOrder(t, g)
	passPriorityAroundTable(t, g)
	if restrictionsOf(t, g, bear)&game.CantBeBlocked == 0 {
		t.Error("the targeted creature can't be blocked this turn")
	}
}

func TestStdB32GetawayGlamerFlickersAtTheNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	victim := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	castModal(t, g, "Getaway Glamer", "Instant", getawayGlamerOracle, []int{0}, twTarget(victim))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the creature should be exiled")
	}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if n := onBattlefieldNamed(g, "Ogre"); n != 1 {
		t.Errorf("%d Ogres on the battlefield, want 1 (returned at the end step)", n)
	}
}

func TestStdB32GetawayGlamerDestroysOnlyTheStrictlyGreatestPower(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	big := b12Creature(g, opp.ID, "Giant", "Creature — Giant", 6, 6)
	tied := b12Creature(g, opp.ID, "Other Giant", "Creature — Giant", 6, 6)
	castModal(t, g, "Getaway Glamer", "Instant", getawayGlamerOracle, []int{1}, twTarget(big))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("no other creature has GREATER power (a tie does not save it), so it is destroyed")
	}
	_ = tied
}

func TestStdB32GetawayGlamerSparesACreatureWithLessPower(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	small := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Giant", "Creature — Giant", 6, 6)
	castModal(t, g, "Getaway Glamer", "Instant", getawayGlamerOracle, []int{1}, twTarget(small))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("another creature has greater power, so the target is not destroyed")
	}
}

func TestStdB32LassoedByTheLawExilesAndMakesAMercenary(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	victim := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	castCatalogSpell(t, g, "Lassoed by the Law", "Enchantment", lassoedByTheLawOracle, nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, g.Seats[0].ID, victim)
	answerTriggerOrderInOfferedOrder(t, g)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the Ogre should be exiled")
	}
	if onBattlefieldNamed(g, "Mercenary") != 1 {
		t.Error("a Mercenary token should have been created")
	}
}

func TestStdB32OneLastJobReturnsACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	castModal(t, g, "One Last Job", "Sorcery", oneLastJobOracle, []int{0}, twTarget(id))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("the creature card should return to the battlefield")
	}
}

func TestStdB32OneLastJobAttachesAnEquipmentToAChosenCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Sword", TypeLine: "Artifact — Equipment", Owner: me.ID, Controller: me.ID})
	castModal(t, g, "One Last Job", "Sorcery", oneLastJobOracle, []int{2}, twTarget(id))
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, host)
	passPriorityAroundTable(t, g)
	c := findBattlefieldCardByID(g, id)
	if c == nil {
		t.Fatal("the Equipment should be on the battlefield")
	}
	if c.AttachedTo.ID != host {
		t.Errorf("attached to %v, want the Bear", c.AttachedTo)
	}
}

func TestStdB32OneLastJobReturnsAnEquipmentUnattachedWithNoCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Sword", TypeLine: "Artifact — Equipment", Owner: me.ID, Controller: me.ID})
	castModal(t, g, "One Last Job", "Sorcery", oneLastJobOracle, []int{2}, twTarget(id))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("with no creature the Equipment returns unattached (the card's ruling)")
	}
}

func TestStdB32AnotherRoundBlinksTheChosenCreaturesXPlusOneTimes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Another Round", TypeLine: "Sorcery", OracleID: anotherRoundOracle, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 1}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, a) // round one: Bear A only
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) {
		t.Fatal("Bear A should have come back as a new object")
	}
	if n := onBattlefieldNamed(g, "Bear A"); n != 1 {
		t.Fatalf("%d Bear A on the battlefield after round one, want 1", n)
	}
	answerChooseCards(t, g, me.ID) // round two (X = 1): choose none
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("X = 1 means two rounds in all")
	}
}

func TestStdB32FleetingReflectionBecomesACopyAndGainsHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { findBattlefieldCardByID(g, mine).Tapped = true })
	theirs := b12Creature(g, opp.ID, "Dragon", "Creature — Dragon", 5, 5)
	castCatalogSpell(t, g, "Fleeting Reflection", "Instant", fleetingReflectionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, mine, "hexproof") {
		t.Error("the creature should have hexproof")
	}
	if findBattlefieldCardByID(g, mine).Tapped {
		t.Error("the creature should be untapped")
	}
	if p, tough := ptOf(t, g, mine); p != 5 || tough != 5 {
		t.Errorf("the copy is %d/%d, want 5/5", p, tough)
	}
}

func TestStdB32FleetingReflectionWithNoSecondTargetStillGivesHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Fleeting Reflection", "Instant", fleetingReflectionOracle, twTarget(mine))
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, mine, "hexproof") {
		t.Error("the creature should have hexproof")
	}
	if p, _ := ptOf(t, g, mine); p != 2 {
		t.Errorf("power %d, want 2 (no copy)", p)
	}
}

func TestStdB32FleetingReflectionRefusesAnOpponentsCreatureAsTheFirstTarget(t *testing.T) {
	g := newCatalogGame(t)
	theirs := b12Creature(g, g.Seats[1].ID, "Dragon", "Creature — Dragon", 5, 5)
	if err := castCatalogSpellErr(t, g, "Fleeting Reflection", "Instant", fleetingReflectionOracle, twTarget(theirs)); err == nil {
		t.Fatal("targeted a creature you don't control")
	}
}

func TestStdB32JailbreakSchemeCountersAndUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castModal(t, g, "Jailbreak Scheme", "Instant", jailbreakSchemeOracle, []int{0}, twTarget(bear))
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, bear); got != 1 {
		t.Errorf("+1/+1 counters %d, want 1", got)
	}
	if restrictionsOf(t, g, bear)&game.CantBeBlocked == 0 {
		t.Error("the creature can't be blocked this turn")
	}
}

func TestStdB32JailbreakSchemeLetsTheOwnerTuckAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	lib := opp.Library.Size()
	castModal(t, g, "Jailbreak Scheme", "Instant", jailbreakSchemeOracle, []int{1}, twTarget(rock))
	passPriorityAroundTable(t, g)
	ask := putInLibraryChoiceFor(g, opp.ID)
	if ask == nil {
		t.Fatal("the owner is not asked for the top or the bottom")
	}
	if err := g.ResolvePutInLibrary(ask.ID, opp.ID, []uuid.UUID{rock}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact should leave the battlefield")
	}
	if opp.Library.Size() != lib+1 {
		t.Errorf("library %d, want %d", opp.Library.Size(), lib+1)
	}
}

func TestStdB32HollowMarauderCostsLessPerCreatureCardInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	if got := selfPricedMV(t, g, me, hollowMarauderOracle, "Creature — Specter", "{4}{B}"); got != 5 {
		t.Errorf("empty graveyard: %d, want 5", got)
	}
	for i := 0; i < 2; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	}
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	if got := selfPricedMV(t, g, me, hollowMarauderOracle, "Creature — Specter", "{4}{B}"); got != 3 {
		t.Errorf("two creature cards: %d, want 3", got)
	}
}

func TestStdB32HollowMarauderDrawsForEachOpponentWhoDidNotDiscardABigCard(t *testing.T) {
	g := newCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	o1.Hand.Cards, o2.Hand.Cards = nil, nil
	o1.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Big", TypeLine: "Creature — Giant", ManaCost: "{4}{R}", Owner: o1.ID, Controller: o1.ID})
	o2.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Small", TypeLine: "Instant", ManaCost: "{R}", Owner: o2.ID, Controller: o2.ID})
	lib := me.Library.Size()
	castCatalogSpell(t, g, "Hollow Marauder", "Creature — Specter", hollowMarauderOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: o1.ID}, {Kind: game.TargetPlayer, ID: o2.ID}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, p := range []*game.Player{o1, o2} {
		if discardChoiceFor(g, p.ID) != nil {
			discardFromHand(t, g, p.ID)
		}
	}
	passPriorityAroundTable(t, g)
	if me.Library.Size() != lib-1 {
		t.Errorf("library %d, want %d (one draw: only the opponent who pitched the small card)", me.Library.Size(), lib-1)
	}
}

func TestStdB32LivelyDirgeSearchesACardIntoTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Bolt", TypeLine: "Instant"})
	castModal(t, g, "Lively Dirge", "Sorcery", livelyDirgeOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, ids[1])
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(ids[1]) {
		t.Error("the searched card should be in the graveyard")
	}
}

func TestStdB32LivelyDirgeReturnsCreaturesWithinTheTotalManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a, b, c := uuid.New(), uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: a, Name: "Two", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: b, Name: "Three", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: c, Name: "Two Again", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	castModal(t, g, "Lively Dirge", "Sorcery", livelyDirgeOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no creature pick")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{a, b}); err == nil {
		t.Fatal("two plus three exceeds total mana value 4")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{a, c}); err != nil {
		t.Fatalf("two plus two is within 4: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(c) || g.Battlefield.Contains(b) {
		t.Error("exactly the two 2-mana creatures should have returned")
	}
}
