package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b24_b_cards_test.go — Star Trek, batch 24 (#4973), part B.

// trekB24ThroughCombat runs the combat the attack declared through its
// damage and settles the triggers it makes.
func trekB24ThroughCombat(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Gowron, Eager for Battle -------------------------------------------------

func TestTrekB24GowronGrowsWhenAnotherCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	gowron := effectstest.B12Push(g, me.ID, "Gowron, Eager for Battle", "Legendary Creature — Klingon Warrior", trekB24GowronOracle, 3, 3)
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B15Destroy(t, g, theirs)
	if got := trekB24Counter(t, g, gowron, game.CounterPlusOne); got != 1 {
		t.Errorf("Gowron has %d counters after an opponent's creature died, want 1", got)
	}
	land := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.B15Destroy(t, g, land)
	if got := trekB24Counter(t, g, gowron, game.CounterPlusOne); got != 1 {
		t.Errorf("Gowron has %d counters after a land died, want 1", got)
	}
}

func TestTrekB24GowronSacrificingAKlingonGivesTrampleACardAndCostsALife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	gowron := effectstest.B12Push(g, me.ID, "Gowron, Eager for Battle", "Legendary Creature — Klingon Warrior", trekB24GowronOracle, 3, 3)
	klingon := effectstest.B16Creature(g, me.ID, "Klingon Grunt", "Creature — Klingon Warrior", 2, 1)
	hand, life := me.Hand.Size(), me.Life
	if err := g.ActivateCatalogAbility(me.ID, gowron, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{klingon}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(klingon) {
		t.Fatal("the Klingon was not sacrificed")
	}
	if !effectstest.HasEffectiveKeyword(t, g, gowron, "trample") {
		t.Error("Gowron has no trample")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if me.Life != life-1 {
		t.Errorf("life = %d, want %d", me.Life, life-1)
	}
	if got := trekB24Counter(t, g, gowron, game.CounterPlusOne); got != 1 {
		t.Errorf("Gowron has %d counters, want 1: the sacrificed creature died", got)
	}
}

func TestTrekB24GowronSacrificingANonKlingonOnlyGivesTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	gowron := effectstest.B12Push(g, me.ID, "Gowron, Eager for Battle", "Legendary Creature — Klingon Warrior", trekB24GowronOracle, 3, 3)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand, life := me.Hand.Size(), me.Life
	if err := g.ActivateCatalogAbility(me.ID, gowron, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{gowron}}); err == nil {
		t.Fatal("Gowron sacrificed itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, gowron, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, gowron, "trample") {
		t.Error("Gowron has no trample")
	}
	if me.Hand.Size() != hand || me.Life != life {
		t.Errorf("hand %d -> %d, life %d -> %d: a non-Klingon draws nothing and costs no life", hand, me.Hand.Size(), life, me.Life)
	}
}

func TestTrekB24GowronCountsAChangelingAsAKlingon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	gowron := effectstest.B12Push(g, me.ID, "Gowron, Eager for Battle", "Legendary Creature — Klingon Warrior", trekB24GowronOracle, 3, 3)
	shape := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Shapeshifter", TypeLine: "Creature — Shapeshifter",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID, Keywords: []string{"changeling"},
	})
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, gowron, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{shape}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d for a changeling, want 1", got)
	}
}

// --- Gray and Adira, Forever Joined -------------------------------------------

