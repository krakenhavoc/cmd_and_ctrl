package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// woe_s_cards_test.go — Wilds of Eldraine, the S batch (#4471). Each test
// plays the real catalog entry and asserts what its text says.

const (
	woeSSaviorOracle      = "412eb5ec-bf13-45fb-94b3-c41e62b19d1d"
	woeSScreamPuffOracle  = "b46ddf4b-7ed8-49c4-8828-f0fbbc023c1f"
	woeSSkybeastOracle    = "2c53a913-8a99-4a15-9c47-e79af026dc05"
	woeSSongOracle        = "5f4d0c87-7d59-41d5-9910-86765a0c144b"
	woeSSpiderFoodOracle  = "38eabfea-5ea5-4baa-8ad6-7a7cb00e1fed"
	woeSStingbladeOracle  = "7938112a-3832-42b4-b123-27a42c06a290"
	woeSBoltOracle        = "1322734c-3c3e-4885-a6d2-d6460a16fca2"
	woeSProwlerOracle     = "d9c0e089-92ca-4fbd-9ddd-1802ac3553c9"
	woeSSuccumbOracle     = "73992788-8084-459c-8375-9ccc8579c5d7"
	woeSSugarRushOracle   = "57085b8e-a85e-4d50-a7d2-9bddfdaf411c"
	woeSWitchOracle       = "635b854d-de71-4c85-9755-8c97bf3cba36"
	woeSShepherdOracle    = shroudedShepherdOracleID
	woeSVanguardOracle    = stormkeldVanguardOracleID
	woeSCreatureTypeOfOne = "Creature — Bear"
)

func TestWoeSEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		woeSSaviorOracle:          "Savior of the Sleeping",
		woeSScreamPuffOracle:      "Scream Puff",
		woeSSkybeastOracle:        "Skybeast Tracker",
		woeSSongOracle:            "Song of Totentanz",
		woeSSpiderFoodOracle:      "Spider Food",
		woeSStingbladeOracle:      "Stingblade Assassin",
		woeSBoltOracle:            "Stonesplitter Bolt",
		woeSProwlerOracle:         "Stormkeld Prowler",
		woeSSuccumbOracle:         "Succumb to the Cold",
		woeSSugarRushOracle:       "Sugar Rush",
		woeSWitchOracle:           "Sweettooth Witch",
		woeSShepherdOracle:        "Shrouded Shepherd",
		woeSShepherdOracle + "#1": "Cleave Shadows",
		woeSVanguardOracle:        "Stormkeld Vanguard",
		woeSVanguardOracle + "#1": "Bear Down",
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

// --- fixtures ------------------------------------------------------------------

// woeSCastBigSpell casts a vanilla {X}{R} sorcery from the active
// player's hand with the given X and resolves everything it triggers.
func woeSCastBigSpell(t *testing.T, g *game.Game, x int) {
	t.Helper()
	castXSpell(t, g, "Red X Spell", "Sorcery", "", "{X}{R}", x, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
}

// --- Savior of the Sleeping ------------------------------------------------------

func TestWoeSSaviorOfTheSleepingGrowsWhenYourEnchantmentGoesToTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	savior := b12Push(g, me.ID, "Savior of the Sleeping", "Creature — Human Knight", woeSSaviorOracle, 2, 3)
	assertKeywords(t, g, savior, "vigilance")
	theirs := b12Permanent(g, opp.ID, "Their Charm", "Enchantment")
	woeB1Destroy(t, g, theirs)
	if n := countersOn(g, savior, game.CounterPlusOne); n != 0 {
		t.Fatalf("%d counters after an opponent's enchantment died, want 0", n)
	}
	mine := b12Permanent(g, me.ID, "My Charm", "Enchantment")
	woeB1Destroy(t, g, mine)
	if n := countersOn(g, savior, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after your enchantment died, want 1", n)
	}
}

// --- Scream Puff -----------------------------------------------------------------

