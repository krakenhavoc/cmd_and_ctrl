package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b24_b_cards_test.go — Star Trek, batch 24 (#4973), part B.

// trekB24ThroughCombat runs the combat the attack declared through its
// damage and settles the triggers it makes.
func trekB24ThroughCombat(t *testing.T, g *game.Game) {
	t.Helper()
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
}

// --- Gowron, Eager for Battle -------------------------------------------------

func TestTrekB24GowronGrowsWhenAnotherCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	gowron := b12Push(g, me.ID, "Gowron, Eager for Battle", "Legendary Creature — Klingon Warrior", trekB24GowronOracle, 3, 3)
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b15Destroy(t, g, theirs)
	if got := trekB24Counter(t, g, gowron, game.CounterPlusOne); got != 1 {
		t.Errorf("Gowron has %d counters after an opponent's creature died, want 1", got)
	}
	land := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	b15Destroy(t, g, land)
	if got := trekB24Counter(t, g, gowron, game.CounterPlusOne); got != 1 {
		t.Errorf("Gowron has %d counters after a land died, want 1", got)
	}
}

func TestTrekB24GowronSacrificingAKlingonGivesTrampleACardAndCostsALife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	gowron := b12Push(g, me.ID, "Gowron, Eager for Battle", "Legendary Creature — Klingon Warrior", trekB24GowronOracle, 3, 3)
	klingon := b16Creature(g, me.ID, "Klingon Grunt", "Creature — Klingon Warrior", 2, 1)
	hand, life := me.Hand.Size(), me.Life
	if err := g.ActivateCatalogAbility(me.ID, gowron, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{klingon}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(klingon) {
		t.Fatal("the Klingon was not sacrificed")
	}
	if !hasEffectiveKeyword(t, g, gowron, "trample") {
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	gowron := b12Push(g, me.ID, "Gowron, Eager for Battle", "Legendary Creature — Klingon Warrior", trekB24GowronOracle, 3, 3)
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand, life := me.Hand.Size(), me.Life
	if err := g.ActivateCatalogAbility(me.ID, gowron, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{gowron}}); err == nil {
		t.Fatal("Gowron sacrificed itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, gowron, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, gowron, "trample") {
		t.Error("Gowron has no trample")
	}
	if me.Hand.Size() != hand || me.Life != life {
		t.Errorf("hand %d -> %d, life %d -> %d: a non-Klingon draws nothing and costs no life", hand, me.Hand.Size(), life, me.Life)
	}
}

func TestTrekB24GowronCountsAChangelingAsAKlingon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	gowron := b12Push(g, me.ID, "Gowron, Eager for Battle", "Legendary Creature — Klingon Warrior", trekB24GowronOracle, 3, 3)
	shape := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Shapeshifter", TypeLine: "Creature — Shapeshifter",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID, Keywords: []string{"changeling"},
	})
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, gowron, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{shape}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d for a changeling, want 1", got)
	}
}

// --- Gray and Adira, Forever Joined -------------------------------------------

func TestTrekB24GrayAndAdiraCopyAfterANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Gray and Adira, Forever Joined", "Legendary Creature — Trill Human Officer", trekB24GrayOracle, 2, 3)
	hero := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Legendary Hero", TypeLine: "Legendary Creature — Human Soldier",
		Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	answerPickTargetIfOpen(t, g, hero)
	passPriorityAroundTable(t, g)
	var token game.Card
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Legendary Hero" && IsToken(c) {
			token, n = c, n+1
		}
	}
	if n != 1 {
		t.Fatalf("%d token copies, want 1", n)
	}
	if token.HasSupertype("Legendary") {
		t.Error("the token copy is legendary")
	}
	if !hasEffectiveKeyword(t, g, token.InstanceID, "haste") {
		t.Error("the token has no haste")
	}
	if !g.Battlefield.Contains(hero) {
		t.Error("the original left: the legend rule should not apply to a nonlegendary copy")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token.InstanceID) {
		t.Error("the token was not sacrificed at the end step")
	}
	if !g.Battlefield.Contains(hero) {
		t.Error("the original was sacrificed")
	}
}

func TestTrekB24GrayAndAdiraDoNothingWithoutANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Gray and Adira, Forever Joined", "Legendary Creature — Trill Human Officer", trekB24GrayOracle, 2, 3)
	b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatal("the trigger asked for a target after only a creature spell")
		}
	}
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if IsToken(c) {
			t.Fatalf("a token %q was made", c.Name)
		}
	}
}

// --- Hemmer, Fixing What Is Broken --------------------------------------------

func TestTrekB24HemmerLootsWhenItBecomesTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hemmer := b12Push(g, me.ID, "Hemmer, Fixing What Is Broken", "Legendary Creature — Aenar Artificer", trekB24HemmerOracle, 2, 3)
	hand := me.Hand.Size()
	grave := len(me.Graveyard.Cards)
	trekB18AAttack(t, g, opp.ID, hemmer)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard prompt after Hemmer became tapped")
	}
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want the same size after a loot", hand, me.Hand.Size())
	}
	if got := len(me.Graveyard.Cards) - grave; got != 1 {
		t.Errorf("%d cards discarded, want 1", got)
	}
}