func TestTrekB24GrayAndAdiraCopyAfterANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Gray and Adira, Forever Joined", "Legendary Creature — Trill Human Officer", trekB24GrayOracle, 2, 3)
	hero := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Legendary Hero", TypeLine: "Legendary Creature — Human Soldier",
		Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.AnswerPickTargetIfOpen(t, g, hero)
	effectstest.PassPriorityAroundTable(t, g)
	var token game.Card
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Legendary Hero" && effects.IsToken(c) {
			token, n = c, n+1
		}
	}
	if n != 1 {
		t.Fatalf("%d token copies, want 1", n)
	}
	if token.HasSupertype("Legendary") {
		t.Error("the token copy is legendary")
	}
	if !effectstest.HasEffectiveKeyword(t, g, token.InstanceID, "haste") {
		t.Error("the token has no haste")
	}
	if !g.Battlefield.Contains(hero) {
		t.Error("the original left: the legend rule should not apply to a nonlegendary copy")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token.InstanceID) {
		t.Error("the token was not sacrificed at the end step")
	}
	if !g.Battlefield.Contains(hero) {
		t.Error("the original was sacrificed")
	}
}

func TestTrekB24GrayAndAdiraDoNothingWithoutANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Gray and Adira, Forever Joined", "Legendary Creature — Trill Human Officer", trekB24GrayOracle, 2, 3)
	effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatal("the trigger asked for a target after only a creature spell")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if effects.IsToken(c) {
			t.Fatalf("a token %q was made", c.Name)
		}
	}
}

// --- Hemmer, Fixing What Is Broken --------------------------------------------

func TestTrekB24HemmerLootsWhenItBecomesTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hemmer := effectstest.B12Push(g, me.ID, "Hemmer, Fixing What Is Broken", "Legendary Creature — Aenar Artificer", trekB24HemmerOracle, 2, 3)
	hand := me.Hand.Size()
	grave := len(me.Graveyard.Cards)
	effectstest.TrekB18AAttack(t, g, opp.ID, hemmer)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard prompt after Hemmer became tapped")
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want the same size after a loot", hand, me.Hand.Size())
	}
	if got := len(me.Graveyard.Cards) - grave; got != 1 {
		t.Errorf("%d cards discarded, want 1", got)
	}
}

func TestTrekB24HemmerGivesSpacecraftHexproofOnlyDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Hemmer, Fixing What Is Broken", "Legendary Creature — Aenar Artificer", trekB24HemmerOracle, 2, 3)
	ship := effectstest.B12Permanent(g, me.ID, "Shuttle", "Artifact — Spacecraft")
	theirShip := effectstest.B12Permanent(g, opp.ID, "Their Shuttle", "Artifact — Spacecraft")
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, ship, "hexproof") {
		t.Error("your Spacecraft has no hexproof during your turn")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirShip, "hexproof") {
		t.Error("an opponent's Spacecraft has hexproof")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("a non-Spacecraft has hexproof")
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, (seat+1)%len(g.Seats))
	if effectstest.HasEffectiveKeyword(t, g, ship, "hexproof") {
		t.Error("your Spacecraft has hexproof during an opponent's turn")
	}
}

// --- Hit It! --------------------------------------------------------------------

func TestTrekB24HitItGivesHasteAndPutsALandFromTheTopFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Ritual", TypeLine: "Sorcery"},
		game.Card{Name: "Relic", TypeLine: "Artifact"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	effectstest.CastCatalogSpell(t, g, "Hit It!", "Enchantment", trekB24HitItOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ids[1]) {
		t.Fatal("the land did not enter")
	}
	if !effectstest.B16Tapped(t, g, ids[1]) {
		t.Error("the land entered untapped")
	}
	for _, i := range []int{0, 2, 3, 4} {
		if !me.Graveyard.Contains(ids[i]) {
			t.Errorf("revealed card %d is not in the graveyard", i)
		}
	}
	if !me.Library.Contains(ids[5]) {
		t.Error("the sixth card left the library")
	}
	bear := effectstest.B16Creature(g, me.ID, "Fresh Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("a creature you control has no haste")
	}
}

