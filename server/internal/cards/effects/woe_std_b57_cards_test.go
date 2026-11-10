package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_std_b57_cards_test.go — Wilds of Eldraine, Standard leftovers
// batches 55–58, second half of the tests (#4471).

// b55Settle passes priority until the table is quiet, answering any pick
// prompt that opens with its first Min cards (or all, with pickAll).
func b55Settle(t *testing.T, g *game.Game, who uuid.UUID, pickAll bool) {
	t.Helper()
	for i := 0; i < 24; i++ {
		passPriorityAroundTable(t, g)
		if sc := latestChoiceOfKindFor(g, game.PendingChoiceSacrifice, who); sc != nil {
			if err := g.ResolveSacrificeChoice(sc.ID, who, sc.SacrificeOptions[0]); err != nil {
				t.Fatalf("ResolveSacrificeChoice: %v", err)
			}
			continue
		}
		if op := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, who); op != nil {
			n := op.ChooseMin
			if pickAll {
				n = len(op.ChooseCards)
			}
			if n > 1 {
				n = 1
			}
			if err := g.ResolveOwnPermanents(op.ID, who, op.ChooseCards[:n]); err != nil {
				t.Fatalf("ResolveOwnPermanents: %v", err)
			}
			continue
		}
		p := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, who)
		if p == nil {
			return
		}
		n := p.ChooseMin
		if pickAll {
			n = len(p.ChooseCards)
		}
		if n > len(p.ChooseCards) {
			n = len(p.ChooseCards)
		}
		if err := g.ResolveChooseCards(p.ID, who, p.ChooseCards[:n]); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
	}
	t.Fatal("the prompts never ran out")
}

// --- Storyteller Pixie ---------------------------------------------------------

func TestB57StorytellerPixieDrawsForAnAdventureSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	b12Push(g, me.ID, "Storyteller Pixie", "Creature — Faerie Wizard", b57Pixie, 2, 3)
	_, id := hobA6HandAdventure(t, g, b55Hollow,
		hobA6Face("Hollow Scavenger", "Creature — Wolf", "{2}{G}", 3, 2, "G"),
		hobA6Face("Bakery Raid", "Sorcery — Adventure", "{G}", 0, 0, "G"))
	hand := b55HandSize(g, me)
	hobA6MustCast(t, g, me, id, 1, nil)
	if got := b55HandSize(g, me); got != hand-1+1 {
		t.Errorf("hand %d -> %d, want the Adventure cast (-1) and one draw (+1)", hand, got)
	}
}

// --- Experimental Confectioner --------------------------------------------------

func TestB57ExperimentalConfectionerMakesAFoodThenARatPerFoodSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	castCatalogSpell(t, g, "Experimental Confectioner", "Creature — Human Peasant", b57Confect, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Fatalf("%d Foods after it entered, want 1", got)
	}
	food := findBattlefieldByName(g, "Food")
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(food); err != nil {
			t.Fatalf("sacrifice: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Rat"); got != 1 {
		t.Errorf("%d Rats after a Food was sacrificed, want 1", got)
	}
	assertRestrictions(t, g, findBattlefieldByName(g, "Rat"), game.CantBlock)
}

// --- Charging Hooligan -----------------------------------------------------------

func TestB57ChargingHooliganGetsPlusOnePerAttackerAndTrampleWithARat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	hooligan := b12Push(g, me.ID, "Charging Hooligan", "Creature — Human Peasant", b57Hooligan, 3, 3)
	rat := b12Creature(g, me.ID, "Rat", "Token Creature — Rat", 1, 1)
	other := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{hooligan, rat, other} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("attack: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if p, _ := effectivePTOf(t, g, hooligan); p != 6 {
		t.Errorf("power %d, want 3 + 3 attackers", p)
	}
	if !b55Has(t, g, hooligan, "trample") {
		t.Error("no trample with a Rat attacking")
	}
}

// --- Farsight Ritual -------------------------------------------------------------

