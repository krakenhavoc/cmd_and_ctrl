package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

func TestGap110SupplyRunnersPutsACounterOnEachOtherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.PushVanillaCreature(g, me.ID, "A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "B", 1, 1)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	runners := effectstest.CastAndResolveCreature(t, g, "Supply Runners", "Creature — Dog", gap110SupplyOracle)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("creature %s has %d counters, want 1", id, got)
		}
	}
	if got := effectstest.CounterCount(g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's creature got %d counters", got)
	}
	if got := effectstest.CounterCount(g, runners, game.CounterPlusOne); got != 0 {
		t.Errorf("Supply Runners put %d counters on itself", got)
	}
}

func TestGap110MakeshiftMunitionsSacrificesAnArtifactOrCreatureForOneDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ench := effectstest.B12Push(g, me.ID, "Makeshift Munitions", "Enchantment", gap110MunitionsOracle, 0, 0)
	rock := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	land := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}

	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, ench, 0, game.ActivateAbilityParams{Strict: true, Targets: target, SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("a land was sacrificed to Makeshift Munitions")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.B16Activate(t, g, me.ID, ench, 0, game.ActivateAbilityParams{Strict: true, Targets: target, SacrificeIDs: []uuid.UUID{rock}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Fatalf("opponent life %d -> %d, want -1", life, opp.Life)
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.B16Activate(t, g, me.ID, ench, 0, game.ActivateAbilityParams{Strict: true, Targets: target, SacrificeIDs: []uuid.UUID{bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 || g.Battlefield.Contains(bear) {
		t.Errorf("creature sacrifice: life %d (want %d), bear still there %v", opp.Life, life-2, g.Battlefield.Contains(bear))
	}
}

func TestGap110OathOfJaceLootsOnEntryAndScriesPerPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 5; i++ {
		effectstest.PushLibraryTyped(me, "Filler", "Instant")
	}
	hand := me.Hand.Size()
	effectstest.CastAndResolveCreature(t, g, "Oath of Jace", "Legendary Enchantment", gap110OathOracle)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.B12Push(g, me.ID, "Oath of Jace", "Legendary Enchantment", gap110OathOracle, 0, 0)
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Jace A", TypeLine: "Legendary Planeswalker — Jace", Counters: map[string]int{game.CounterLoyalty: 3}})
	effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Jace B", TypeLine: "Legendary Planeswalker — Jace", Counters: map[string]int{game.CounterLoyalty: 3}})
	for i := 0; i < 5; i++ {
		effectstest.PushLibraryTyped(me, "Filler", "Instant")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	scry := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID)
	if scry == nil {
		t.Fatalf("no scry prompt at upkeep with two planeswalkers: %+v", g.PendingChoices)
	}
}

func TestGap110CacophonyUnleashedWipesNonenchantmentCreaturesOnlyWhenCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	nymph := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Nymph", TypeLine: "Enchantment Creature — Nymph", Power: 2, Toughness: 2})
	cac := effectstest.CastAndResolveCreature(t, g, "Cacophony Unleashed", "Enchantment", gap110CacophonyOracle)
	effectstest.MonSettle(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("a nonenchantment creature survived the wipe")
	}
	if !g.Battlefield.Contains(nymph) {
		t.Error("an enchantment creature was destroyed")
	}
	c := effectstest.GaLayered(t, g, cac)
	if !c.IsCreature() || !c.HasCardType("enchantment") || !c.HasSubtype("God") || !c.HasSubtype("Nightmare") || effectstest.EffectivePower(t, g, cac) != 6 {
		t.Errorf("after entering it is not a 6/6 enchantment Nightmare God creature (power %d)", effectstest.EffectivePower(t, g, cac))
	}
	if !effectstest.HasEffectiveKeyword(t, g, cac, "menace") || !effectstest.HasEffectiveKeyword(t, g, cac, "deathtouch") {
		t.Error("missing menace or deathtouch")
	}
}

func TestGap110CacophonyUnleashedDoesNotWipeWhenItWasNotCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Cacophony Unleashed", TypeLine: "Enchantment", OracleID: gap110CacophonyOracle, Owner: me.ID, Controller: me.ID})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("a Cacophony Unleashed that was not cast wiped the board")
	}
}

func TestGap110CacophonyUnleashedAnimatesWhenAnotherEnchantmentEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cac := effectstest.B12Push(g, me.ID, "Cacophony Unleashed", "Enchantment", gap110CacophonyOracle, 0, 0)
	if c := effectstest.GaLayered(t, g, cac); c.IsCreature() {
		t.Fatal("Cacophony Unleashed is a creature before any enchantment entered")
	}
	effectstest.CastAndResolveCreature(t, g, "Plain Aura", "Enchantment", "")
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.GaLayered(t, g, cac); !c.IsCreature() || effectstest.EffectivePower(t, g, cac) != 6 {
		t.Errorf("not animated by another enchantment entering: creature %v power %d", c.IsCreature(), effectstest.EffectivePower(t, g, cac))
	}
}

