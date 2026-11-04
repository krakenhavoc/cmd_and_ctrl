package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jmp_b6_cards_test.go — the Jumpstart Selesnya deck request, batch 6
// (#4893). Each test plays the real catalog entry and asserts what its
// text says, and what it refuses.

const (
	jmpB6PatronOracle    = "4d36ea21-e8cb-442d-ba3f-c399289ac208"
	jmpB6MuttOracle      = "48a53d6e-0320-4b9c-b692-6d0fc00e46cc"
	jmpB6UnicornOracle   = "3c6913cc-c543-45f8-ad3b-ccd282efa6f7"
	jmpB6SaviorOracle    = "49dc3434-2501-4788-84f2-93e55730f67c"
	jmpB6SurveyorOracle  = "85364398-923c-4ee9-8519-576fce1c26f7"
	jmpB6SwiftOracle     = "5820d760-8ed7-4abf-a959-c06ed86a3730"
	jmpB6RangerOracle    = "29a90fda-541e-4923-b3ed-e2b35d02c74b"
	jmpB6ThragtuskOracle = "0dd0e91a-d16b-4718-8d11-1a3fcf8e0753"
	jmpB6TrufflesOracle  = "8a15841a-dcb6-454a-b865-43a067f42b4e"
	jmpB6RetrieverOracle = "06a2a50d-f54f-4eb7-9cf8-ce37d802794a"
	jmpB6SteedOracle     = "4b5f30a6-b0fb-4dcb-85d5-1db1bba73737"
)

func TestJmpB6EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		jmpB6PatronOracle:    "Patron of the Valiant",
		jmpB6MuttOracle:      "Rambunctious Mutt",
		jmpB6UnicornOracle:   "Ronom Unicorn",
		jmpB6SaviorOracle:    "Selfless Savior",
		jmpB6SurveyorOracle:  "Skittering Surveyor",
		jmpB6SwiftOracle:     "Swift Response",
		jmpB6RangerOracle:    "Sylvan Ranger",
		jmpB6ThragtuskOracle: "Thragtusk",
		jmpB6TrufflesOracle:  "Trufflesnout",
		jmpB6RetrieverOracle: "Trusty Retriever",
		jmpB6SteedOracle:     "Valorous Steed",
	}
	for oracle, name := range want {
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

// jmpB6PickMode answers the chooser's open mode prompt.
func jmpB6PickMode(t *testing.T, g *game.Game, chooser uuid.UUID, modes ...int) error {
	t.Helper()
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, chooser)
	if pick == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	return g.ResolveModePick(pick.ID, chooser, modes)
}

// --- Patron of the Valiant ---

func TestJmpB6PatronAddsACounterToEachOfYourCountedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	countered := b12Creature(g, me.ID, "Countered", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	setCounters(g, countered, map[string]int{game.CounterPlusOne: 2})
	setCounters(g, theirs, map[string]int{game.CounterPlusOne: 1})

	patron := castHoldCreature(t, g, "Patron of the Valiant", "Creature — Angel", jmpB6PatronOracle, 4, 4)
	passPriorityAroundTable(t, g)
	if got := b18Counters(t, g, countered, game.CounterPlusOne); got != 3 {
		t.Errorf("the creature with counters has %d, want 3", got)
	}
	if got := b18Counters(t, g, plain, game.CounterPlusOne); got != 0 {
		t.Errorf("the creature without counters got %d, want 0", got)
	}
	if got := b18Counters(t, g, theirs, game.CounterPlusOne); got != 1 {
		t.Errorf("an opponent's creature has %d, want 1 (untouched)", got)
	}
	if got := b18Counters(t, g, patron, game.CounterPlusOne); got != 0 {
		t.Errorf("Patron entered without counters and got %d", got)
	}
}

// --- Rambunctious Mutt ---

func TestJmpB6MuttDestroysAnOpponentsArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Permanent(g, me.ID, "My Relic", "Artifact")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	ench := b12Permanent(g, opp.ID, "Their Aura", "Enchantment")

	castHoldCreature(t, g, "Rambunctious Mutt", "Creature — Dog", jmpB6MuttOracle, 3, 4)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	for _, bad := range []uuid.UUID{mine, bear} {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Error("your own artifact or an opponent's creature was accepted")
		}
	}
	pickCard(t, g, me.ID, ench)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the opponent's enchantment survived")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(bear) {
		t.Error("something else was destroyed")
	}
}

func TestJmpB6MuttWithNoLegalTargetDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Permanent(g, me.ID, "My Relic", "Artifact")
	castHoldCreature(t, g, "Rambunctious Mutt", "Creature — Dog", jmpB6MuttOracle, 3, 4)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatal("a target prompt opened with only your own artifact")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Error("your own artifact was destroyed")
	}
}

// --- Ronom Unicorn ---

func TestJmpB6RonomUnicornSacrificesToDestroyAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	unicorn := b12Push(g, me.ID, "Ronom Unicorn", "Creature — Unicorn", jmpB6UnicornOracle, 2, 2)
	relic := b12Permanent(g, opp.ID, "Relic", "Artifact")
	ench := b12Permanent(g, opp.ID, "Aura", "Enchantment")
	if err := g.ActivateCatalogAbility(me.ID, unicorn, 0, game.ActivateAbilityParams{Targets: b16TargetCard(relic)}); err == nil {
		t.Fatal("an artifact is not an enchantment")
	}
	b16Activate(t, g, me.ID, unicorn, 0, game.ActivateAbilityParams{Targets: b16TargetCard(ench)})
	if g.Battlefield.Contains(ench) || g.Battlefield.Contains(unicorn) {
		t.Error("the enchantment is not destroyed or the Unicorn is not sacrificed")
	}
	if !g.Battlefield.Contains(relic) {
		t.Error("the artifact was destroyed")
	}
}

// --- Selfless Savior ---

func TestJmpB6SelflessSaviorMakesAnotherCreatureIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	savior := b12Push(g, me.ID, "Selfless Savior", "Creature — Dog", jmpB6SaviorOracle, 1, 1)
	friend := b12Creature(g, me.ID, "Friend", "Creature — Bear", 2, 2)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	for _, bad := range []uuid.UUID{savior, foe} {
		if err := g.ActivateCatalogAbility(me.ID, savior, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bad)}); err == nil {
			t.Fatal("the Savior itself or an opponent's creature was accepted")
		}
	}
	b16Activate(t, g, me.ID, savior, 0, game.ActivateAbilityParams{Targets: b16TargetCard(friend)})
	if g.Battlefield.Contains(savior) {
		t.Error("the Savior was not sacrificed")
	}
	if !p1g2HasKeyword(g, friend, "indestructible") {
		t.Fatal("the friend did not gain indestructible")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(friend) })
	if !g.Battlefield.Contains(friend) {
		t.Error("the indestructible friend was destroyed")
	}
}

// --- Skittering Surveyor and Sylvan Ranger ---

func jmpB6FetchesABasicToHand(t *testing.T, name, typeLine, oracle string, power, toughness int) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Temple Garden", "Land — Forest Plains"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	castHoldCreature(t, g, name, typeLine, oracle, power, toughness)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	for _, bad := range []string{"Temple Garden", "Bear"} {
		if searchOptionNamed(g, c, bad) != uuid.Nil {
			t.Errorf("the search offered %s, which is not a basic land card", bad)
		}
	}
	handBefore := me.Hand.Size()
	answerSearchNamed(t, g, me.ID, "Forest")
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d -> %d, want +1", handBefore, me.Hand.Size())
	}
	if me.Library.Size() != 2 {
		t.Errorf("library has %d cards, want 2", me.Library.Size())
	}
}

func jmpB6DeclinesTheSearch(t *testing.T, name, typeLine, oracle string, power, toughness int) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	before := libraryIDs(me)
	castHoldCreature(t, g, name, typeLine, oracle, power, toughness)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if err := answerSearchApply(g, c, me.ID, me.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !sameIDList(before, libraryIDs(me)) {
		t.Error("declining the search still moved or shuffled the library")
	}
}

