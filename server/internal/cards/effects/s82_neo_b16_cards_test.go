package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	neoSharedPurposeOracle  = "4cab6ea8-4a80-4b34-b758-7f2aa4d540e4"
	neoHiddenCrueltyOracle  = "ff8cf669-a38a-44ae-baff-420f59492371"
	neoAncientWarsOracle    = "d145e4d4-a190-4c0b-8dfd-401c0111e995"
	neoBoundlessVigorOracle = "183cfd35-c18a-4838-94e5-1ad66e6bd325"
	neoLostWisdomOracle     = "f6a2092a-093e-4cd6-9eb0-d8ed32da896c"
	neoLifesOriginOracle    = "6dbcd542-ddb0-434e-b8ee-3e31e1a8b051"
	neoMyojinDreamsOracle   = "34da0b06-565f-4964-b8bc-dd106278e086"
	neoMyojinBladesOracle   = "20d20254-0f77-429a-a74e-0aec08d18f2d"
	neoMyojinBetrayalOracle = "896fc91a-1f57-45e2-b135-37bc9c2c80a7"
	neoMyojinMightOracle    = "1cb42cf5-ae69-46c5-b2ff-d2c1cd60099b"
	neoMyojinDawnOracle     = "bfb75b28-0e51-4829-b104-f1679e63c126"
)

func TestNeoB16CardsAreRegistered(t *testing.T) {
	want := map[string]string{
		neoSharedPurposeOracle: "Go-Shintai of Shared Purpose", neoHiddenCrueltyOracle: "Go-Shintai of Hidden Cruelty",
		neoAncientWarsOracle: "Go-Shintai of Ancient Wars", neoBoundlessVigorOracle: "Go-Shintai of Boundless Vigor",
		neoLostWisdomOracle: "Go-Shintai of Lost Wisdom", neoLifesOriginOracle: "Go-Shintai of Life's Origin",
		neoMyojinDreamsOracle: "Myojin of Cryptic Dreams", neoMyojinBladesOracle: "Myojin of Roaring Blades",
		neoMyojinBetrayalOracle: "Myojin of Grim Betrayal", neoMyojinMightOracle: "Myojin of Towering Might",
		neoMyojinDawnOracle: "Myojin of Blooming Dawn",
	}
	for id, name := range want {
		spec, ok := Lookup(id)
		if !ok || spec.Name != name {
			t.Errorf("%s: registered as %q (found %v)", name, spec.Name, ok)
			continue
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s is not Full", name)
		}
	}
}

// --- the Go-Shintai --------------------------------------------------------------------

const neoShrineLine = "Legendary Enchantment Creature — Shrine"

// neoShrineBoard puts the named Go-Shintai and one other Shrine under seat 0,
// and a Shrine under seat 1 that must not count, then moves to seat 0's end step.
func neoShrineBoard(t *testing.T, name, oracle string, p, tough int) (g *game.Game, me, opp *game.Player, self, other uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	self = b12Push(g, me.ID, name, neoShrineLine, oracle, p, tough)
	other = b12Permanent(g, me.ID, "Other Shrine", "Legendary Enchantment — Shrine")
	b12Permanent(g, me.ID, "Plain Enchantment", "Enchantment")
	b12Permanent(g, opp.ID, "Their Shrine", "Legendary Enchantment — Shrine")
	advanceToEndStepOf(t, g, 0)
	return g, me, opp, self, other
}

func TestNeoB16SharedPurposeMakesASpiritPerShrineWhenPaid(t *testing.T) {
	g, me, _, self, _ := neoShrineBoard(t, "Go-Shintai of Shared Purpose", neoSharedPurposeOracle, 1, 3)
	if !hasEffectiveKeyword(t, g, self, "vigilance") {
		t.Error("no vigilance")
	}
	b06AddMana(me, "C")
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Spirit"); n != 2 {
		t.Errorf("%d Spirits, want 2 (two Shrines you control; theirs and the plain enchantment don't count)", n)
	}
}

