package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b110_test.go — S82 GAP batch 110 (Commander Masters creatures and
// enchantments): every test drives the real cast, trigger and layer paths.

const (
	gap110SupplyOracle    = "4d6da25d-7b62-4a64-bbaa-ae8edab1dd51"
	gap110CacophonyOracle = "1f6056be-67c8-40a9-86ef-1606961644eb"
	gap110BretagardOracle = "8294539d-7982-4ec9-8392-99eed78542f5"
	gap110MunitionsOracle = "2421bec6-7647-4684-be61-d1aa951c6b4f"
	gap110OathOracle      = "dc92b420-8ac7-4478-a287-3d897357be08"
	gap110HelvaultOracle  = "076584c8-ec58-4711-8f59-4839a4fa88ad"
)

func TestGap110CardsAreRegisteredComplete(t *testing.T) {
	for oracle, name := range map[string]string{
		gap110SupplyOracle:    "Supply Runners",
		gap110CacophonyOracle: "Cacophony Unleashed",
		gap110BretagardOracle: "Battle for Bretagard",
		gap110MunitionsOracle: "Makeshift Munitions",
		gap110OathOracle:      "Oath of Jace",
		gap110HelvaultOracle:  "Battle at the Helvault",
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

func TestGap110SupplyRunnersPutsACounterOnEachOtherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushVanillaCreature(g, me.ID, "A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "B", 1, 1)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	runners := castAndResolveCreature(t, g, "Supply Runners", "Creature — Dog", gap110SupplyOracle)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := counterCount(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("creature %s has %d counters, want 1", id, got)
		}
	}
	if got := counterCount(g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's creature got %d counters", got)
	}
	if got := counterCount(g, runners, game.CounterPlusOne); got != 0 {
		t.Errorf("Supply Runners put %d counters on itself", got)
	}
}

func TestGap110MakeshiftMunitionsSacrificesAnArtifactOrCreatureForOneDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ench := b12Push(g, me.ID, "Makeshift Munitions", "Enchantment", gap110MunitionsOracle, 0, 0)
	rock := pushTestPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	land := pushTestPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	advanceToMain(t, g)
	life := opp.Life
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}

	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, ench, 0, game.ActivateAbilityParams{Strict: true, Targets: target, SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("a land was sacrificed to Makeshift Munitions")
	}
	floatMana(t, g, me, "{C}")
	b16Activate(t, g, me.ID, ench, 0, game.ActivateAbilityParams{Strict: true, Targets: target, SacrificeIDs: []uuid.UUID{rock}})
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Fatalf("opponent life %d -> %d, want -1", life, opp.Life)
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	floatMana(t, g, me, "{C}")
	b16Activate(t, g, me.ID, ench, 0, game.ActivateAbilityParams{Strict: true, Targets: target, SacrificeIDs: []uuid.UUID{bear}})
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 || g.Battlefield.Contains(bear) {
		t.Errorf("creature sacrifice: life %d (want %d), bear still there %v", opp.Life, life-2, g.Battlefield.Contains(bear))
	}
}

func TestGap110OathOfJaceLootsOnEntryAndScriesPerPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 5; i++ {
		pushLibraryTyped(me, "Filler", "Instant")
	}
	hand := me.Hand.Size()
	castAndResolveCreature(t, g, "Oath of Jace", "Legendary Enchantment", gap110OathOracle)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+3 {
		t.Fatalf("hand %d -> %d, want +3 before discarding", hand, got)
	}
	if len(g.PendingChoices) == 0 {
		t.Fatal("no discard prompt after drawing three")
	}
	if c := g.PendingChoices[0]; c.Chooser != me.ID || c.Count != 2 {
		t.Errorf("first prompt %+v, want me discarding two", c)
	}
}

func TestGap110OathOfJaceUpkeepScriesForEachPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	b12Push(g, me.ID, "Oath of Jace", "Legendary Enchantment", gap110OathOracle, 0, 0)
	pushTestPermanent(g, me.ID, game.Card{Name: "Jace A", TypeLine: "Legendary Planeswalker — Jace", Counters: map[string]int{game.CounterLoyalty: 3}})
	pushTestPermanent(g, me.ID, game.Card{Name: "Jace B", TypeLine: "Legendary Planeswalker — Jace", Counters: map[string]int{game.CounterLoyalty: 3}})
	for i := 0; i < 5; i++ {
		pushLibraryTyped(me, "Filler", "Instant")
	}
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	scry := latestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID)
	if scry == nil {
		t.Fatalf("no scry prompt at upkeep with two planeswalkers: %+v", g.PendingChoices)
	}
}

func TestGap110CacophonyUnleashedWipesNonenchantmentCreaturesOnlyWhenCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	nymph := pushTestPermanent(g, opp.ID, game.Card{Name: "Nymph", TypeLine: "Enchantment Creature — Nymph", Power: 2, Toughness: 2})
	cac := castAndResolveCreature(t, g, "Cacophony Unleashed", "Enchantment", gap110CacophonyOracle)
	monSettle(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("a nonenchantment creature survived the wipe")
	}
	if !g.Battlefield.Contains(nymph) {
		t.Error("an enchantment creature was destroyed")
	}
	c := gaLayered(t, g, cac)
	if !c.IsCreature() || !c.HasCardType("enchantment") || !c.HasSubtype("God") || !c.HasSubtype("Nightmare") || effectivePower(t, g, cac) != 6 {
		t.Errorf("after entering it is not a 6/6 enchantment Nightmare God creature (power %d)", effectivePower(t, g, cac))
	}
	if !hasEffectiveKeyword(t, g, cac, "menace") || !hasEffectiveKeyword(t, g, cac, "deathtouch") {
		t.Error("missing menace or deathtouch")
	}
}

