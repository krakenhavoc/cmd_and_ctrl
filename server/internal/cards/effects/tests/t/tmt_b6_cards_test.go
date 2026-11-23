package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmt_b6_cards_test.go — Teenage Mutant Ninja Turtles, batch 6 (#4457).
// Each test plays the real catalog entry and asserts what its text says.

const (
	tmtB6PutridPalsOracle  = "3aa3b908-e8d3-4944-8ced-5219d4129594"
	tmtB6RaptorOracle      = "71b81f3f-6092-4819-8529-e27bccfcf088"
	tmtB6RaphaelOracle     = "2c72e6d2-bbb0-42de-b465-438d584cbddc"
	tmtB6RetroOracle       = "d8b5de6d-8654-4550-bccb-79de6cbef4d5"
	tmtB6RockSoldiersOrcl  = "0217da8e-e74f-4f5c-ab75-f17600f94405"
	tmtB6RocksteadyOracle  = "f9330a7b-8a65-4e04-a777-bf6512332850"
	tmtB6SavedOracle       = "7ef5c812-ad7d-4d6c-8ee5-68561e7e4bd7"
	tmtB6CamOracle         = "25c83d45-581f-48ba-b6bb-feb66e8a65c7"
	tmtB6RevengeOracle     = "5dd31fcd-72db-4daf-b127-ceb27bf25a94"
	tmtB6SkateboardOracle  = "decfde5e-ea53-4b15-8f69-51da7ef39d15"
	tmtB6PizzaOracle       = "9665f940-c0a4-4d3e-b889-8049cf549a4f"
	tmtB6SquirrelOracle    = "315ddc7a-c39c-43d9-b134-23acbd865e3e"
	tmtB6StockmanOracle    = "baa4cb1b-6fb5-46df-9cdd-54421ee0fc88"
	tmtB6StompedOracle     = "10185447-a13f-40d8-a8b6-39ce801f2fc0"
	tmtB6StockmanTypeLine  = "Legendary Creature — Insect Mutant Scientist"
	tmtB6RocksteadyTypeLn  = "Legendary Creature — Rhino Mutant"
	tmtB6RetroTypeLine     = "Enchantment — Aura"
	tmtB6SquirrelTypeLine  = "Creature — Squirrel Mutant"
	tmtB6PutridPalsTypeLn  = "Creature — Human Ooze Mutant"
	tmtB6SkateboardTypeLn  = "Artifact — Equipment"
	tmtB6PizzaTypeLine     = "Artifact — Food"
	tmtB6RockSoldiersTypeL = "Artifact Creature — Elemental Soldier"
)

func TestTmtB6EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tmtB6PutridPalsOracle: "Putrid Pals",
		tmtB6RaptorOracle:     "Ragamuffin Raptor",
		tmtB6RaphaelOracle:    "Raphael, Tough Turtle",
		tmtB6RetroOracle:      "Retro-Mutation",
		tmtB6RockSoldiersOrcl: "Rock Soldiers",
		tmtB6RocksteadyOracle: "Rocksteady, Crash Courser",
		tmtB6SavedOracle:      "Saved by the Shell",
		tmtB6CamOracle:        "Sewer-veillance Cam",
		tmtB6RevengeOracle:    "Shredder's Revenge",
		tmtB6SkateboardOracle: "Skateboard",
		tmtB6PizzaOracle:      "Spicy Oatmeal Pizza",
		tmtB6SquirrelOracle:   "Squirrelanoids",
		tmtB6StockmanOracle:   "Stockman, Mad Fly-entist",
		tmtB6StompedOracle:    "Stomped by the Foot",
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

// --- Putrid Pals -------------------------------------------------------------

func TestTmtB6PutridPalsEntersWithTwoCountersOnlyWithDisappear(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	plain := effectstest.CastAndResolveCreature(t, g, "Putrid Pals", tmtB6PutridPalsTypeLn, tmtB6PutridPalsOracle)
	if !g.Battlefield.Contains(plain) {
		t.Fatal("Putrid Pals is not on the battlefield")
	}
	if n := effectstest.BattlefieldPlusOnes(g, plain); n != 0 {
		t.Errorf("no permanent left: %d counters, want 0", n)
	}
	effectstest.AssertKeywords(t, g, plain, "deathtouch")
	effectstest.RevoltFor(t, g, me.ID)
	boosted := effectstest.CastAndResolveCreature(t, g, "Putrid Pals", tmtB6PutridPalsTypeLn, tmtB6PutridPalsOracle)
	if n := effectstest.BattlefieldPlusOnes(g, boosted); n != 2 {
		t.Errorf("a permanent left: %d counters, want 2", n)
	}
}

// --- Ragamuffin Raptor ---------------------------------------------------------

func TestTmtB6RagamuffinRaptorReturnsACreatureOrFoodCardFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	food := effectstest.PushGraveyardCardTyped(me, "Old Pizza", "Artifact — Food")
	bear := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	shock := effectstest.PushGraveyardCardTyped(me, "Shock", "Instant")
	theirs := effectstest.PushGraveyardCardTyped(opp, "Their Bear", "Creature — Bear")

	effectstest.CastCatalogSpell(t, g, "Ragamuffin Raptor", "Creature — Dinosaur", tmtB6RaptorOracle, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	for _, bad := range []uuid.UUID{shock, theirs} {
		if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(bad)); err == nil {
			t.Fatalf("Ragamuffin Raptor accepted %s; it takes a creature or Food card from your graveyard", bad)
		}
	}
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(food)...)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(food) || me.Graveyard.Contains(food) {
		t.Error("the Food card did not return to the hand")
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the untargeted creature card left the graveyard")
	}
}

