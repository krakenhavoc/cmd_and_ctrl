package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardssh "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/sh"
	cardsst "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/st"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	woeSShepherdOracle    = cardssh.ShroudedShepherdOracleID
	woeSVanguardOracle    = cardsst.StormkeldVanguardOracleID
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
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- fixtures ------------------------------------------------------------------

// woeSCastBigSpell casts a vanilla {X}{R} sorcery from the active
// player's hand with the given X and resolves everything it triggers.
func woeSCastBigSpell(t *testing.T, g *game.Game, x int) {
	t.Helper()
	effectstest.CastXSpell(t, g, "Red X Spell", "Sorcery", "", "{X}{R}", x, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Savior of the Sleeping ------------------------------------------------------

func TestWoeSSaviorOfTheSleepingGrowsWhenYourEnchantmentGoesToTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	savior := effectstest.B12Push(g, me.ID, "Savior of the Sleeping", "Creature — Human Knight", woeSSaviorOracle, 2, 3)
	effectstest.AssertKeywords(t, g, savior, "vigilance")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Charm", "Enchantment")
	effectstest.WoeB1Destroy(t, g, theirs)
	if n := effectstest.CountersOn(g, savior, game.CounterPlusOne); n != 0 {
		t.Fatalf("%d counters after an opponent's enchantment died, want 0", n)
	}
	mine := effectstest.B12Permanent(g, me.ID, "My Charm", "Enchantment")
	effectstest.WoeB1Destroy(t, g, mine)
	if n := effectstest.CountersOn(g, savior, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after your enchantment died, want 1", n)
	}
}

// --- Scream Puff -----------------------------------------------------------------

func TestWoeSScreamPuffMakesAFoodWhenItConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	puff := effectstest.B12Push(g, me.ID, "Scream Puff", "Creature — Horror", woeSScreamPuffOracle, 4, 5)
	effectstest.AssertKeywords(t, g, puff, "deathtouch")
	effectstest.DealNoncombatDamage(g, me.ID, puff, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 0 {
		t.Fatalf("%d Food after noncombat damage, want 0", got)
	}
	effectstest.DealCombatDamage(g, me.ID, puff, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food after combat damage to a player, want 1", got)
	}
}

// --- Skybeast Tracker and Stormkeld Prowler -------------------------------------

func TestWoeSSkybeastTrackerMakesAFoodForAFiveDrop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	tracker := effectstest.B12Push(g, me.ID, "Skybeast Tracker", "Creature — Giant Archer", woeSSkybeastOracle, 2, 4)
	effectstest.AssertKeywords(t, g, tracker, "reach")
	woeSCastBigSpell(t, g, 3) // mana value 4 on the stack
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 0 {
		t.Fatalf("%d Food after a four-drop, want 0", got)
	}
	woeSCastBigSpell(t, g, 4) // mana value 5 on the stack (CR 202.3e)
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food after a five-drop, want 1", got)
	}
}

func TestWoeSStormkeldProwlerGetsTwoCountersForAFiveDrop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	prowler := effectstest.B12Push(g, me.ID, "Stormkeld Prowler", "Creature — Human Rogue", woeSProwlerOracle, 2, 1)
	woeSCastBigSpell(t, g, 3)
	if n := effectstest.CountersOn(g, prowler, game.CounterPlusOne); n != 0 {
		t.Fatalf("%d counters after a four-drop, want 0", n)
	}
	woeSCastBigSpell(t, g, 4)
	if n := effectstest.CountersOn(g, prowler, game.CounterPlusOne); n != 2 {
		t.Errorf("%d counters after a five-drop, want 2", n)
	}
}

// --- Song of Totentanz -----------------------------------------------------------

func TestWoeSSongOfTotentanzMakesXRatsAndHastesTheTeam(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", woeSCreatureTypeOfOne, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", woeSCreatureTypeOfOne, 2, 2)
	effectstest.CastXSpell(t, g, "Song of Totentanz", "Sorcery", woeSSongOracle, "{X}{R}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Rat"); got != 2 {
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
		if !effectstest.HasEffectiveKeyword(t, g, rat, "haste") {
			t.Error("a new Rat does not have haste")
		}
		effectstest.AssertRestrictions(t, g, rat, game.CantBlock)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("your creature does not have haste")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature has haste")
	}
}

