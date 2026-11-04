package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmt_b8_cards_test.go — Teenage Mutant Ninja Turtles, batch 8 (#4457).
// Each test plays the real catalog entry and asserts what its text says.

const (
	tmtB8BebopOracle      = "a08b0a07-6da2-4ff6-88e5-88355dd84589"
	tmtB8BrillianceOracle = "b047cd3f-9129-41cc-8b91-b0dc970f5068"
	tmtB8CaseyOracle      = "bc163b5a-0b1a-4257-8c07-5c7d5284015f"
	tmtB8CourierOracle    = "049d8b06-db90-4b9c-8195-d987b3ef1005"
	tmtB8DoesOracle       = "df54b79d-f31d-4582-86c0-1d34ec77be0f"
	tmtB8DonLeoOracle     = "a4ae71c8-ea1b-4fa4-9e47-cf00ab60c5ee"
	tmtB8DonRaphOracle    = "01758a4a-ac29-47ad-8e88-158374f1ae4c"
	tmtB8DonatelloOracle  = "f6a7838d-380f-4752-a647-c55c2a0eb6e2"
	tmtB8PizzaOracle      = "aa481983-8f1a-4519-9371-5e7edcd3274f"
	tmtB8DroidOracle      = "c04d4fd4-e5aa-431d-824b-8bc94245f103"
	tmtB8KitsuneOracle    = "9be19bdf-db1e-433e-8be4-8b985a568211"
	tmtB8KoyaOracle       = "e7cd87d4-028c-49d9-9416-69f57e7dfec5"
	tmtB8KrangOracle      = "2385c8fb-9c38-4ff3-8f61-4e25a8c7d46b"
)

func TestTmtB8EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tmtB8BebopOracle:      "Bebop & Rocksteady",
		tmtB8BrillianceOracle: "Brilliance Unleashed",
		tmtB8CaseyOracle:      "Casey Jones, Vigilante",
		tmtB8CourierOracle:    "Courier of Comestibles",
		tmtB8DoesOracle:       "Does Machines",
		tmtB8DonLeoOracle:     "Don & Leo, Problem Solvers",
		tmtB8DonRaphOracle:    "Don & Raph, Hard Science",
		tmtB8DonatelloOracle:  "Donatello, Mutant Mechanic",
		tmtB8PizzaOracle:      "Everything Pizza",
		tmtB8DroidOracle:      "Fugitive Droid",
		tmtB8KitsuneOracle:    "Kitsune, Dragon's Daughter",
		tmtB8KoyaOracle:       "Koya, Death from Above",
		tmtB8KrangOracle:      "Krang & Shredder",
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

// tmtB8Answer resolves the latest choose-cards prompt for `who` with the
// given picks.
func tmtB8Answer(t *testing.T, g *game.Game, who uuid.UUID, picks ...uuid.UUID) {
	t.Helper()
	c := latestChooseCardsFor(g, who)
	if c == nil {
		t.Fatalf("no choose-cards prompt for %s: %+v", who, g.PendingChoices)
	}
	if err := g.ResolveChooseCards(c.ID, who, picks); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
}

// tmtB8Graveyard puts a card into a player's graveyard.
func tmtB8Graveyard(p *game.Player, name, typeLine string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID})
	return id
}

// tmtB8Attack declares `attacker` against `defender` and resolves what
// the declaration triggered, up to the first prompt.
func tmtB8Attack(t *testing.T, g *game.Game, attacker, defender uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	// Leaving the declaration puts the attack triggers on the stack; a
	// prompt one of them asks stops the walk.
	for i := 0; i < 8 && g.Turn.Step == game.StepDeclareAttackers && len(g.PendingChoices) == 0; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			break
		}
	}
	passPriorityAroundTable(t, g)
}

// tmtB8PT is a battlefield permanent's power and toughness with its
// counters.
func tmtB8PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	c := tmtB1Card(t, g, id)
	return c.CurrentPower(), c.CurrentToughness()
}

// --- Bebop & Rocksteady ------------------------------------------------------