func TestWoeSScreamPuffMakesAFoodWhenItConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	puff := b12Push(g, me.ID, "Scream Puff", "Creature — Horror", woeSScreamPuffOracle, 4, 5)
	assertKeywords(t, g, puff, "deathtouch")
	dealNoncombatDamage(g, me.ID, puff, opp.ID)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Food"); got != 0 {
		t.Fatalf("%d Food after noncombat damage, want 0", got)
	}
	dealCombatDamage(g, me.ID, puff, opp.ID)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food after combat damage to a player, want 1", got)
	}
}

// --- Skybeast Tracker and Stormkeld Prowler -------------------------------------

func TestWoeSSkybeastTrackerMakesAFoodForAFiveDrop(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	tracker := b12Push(g, me.ID, "Skybeast Tracker", "Creature — Giant Archer", woeSSkybeastOracle, 2, 4)
	assertKeywords(t, g, tracker, "reach")
	woeSCastBigSpell(t, g, 3) // mana value 4 on the stack
	if got := b27CountNamed(g, me.ID, "Food"); got != 0 {
		t.Fatalf("%d Food after a four-drop, want 0", got)
	}
	woeSCastBigSpell(t, g, 4) // mana value 5 on the stack (CR 202.3e)
	if got := b27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food after a five-drop, want 1", got)
	}
}

func TestWoeSStormkeldProwlerGetsTwoCountersForAFiveDrop(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	prowler := b12Push(g, me.ID, "Stormkeld Prowler", "Creature — Human Rogue", woeSProwlerOracle, 2, 1)
	woeSCastBigSpell(t, g, 3)
	if n := countersOn(g, prowler, game.CounterPlusOne); n != 0 {
		t.Fatalf("%d counters after a four-drop, want 0", n)
	}
	woeSCastBigSpell(t, g, 4)
	if n := countersOn(g, prowler, game.CounterPlusOne); n != 2 {
		t.Errorf("%d counters after a five-drop, want 2", n)
	}
}

// --- Song of Totentanz -----------------------------------------------------------

func TestWoeSSongOfTotentanzMakesXRatsAndHastesTheTeam(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	bear := b12Creature(g, me.ID, "My Bear", woeSCreatureTypeOfOne, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", woeSCreatureTypeOfOne, 2, 2)
	castXSpell(t, g, "Song of Totentanz", "Sorcery", woeSSongOracle, "{X}{R}", 2, nil)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Rat"); got != 2 {
		t.Fatalf("%d Rats, want 2", got)
	}
	var rats []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Rat" && c.Controller == me.ID {
				rats = append(rats, c.InstanceID)
			}
		}
	})
	for _, rat := range rats {
		if !hasEffectiveKeyword(t, g, rat, "haste") {
			t.Error("a new Rat does not have haste")
		}
		assertRestrictions(t, g, rat, game.CantBlock)
	}
	if !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("your creature does not have haste")
	}
	if hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature has haste")
	}
}

func TestWoeSSongOfTotentanzAtXZeroStillHastesTheTeam(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	bear := b12Creature(g, me.ID, "My Bear", woeSCreatureTypeOfOne, 2, 2)
	castXSpell(t, g, "Song of Totentanz", "Sorcery", woeSSongOracle, "{X}{R}", 0, nil)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Rat"); got != 0 {
		t.Errorf("%d Rats at X=0, want 0", got)
	}
	if !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("your creature does not have haste")
	}
}

// --- Spider Food -----------------------------------------------------------------

func TestWoeSSpiderFoodDestroysAFlierAndMakesAFood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	bird := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	bear := b12Creature(g, opp.ID, "Their Bear", woeSCreatureTypeOfOne, 2, 2)
	if err := castCatalogSpellErr(t, g, "Spider Food", "Sorcery", woeSSpiderFoodOracle, cardRefs(bear)); err == nil {
		t.Fatal("a creature without flying was a legal target")
	}
	castCatalogSpell(t, g, "Spider Food", "Sorcery", woeSSpiderFoodOracle, cardRefs(bird))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bird) {
		t.Error("the flier was not destroyed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the bear was destroyed")
	}
	if got := b27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food, want 1", got)
	}
}