func TestNeoB16SharedPurposeDecliningOrNotPayingMakesNothing(t *testing.T) {
	g, me, _, _, _ := neoShrineBoard(t, "Go-Shintai of Shared Purpose", neoSharedPurposeOracle, 1, 3)
	b06AddMana(me, "C")
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Spirit"); n != 0 {
		t.Errorf("%d Spirits after declining, want 0", n)
	}
}

func TestNeoB16HiddenCrueltyDestroysACreatureWithToughnessAtMostTheShrines(t *testing.T) {
	g, me, opp, self, _ := neoShrineBoard(t, "Go-Shintai of Hidden Cruelty", neoHiddenCrueltyOracle, 2, 2)
	if !hasEffectiveKeyword(t, g, self, "deathtouch") {
		t.Error("no deathtouch")
	}
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Big", "Creature — Giant", 4, 3)
	b06AddMana(me, "C")
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the reflexive trigger")
	}
	if hasID(pick.PickTargetCards, big) || !hasID(pick.PickTargetCards, small) {
		t.Errorf("offered %v; want toughness 2 or less (X = 2 Shrines) only", pick.PickTargetCards)
	}
	pickTriggerTarget(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the 2/2 was not destroyed")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the 4/3 was destroyed")
	}
}

func TestNeoB16HiddenCrueltyAsksNothingWhenDeclined(t *testing.T) {
	g, me, opp, _, _ := neoShrineBoard(t, "Go-Shintai of Hidden Cruelty", neoHiddenCrueltyOracle, 2, 2)
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 1, 1)
	b06AddMana(me, "C")
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || !g.Battlefield.Contains(small) {
		t.Error("the reflexive trigger happened without the payment")
	}
}

func TestNeoB16AncientWarsDealsDamageEqualToTheShrines(t *testing.T) {
	g, me, opp, self, _ := neoShrineBoard(t, "Go-Shintai of Ancient Wars", neoAncientWarsOracle, 2, 2)
	if !hasEffectiveKeyword(t, g, self, "first strike") {
		t.Error("no first strike")
	}
	life := opp.Life
	b06AddMana(me, "C")
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life %d → %d, want -2 (two Shrines)", life, opp.Life)
	}
}

func TestNeoB16BoundlessVigorPutsCountersOnATargetShrine(t *testing.T) {
	g, me, _, self, other := neoShrineBoard(t, "Go-Shintai of Boundless Vigor", neoBoundlessVigorOracle, 1, 1)
	if !hasEffectiveKeyword(t, g, self, "trample") {
		t.Error("no trample")
	}
	b06AddMana(me, "C")
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, other)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, other, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters on the other Shrine: %d, want 2", got)
	}
}

func TestNeoB16LostWisdomMillsXCards(t *testing.T) {
	g, me, opp, self, _ := neoShrineBoard(t, "Go-Shintai of Lost Wisdom", neoLostWisdomOracle, 0, 4)
	if !hasEffectiveKeyword(t, g, self, "flying") {
		t.Error("no flying")
	}
	lib := opp.Library.Size()
	b06AddMana(me, "C")
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if got := lib - opp.Library.Size(); got != 2 {
		t.Errorf("milled %d, want 2", got)
	}
}

func TestNeoB16LifesOriginMakesAShrineTokenForItselfAndOtherNontokenShrines(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	floatForTest(g, me, "GGGG")
	castFromHand(t, g, game.Card{Name: "Go-Shintai of Life's Origin", TypeLine: neoShrineLine, ManaCost: "{3}{G}",
		OracleID: neoLifesOriginOracle, Power: 3, Toughness: 4, Colors: []string{"G"}}, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Shrine"); n != 1 {
		t.Fatalf("%d Shrine tokens after it entered, want 1", n)
	}
	// Another nontoken Shrine entering triggers it; the token entering did not (a second trigger).
	hobEnter(t, g, me.ID, game.Card{Name: "Honden", TypeLine: "Legendary Enchantment — Shrine"})
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Shrine"); n != 2 {
		t.Errorf("%d Shrine tokens, want 2", n)
	}
	// A nontoken non-Shrine does not.
	hobEnter(t, g, me.ID, game.Card{Name: "Plain", TypeLine: "Enchantment"})
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Shrine"); n != 2 {
		t.Errorf("%d Shrine tokens after a plain enchantment, want 2", n)
	}
	// An opponent's Shrine does not.
	hobEnter(t, g, g.Seats[1].ID, game.Card{Name: "Their Honden", TypeLine: "Legendary Enchantment — Shrine"})
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Shrine"); n != 2 {
		t.Errorf("%d Shrine tokens after an opposing Shrine, want 2", n)
	}
}

