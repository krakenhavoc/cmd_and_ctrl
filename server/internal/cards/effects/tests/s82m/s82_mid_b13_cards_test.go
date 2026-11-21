package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mid_b13_cards_test.go — S82 Midnight Hunt Commander, batch 13 (#5497):
// Leinore, Autumn Sovereign; Sungold Sentinel; Sigarda, Champion of Light;
// Wall of Mourning; Sigarda's Vanguard; Sigardian Zealot; Celestial Judgment;
// Celebrate the Harvest.

const (
	midB13Leinore   = "cd47eedd-57cb-460d-b339-4785f63a5475"
	midB13Sentinel  = "410f5067-3380-4afb-9881-4b3bf0268493"
	midB13Sigarda   = "24919141-82d2-42ad-963c-e53511a868b3"
	midB13Wall      = "585c9671-7451-4506-8b21-0d64106c4cd7"
	midB13Vanguard  = "147d8a1f-fe82-4078-a281-29a2eed4c750"
	midB13Zealot    = "30aecffc-bbbf-44d0-8002-0ee5982f27b9"
	midB13Judgment  = "d18c40cd-0f38-498b-8882-10381a831751"
	midB13Celebrate = "de3764b9-9572-4f5c-8c90-82dedec47557"
)

func TestMidB13CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		midB13Leinore:   "Leinore, Autumn Sovereign",
		midB13Sentinel:  "Sungold Sentinel",
		midB13Sigarda:   "Sigarda, Champion of Light",
		midB13Wall:      "Wall of Mourning",
		midB13Vanguard:  "Sigarda's Vanguard",
		midB13Zealot:    "Sigardian Zealot",
		midB13Judgment:  "Celestial Judgment",
		midB13Celebrate: "Celebrate the Harvest",
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

// midB13Bears seeds one vanilla creature per power under p.
func midB13Bears(g *game.Game, p *game.Player, powers ...int) []uuid.UUID {
	var ids []uuid.UUID
	for _, pw := range powers {
		ids = append(ids, effectstest.B12Creature(g, p.ID, "Bear", "Creature — Bear", pw, 3))
	}
	return ids
}

// --- Celebrate the Harvest -------------------------------------------------

func TestMidB13CelebrateTheHarvestFetchesOneBasicPerDifferentPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	midB13Bears(g, me, 1, 2, 2, 3)
	// An opponent's creature of another power is not counted.
	midB13Bears(g, g.Seats[1], 7)
	forest := effectstest.B23Library(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Wastes Relic", TypeLine: "Land"})
	effectstest.CastCatalogSpell(t, g, "Celebrate the Harvest", "Sorcery", midB13Celebrate, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	c = effectstest.SearchChoiceFor(g, me.ID)
	if c.SearchMax != 3 {
		t.Fatalf("search max = %d, want 3 (three different powers)", c.SearchMax)
	}
	for _, id := range c.SearchCards {
		if id == forest[4] {
			t.Error("a nonbasic land was offered")
		}
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, forest[:3]); err != nil {
		t.Fatalf("take three basics: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range forest[:3] {
		card := effectstest.BattlefieldCardFor(g, id)
		if card == nil {
			t.Fatalf("%s did not reach the battlefield", id)
		}
		if !card.Tapped {
			t.Error("a fetched basic entered untapped")
		}
	}
	if effectstest.BattlefieldCardFor(g, forest[3]) != nil {
		t.Error("a fourth basic was fetched")
	}
}

func TestMidB13CelebrateTheHarvestWithNoCreaturesOnlyShuffles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B23Library(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.CastCatalogSpell(t, g, "Celebrate the Harvest", "Sorcery", midB13Celebrate, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("a search was offered with X = 0")
	}
}

// --- Leinore, Autumn Sovereign ---------------------------------------------

func TestMidB13LeinoreCounterCanMakeTheThirdPowerAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Leinore, Autumn Sovereign", "Legendary Creature — Human Noble", midB13Leinore, 0, 4)
	bears := midB13Bears(g, me, 1, 1)
	hand := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.AnswerTriggerTargets(t, g, me.ID, bears[1])
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Card(t, g, bears[1]).Counters[game.CounterPlusOne]; n != 1 {
		t.Fatalf("the target has %d +1/+1 counters, want 1", n)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card (powers 0, 1 and 2)", hand, got)
	}
}

func TestMidB13LeinoreWithoutThreePowersGrowsButDoesNotDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Leinore, Autumn Sovereign", "Legendary Creature — Human Noble", midB13Leinore, 0, 4)
	bears := midB13Bears(g, me, 3)
	hand := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.AnswerTriggerTargets(t, g, me.ID, bears[0])
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Card(t, g, bears[0]).Counters[game.CounterPlusOne]; n != 1 {
		t.Fatalf("counters = %d, want 1", n)
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d with two powers", hand, got)
	}
}

// --- Sungold Sentinel -------------------------------------------------------

func TestMidB13SungoldSentinelExilesAGraveyardCardOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	victim := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: victim, Name: "Gone", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	effectstest.CastHoldCreature(t, g, "Sungold Sentinel", "Creature — Human Soldier", midB13Sentinel, 3, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(victim) {
		t.Error("the targeted card is still in the graveyard")
	}
}

func TestMidB13SungoldSentinelCovenAbilityNeedsThreePowersAndGrantsColorEvasion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	sentinel := effectstest.B12Push(g, me.ID, "Sungold Sentinel", "Creature — Human Soldier", midB13Sentinel, 3, 2)
	effectstest.FillPoolColored(me, "W", 2)
	if err := g.ActivateCatalogAbility(me.ID, sentinel, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with only one power among creatures")
	}
	midB13Bears(g, me, 1, 2)
	effectstest.FillPoolColored(me, "W", 2)
	if err := g.ActivateCatalogAbility(me.ID, sentinel, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "R")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, sentinel, game.HexproofFromColor("R")) {
		t.Error("no hexproof from red after the ability resolved")
	}
	if effectstest.HasEffectiveKeyword(t, g, sentinel, game.HexproofFromColor("G")) {
		t.Error("hexproof from a color that was not chosen")
	}
	_ = opp
}