func TestWoeSSpiderFoodWithNoTargetStillMakesAFood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	relic := b12Permanent(g, opp.ID, "Relic", "Artifact")
	castCatalogSpell(t, g, "Spider Food", "Sorcery", woeSSpiderFoodOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(relic) {
		t.Error("an untargeted artifact was destroyed")
	}
	if got := b27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food, want 1", got)
	}
}

// --- Stingblade Assassin ---------------------------------------------------------

func TestWoeSStingbladeAssassinDestroysADamagedOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	hurt := b12Creature(g, opp.ID, "Hurt Bear", woeSCreatureTypeOfOne, 2, 2)
	fresh := b12Creature(g, opp.ID, "Fresh Bear", woeSCreatureTypeOfOne, 2, 2)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, hurt, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	assassin := castAndResolveCreature(t, g, "Stingblade Assassin", "Creature — Faerie Assassin", woeSStingbladeOracle)
	assertKeywords(t, g, assassin, "flash", "flying")
	if p := latestPickTarget(g, me.ID); p != nil {
		if hasID(p.PickTargetCards, fresh) {
			t.Errorf("the Assassin offers an undamaged creature: %v", p.PickTargetCards)
		}
		pickCard(t, g, me.ID, hurt)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hurt) {
		t.Error("the damaged creature was not destroyed")
	}
	if !g.Battlefield.Contains(fresh) {
		t.Error("the undamaged creature was destroyed")
	}
}

// --- Stonesplitter Bolt ----------------------------------------------------------

func TestWoeSStonesplitterBoltDealsXOrTwiceXBargained(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	first := b12Creature(g, opp.ID, "Big Bear", woeSCreatureTypeOfOne, 9, 9)
	castXSpell(t, g, "Stonesplitter Bolt", "Instant", woeSBoltOracle, "{X}{R}", 2, cardRefs(first))
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, first); got != 2 {
		t.Errorf("unbargained: %d damage, want 2", got)
	}

	second := b12Creature(g, opp.ID, "Bigger Bear", woeSCreatureTypeOfOne, 9, 9)
	token := woeILToken(g, me.ID)
	p1g2Cast(t, g, "Stonesplitter Bolt", "Instant", woeSBoltOracle, "{X}{R}", game.CastSpellParams{
		Targets: cardRefs(second), XValue: 2, OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	if g.Battlefield.Contains(token) {
		t.Fatal("the bargained token was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, second); got != 4 {
		t.Errorf("bargained: %d damage, want 4", got)
	}
}

// --- Succumb to the Cold ---------------------------------------------------------

func TestWoeSSuccumbToTheColdTapsAndStunsTwoOpposingCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	a := b12Creature(g, opp.ID, "Bear A", woeSCreatureTypeOfOne, 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", woeSCreatureTypeOfOne, 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", woeSCreatureTypeOfOne, 2, 2)
	if err := castCatalogSpellErr(t, g, "Succumb to the Cold", "Instant", woeSSuccumbOracle, cardRefs(a, mine)); err == nil {
		t.Fatal("a creature you control was a legal target")
	}
	castCatalogSpell(t, g, "Succumb to the Cold", "Instant", woeSSuccumbOracle, cardRefs(a, b))
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if !isTapped(g, id) {
			t.Error("a target is untapped")
		}
		if n := countersOn(g, id, game.CounterStun); n != 1 {
			t.Errorf("%d stun counters on a target, want 1", n)
		}
	}
}

// --- Sugar Rush ------------------------------------------------------------------

func TestWoeSSugarRushPumpsThreeAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	seedLibrary(me, "A", "B")
	bear := b12Creature(g, me.ID, "My Bear", woeSCreatureTypeOfOne, 2, 2)
	castCatalogSpell(t, g, "Sugar Rush", "Instant", woeSSugarRushOracle, cardRefs(bear))
	hand := handSize(me)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 5 || tough != 2 {
		t.Errorf("the target is %d/%d, want 5/2", p, tough)
	}
	if got := handSize(me); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

// --- Sweettooth Witch ------------------------------------------------------------