func TestTmtB8BebopAndRocksteadyAttackDiscardsWhenYouChooseTo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	duo := b27Push(g, me.ID, "Bebop & Rocksteady", "Legendary Creature — Boar Rhino Mutant", tmtB8BebopOracle, "{1}{B/G}{B/G}", 7, 5, "B", "G")
	land := b12Permanent(g, me.ID, "My Land", "Land")
	tmtB8Attack(t, g, duo, opp.ID)
	hand, graves := me.Hand.Size(), me.Graveyard.Size()

	answerMayChoice(t, g, me.ID, true)
	c := latestChooseCardsFor(g, me.ID)
	if c == nil || len(c.ChooseCards) == 0 {
		t.Fatalf("the discard branch asks which card: %+v", g.PendingChoices)
	}
	tmtB8Answer(t, g, me.ID, c.ChooseCards[0])
	if me.Hand.Size() != hand-1 || me.Graveyard.Size() != graves+1 {
		t.Errorf("hand %d → %d, graveyard %d → %d; want one card discarded", hand, me.Hand.Size(), graves, me.Graveyard.Size())
	}
	if !g.Battlefield.Contains(land) || !g.Battlefield.Contains(duo) {
		t.Error("a permanent was sacrificed although a card was discarded")
	}
}

func TestTmtB8BebopAndRocksteadySacrificesAPermanentOtherwise(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	duo := b27Push(g, me.ID, "Bebop & Rocksteady", "Legendary Creature — Boar Rhino Mutant", tmtB8BebopOracle, "{1}{B/G}{B/G}", 7, 5, "B", "G")
	land := b12Permanent(g, me.ID, "My Land", "Land")
	tmtB8Attack(t, g, duo, opp.ID)
	hand := me.Hand.Size()

	answerMayChoice(t, g, me.ID, false)
	tmtB8Answer(t, g, me.ID, land)
	if g.Battlefield.Contains(land) {
		t.Error("the chosen land was not sacrificed")
	}
	if me.Hand.Size() != hand {
		t.Errorf("a card was discarded: hand %d → %d", hand, me.Hand.Size())
	}
}

func TestTmtB8BebopAndRocksteadyBlockWithAnEmptyHandAsksOnlyForTheSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	attacker, me := g.Seats[0], g.Seats[1]
	duo := b27Push(g, me.ID, "Bebop & Rocksteady", "Legendary Creature — Boar Rhino Mutant", tmtB8BebopOracle, "{1}{B/G}{B/G}", 7, 5, "B", "G")
	land := b12Permanent(g, me.ID, "My Land", "Land")
	bear := b12Creature(g, attacker.ID, "Their Bear", "Creature — Bear", 2, 2)
	me.Hand.Cards = nil
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, me.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(duo, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	finishBlockDeclarations(t, g)
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		t.Error("an empty hand was asked whether to discard")
	}
	tmtB8Answer(t, g, me.ID, land)
	if g.Battlefield.Contains(land) {
		t.Error("the block did not cost a permanent")
	}
}

// --- Brilliance Unleashed ----------------------------------------------------

func TestTmtB8BrillianceUnleashedReturnsANoncreatureArtifactAsAFlyingRobot(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gear := tmtB8Graveyard(me, "Spare Gear", "Artifact — Equipment", 0, 0)

	castModal(t, g, "Brilliance Unleashed", "Sorcery", tmtB8BrillianceOracle, []int{1}, cardRefs(gear))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(gear) {
		t.Fatal("the artifact card did not return to the battlefield")
	}
	c := tmtB1Card(t, g, gear)
	ch := effectiveOf(t, g, gear)
	// The Robot subtype replaces creature types only; the artifact keeps
	// its artifact type Equipment (CR 205.1a).
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Robot") || !c.HasSubtype("Equipment") {
		t.Errorf("types %v %v; want a Robot artifact creature that is still an Equipment", ch.Types, ch.Subtypes)
	}
	if ch.Power != 3 || ch.Toughness != 3 || !effectiveAbilitiesContain(t, g, gear, "flying") {
		t.Errorf("%d/%d flying=%v; want a 3/3 with flying", ch.Power, ch.Toughness, effectiveAbilitiesContain(t, g, gear, "flying"))
	}
	if controllerOf(t, g, gear) != me.ID {
		t.Error("the returned card is not the caster's")
	}
}