// --- Sigarda, Champion of Light --------------------------------------------

func TestMidB13SigardaAnthemsHumansOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Sigarda, Champion of Light", "Legendary Creature — Angel", midB13Sigarda, 4, 4)
	human := effectstest.B12Creature(g, me.ID, "Villager", "Creature — Human Peasant", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Their Villager", "Creature — Human Peasant", 1, 1)
	if p := effectstest.PowerOf(t, g, human); p != 2 {
		t.Errorf("a Human you control has power %d, want 2", p)
	}
	if p := effectstest.PowerOf(t, g, bear); p != 2 {
		t.Errorf("a Bear has power %d, want 2", p)
	}
	if p := effectstest.PowerOf(t, g, theirs); p != 1 {
		t.Errorf("an opponent's Human has power %d, want 1", p)
	}
}

func TestMidB13SigardaCovenAttackLooksAtFiveAndTakesAHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	sigarda := effectstest.B12Push(g, me.ID, "Sigarda, Champion of Light", "Legendary Creature — Angel", midB13Sigarda, 4, 4)
	midB13Bears(g, me, 1, 2)
	ids := effectstest.B23Library(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Villager", TypeLine: "Creature — Human Peasant"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Deep Human", TypeLine: "Creature — Human Peasant"})
	// b23Library pushes top-last: ids[0] is the bottom... the top five are ids[1:].
	human := ids[1]
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(sigarda, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no pick: %+v", g.PendingChoices)
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != human {
		t.Fatalf("offered %v, want only the Human among the top five", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{human}); err != nil {
		t.Fatalf("take the Human: %v", err)
	}
	if !me.Hand.Contains(human) {
		t.Error("the Human is not in hand")
	}
	if !me.Library.Contains(ids[5]) || me.Hand.Contains(ids[5]) {
		t.Error("the sixth card left the library")
	}
}

func TestMidB13SigardaAttackWithoutCovenDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	sigarda := effectstest.B12Push(g, me.ID, "Sigarda, Champion of Light", "Legendary Creature — Angel", midB13Sigarda, 4, 4)
	midB13Bears(g, me, 1)
	effectstest.B23Library(me, game.Card{Name: "Villager", TypeLine: "Creature — Human Peasant"})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(sigarda, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Error("two powers still triggered coven")
	}
}

// --- Wall of Mourning -------------------------------------------------------

func TestMidB13WallOfMourningExilesOnePerOpponentThenCovenReturnsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.B23Library(me,
		game.Card{Name: "A", TypeLine: "Creature — Bear"}, game.Card{Name: "B", TypeLine: "Creature — Bear"},
		game.Card{Name: "C", TypeLine: "Creature — Bear"}, game.Card{Name: "D", TypeLine: "Creature — Bear"})
	wall := effectstest.CastHoldCreature(t, g, "Wall of Mourning", "Creature — Wall", midB13Wall, 0, 4)
	effectstest.PassPriorityAroundTable(t, g)
	exiled := 0
	for _, id := range ids {
		if g.Exile.Contains(id) {
			exiled++
		}
	}
	if exiled != 3 {
		t.Fatalf("%d cards exiled, want one per opponent (3)", exiled)
	}
	if len(g.ExiledWithForEffect(game.ObjectRefOf(effectstest.B12Card(t, g, wall)))) != 3 {
		t.Fatal("the exiled cards are not linked to the Wall")
	}
	// Coven: powers 0 (the Wall), 1, 2.
	midB13Bears(g, me, 1, 2)
	hand := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no pick among the linked cards: %+v", g.PendingChoices)
	}
	if len(pick.ChooseCards) != 3 {
		t.Fatalf("offered %d cards, want 3", len(pick.ChooseCards))
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, pick.ChooseCards[:1]); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, got)
	}
}