func TestJmpB6SkitteringSurveyorFetchesABasicLand(t *testing.T) {
	jmpB6FetchesABasicToHand(t, "Skittering Surveyor", "Artifact Creature — Construct", jmpB6SurveyorOracle, 1, 2)
}

func TestJmpB6SkitteringSurveyorMayDecline(t *testing.T) {
	jmpB6DeclinesTheSearch(t, "Skittering Surveyor", "Artifact Creature — Construct", jmpB6SurveyorOracle, 1, 2)
}

func TestJmpB6SylvanRangerFetchesABasicLand(t *testing.T) {
	jmpB6FetchesABasicToHand(t, "Sylvan Ranger", "Creature — Elf Scout Ranger", jmpB6RangerOracle, 1, 1)
}

func TestJmpB6SylvanRangerMayDecline(t *testing.T) {
	jmpB6DeclinesTheSearch(t, "Sylvan Ranger", "Creature — Elf Scout Ranger", jmpB6RangerOracle, 1, 1)
}

// --- Swift Response ---

func TestJmpB6SwiftResponseDestroysOnlyATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	untapped := b12Creature(g, opp.ID, "Awake", "Creature — Bear", 2, 2)
	tapped := b12Creature(g, opp.ID, "Asleep", "Creature — Bear", 2, 2)
	tapCard(g, tapped)
	if err := castCatalogSpellErr(t, g, "Swift Response", "Instant", jmpB6SwiftOracle, b16TargetCard(untapped)); err == nil {
		t.Fatal("an untapped creature was accepted")
	}
	castCatalogSpell(t, g, "Swift Response", "Instant", jmpB6SwiftOracle, b16TargetCard(tapped))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped creature survived")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("the untapped creature was destroyed")
	}
}

func TestJmpB6SwiftResponseFizzlesIfTheTargetUntaps(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	tapped := b12Creature(g, opp.ID, "Asleep", "Creature — Bear", 2, 2)
	tapCard(g, tapped)
	castCatalogSpell(t, g, "Swift Response", "Instant", jmpB6SwiftOracle, b16TargetCard(tapped))
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == tapped {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(tapped) {
		t.Error("a creature untapped in response was still destroyed")
	}
}

// --- Thragtusk ---

func TestJmpB6ThragtuskGainsFiveAndLeavesABeast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	tusk := castHoldCreature(t, g, "Thragtusk", "Creature — Beast", jmpB6ThragtuskOracle, 5, 3)
	passPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life %d -> %d, want +5", life, me.Life)
	}
	if n, _ := countTokensNamed(g, me.ID, "Beast"); n != 0 {
		t.Fatalf("a Beast appeared before Thragtusk left (%d)", n)
	}
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(tusk) })
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Beast"); n != 1 {
		t.Errorf("bouncing Thragtusk made %d Beasts, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Beast" && IsToken(c) && (c.Power != 3 || c.Toughness != 3) {
			t.Errorf("the Beast is %d/%d, want 3/3", c.Power, c.Toughness)
		}
	}
	if me.Life != life+5 {
		t.Errorf("leaving gained life too: %d", me.Life)
	}
}

func TestJmpB6ThragtuskDyingAlsoLeavesABeast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tusk := castHoldCreature(t, g, "Thragtusk", "Creature — Beast", jmpB6ThragtuskOracle, 5, 3)
	passPriorityAroundTable(t, g)
	b18Kill(t, g, tusk)
	if n, _ := countTokensNamed(g, me.ID, "Beast"); n != 1 {
		t.Errorf("Thragtusk dying made %d Beasts, want 1", n)
	}
}

// --- Trufflesnout ---

func TestJmpB6TrufflesnoutCounterMode(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	snout := castHoldCreature(t, g, "Trufflesnout", "Creature — Boar", jmpB6TrufflesOracle, 2, 2)
	if err := jmpB6PickMode(t, g, me.ID, 0); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b18Counters(t, g, snout, game.CounterPlusOne); got != 1 {
		t.Errorf("Trufflesnout has %d counters, want 1", got)
	}
	if me.Life != life {
		t.Errorf("the counter mode also gained life: %d -> %d", life, me.Life)
	}
}