func TestTmtB8BrillianceUnleashedBothModesReturnAnArtifactCreatureAsItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	golem := tmtB8Graveyard(me, "Old Golem", "Artifact Creature — Golem", 5, 5)
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 5, 5)

	castModal(t, g, "Brilliance Unleashed", "Sorcery", tmtB8BrillianceOracle, []int{0, 1}, cardRefs(ogre, golem))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("5 damage did not kill the 5/5")
	}
	if !g.Battlefield.Contains(golem) {
		t.Fatal("the artifact creature card did not return")
	}
	ch := effectiveOf(t, g, golem)
	if ch.Power != 5 || ch.Toughness != 5 || !tmtB1Card(t, g, golem).HasSubtype("Golem") || effectiveAbilitiesContain(t, g, golem, "flying") {
		t.Errorf("the artifact creature changed: %d/%d %v", ch.Power, ch.Toughness, ch.Subtypes)
	}
}

// --- Casey Jones, Vigilante --------------------------------------------------

func TestTmtB8CaseyJonesDrawsThreeThenDiscardsThreeAtRandomAtYourNextUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	before := me.Hand.Size()
	casey := castCatalogSpell(t, g, "Casey Jones, Vigilante", "Legendary Creature — Human Berserker", tmtB8CaseyOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(casey) {
		t.Fatal("Casey Jones did not resolve")
	}
	if got := me.Hand.Size(); got != before+3 {
		t.Fatalf("hand %d → %d; want three drawn", before, got)
	}

	advanceToStepOf(t, g, 1, game.StepUpkeep)
	passPriorityAroundTable(t, g)
	held := me.Hand.Size()
	if held != before+3 {
		t.Errorf("an opponent's upkeep discarded: %d", held)
	}

	advanceToStepOf(t, g, 0, game.StepUpkeep)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != held-3 {
		t.Errorf("hand at your next upkeep %d → %d; want three discarded", held, got)
	}
}

// --- Courier of Comestibles --------------------------------------------------

func TestTmtB8CourierOfComestiblesFindsAFoodCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pizza := pushLibraryCardForTest(me, game.Card{Name: "Leftover Pizza", TypeLine: "Artifact — Food"})
	castCatalogSpell(t, g, "Courier of Comestibles", "Creature — Human Citizen", tmtB8CourierOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, pizza)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(pizza) {
		t.Error("the Food card is not in hand")
	}
	if n := tmtB1Count(g, me.ID, "Food"); n != 0 {
		t.Errorf("made %d Food tokens although a card was found", n)
	}
}