func TestTrekB24HemmerGivesSpacecraftHexproofOnlyDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	b12Push(g, me.ID, "Hemmer, Fixing What Is Broken", "Legendary Creature — Aenar Artificer", trekB24HemmerOracle, 2, 3)
	ship := b12Permanent(g, me.ID, "Shuttle", "Artifact — Spacecraft")
	theirShip := b12Permanent(g, opp.ID, "Their Shuttle", "Artifact — Spacecraft")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, ship, "hexproof") {
		t.Error("your Spacecraft has no hexproof during your turn")
	}
	if hasEffectiveKeyword(t, g, theirShip, "hexproof") {
		t.Error("an opponent's Spacecraft has hexproof")
	}
	if hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("a non-Spacecraft has hexproof")
	}
	advanceToPrecombatMainOf(t, g, (seat+1)%len(g.Seats))
	if hasEffectiveKeyword(t, g, ship, "hexproof") {
		t.Error("your Spacecraft has hexproof during an opponent's turn")
	}
}

// --- Hit It! --------------------------------------------------------------------

func TestTrekB24HitItGivesHasteAndPutsALandFromTheTopFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Ritual", TypeLine: "Sorcery"},
		game.Card{Name: "Relic", TypeLine: "Artifact"},
		game.Card{Name: "Deep Card", TypeLine: "Sorcery"},
	)
	castCatalogSpell(t, g, "Hit It!", "Enchantment", trekB24HitItOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ids[1]) {
		t.Fatal("the land did not enter")
	}
	if !b16Tapped(t, g, ids[1]) {
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
	bear := b16Creature(g, me.ID, "Fresh Bear", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("a creature you control has no haste")
	}
}

func TestTrekB24HitItWithNoLandPutsAllFiveInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ids := seedSearchLibrary(me,
		game.Card{Name: "A", TypeLine: "Instant"},
		game.Card{Name: "B", TypeLine: "Instant"},
		game.Card{Name: "C", TypeLine: "Instant"},
		game.Card{Name: "D", TypeLine: "Instant"},
		game.Card{Name: "E", TypeLine: "Instant"},
	)
	castCatalogSpell(t, g, "Hit It!", "Enchantment", trekB24HitItOracle, nil)
	passPriorityAroundTable(t, g)
	for i, id := range ids {
		if !me.Graveyard.Contains(id) {
			t.Errorf("card %d is not in the graveyard", i)
		}
	}
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature has haste")
	}
}

// --- The Inner Light ------------------------------------------------------------

func TestTrekB24TheInnerLightChapterOneReturnsAGreenOrWhiteCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	green := trekB24Graveyard(g, me, game.Card{Name: "Elf", TypeLine: "Creature — Elf", Colors: []string{"G"}, Power: 1, Toughness: 1})
	red := trekB24Graveyard(g, me, game.Card{Name: "Goblin", TypeLine: "Creature — Goblin", Colors: []string{"R"}, Power: 1, Toughness: 1})
	saga := castCatalogSpell(t, g, "The Inner Light", "Enchantment — Saga", trekB24LightOracle, nil)
	passPriorityAroundTable(t, g)
	if loreCountersOn(g, saga) != 1 {
		t.Fatalf("lore counters = %d, want 1", loreCountersOn(g, saga))
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
	answerPickTargetIfOpen(t, g, green)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(green) {
		t.Fatal("the green creature card did not return")
	}
	if !me.Graveyard.Contains(red) {
		t.Error("the red creature card left the graveyard")
	}
}

func TestTrekB24TheInnerLightChapterTwoDrawsThatManyOnEachHit(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Inner Light", TypeLine: "Enchantment — Saga", OracleID: trekB24LightOracle,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLore: 1},
	})
	hitter := b16Creature(g, me.ID, "Big Beast", "Creature — Beast", 3, 3)
	other := b16Creature(g, me.ID, "Other Beast", "Creature — Beast", 2, 2)
	advanceToPrecombatMainOf(t, g, seat)
	answerPickTargetIfOpen(t, g, hitter)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	trekB18AAttack(t, g, opp.ID, hitter, other)
	trekB24ThroughCombat(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3: the chosen creature dealt 3 and the other one is not watched", got)
	}
}

func TestTrekB24TheInnerLightChapterThreeGainsTheGreatestPowerAndMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saga := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Inner Light", TypeLine: "Enchantment — Saga", OracleID: trekB24LightOracle,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLore: 2},
	})
	b16Creature(g, me.ID, "Big Beast", "Creature — Beast", 5, 5)
	b16Creature(g, me.ID, "Small Beast", "Creature — Beast", 2, 2)
	advanceToPrecombatMainOf(t, g, seat)
	life := me.Life
	passPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life = %d, want %d", me.Life, life+5)
	}
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasures, want 1", n)
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}