// --- Raphael, Tough Turtle -------------------------------------------------------

func TestTmtB6RaphaelPingsTargetOpponentWhenAnotherCreatureOfYoursEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "Raphael, Tough Turtle", "Legendary Creature — Mutant Ninja Turtle", tmtB6RaphaelOracle, "{1}{R}", 1, 3, "R")
	oppLife, myLife := opp.Life, me.Life

	// An opponent's creature entering does nothing.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TokenCard("2/2 red Mutant"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || opp.Life != oppLife {
		t.Fatal("an opponent's creature entering triggered Raphael")
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("2/2 red Mutant"), 1) })
	effectstest.B04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTargets(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, []game.TargetRef{effectstest.PlayerTarget(me.ID)}); err == nil {
		t.Fatal("Raphael targeted his own controller; the target is an opponent")
	}
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.PlayerTarget(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if got := oppLife - opp.Life; got != 1 {
		t.Errorf("the opponent lost %d life, want 1", got)
	}
	if me.Life != myLife {
		t.Errorf("Raphael's controller lost %d life", myLife-me.Life)
	}
}

// --- Retro-Mutation ----------------------------------------------------------------

func TestTmtB6RetroMutationMakesAnAbilitylessZeroOneTurtleThatCantAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rats := effectstest.B27Push(g, me.ID, "Squirrelanoids", tmtB6SquirrelTypeLine, tmtB6SquirrelOracle, "{B}", 1, 1, "B")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(rats, game.CounterPlusOne, 1) })
	effectstest.AssertKeywords(t, g, rats, "deathtouch")

	aura := effectstest.CastCatalogSpell(t, g, "Retro-Mutation", tmtB6RetroTypeLine, tmtB6RetroOracle, effectstest.CardRefs(rats))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatal("Retro-Mutation is not on the battlefield")
	}
	// Base 0/1 is set in layer 7b; the +1/+1 counter still applies on top.
	if p, tough := effectstest.EffectivePower(t, g, rats), effectstest.EffectiveToughness(t, g, rats); p != 0 || tough != 1 {
		t.Errorf("enchanted creature's base is %d/%d, want 0/1", p, tough)
	}
	if got := effectstest.EffectivePowerWithCounters(t, g, rats); got != 1 {
		t.Errorf("power with its counter = %d, want 1", got)
	}
	if got := effectstest.EffectiveSubtypes(t, g, rats); len(got) != 1 || got[0] != "Turtle" {
		t.Errorf("subtypes = %v, want only Turtle", got)
	}
	if !effectstest.TmtB1Card(t, g, rats).IsCreature() {
		t.Error("the enchanted permanent stopped being a creature")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, rats, "deathtouch") {
		t.Error("the enchanted creature kept deathtouch")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(rats, opp.ID); err == nil {
		t.Error("the enchanted creature attacked")
	}
}