func TestTmtB8CourierOfComestiblesMakesAFoodWhenNothingIsFound(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushLibraryCardForTest(me, game.Card{Name: "Leftover Pizza", TypeLine: "Artifact — Food"})
	castCatalogSpell(t, g, "Courier of Comestibles", "Creature — Human Citizen", tmtB8CourierOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchFailToFind(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if n := tmtB1Count(g, me.ID, "Food"); n != 1 {
		t.Errorf("made %d Food tokens, want 1", n)
	}
}

func TestTmtB8CourierOfComestiblesMakesAFoodWhenTheSearchIsDeclined(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushLibraryCardForTest(me, game.Card{Name: "Leftover Pizza", TypeLine: "Artifact — Food"})
	castCatalogSpell(t, g, "Courier of Comestibles", "Creature — Human Citizen", tmtB8CourierOracle, nil)
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if err := g.DeclineSearchLibrary(c.ID, me.ID); err != nil {
		t.Fatalf("DeclineSearchLibrary: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := tmtB1Count(g, me.ID, "Food"); n != 1 {
		t.Errorf("made %d Food tokens, want 1", n)
	}
}

// --- Does Machines -----------------------------------------------------------

func TestTmtB8DoesMachinesEntersMillsDrawsThenDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	hand, lib, graves := me.Hand.Size(), me.Library.Size(), me.Graveyard.Size()
	castCatalogSpell(t, g, "Does Machines", "Enchantment — Class", tmtB8DoesOracle, nil)
	passPriorityAroundTable(t, g)
	c := latestChooseCardsFor(g, me.ID)
	if c == nil || len(c.ChooseCards) < 2 {
		t.Fatalf("no discard-two prompt: %+v", g.PendingChoices)
	}
	tmtB8Answer(t, g, me.ID, c.ChooseCards[0], c.ChooseCards[1])
	if got := lib - me.Library.Size(); got != 4 {
		t.Errorf("library lost %d cards, want 4 (two milled, two drawn)", got)
	}
	if got := me.Graveyard.Size() - graves; got != 4 {
		t.Errorf("graveyard gained %d, want 4 (two milled, two discarded)", got)
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d → %d; want two drawn and two discarded", hand, got)
	}
}

func TestTmtB8DoesMachinesLevelTwoReturnsArtifactsAndLevelThreeMakesARobot(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	class := b27Push(g, me.ID, "Does Machines", "Enchantment — Class", tmtB8DoesOracle, "{1}{U}", 0, 0, "U")
	a := tmtB8Graveyard(me, "Gear A", "Artifact", 0, 0)
	b := tmtB8Graveyard(me, "Gear B", "Artifact", 0, 0)
	bear := tmtB8Graveyard(me, "Dead Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")

	if err := hobB4Activate(t, g, me.ID, class, 0, "{C}{U}", nil); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for level 2")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(a, bear)); err == nil {
		t.Error("a creature card was accepted as an artifact card")
	}
	pickTargetRefs(t, g, me.ID, nil, cardRefs(a, b)...)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Error("the two artifact cards did not return to hand")
	}

	if err := hobB4Activate(t, g, me.ID, class, 1, "{C}{C}{C}{C}{U}", nil); err != nil {
		t.Fatalf("level 3: %v", err)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	pickTargetRefs(t, g, me.ID, nil, cardRefs(rock)...)
	passPriorityAroundTable(t, g)
	c := tmtB1Card(t, g, rock)
	ch := effectiveOf(t, g, rock)
	pw, tg := tmtB8PT(t, g, rock)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Robot") || pw != 3 || tg != 3 {
		t.Errorf("the artifact is %v %v %d/%d; want a 3/3 Robot artifact creature", ch.Types, ch.Subtypes, pw, tg)
	}
}

// --- Don & Leo, Problem Solvers ----------------------------------------------

func TestTmtB8DonAndLeoFlickersAnArtifactAndACreatureAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Don & Leo, Problem Solvers", "Legendary Creature — Mutant Ninja Turtle", tmtB8DonLeoOracle, "{3}{W/U}{W/U}", 4, 6, "W", "U")
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })

	advanceToStepOf(t, g, 0, game.StepEnd)
	pickTargetRefs(t, g, me.ID, nil, game.TargetRef{Kind: game.TargetCard, ID: rock})
	pickTargetRefs(t, g, me.ID, nil, game.TargetRef{Kind: game.TargetCard, ID: bear, Slot: 1})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(bear) {
		t.Error("the old objects are still on the battlefield")
	}
	rocks, bears := tmtB1Named(g, "My Rock"), tmtB1Named(g, "My Bear")
	if len(rocks) != 1 || len(bears) != 1 {
		t.Fatalf("returned %d Rocks and %d Bears, want one each", len(rocks), len(bears))
	}
	if plusOnes(g, bears[0]) != 0 {
		t.Error("the returned Bear kept its counter")
	}
}

// --- Don & Raph, Hard Science ------------------------------------------------