func TestJmpB6TrufflesnoutLifeModeAndOnlyOneMode(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	snout := castHoldCreature(t, g, "Trufflesnout", "Creature — Boar", jmpB6TrufflesOracle, 2, 2)
	if err := jmpB6PickMode(t, g, me.ID, 0, 1); err == nil {
		t.Fatal("both modes were accepted on a choose-one")
	}
	if err := jmpB6PickMode(t, g, me.ID, 1); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d -> %d, want +4", life, me.Life)
	}
	if got := b18Counters(t, g, snout, game.CounterPlusOne); got != 0 {
		t.Errorf("the life mode also put %d counters", got)
	}
}

func TestJmpB6TrufflesnoutCounterModeAfterItLeftDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	snout := castHoldCreature(t, g, "Trufflesnout", "Creature — Boar", jmpB6TrufflesOracle, 2, 2)
	if err := jmpB6PickMode(t, g, me.ID, 0); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(snout) })
	passPriorityAroundTable(t, g)
	for _, c := range me.Hand.Cards {
		if c.InstanceID == snout && c.Counters[game.CounterPlusOne] != 0 {
			t.Error("the bounced Trufflesnout kept a counter")
		}
	}
}

// --- Trusty Retriever ---

func TestJmpB6TrustyRetrieverCounterMode(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dog := castHoldCreature(t, g, "Trusty Retriever", "Creature — Dog", jmpB6RetrieverOracle, 2, 3)
	if err := jmpB6PickMode(t, g, me.ID, 0); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b18Counters(t, g, dog, game.CounterPlusOne); got != 1 {
		t.Errorf("Trusty Retriever has %d counters, want 1", got)
	}
}

func TestJmpB6TrustyRetrieverReturnsAnArtifactOrEnchantmentCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	relic := gyCard(me, "Relic", "Artifact", "{2}", "")
	bear := gyCard(me, "Bear", "Creature — Bear", "{1}{G}", "")
	theirs := gyCard(opp, "Their Aura", "Enchantment — Aura", "{W}", "")
	castHoldCreature(t, g, "Trusty Retriever", "Creature — Dog", jmpB6RetrieverOracle, 2, 3)
	if err := jmpB6PickMode(t, g, me.ID, 1); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	for _, bad := range []uuid.UUID{bear, theirs} {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Error("a creature card or an opponent's card was accepted")
		}
	}
	pickCard(t, g, me.ID, relic)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(relic) || me.Graveyard.Contains(relic) {
		t.Error("the artifact card did not return to hand")
	}
	if !me.Graveyard.Contains(bear) || !opp.Graveyard.Contains(theirs) {
		t.Error("another card left a graveyard")
	}
}

// --- Valorous Steed ---

func TestJmpB6ValorousSteedMakesAVigilantKnight(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castHoldCreature(t, g, "Valorous Steed", "Creature — Unicorn", jmpB6SteedOracle, 3, 3)
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && IsToken(c) && c.Name == "Knight" {
			n++
			if c.Power != 2 || c.Toughness != 2 || !p1g2HasKeyword(g, c.InstanceID, "vigilance") {
				t.Errorf("the Knight is %d/%d without vigilance", c.Power, c.Toughness)
			}
		}
	}
	if n != 1 {
		t.Errorf("%d Knights, want 1", n)
	}
}

func TestJmpB6TrustyRetrieverReturnModeNeedsATarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gyCard(me, "Bear", "Creature — Bear", "{1}{G}", "")
	castHoldCreature(t, g, "Trusty Retriever", "Creature — Dog", jmpB6RetrieverOracle, 2, 3)
	if err := jmpB6PickMode(t, g, me.ID, 1); err == nil {
		t.Error("the return mode was chosen with no artifact or enchantment card in your graveyard")
	}
}