// --- Rock Soldiers -------------------------------------------------------------------

func TestTmtB6RockSoldiersDestroysUpToOneNoncreatureArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	golem := effectstest.B12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 3, 3)

	soldiers := effectstest.CastCatalogSpell(t, g, "Rock Soldiers", tmtB6RockSoldiersTypeL, tmtB6RockSoldiersOrcl, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTargets(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, effectstest.CardRefs(golem)); err == nil {
		t.Fatal("Rock Soldiers targeted an artifact creature")
	}
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(rock)...)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the noncreature artifact survived")
	}
	if !g.Battlefield.Contains(golem) || !g.Battlefield.Contains(soldiers) {
		t.Error("the artifact creature or Rock Soldiers left the battlefield")
	}
}

// --- Rocksteady, Crash Courser -----------------------------------------------------

func TestTmtB6RocksteadyAndYourBoarsCantBeBlockedByMoreThanOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rocksteady := effectstest.B27Push(g, me.ID, "Rocksteady, Crash Courser", tmtB6RocksteadyTypeLn, tmtB6RocksteadyOracle, "{4}{G}{G}", 7, 7, "G")
	boar := effectstest.B12Creature(g, me.ID, "My Boar", "Creature — Boar", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	a := effectstest.B12Creature(g, opp.ID, "Blocker A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, opp.ID, "Blocker B", "Creature — Bear", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{rocksteady, boar, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	for _, limited := range []uuid.UUID{rocksteady, boar} {
		effectstest.BrRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
			{Blocker: a, Attacker: limited}, {Blocker: b, Attacker: limited},
		}), game.BlockReasonTooManyBlockers)
	}
	if err := g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: bear}, {Blocker: b, Attacker: bear},
	}); err != nil {
		t.Errorf("a double block on a creature that is not a Boar was refused: %v", err)
	}
}

func TestTmtB6RocksteadyForestcyclingFetchesAForestCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	id, _ := effectstest.CycleFromHand(t, g, "Rocksteady, Crash Courser", tmtB6RocksteadyTypeLn, tmtB6RocksteadyOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the forestcycled card is not in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil || len(c.SearchCards) != 1 || c.SearchCards[0] != ids[0] {
		t.Fatalf("search = %+v, want only the Forest", c)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the Forest did not reach the hand")
	}
}

// --- Saved by the Shell ----------------------------------------------------------

func TestTmtB6SavedByTheShellCostsLessWithATurtle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if got := effectstest.SelfPricedMV(t, g, me, tmtB6SavedOracle, "Instant", "{1}{G}"); got != 2 {
		t.Errorf("without a Turtle it costs %d, want 2", got)
	}
	effectstest.B12Creature(g, me.ID, "Box Turtle", "Creature — Turtle", 1, 1)
	if got := effectstest.SelfPricedMV(t, g, me, tmtB6SavedOracle, "Instant", "{1}{G}"); got != 1 {
		t.Errorf("with a Turtle it costs %d, want 1", got)
	}
}