func TestTmtB8DonAndRaphAttackGivesTheNextNoncreatureSpellAffinity(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	duo := b27Push(g, me.ID, "Don & Raph, Hard Science", "Legendary Creature — Mutant Ninja Turtle", tmtB8DonRaphOracle, "{1}{U/R}{U/R}", 2, 4, "U", "R")
	b12Permanent(g, me.ID, "Rock A", "Artifact")
	b12Permanent(g, me.ID, "Rock B", "Artifact")
	assertKeywords(t, g, duo, "menace")
	sorcery := game.Card{Name: "Test Sorcery", TypeLine: "Sorcery", ManaCost: "{4}{R}", Owner: me.ID}
	if p := nsPrice(t, g, me, sorcery); p.Generic != 4 {
		t.Fatalf("before the attack: %+v", p)
	}

	tmtB8Attack(t, g, duo, opp.ID)
	if p := nsPrice(t, g, me, sorcery); p.Generic != 2 {
		t.Errorf("a noncreature spell with two artifacts: %+v, want {2}{R}", p)
	}
	creature := game.Card{Name: "Test Bear", TypeLine: "Creature — Bear", ManaCost: "{4}{R}", Owner: me.ID}
	if p := nsPrice(t, g, me, creature); p.Generic != 4 {
		t.Errorf("a creature spell got affinity: %+v", p)
	}
}

// --- Donatello, Mutant Mechanic ----------------------------------------------

func TestTmtB8DonatelloAnimatesAnArtifactAndMovesItsCountersWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	don := b27Push(g, me.ID, "Donatello, Mutant Mechanic", "Legendary Creature — Mutant Ninja Turtle", tmtB8DonatelloOracle, "{3}{U}", 3, 5, "U")
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := hobB4Activate(t, g, me.ID, don, 0, "", cardRefs(bear)); err == nil {
		t.Error("Donatello targeted a nonartifact creature")
	}
	if err := hobB4Activate(t, g, me.ID, don, 0, "", cardRefs(rock)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := tmtB1Card(t, g, rock)
	ch := effectiveOf(t, g, rock)
	pw, tg := tmtB8PT(t, g, rock)
	if !c.IsCreature() || !c.HasSubtype("Robot") || plusOnes(g, rock) != 3 || pw != 3 || tg != 3 {
		t.Fatalf("the Rock is %v %v with %d counters, %d/%d; want a 3/3 Robot creature", ch.Types, ch.Subtypes, plusOnes(g, rock), pw, tg)
	}

	b27Kill(g, rock)
	pickTargetRefs(t, g, me.ID, nil, cardRefs(bear)...)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, bear); got != 3 {
		t.Errorf("the Bear has %d +1/+1 counters, want the Rock's 3", got)
	}
}

func TestTmtB8DonatelloIgnoresAnArtifactWithNoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Donatello, Mutant Mechanic", "Legendary Creature — Mutant Ninja Turtle", tmtB8DonatelloOracle, "{3}{U}", 3, 5, "U")
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b27Kill(g, rock)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("an artifact with no counters triggered Donatello")
	}
}

// --- Everything Pizza --------------------------------------------------------