func TestMidB13WallOfMourningWithoutCovenKeepsTheCardsExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B23Library(me,
		game.Card{Name: "A", TypeLine: "Creature — Bear"}, game.Card{Name: "B", TypeLine: "Creature — Bear"},
		game.Card{Name: "C", TypeLine: "Creature — Bear"})
	effectstest.CastHoldCreature(t, g, "Wall of Mourning", "Creature — Wall", midB13Wall, 0, 4)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil || me.Hand.Size() != hand {
		t.Error("a card came back with one power among creatures")
	}
}

// --- Sigarda's Vanguard ------------------------------------------------------

func TestMidB13VanguardGivesDoubleStrikeToChosenCreaturesWithDifferentPowers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bears := midB13Bears(g, me, 2, 2)
	theirs := midB13Bears(g, g.Seats[1], 5)
	effectstest.CastHoldCreature(t, g, "Sigarda's Vanguard", "Creature — Angel", midB13Vanguard, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no pick: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bears[0], bears[1]}); err == nil {
		t.Fatal("two creatures of one power were accepted")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bears[0], theirs[0]}); err != nil {
		t.Fatalf("choose a Bear and an opponent's creature: %v", err)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bears[0], "double strike") || !effectstest.HasEffectiveKeyword(t, g, theirs[0], "double strike") {
		t.Error("a chosen creature did not gain double strike")
	}
	if effectstest.HasEffectiveKeyword(t, g, bears[1], "double strike") {
		t.Error("the unchosen Bear gained double strike")
	}
}

// --- Sigardian Zealot --------------------------------------------------------

func TestMidB13ZealotPumpsChosenCreaturesByItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	zealot := effectstest.B12Push(g, me.ID, "Sigardian Zealot", "Creature — Human Cleric", midB13Zealot, 3, 3)
	bears := midB13Bears(g, me, 1, 1)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no pick: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bears[0], zealot}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	if p := effectstest.PowerOf(t, g, bears[0]); p != 4 {
		t.Errorf("the chosen Bear has power %d, want 1+3", p)
	}
	if p := effectstest.PowerOf(t, g, zealot); p != 6 {
		t.Errorf("the Zealot has power %d, want 3+3 (X read once)", p)
	}
	if p := effectstest.PowerOf(t, g, bears[1]); p != 1 {
		t.Errorf("the unchosen Bear has power %d", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bears[0], "vigilance") {
		t.Error("the chosen Bear has no vigilance")
	}
}

// --- Celestial Judgment ------------------------------------------------------

func TestMidB13CelestialJudgmentKeepsOneCreaturePerPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	mine := midB13Bears(g, me, 1, 1, 2)
	theirs := midB13Bears(g, g.Seats[1], 2, 3, 3, 3)
	effectstest.CastCatalogSpell(t, g, "Celestial Judgment", "Sorcery", midB13Judgment, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no pick: %+v", g.PendingChoices)
	}
	if pick.ChooseMin != 3 || pick.ChooseMax != 3 {
		t.Fatalf("pick bounds %d..%d, want exactly 3 (powers 1, 2, 3)", pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{mine[0], mine[1], theirs[1]}); err == nil {
		t.Fatal("two creatures of power 1 were accepted")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{mine[0], theirs[0], theirs[1]}); err != nil {
		t.Fatalf("choose one per power: %v", err)
	}
	g.RunStateChecksForTest()
	for _, id := range []uuid.UUID{mine[0], theirs[0], theirs[1]} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("a chosen creature %s was destroyed", id)
		}
	}
	for _, id := range []uuid.UUID{mine[1], mine[2], theirs[2], theirs[3]} {
		if g.Battlefield.Contains(id) {
			t.Errorf("an unchosen creature %s survived", id)
		}
	}
}

func TestMidB13CelestialJudgmentWithAllPowersDifferentDestroysNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	mine := midB13Bears(g, me, 1, 2, 3)
	effectstest.CastCatalogSpell(t, g, "Celestial Judgment", "Sorcery", midB13Judgment, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Error("a pick was asked with every power different")
	}
	g.RunStateChecksForTest()
	for _, id := range mine {
		if !g.Battlefield.Contains(id) {
			t.Error("a creature was destroyed")
		}
	}
}