func TestTmtB6SavedByTheShellCountersAndProtectsYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Saved by the Shell", "Instant", tmtB6SavedOracle, effectstest.CardRefs(theirs)); err == nil {
		t.Fatal("Saved by the Shell targeted an opponent's creature")
	}
	effectstest.CastCatalogSpell(t, g, "Saved by the Shell", "Instant", tmtB6SavedOracle, effectstest.CardRefs(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, mine); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	effectstest.AssertKeywords(t, g, mine, "trample", "hexproof", "indestructible")

	effectstest.AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	for _, kw := range []string{"trample", "hexproof", "indestructible"} {
		if effectstest.EffectiveAbilitiesContain(t, g, mine, kw) {
			t.Errorf("%s outlasted the turn", kw)
		}
	}
	if got := effectstest.PlusOnes(g, mine); got != 1 {
		t.Errorf("the counter is gone after the turn: %d", got)
	}
}

// --- Sewer-veillance Cam ----------------------------------------------------------

func TestTmtB6SewerveillanceCamTapsOnEntryUntapsOnLeavingAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	cam := effectstest.CastCatalogSpell(t, g, "Sewer-veillance Cam", "Artifact", tmtB6CamOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, cam, "flash")
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	if !effectstest.B16Tapped(t, g, victim) {
		t.Fatal("\"tap it\" on entry did not tap the creature")
	}

	hand := me.Hand.Size()
	if err := effectstest.HobB4Activate(t, g, me.ID, cam, 0, "{C}{C}{C}{U}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(cam) {
		t.Fatal("the Cam was not sacrificed")
	}
	// The leaving half triggers on the sacrifice.
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, victim) {
		t.Error("\"untap it\" on leaving did not untap the creature")
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

// --- Shredder's Revenge -------------------------------------------------------------

func TestTmtB6ShreddersRevengeMakesTargetPlayerDiscardTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.B02cCastModalSpell(t, g, "Shredder's Revenge", "Sorcery", tmtB6RevengeOracle, []int{0},
		[]game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DiscardOwed(g, opp.ID); got != 2 {
		t.Errorf("the target owes %d discards, want 2", got)
	}
}

func TestTmtB6ShreddersRevengeMakesTargetPlayerDrawTwoAndLoseTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.SeedLibrary(opp, "One", "Two", "Three")
	hand, life := opp.Hand.Size(), opp.Life
	effectstest.B02cCastModalSpell(t, g, "Shredder's Revenge", "Sorcery", tmtB6RevengeOracle, []int{1},
		[]game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - hand; got != 2 {
		t.Errorf("the target drew %d, want 2", got)
	}
	if got := life - opp.Life; got != 2 {
		t.Errorf("the target lost %d life, want 2", got)
	}
}

// --- Skateboard ----------------------------------------------------------------------

func TestTmtB6SkateboardTapsOnEntryAndGivesPowerAndHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	board := effectstest.CastCatalogSpell(t, g, "Skateboard", tmtB6SkateboardTypeLn, tmtB6SkateboardOracle, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(theirs)...)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Fatal("the enters trigger did not tap its target")
	}

	effectstest.EquipTo(t, g, me.ID, board, mine)
	if got := effectstest.EffectivePower(t, g, mine); got != 3 {
		t.Errorf("equipped power = %d, want 3", got)
	}
	if got := effectstest.EffectiveToughness(t, g, mine); got != 2 {
		t.Errorf("equipped toughness = %d, want 2", got)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, mine, "haste") {
		t.Error("the equipped creature does not have haste")
	}
}

// --- Spicy Oatmeal Pizza -----------------------------------------------------------