func TestB57FarsightRitualTakesTwoOfTheTopFourOrEight(t *testing.T) {
	for _, bargained := range []bool{false, true} {
		g := newCatalogGame(t)
		me, _ := woeB1Seats(g)
		lib := len(me.Library.Cards)
		hand := b55HandSize(g, me)
		params := game.CastSpellParams{}
		if bargained {
			params.OptionalCosts = []int{0}
			params.SacrificeIDs = []uuid.UUID{woeILToken(g, me.ID)}
		}
		advanceTo(t, g, game.StepPrecombatMain)
		floatForTest(g, me, "UUCC")
		p1g2Cast(t, g, "Farsight Ritual", "Instant", b57Farsight, "{2}{U}{U}", params)
		b55Settle(t, g, me.ID, false)
		if got := b55HandSize(g, me); got != hand+2-1+0 && got != hand+2 {
			t.Errorf("bargained=%v: hand %d -> %d, want +2 (less the spell)", bargained, hand, got)
		}
		if got := len(me.Library.Cards); got != lib-2 {
			t.Errorf("bargained=%v: library %d -> %d, want two taken", bargained, lib, got)
		}
	}
}

// --- The End ---------------------------------------------------------------------

func TestB57TheEndCostsTwoLessAtFiveLifeOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	if got := woeILPrice(t, g, me.ID, "The End", "Instant", "{2}{B}{B}", b57End, nil); got != 4 {
		t.Errorf("price at full life %d, want 4", got)
	}
	g.WithWriteLock(func() { me.Life = 5 })
	if got := woeILPrice(t, g, me.ID, "The End", "Instant", "{2}{B}{B}", b57End, nil); got != 2 {
		t.Errorf("price at 5 life %d, want 2", got)
	}
}

func TestB57TheEndExilesAndSearchesThreeZonesThenVictimDrawsForTheHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	bear := b12Creature(g, opp.ID, "Grizzly", "Creature — Bear", 2, 2)
	named := func(zoneName string) game.Card {
		return game.Card{InstanceID: uuid.New(), Name: "Grizzly", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID}
	}
	hand1, hand2 := named("hand"), named("hand")
	opp.Hand.PushTop(hand1)
	opp.Hand.PushTop(hand2)
	opp.Graveyard.PushTop(named("gy"))
	opp.Library.PushTop(named("lib"))
	handBefore := b55HandSize(g, opp)
	libBefore := len(opp.Library.Cards)
	advanceTo(t, g, game.StepPrecombatMain)
	floatForTest(g, me, "BBCC")
	p1g2Cast(t, g, "The End", "Instant", b57End, "{2}{B}{B}", game.CastSpellParams{Targets: []game.TargetRef{hobA6Card(bear)}})
	b55SettleSearch(t, g, me.ID)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not exiled")
	}
	for _, c := range append(append([]game.Card{}, opp.Hand.Cards...), opp.Graveyard.Cards...) {
		if c.Name == "Grizzly" {
			t.Errorf("a Grizzly is still in a hand or graveyard")
		}
	}
	// Two cards left the hand, so the victim drew two.
	if got := b55HandSize(g, opp); got != handBefore-2+2 {
		t.Errorf("victim hand %d -> %d, want %d", handBefore, got, handBefore)
	}
	if len(opp.Library.Cards) >= libBefore {
		t.Logf("library %d -> %d (the draws and the exiled copy)", libBefore, len(opp.Library.Cards))
	}
}

// b55SettleSearch answers search prompts by taking every offered card.
func b55SettleSearch(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	for i := 0; i < 24; i++ {
		passPriorityAroundTable(t, g)
		if p := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, who); p != nil {
			if err := g.ResolveChooseCards(p.ID, who, p.ChooseCards); err != nil {
				t.Fatalf("ResolveChooseCards: %v", err)
			}
			continue
		}
		if p := latestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, who); p != nil {
			if err := g.ResolveSearchLibrary(p.ID, who, p.SearchCards); err != nil {
				t.Fatalf("ResolveSearch: %v", err)
			}
			continue
		}
		return
	}
}