func TestTmtB8EverythingPizzaFetchesABasicAndDoesEverything(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	forest := pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	pizza := castCatalogSpell(t, g, "Everything Pizza", "Artifact — Food", tmtB8PizzaOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, forest)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(forest) {
		t.Fatal("the basic land is not in hand")
	}

	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	life, hand := me.Life, me.Hand.Size()
	oppHands := map[uuid.UUID]int{}
	for _, p := range g.Seats[1:] {
		oppHands[p.ID] = p.Hand.Size()
	}
	if err := hobB4Activate(t, g, me.ID, pizza, 0, "{C}{C}{W}{U}{B}{R}{G}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}, {Kind: game.TargetCard, ID: ogre, Slot: 1}, {Kind: game.TargetCard, ID: bear, Slot: 2}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(pizza) {
		t.Error("Everything Pizza was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		if c := latestChooseCardsFor(g, p.ID); c != nil && len(c.ChooseCards) > 0 {
			tmtB8Answer(t, g, p.ID, c.ChooseCards[0])
		}
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+3 || me.Hand.Size() != hand+1 {
		t.Errorf("life %d → %d, hand %d → %d; want +3 and one card", life, me.Life, hand, me.Hand.Size())
	}
	for _, p := range g.Seats[1:] {
		if p.Hand.Size() != oppHands[p.ID]-1 {
			t.Errorf("opponent %s hand %d → %d; want one discarded", p.Name, oppHands[p.ID], p.Hand.Size())
		}
	}
	if g.Battlefield.Contains(ogre) {
		t.Error("3 damage did not kill the 3/3")
	}
	if got := plusOnes(g, bear); got != 3 {
		t.Errorf("the Bear has %d counters, want 3", got)
	}
}

// --- Fugitive Droid ----------------------------------------------------------

func TestTmtB8FugitiveDroidIsUnblockableOnlyAfterAnArtifactEntered(t *testing.T) {
	for _, entered := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		droid := b27Push(g, me.ID, "Fugitive Droid", "Artifact Creature — Robot Scientist", tmtB8DroidOracle, "{U}", 1, 1, "U")
		wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
		advanceToMain(t, g)
		if entered {
			g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1) })
		}
		advanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(droid, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		advanceTo(t, g, game.StepDeclareBlockers)
		err := g.DeclareBlocker(wall, droid)
		if entered && err == nil {
			t.Error("blocked after an artifact entered under your control this turn")
		}
		if !entered && err != nil {
			t.Errorf("could not be blocked with no artifact entered: %v", err)
		}
	}
}

func TestTmtB8FugitiveDroidCountersASpellAimedAtYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	droid := b27Push(g, me.ID, "Fugitive Droid", "Artifact Creature — Robot Scientist", tmtB8DroidOracle, "{U}", 1, 1, "U")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)

	face := b10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	if err := hobB4Activate(t, g, me.ID, droid, 0, "{U}", cardRefs(face)); err == nil {
		t.Error("countered a spell that targets a player")
	}
	passPriorityAroundTable(t, g)

	bolt := b10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetCard, ID: bear})
	if err := hobB4Activate(t, g, me.ID, droid, 0, "{U}", cardRefs(bolt)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(droid) {
		t.Error("Fugitive Droid was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bolt resolved")
	}
	if b12ZoneOf(g, bolt) != game.ZoneGraveyard {
		t.Error("the countered Bolt is not in a graveyard")
	}
}

// --- Kitsune, Dragon's Daughter ----------------------------------------------

func TestTmtB8KitsuneExchangesTwoCreaturesOfDifferentPlayers(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	x := b12Creature(g, a.ID, "Creature X", "Creature — Bear", 2, 2)
	x2 := b12Creature(g, a.ID, "Creature X2", "Creature — Bear", 2, 2)
	y := b12Creature(g, b.ID, "Creature Y", "Creature — Bear", 2, 2)
	kitsune := castCatalogSpell(t, g, "Kitsune, Dragon's Daughter", "Legendary Creature — Fox Warlock Avatar", tmtB8KitsuneOracle, nil)
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, kitsune, "vigilance")
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for Kitsune's trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(x, x2)); err == nil {
		t.Error("two creatures controlled by the same player were accepted")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(kitsune, y)); err == nil {
		t.Error("Kitsune itself was accepted as a target")
	}
	pickTargetRefs(t, g, me.ID, nil, cardRefs(x, y)...)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if controllerOf(t, g, x) != b.ID || controllerOf(t, g, y) != a.ID {
		t.Error("control of the two creatures was not exchanged")
	}
	if controllerOf(t, g, x2) != a.ID {
		t.Error("an untargeted creature changed hands")
	}
}

// --- Koya, Death from Above --------------------------------------------------