func TestNeoB16LifesOriginReturnsAnEnchantmentCardForFiveColors(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	go1 := b12Push(g, me.ID, "Go-Shintai of Life's Origin", neoShrineLine, neoLifesOriginOracle, 3, 4)
	ench := pushGraveyardCardTyped(me, "Old Enchantment", "Enchantment")
	art := pushGraveyardCardTyped(me, "Old Artifact", "Artifact")
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, go1, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ench}}}); err == nil {
		t.Error("activated without the five colors of mana")
	}
	floatForTest(g, me, "WUBRG")
	if err := g.ActivateCatalogAbility(me.ID, go1, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: art}}}); err == nil {
		t.Error("an artifact card was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, go1, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ench}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ench) {
		t.Error("the enchantment card was not returned to the battlefield")
	}
}

// --- the Myojins ---------------------------------------------------------------------------

// neoCastMyojin casts a Myojin from hand with its counter and returns it.
func neoCastMyojin(t *testing.T, g *game.Game, name, oracle, cost, color string, p, tough int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	floatForTest(g, me, color+color+color+color+color+color+color+color)
	id := castFromHand(t, g, game.Card{Name: name, TypeLine: "Legendary Creature — Spirit", ManaCost: cost,
		OracleID: oracle, Power: p, Toughness: tough, Colors: []string{color}}, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	if got := counterCount(g, id, game.CounterIndestructible); got != 1 {
		t.Fatalf("%s: %d indestructible counters after a hand cast, want 1", name, got)
	}
	if !effectiveAbilitiesContain(t, g, id, "indestructible") {
		t.Fatalf("%s: no indestructible with the counter", name)
	}
	return id
}

func TestNeoB16MyojinsEnterWithNoCounterUnlessCastFromHand(t *testing.T) {
	for _, c := range []struct{ name, oracle string }{
		{"Myojin of Cryptic Dreams", neoMyojinDreamsOracle}, {"Myojin of Roaring Blades", neoMyojinBladesOracle},
		{"Myojin of Grim Betrayal", neoMyojinBetrayalOracle}, {"Myojin of Towering Might", neoMyojinMightOracle},
		{"Myojin of Blooming Dawn", neoMyojinDawnOracle},
	} {
		g := newCatalogGame(t)
		id := hobEnter(t, g, g.Seats[0].ID, game.Card{Name: c.name, OracleID: c.oracle, TypeLine: "Legendary Creature — Spirit", Power: 3, Toughness: 3})
		if got := counterCount(g, id, game.CounterIndestructible); got != 0 {
			t.Errorf("%s: %d counters without being cast", c.name, got)
		}
		if effectiveAbilitiesContain(t, g, id, "indestructible") {
			t.Errorf("%s: indestructible without a counter", c.name)
		}
		if err := g.ActivateCatalogAbility(g.Seats[0].ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
			t.Errorf("%s: activated with no counter to remove", c.name)
		}
	}
}

func TestNeoB16BloomingDawnMakesASpiritPerPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Permanent(g, me.ID, "Ring", "Artifact")
	b12Permanent(g, g.Seats[1].ID, "Theirs", "Artifact")
	id := neoCastMyojin(t, g, "Myojin of Blooming Dawn", neoMyojinDawnOracle, "{5}{W}{W}{W}", "W", 4, 6)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := counterCount(g, id, game.CounterIndestructible); got != 0 {
		t.Errorf("the counter is the cost: %d left", got)
	}
	passPriorityAroundTable(t, g)
	// Bear, Ring and the Myojin on the battlefield when the ability resolves.
	if n, _ := countTokensNamed(g, me.ID, "Spirit"); n != 3 {
		t.Errorf("%d Spirits, want 3", n)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a second activation with no counter was accepted")
	}
}