func TestTmtB6SpicyOatmealPizzaDealsFourAndThreeThenIsAFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oppLife, myLife := opp.Life, me.Life

	pizza := effectstest.CastCatalogSpell(t, g, "Spicy Oatmeal Pizza", tmtB6PizzaTypeLine, tmtB6PizzaOracle, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.PlayerTarget(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if got := oppLife - opp.Life; got != 4 {
		t.Errorf("the target lost %d life, want 4", got)
	}
	if got := myLife - me.Life; got != 3 {
		t.Errorf("the controller lost %d life, want 3", got)
	}

	life := me.Life
	if err := effectstest.HobB4Activate(t, g, me.ID, pizza, 0, "{C}{C}", nil); err != nil {
		t.Fatalf("activate the Food ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pizza) {
		t.Error("the Pizza was not sacrificed")
	}
	if got := me.Life - life; got != 3 {
		t.Errorf("gained %d life, want 3", got)
	}
}

func TestTmtB6SpicyOatmealPizzaDealsNothingWhenItsTargetIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	myLife := me.Life
	effectstest.CastCatalogSpell(t, g, "Spicy Oatmeal Pizza", tmtB6PizzaTypeLine, tmtB6PizzaOracle, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(bear)...)
	effectstest.B27Kill(g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife {
		t.Errorf("with its only target gone the trigger still dealt %d to its controller", myLife-me.Life)
	}
}

// --- Squirrelanoids ------------------------------------------------------------------

func TestTmtB6SquirrelanoidsHasDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B27Push(g, me.ID, "Squirrelanoids", tmtB6SquirrelTypeLine, tmtB6SquirrelOracle, "{B}", 1, 1, "B")
	effectstest.AssertKeywords(t, g, id, "deathtouch")
}

// --- Stockman, Mad Fly-entist ------------------------------------------------------

func TestTmtB6StockmanLootsOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	stockman := effectstest.CastCatalogSpell(t, g, "Stockman, Mad Fly-entist", tmtB6StockmanTypeLine, tmtB6StockmanOracle, nil)
	hand, grave := me.Hand.Size(), me.Graveyard.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DiscardOwed(g, me.ID); got != 1 {
		t.Fatalf("owes %d discards after the draw, want 1", got)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Graveyard.Size() != grave+1 {
		t.Errorf("hand %d -> %d, graveyard %d -> %d; want a draw and a discard", hand, me.Hand.Size(), grave, me.Graveyard.Size())
	}
	effectstest.AssertKeywords(t, g, stockman, "flying")
}

func TestTmtB6StockmanIslandcyclingFetchesAnIslandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
	)
	id, _ := effectstest.CycleFromHand(t, g, "Stockman, Mad Fly-entist", tmtB6StockmanTypeLine, tmtB6StockmanOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the islandcycled card is not in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil || len(c.SearchCards) != 1 || c.SearchCards[0] != ids[0] {
		t.Fatalf("search = %+v, want only the Island", c)
	}
}

// --- Stomped by the Foot ---------------------------------------------------------

func TestTmtB6StompedByTheFootGivesMinusTwoUnkicked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	big := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Stomped by the Foot", "Instant", tmtB6StompedOracle, effectstest.CardRefs(big))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, big), effectstest.EffectiveToughness(t, g, big); p != 2 || tough != 2 {
		t.Errorf("unkicked: %d/%d, want 2/2", p, tough)
	}
}

func TestTmtB6StompedByTheFootKickedBySacrificeGivesMinusFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	big := effectstest.B12Creature(g, opp.ID, "Huge Bear", "Creature — Bear", 6, 6)
	fodder := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	effectstest.Pr7bCastNow(t, g, "Stomped by the Foot", "Instant", tmtB6StompedOracle, game.CastSpellParams{
		OptionalCosts: []int{0},
		SacrificeIDs:  []uuid.UUID{fodder},
		Targets:       effectstest.CardRefs(big),
	})
	if g.Battlefield.Contains(fodder) {
		t.Error("the kicker's sacrifice was not paid")
	}
	if p, tough := effectstest.EffectivePower(t, g, big), effectstest.EffectiveToughness(t, g, big); p != 1 || tough != 1 {
		t.Errorf("kicked: %d/%d, want 1/1 (-5/-5 instead of -2/-2)", p, tough)
	}
}