func TestTrekB24HitItWithNoLandPutsAllFiveInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "A", TypeLine: "Instant"},
		game.Card{Name: "B", TypeLine: "Instant"},
		game.Card{Name: "C", TypeLine: "Instant"},
		game.Card{Name: "D", TypeLine: "Instant"},
		game.Card{Name: "E", TypeLine: "Instant"},
	)
	effectstest.CastCatalogSpell(t, g, "Hit It!", "Enchantment", trekB24HitItOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, id := range ids {
		if !me.Graveyard.Contains(id) {
			t.Errorf("card %d is not in the graveyard", i)
		}
	}
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature has haste")
	}
}

// #5949: the Oracle text reads "look at the top five cards", not "reveal"
// (CR 701.20e). While its controller picks between two lands, only they
// know the five cards, and no reveal is logged.
func TestTrekB24HitItLooksAtTheTopFiveWithoutRevealing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Ritual", TypeLine: "Sorcery"},
		game.Card{Name: "Relic", TypeLine: "Artifact"},
	)
	effectstest.CastCatalogSpell(t, g, "Hit It!", "Enchantment", trekB24HitItOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) == 0 {
		t.Fatal("no land prompt is open with two lands among the five")
	}
	for i, id := range ids {
		c, ok := g.LookupCardForEffect(id)
		if !ok || !me.Library.Contains(id) {
			t.Fatalf("card %d left the library before the pick", i)
		}
		if !c.IsKnownTo(me.ID) {
			t.Errorf("card %d is not known to its controller", i)
		}
		if c.IsKnownTo(opp.ID) {
			t.Errorf("card %d is known to an opponent", i)
		}
	}
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventRevealCards {
			t.Errorf("a reveal was logged for %v", ev.CardID)
		}
	}
}

// --- The Inner Light ------------------------------------------------------------

func TestTrekB24TheInnerLightChapterOneReturnsAGreenOrWhiteCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	green := trekB24Graveyard(g, me, game.Card{Name: "Elf", TypeLine: "Creature — Elf", Colors: []string{"G"}, Power: 1, Toughness: 1})
	red := trekB24Graveyard(g, me, game.Card{Name: "Goblin", TypeLine: "Creature — Goblin", Colors: []string{"R"}, Power: 1, Toughness: 1})
	saga := effectstest.CastCatalogSpell(t, g, "The Inner Light", "Enchantment — Saga", trekB24LightOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LoreCountersOn(g, saga) != 1 {
		t.Fatalf("lore counters = %d, want 1", effectstest.LoreCountersOn(g, saga))
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			for _, id := range c.PickTargetCards {
				if id == red {
					t.Fatal("a red creature card is offered")
				}
			}
		}
	}
	effectstest.AnswerPickTargetIfOpen(t, g, green)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(green) {
		t.Fatal("the green creature card did not return")
	}
	if !me.Graveyard.Contains(red) {
		t.Error("the red creature card left the graveyard")
	}
}

func TestTrekB24TheInnerLightChapterTwoDrawsThatManyOnEachHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Inner Light", TypeLine: "Enchantment — Saga", OracleID: trekB24LightOracle,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLore: 1},
	})
	hitter := effectstest.B16Creature(g, me.ID, "Big Beast", "Creature — Beast", 3, 3)
	other := effectstest.B16Creature(g, me.ID, "Other Beast", "Creature — Beast", 2, 2)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.AnswerPickTargetIfOpen(t, g, hitter)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.TrekB18AAttack(t, g, opp.ID, hitter, other)
	trekB24ThroughCombat(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3: the chosen creature dealt 3 and the other one is not watched", got)
	}
}

func TestTrekB24TheInnerLightChapterThreeGainsTheGreatestPowerAndMakesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saga := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Inner Light", TypeLine: "Enchantment — Saga", OracleID: trekB24LightOracle,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLore: 2},
	})
	effectstest.B16Creature(g, me.ID, "Big Beast", "Creature — Beast", 5, 5)
	effectstest.B16Creature(g, me.ID, "Small Beast", "Creature — Beast", 2, 2)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	life := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life = %d, want %d", me.Life, life+5)
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasures, want 1", n)
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}