func TestGap110CacophonyUnleashedDoesNotWipeWhenItWasNotCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Cacophony Unleashed", TypeLine: "Enchantment", OracleID: gap110CacophonyOracle, Owner: me.ID, Controller: me.ID})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("a Cacophony Unleashed that was not cast wiped the board")
	}
}

func TestGap110CacophonyUnleashedAnimatesWhenAnotherEnchantmentEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cac := b12Push(g, me.ID, "Cacophony Unleashed", "Enchantment", gap110CacophonyOracle, 0, 0)
	if c := gaLayered(t, g, cac); c.IsCreature() {
		t.Fatal("Cacophony Unleashed is a creature before any enchantment entered")
	}
	castAndResolveCreature(t, g, "Plain Aura", "Enchantment", "")
	passPriorityAroundTable(t, g)
	if c := gaLayered(t, g, cac); !c.IsCreature() || effectivePower(t, g, cac) != 6 {
		t.Errorf("not animated by another enchantment entering: creature %v power %d", c.IsCreature(), effectivePower(t, g, cac))
	}
}

func TestGap110BattleForBretagardMakesTokensThenCopiesDifferentlyNamedOnes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	saga := castCatalogSpell(t, g, "Battle for Bretagard", "Enchantment — Saga", gap110BretagardOracle, nil)
	passPriorityAroundTable(t, g)
	if n := battlefieldCountNamed(g, me.ID, "Human Warrior"); n != 1 {
		t.Fatalf("chapter I: %d Human Warriors, want 1", n)
	}
	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := battlefieldCountNamed(g, me.ID, "Elf Warrior"); n != 1 {
		t.Fatalf("chapter II: %d Elf Warriors, want 1", n)
	}
	// A second Human Warrior shares a name with the first: two of them cannot both be chosen.
	humans := []uuid.UUID{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.Name == "Human Warrior" {
			humans = append(humans, c.InstanceID)
		}
	}
	extra := pushTestPermanent(g, me.ID, game.Card{Name: "Human Warrior", TypeLine: "Token Creature — Human Warrior", Power: 1, Toughness: 1})
	elf := uuid.Nil
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.Name == "Elf Warrior" {
			elf = c.InstanceID
		}
	}
	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatalf("chapter III asked nothing: %+v", g.PendingChoices)
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{humans[0], extra}); err == nil {
		t.Fatal("two tokens with the same name were accepted")
	}
	answerPermanentsPick(t, g, me.ID, humans[0], elf)
	passPriorityAroundTable(t, g)
	if n := battlefieldCountNamed(g, me.ID, "Human Warrior"); n != 3 {
		t.Errorf("%d Human Warriors, want 3 (two plus one copy)", n)
	}
	if n := battlefieldCountNamed(g, me.ID, "Elf Warrior"); n != 2 {
		t.Errorf("%d Elf Warriors, want 2", n)
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

func TestGap110BattleAtTheHelvaultExilesOnePerPlayerUntilItLeavesThenMakesAvacyn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	land := pushTestPermanent(g, opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	castCatalogSpell(t, g, "Battle at the Helvault", "Enchantment — Saga", gap110HelvaultOracle, nil)
	passPriorityAroundTable(t, g)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatalf("no target prompt for chapter I: %+v", g.PendingChoices)
	}
	if hasID(prompt.PickTargetCards, land) {
		t.Error("a land was offered")
	}
	pickCard(t, g, me.ID, mine)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Fatal("chapter I did not exile one permanent per player")
	}
	if !g.Battlefield.Contains(land) {
		t.Fatal("the land was exiled")
	}
	// Chapters II and III: after III the Saga is sacrificed and everything returns, with Avacyn.
	for i := 0; i < 2; i++ {
		advanceToPrecombatMainOf(t, g, 0)
		passPriorityAroundTable(t, g)
		for latestPickTarget(g, me.ID) != nil {
			if err := g.ResolvePickTargets(latestPickTarget(g, me.ID).ID, me.ID, nil); err != nil {
				t.Fatalf("declining chapter II targets: %v", err)
			}
		}
		passPriorityAroundTable(t, g)
	}
	if n := namedOnBattlefield(g, "My Bear", me.ID); n != 1 {
		t.Errorf("my exiled creature: %d back on the battlefield, want 1", n)
	}
	if n := namedOnBattlefield(g, "Their Bear", opp.ID); n != 1 {
		t.Errorf("their exiled creature: %d back on the battlefield, want 1", n)
	}
	if n := battlefieldCountNamed(g, me.ID, "Avacyn"); n != 1 {
		t.Errorf("%d Avacyn tokens, want 1", n)
	}
}