func TestNeoB16RoaringBladesDealsSevenToEachOfUpToThreeTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	giant := b12Creature(g, opp.ID, "Giant", "Creature — Giant", 6, 7)
	id := neoCastMyojin(t, g, "Myojin of Roaring Blades", neoMyojinBladesOracle, "{5}{R}{R}{R}", "R", 7, 4)
	l1, l2 := opp.Life, opp2.Life
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetPlayer, ID: opp2.ID}, {Kind: game.TargetCard, ID: giant}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != l1-7 || opp2.Life != l2-7 {
		t.Errorf("life %d→%d and %d→%d, want 7 each", l1, opp.Life, l2, opp2.Life)
	}
	if g.Battlefield.Contains(giant) {
		t.Error("the 6/7 survived 7 damage")
	}
	if !g.Battlefield.Contains(id) {
		t.Error("the Myojin is not on the battlefield")
	}
}

func TestNeoB16RoaringBladesRefusesFourTargets(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := neoCastMyojin(t, g, "Myojin of Roaring Blades", neoMyojinBladesOracle, "{5}{R}{R}{R}", "R", 7, 4)
	err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{
		{Kind: game.TargetPlayer, ID: g.Seats[1].ID}, {Kind: game.TargetPlayer, ID: g.Seats[2].ID},
		{Kind: game.TargetPlayer, ID: g.Seats[3].ID}, {Kind: game.TargetPlayer, ID: me.ID}}})
	if err == nil {
		t.Error("four targets were accepted")
	}
}

func TestNeoB16ToweringMightDistributesEightCountersAndGivesTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	id := neoCastMyojin(t, g, "Myojin of Towering Might", neoMyojinMightOracle, "{5}{G}{G}{G}", "G", 8, 8)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true,
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
		Distribution: map[uuid.UUID]int{theirs: 8}}); err == nil {
		t.Error("an opponent's creature was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true,
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
		Distribution: map[uuid.UUID]int{a: 5, b: 3}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if counterCount(g, a, game.CounterPlusOne) != 5 || counterCount(g, b, game.CounterPlusOne) != 3 {
		t.Errorf("counters a=%d b=%d, want 5 and 3", counterCount(g, a, game.CounterPlusOne), counterCount(g, b, game.CounterPlusOne))
	}
	for _, c := range []uuid.UUID{a, b} {
		if !hasEffectiveKeyword(t, g, c, "trample") {
			t.Error("a target did not gain trample")
		}
	}
	if hasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an untargeted creature gained trample")
	}
}

func TestNeoB16GrimBetrayalTakesCreatureCardsPutIntoGraveyardsThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	old := pushGraveyardCardTyped(opp, "Old Corpse", "Creature — Zombie") // there since before this turn
	myDead := b12Creature(g, me.ID, "My Dead", "Creature — Bear", 2, 2)
	theirDead := b12Creature(g, opp.ID, "Their Dead", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, opp.ID, "Dead Rock", "Artifact")
	spirit := pushTestPermanent(g, opp.ID, game.Card{Name: "Dead Token", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1})
	id := neoCastMyojin(t, g, "Myojin of Grim Betrayal", neoMyojinBetrayalOracle, "{5}{B}{B}{B}", "B", 5, 2)
	// The Old Corpse was seeded into a graveyard with no zone move this turn;
	// the creatures below die this turn.
	for _, c := range []uuid.UUID{myDead, theirDead, rock, spirit} {
		p1Destroy(t, g, c)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, c := range []uuid.UUID{myDead, theirDead} {
		if !g.Battlefield.Contains(c) {
			t.Errorf("a creature card put into a graveyard this turn did not return")
		}
	}
	for _, c := range g.Battlefield.Cards {
		if (c.InstanceID == myDead || c.InstanceID == theirDead) && c.Controller != me.ID {
			t.Errorf("%s returned under %s's control, want yours", c.Name, c.Controller)
		}
	}
	if g.Battlefield.Contains(old) || g.Battlefield.Contains(rock) || g.Battlefield.Contains(spirit) {
		t.Error("a card that did not go to a graveyard this turn, a noncreature card, or a token came back")
	}
}