// --- Rankle's Prank --------------------------------------------------------------

func TestB57RanklesPrankLifeLossHitsEveryPlayer(t *testing.T) {
	g := newCatalogGame(t)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = lifeOf(g, p.ID)
	}
	castCatalogSpellWithModes(t, g, "Rankle's Prank", "Sorcery", b57Prank, []int{1})
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if got := lifeOf(g, p.ID); got != before[p.ID]-4 {
			t.Errorf("%s at %d, want %d", p.Name, got, before[p.ID]-4)
		}
	}
}

func TestB57RanklesPrankSacrificesTwoCreaturesEach(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	mine := []uuid.UUID{b12Creature(g, me.ID, "A", "Creature — Bear", 1, 1), b12Creature(g, me.ID, "B", "Creature — Bear", 1, 1), b12Creature(g, me.ID, "C", "Creature — Bear", 1, 1)}
	theirs := []uuid.UUID{b12Creature(g, opp.ID, "D", "Creature — Bear", 1, 1)}
	castCatalogSpellWithModes(t, g, "Rankle's Prank", "Sorcery", b57Prank, []int{2})
	for _, p := range g.Seats {
		b55Settle(t, g, p.ID, false)
	}
	left := 0
	for _, id := range append(mine, theirs...) {
		if g.Battlefield.Contains(id) {
			left++
		}
	}
	if left != 1 {
		t.Errorf("%d creatures remain, want 1 (3 -> 1 for me, 1 -> 0 for them)", left)
	}
}

// --- Specter of Mortality --------------------------------------------------------

func TestB57SpecterOfMortalityShrinksOtherCreaturesByTheCardsExiled(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	for i := 0; i < 2; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID, Power: 1, Toughness: 1})
	}
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Big", "Creature — Bear", 5, 5)
	spec := castCatalogSpell(t, g, "Specter of Mortality", "Creature — Specter", b57Specter, nil)
	b55Settle(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the 2/2 survived -2/-2")
	}
	if p, tough := effectivePTOf(t, g, big); p != 3 || tough != 3 {
		t.Errorf("the 5/5 is %d/%d, want 3/3", p, tough)
	}
	if !g.Battlefield.Contains(spec) {
		t.Error("the Specter shrank itself")
	}
	if len(me.Graveyard.Cards) != 0 {
		t.Errorf("%d cards left in the graveyard", len(me.Graveyard.Cards))
	}
}

// --- Tangled Colony --------------------------------------------------------------

func TestB57TangledColonyLeavesARatPerDamageDealtToIt(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	colony := b12Push(g, me.ID, "Tangled Colony", "Creature — Rat", b57Colony, 2, 2)
	assertRestrictions(t, g, colony, game.CantBlock)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.New(), colony, 3); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(colony) {
		t.Fatal("the Colony survived 3 damage")
	}
	if got := b27CountNamed(g, me.ID, "Rat"); got != 3 {
		t.Errorf("%d Rats, want 3", got)
	}
}

// --- Rowan, Scion of War and Will, Scion of Peace -----------------------------------

func TestB55RowanDiscountsBlackAndRedSpellsByLifeLost(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	rowan := b12Push(g, me.ID, "Rowan, Scion of War", "Legendary Creature — Human Wizard", b55Rowan, 3, 3)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.New(), me.ID, -2); err != nil {
			t.Fatal(err)
		}
	})
	advanceTo(t, g, game.StepPrecombatMain)
	if err := hobB4Activate(t, g, me.ID, rowan, 0, "", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := woeILPrice(t, g, me.ID, "Black Thing", "Sorcery", "{3}{B}", "x", nil); got != 2 {
		t.Errorf("a {3}{B} black spell costs %d, want 2", got)
	}
	if got := woeILPrice(t, g, me.ID, "Green Thing", "Sorcery", "{3}{G}", "x", nil); got != 4 {
		t.Errorf("a {3}{G} green spell costs %d, want 4", got)
	}
}