func TestWoeSSongOfTotentanzAtXZeroStillHastesTheTeam(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", woeSCreatureTypeOfOne, 2, 2)
	effectstest.CastXSpell(t, g, "Song of Totentanz", "Sorcery", woeSSongOracle, "{X}{R}", 0, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Rat"); got != 0 {
		t.Errorf("%d Rats at X=0, want 0", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("your creature does not have haste")
	}
}

// --- Spider Food -----------------------------------------------------------------

func TestWoeSSpiderFoodDestroysAFlierAndMakesAFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	bird := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", woeSCreatureTypeOfOne, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Spider Food", "Sorcery", woeSSpiderFoodOracle, effectstest.CardRefs(bear)); err == nil {
		t.Fatal("a creature without flying was a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Spider Food", "Sorcery", woeSSpiderFoodOracle, effectstest.CardRefs(bird))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bird) {
		t.Error("the flier was not destroyed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the bear was destroyed")
	}
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food, want 1", got)
	}
}

func TestWoeSSpiderFoodWithNoTargetStillMakesAFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	relic := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Spider Food", "Sorcery", woeSSpiderFoodOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(relic) {
		t.Error("an untargeted artifact was destroyed")
	}
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Errorf("%d Food, want 1", got)
	}
}

// --- Stingblade Assassin ---------------------------------------------------------

func TestWoeSStingbladeAssassinDestroysADamagedOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	hurt := effectstest.B12Creature(g, opp.ID, "Hurt Bear", woeSCreatureTypeOfOne, 2, 2)
	fresh := effectstest.B12Creature(g, opp.ID, "Fresh Bear", woeSCreatureTypeOfOne, 2, 2)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, hurt, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	assassin := effectstest.CastAndResolveCreature(t, g, "Stingblade Assassin", "Creature — Faerie Assassin", woeSStingbladeOracle)
	effectstest.AssertKeywords(t, g, assassin, "flash", "flying")
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if effectstest.HasID(p.PickTargetCards, fresh) {
			t.Errorf("the Assassin offers an undamaged creature: %v", p.PickTargetCards)
		}
		effectstest.PickCard(t, g, me.ID, hurt)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hurt) {
		t.Error("the damaged creature was not destroyed")
	}
	if !g.Battlefield.Contains(fresh) {
		t.Error("the undamaged creature was destroyed")
	}
}

// --- Stonesplitter Bolt ----------------------------------------------------------

func TestWoeSStonesplitterBoltDealsXOrTwiceXBargained(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	first := effectstest.B12Creature(g, opp.ID, "Big Bear", woeSCreatureTypeOfOne, 9, 9)
	effectstest.CastXSpell(t, g, "Stonesplitter Bolt", "Instant", woeSBoltOracle, "{X}{R}", 2, effectstest.CardRefs(first))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, first); got != 2 {
		t.Errorf("unbargained: %d damage, want 2", got)
	}

	second := effectstest.B12Creature(g, opp.ID, "Bigger Bear", woeSCreatureTypeOfOne, 9, 9)
	token := woeILToken(g, me.ID)
	effectstest.P1g2Cast(t, g, "Stonesplitter Bolt", "Instant", woeSBoltOracle, "{X}{R}", game.CastSpellParams{
		Targets: effectstest.CardRefs(second), XValue: 2, OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{token},
	})
	if g.Battlefield.Contains(token) {
		t.Fatal("the bargained token was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, second); got != 4 {
		t.Errorf("bargained: %d damage, want 4", got)
	}
}

// --- Succumb to the Cold ---------------------------------------------------------

func TestWoeSSuccumbToTheColdTapsAndStunsTwoOpposingCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	a := effectstest.B12Creature(g, opp.ID, "Bear A", woeSCreatureTypeOfOne, 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", woeSCreatureTypeOfOne, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", woeSCreatureTypeOfOne, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Succumb to the Cold", "Instant", woeSSuccumbOracle, effectstest.CardRefs(a, mine)); err == nil {
		t.Fatal("a creature you control was a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Succumb to the Cold", "Instant", woeSSuccumbOracle, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if !effectstest.IsTapped(g, id) {
			t.Error("a target is untapped")
		}
		if n := effectstest.CountersOn(g, id, game.CounterStun); n != 1 {
			t.Errorf("%d stun counters on a target, want 1", n)
		}
	}
}

// --- Sugar Rush ------------------------------------------------------------------

func TestWoeSSugarRushPumpsThreeAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	effectstest.SeedLibrary(me, "A", "B")
	bear := effectstest.B12Creature(g, me.ID, "My Bear", woeSCreatureTypeOfOne, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Sugar Rush", "Instant", woeSSugarRushOracle, effectstest.CardRefs(bear))
	hand := effectstest.HandSize(me)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 5 || tough != 2 {
		t.Errorf("the target is %d/%d, want 5/2", p, tough)
	}
	if got := effectstest.HandSize(me); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

// --- Sweettooth Witch ------------------------------------------------------------

func TestWoeSSweettoothWitchMakesAFoodAndEatsItToDrain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	witch := effectstest.CastCatalogSpell(t, g, "Sweettooth Witch", "Creature — Human Warlock", woeSWitchOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 1 {
		t.Fatalf("%d Food after the Witch entered, want 1", got)
	}
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	them := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := g.ActivateCatalogAbility(me.ID, witch, 0, game.ActivateAbilityParams{Targets: them, SacrificeIDs: []uuid.UUID{relic}}); err == nil {
		t.Fatal("an artifact that is not a Food paid the cost")
	}
	life := effectstest.LifeOf(g, opp.ID)
	food := effectstest.FindBattlefieldByName(g, "Food")
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, witch, 0, game.ActivateAbilityParams{Strict: true, Targets: them, SacrificeIDs: []uuid.UUID{food}}); err != nil {
		t.Fatalf("{2}, sacrifice a Food (while summoning sick): %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-2 {
		t.Errorf("the target player is at %d, want %d", got, life-2)
	}
	if got := effectstest.B27CountNamed(g, me.ID, "Food"); got != 0 {
		t.Errorf("%d Food left, want 0", got)
	}
}

// --- Shrouded Shepherd // Cleave Shadows -----------------------------------------

func TestWoeSShroudedShepherdPumpsACreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.WoeB1Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", woeSCreatureTypeOfOne, 2, 2)
	effectstest.CastAndResolveCreature(t, g, "Shrouded Shepherd", "Creature — Spirit Warrior", woeSShepherdOracle)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, tough)
	}
}

func TestWoeSCleaveShadowsShrinksOnlyOpposingCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	small := effectstest.B12Creature(g, opp.ID, "Their Squirrel", woeSCreatureTypeOfOne, 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", woeSCreatureTypeOfOne, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", woeSCreatureTypeOfOne, 2, 2)
	card := woeB1Adventure(me, woeSShepherdOracle,
		game.Face{Name: "Shrouded Shepherd", TypeLine: "Creature — Spirit Warrior", ManaCost: "{1}{W}", Colors: []string{"W"}, Power: 2, Toughness: 2},
		game.Face{Name: "Cleave Shadows", TypeLine: "Sorcery — Adventure", ManaCost: "{1}{B}", Colors: []string{"B"}})
	woeB1CastAdventure(t, g, me, card, nil)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(small) {
		t.Error("the opposing 1/1 survived -1/-1")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("the opposing bear is %d/%d, want 1/1", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 2 || tough != 2 {
		t.Errorf("your bear is %d/%d, want 2/2", p, tough)
	}
}

// --- Stormkeld Vanguard // Bear Down ---------------------------------------------

func TestWoeSStormkeldVanguardCantBeBlockedBySmallCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	giant := effectstest.B12Push(g, me.ID, "Stormkeld Vanguard", "Creature — Giant Warrior", woeSVanguardOracle, 6, 7)
	small := effectstest.B12Creature(g, opp.ID, "Two Power", woeSCreatureTypeOfOne, 2, 5)
	big := effectstest.B12Creature(g, opp.ID, "Three Power", woeSCreatureTypeOfOne, 3, 3)
	woeILAttacking(g, giant, opp.ID)
	if !woeILBlockRefused(g, giant, small) {
		t.Error("a creature with power 2 may block it")
	}
	if woeILBlockRefused(g, giant, big) {
		t.Error("a creature with power 3 may not block it")
	}
}

func TestWoeSBearDownDestroysAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.WoeB1Seats(g)
	relic := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	card := woeB1Adventure(me, woeSVanguardOracle,
		game.Face{Name: "Stormkeld Vanguard", TypeLine: "Creature — Giant Warrior", ManaCost: "{4}{G}{G}", Colors: []string{"G"}, Power: 6, Toughness: 7},
		game.Face{Name: "Bear Down", TypeLine: "Sorcery — Adventure", ManaCost: "{1}{G}", Colors: []string{"G"}})
	woeB1CastAdventure(t, g, me, card, effectstest.CardRefs(relic))
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact was not destroyed")
	}
}