func TestGap110BattleForBretagardMakesTokensThenCopiesDifferentlyNamedOnes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	saga := effectstest.CastCatalogSpell(t, g, "Battle for Bretagard", "Enchantment — Saga", gap110BretagardOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Human Warrior"); n != 1 {
		t.Fatalf("chapter I: %d Human Warriors, want 1", n)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Elf Warrior"); n != 1 {
		t.Fatalf("chapter II: %d Elf Warriors, want 1", n)
	}
	// A second Human Warrior shares a name with the first: two of them cannot both be chosen.
	humans := []uuid.UUID{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.Name == "Human Warrior" {
			humans = append(humans, c.InstanceID)
		}
	}
	extra := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Human Warrior", TypeLine: "Token Creature — Human Warrior", Power: 1, Toughness: 1})
	elf := uuid.Nil
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.Name == "Elf Warrior" {
			elf = c.InstanceID
		}
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatalf("chapter III asked nothing: %+v", g.PendingChoices)
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{humans[0], extra}); err == nil {
		t.Fatal("two tokens with the same name were accepted")
	}
	answerPermanentsPick(t, g, me.ID, humans[0], elf)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Human Warrior"); n != 3 {
		t.Errorf("%d Human Warriors, want 3 (two plus one copy)", n)
	}
	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Elf Warrior"); n != 2 {
		t.Errorf("%d Elf Warriors, want 2", n)
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

func TestGap110BattleAtTheHelvaultExilesOnePerPlayerUntilItLeavesThenMakesAvacyn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	land := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.CastCatalogSpell(t, g, "Battle at the Helvault", "Enchantment — Saga", gap110HelvaultOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatalf("no target prompt for chapter I: %+v", g.PendingChoices)
	}
	if effectstest.HasID(prompt.PickTargetCards, land) {
		t.Error("a land was offered")
	}
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Fatal("chapter I did not exile one permanent per player")
	}
	if !g.Battlefield.Contains(land) {
		t.Fatal("the land was exiled")
	}
	// Chapters II and III: after III the Saga is sacrificed and everything returns, with Avacyn.
	for i := 0; i < 2; i++ {
		effectstest.AdvanceToPrecombatMainOf(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		for effectstest.LatestPickTarget(g, me.ID) != nil {
			if err := g.ResolvePickTargets(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, nil); err != nil {
				t.Fatalf("declining chapter II targets: %v", err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if n := effectstest.NamedOnBattlefield(g, "My Bear", me.ID); n != 1 {
		t.Errorf("my exiled creature: %d back on the battlefield, want 1", n)
	}
	if n := effectstest.NamedOnBattlefield(g, "Their Bear", opp.ID); n != 1 {
		t.Errorf("their exiled creature: %d back on the battlefield, want 1", n)
	}
	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Avacyn"); n != 1 {
		t.Errorf("%d Avacyn tokens, want 1", n)
	}
}