func TestWoeSSweettoothWitchMakesAFoodAndEatsItToDrain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	witch := castCatalogSpell(t, g, "Sweettooth Witch", "Creature — Human Warlock", woeSWitchOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := b27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Fatalf("%d Food after the Witch entered, want 1", got)
	}
	relic := b12Permanent(g, me.ID, "Relic", "Artifact")
	them := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := g.ActivateCatalogAbility(me.ID, witch, 0, game.ActivateAbilityParams{Targets: them, SacrificeIDs: []uuid.UUID{relic}}); err == nil {
		t.Fatal("an artifact that is not a Food paid the cost")
	}
	life := lifeOf(g, opp.ID)
	food := findBattlefieldByName(g, "Food")
	floatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, witch, 0, game.ActivateAbilityParams{Strict: true, Targets: them, SacrificeIDs: []uuid.UUID{food}}); err != nil {
		t.Fatalf("{2}, sacrifice a Food (while summoning sick): %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-2 {
		t.Errorf("the target player is at %d, want %d", got, life-2)
	}
	if got := b27CountNamed(g, me.ID, "Food"); got != 0 {
		t.Errorf("%d Food left, want 0", got)
	}
}

// --- Shrouded Shepherd // Cleave Shadows -----------------------------------------

func TestWoeSShroudedShepherdPumpsACreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := woeB1Seats(g)
	bear := b12Creature(g, me.ID, "My Bear", woeSCreatureTypeOfOne, 2, 2)
	castAndResolveCreature(t, g, "Shrouded Shepherd", "Creature — Spirit Warrior", woeSShepherdOracle)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, tough)
	}
}

func TestWoeSCleaveShadowsShrinksOnlyOpposingCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	small := b12Creature(g, opp.ID, "Their Squirrel", woeSCreatureTypeOfOne, 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", woeSCreatureTypeOfOne, 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", woeSCreatureTypeOfOne, 2, 2)
	card := woeB1Adventure(me, woeSShepherdOracle,
		game.Face{Name: "Shrouded Shepherd", TypeLine: "Creature — Spirit Warrior", ManaCost: "{1}{W}", Colors: []string{"W"}, Power: 2, Toughness: 2},
		game.Face{Name: "Cleave Shadows", TypeLine: "Sorcery — Adventure", ManaCost: "{1}{B}", Colors: []string{"B"}})
	woeB1CastAdventure(t, g, me, card, nil)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(small) {
		t.Error("the opposing 1/1 survived -1/-1")
	}
	if p, tough := effectivePTOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("the opposing bear is %d/%d, want 1/1", p, tough)
	}
	if p, tough := effectivePTOf(t, g, mine); p != 2 || tough != 2 {
		t.Errorf("your bear is %d/%d, want 2/2", p, tough)
	}
}

// --- Stormkeld Vanguard // Bear Down ---------------------------------------------

func TestWoeSStormkeldVanguardCantBeBlockedBySmallCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	giant := b12Push(g, me.ID, "Stormkeld Vanguard", "Creature — Giant Warrior", woeSVanguardOracle, 6, 7)
	small := b12Creature(g, opp.ID, "Two Power", woeSCreatureTypeOfOne, 2, 5)
	big := b12Creature(g, opp.ID, "Three Power", woeSCreatureTypeOfOne, 3, 3)
	woeILAttacking(g, giant, opp.ID)
	if !woeILBlockRefused(g, giant, small) {
		t.Error("a creature with power 2 may block it")
	}
	if woeILBlockRefused(g, giant, big) {
		t.Error("a creature with power 3 may not block it")
	}
}

func TestWoeSBearDownDestroysAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := woeB1Seats(g)
	relic := b12Permanent(g, opp.ID, "Relic", "Artifact")
	card := woeB1Adventure(me, woeSVanguardOracle,
		game.Face{Name: "Stormkeld Vanguard", TypeLine: "Creature — Giant Warrior", ManaCost: "{4}{G}{G}", Colors: []string{"G"}, Power: 6, Toughness: 7},
		game.Face{Name: "Bear Down", TypeLine: "Sorcery — Adventure", ManaCost: "{1}{G}", Colors: []string{"G"}})
	woeB1CastAdventure(t, g, me, card, cardRefs(relic))
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact was not destroyed")
	}
}