func TestTmtB8KoyaReturnsTheExiledCreatureUnlessYouPay(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		advanceToMain(t, g)
		ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
		koya := castCatalogSpell(t, g, "Koya, Death from Above", "Legendary Creature — Mutant Ninja Bird", tmtB8KoyaOracle, nil)
		passPriorityAroundTable(t, g)
		assertKeywords(t, g, koya, "flying")
		pickTargetRefs(t, g, me.ID, nil, cardRefs(ogre)...)
		passPriorityAroundTable(t, g)
		if b12ZoneOf(g, ogre) != game.ZoneExile {
			t.Fatal("the creature was not exiled")
		}

		advanceToStepOf(t, g, 0, game.StepEnd)
		passPriorityAroundTable(t, g)
		if pay {
			if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{B}"); err != nil {
				t.Fatalf("AddManaForEffect: %v", err)
			}
		}
		answerPayUnless(t, g, me.ID, pay)
		passPriorityAroundTable(t, g)
		back := tmtB1Named(g, "Their Ogre")
		if pay && (len(back) != 0 || b12ZoneOf(g, ogre) != game.ZoneExile) {
			t.Error("paying {3}{B} did not keep the creature in exile")
		}
		if !pay {
			if len(back) != 1 {
				t.Fatal("declining did not return the creature")
			}
			if controllerOf(t, g, back[0]) != opp.ID {
				t.Error("the creature did not return under its owner's control")
			}
		}
	}
}

// --- Krang & Shredder --------------------------------------------------------

func TestTmtB8KrangAndShredderExilesUntilNonlandAndCastsOneFree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	var spells []uuid.UUID
	var lands []uuid.UUID
	for _, p := range g.Seats[1:] {
		bottom := game.Card{InstanceID: uuid.New(), Name: "Deep Card", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID}
		spell := game.Card{InstanceID: uuid.New(), Name: "Free Bear " + p.Name, TypeLine: "Creature — Bear", ManaCost: "{5}{G}", Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID}
		land1 := game.Card{InstanceID: uuid.New(), Name: "Top Land", TypeLine: "Land", Owner: p.ID, Controller: p.ID}
		land2 := game.Card{InstanceID: uuid.New(), Name: "Second Land", TypeLine: "Land", Owner: p.ID, Controller: p.ID}
		p.Library.Cards = []game.Card{bottom, spell, land2, land1}
		spells = append(spells, spell.InstanceID)
		lands = append(lands, land1.InstanceID, land2.InstanceID)
	}
	krang := castCatalogSpell(t, g, "Krang & Shredder", "Legendary Creature — Utrom Human Ninja", tmtB8KrangOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range append(append([]uuid.UUID(nil), spells...), lands...) {
		if b12ZoneOf(g, id) != game.ZoneExile {
			t.Errorf("%s was not exiled", id)
		}
	}
	for _, p := range g.Seats[1:] {
		if p.Library.Size() != 1 {
			t.Errorf("%s has %d cards left, want 1: the run stops at the first nonland card", p.Name, p.Library.Size())
		}
	}

	// Disappear needs a permanent to have left under your control.
	b27Kill(g, b12Permanent(g, me.ID, "My Rock", "Artifact"))
	advanceToStepOf(t, g, 0, game.StepEnd)
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, spells...)
	if err := castThroughResolutionCast(t, g, me.ID, spells[0], game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	if resolutionCastOfferFor(g, me.ID) != nil {
		if err := castThroughResolutionCast(t, g, me.ID, spells[1], game.CastSpellParams{Strict: true}); err == nil {
			t.Error("a second card was cast; Disappear casts one")
		}
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(spells[0]) || controllerOf(t, g, spells[0]) != me.ID {
		t.Error("the free creature spell did not resolve under your control")
	}
	if !g.Battlefield.Contains(krang) {
		t.Error("Krang & Shredder left")
	}
}

func TestTmtB8KrangAndShredderDisappearNeedsAPermanentToHaveLeft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Krang & Shredder", "Legendary Creature — Utrom Human Ninja", tmtB8KrangOracle, nil)
	passPriorityAroundTable(t, g)
	advanceToStepOf(t, g, 0, game.StepEnd)
	passPriorityAroundTable(t, g)
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Error("Disappear offered a cast with no permanent having left")
	}
}
