package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmc_b7_cards_test.go — TMNT Commander batch 7 (#5011). Each test plays
// the real catalog entry and asserts what its text says.

const (
	tmcB7SplinterLeoOracle  = "c01095ba-b9c9-44e0-97d1-35cc47c4ed04"
	tmcB7SplinterAgedOracle = "45a87131-ae8a-4e6d-9ac0-e52b4795691b"
	tmcB7RegentaOracle      = "2650d2d5-576e-4b9a-95ac-0f5307f1580f"
	tmcB7CaneOracle         = "d41eec6c-de14-4809-abae-b94355674a33"
	tmcB7AprilOracle        = "020963ef-24c0-48cc-8776-bc257df684bc"
	tmcB7ContainmentOracle  = "d33c7ea2-4d17-4cdb-aa58-c25c57836297"
	tmcB7DonatelloOracle    = "671209d5-8b56-4ce7-94a0-d770d698c0df"
	tmcB7LessonOracle       = "cd61a01c-dfe1-4c1e-8063-7f90fcd0c452"
	tmcB7DonnieAprilOracle  = "7b33368f-0668-4233-8bbf-725c66c771cb"
	tmcB7UtromOracle        = "976e1fe6-ac32-4ea9-9828-acd3c4791e01"
)

func TestTmcB7EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tmcB7SplinterLeoOracle:  "Splinter & Leo, Father & Son",
		tmcB7SplinterAgedOracle: "Splinter, Aging Champion",
		tmcB7RegentaOracle:      "Triceraton Regenta",
		tmcB7CaneOracle:         "Wooden Cane",
		tmcB7AprilOracle:        "April O'Neil, Human Element",
		tmcB7ContainmentOracle:  "Containment Protocol",
		tmcB7DonatelloOracle:    "Donatello, Rad Scientist",
		tmcB7LessonOracle:       "Donatello's Science Lesson",
		tmcB7DonnieAprilOracle:  "Donnie & April, Adorkable Duo",
		tmcB7UtromOracle:        "Utrom Monitor",
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

// tmcB7PlayerRef is a player target ref.
func tmcB7PlayerRef(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetPlayer, ID: id}
}

// tmcB7TryPickPlayer answers the open pick_target prompt for `chooser`
// with a player and returns the engine's answer.
func tmcB7TryPickPlayer(t *testing.T, g *game.Game, chooser, target uuid.UUID) error {
	t.Helper()
	b04WaitForPick(t, g, chooser)
	p := latestPickTarget(g, chooser)
	return g.ResolvePickTarget(p.ID, chooser, tmcB7PlayerRef(target))
}

// tmcB7Tap taps a battlefield permanent directly.
func tmcB7Tap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
}

// tmcB7Hand is the size of p's hand.
func tmcB7Hand(g *game.Game, p *game.Player) int {
	n := 0
	g.ReadSnapshot(func() { n = p.Hand.Size() })
	return n
}

// --- Splinter & Leo, Father & Son ---------------------------------------------

func TestTmcB7SplinterAndLeoBothModesTargetDifferentPlayers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	duo := castHoldCreature(t, g, "Splinter & Leo, Father & Son", "Legendary Creature — Mutant Ninja Rat Turtle", tmcB7SplinterLeoOracle, 2, 2)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enter trigger did not ask for modes")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0, 1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	if err := tmcB7TryPickPlayer(t, g, me.ID, opp.ID); err != nil {
		t.Fatalf("the first bullet refused the opponent: %v", err)
	}
	// Each mode must target a different player: the opponent is taken.
	if err := tmcB7TryPickPlayer(t, g, me.ID, opp.ID); err == nil {
		t.Fatal("the second bullet accepted the player the first bullet targets")
	}
	if err := tmcB7TryPickPlayer(t, g, me.ID, me.ID); err != nil {
		t.Fatalf("the second bullet refused me: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := tmtB1Count(g, opp.ID, "Mutant"); got != 1 {
		t.Errorf("the opponent has %d Mutant tokens, want 1", got)
	}
	if got := tmtB1Count(g, me.ID, "Mutant"); got != 0 {
		t.Errorf("I have %d Mutant tokens, want 0", got)
	}
	if got := plusOnes(g, mine); got != 1 {
		t.Errorf("my other creature has %d +1/+1 counters, want 1", got)
	}
	if got := plusOnes(g, duo); got != 0 {
		t.Errorf("Splinter & Leo has %d +1/+1 counters; it is not an \"other\" creature", got)
	}
	if got := plusOnes(g, theirs); got != 0 {
		t.Errorf("the opponent's creature has %d counters, want 0", got)
	}
}

func TestTmcB7SplinterAndLeoOneModeMayTargetAnyone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Splinter & Leo, Father & Son", "Legendary Creature — Mutant Ninja Rat Turtle", tmcB7SplinterLeoOracle, 2, 2)
	answerMode(t, g, modePickChoiceFor(g, me.ID), me.ID, 1)
	if err := tmcB7TryPickPlayer(t, g, me.ID, opp.ID); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, theirs); got != 1 {
		t.Errorf("the opponent's creature has %d counters, want 1", got)
	}
	if got := tmtB1Count(g, opp.ID, "Mutant") + tmtB1Count(g, me.ID, "Mutant"); got != 0 {
		t.Errorf("%d Mutants were made with only the counter bullet chosen", got)
	}
}

// --- Splinter, Aging Champion -------------------------------------------------

func TestTmcB7SplinterAgingChampionDestroysOnlyATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	tapped := b12Creature(g, opp.ID, "Tapped Bear", "Creature — Bear", 2, 2)
	untapped := b12Creature(g, opp.ID, "Untapped Bear", "Creature — Bear", 2, 2)
	tmcB7Tap(g, tapped)
	castHoldCreature(t, g, "Splinter, Aging Champion", "Legendary Creature — Mutant Ninja Rat", tmcB7SplinterAgedOracle, 2, 2)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: untapped}); err == nil {
		t.Fatal("an untapped creature was accepted as a target")
	}
	pickTargetRefs(t, g, me.ID, nil, cardRefs(tapped)...)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("the tapped creature survived")
	}
	if !g.Battlefield.Contains(untapped) {
		t.Error("the untapped creature was destroyed")
	}
}

func TestTmcB7SplinterAgingChampionLeavingDrawsForYouAndAnotherPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	advanceToMain(t, g)
	splinter := b27Push(g, me.ID, "Splinter, Aging Champion", "Legendary Creature — Mutant Ninja Rat", tmcB7SplinterAgedOracle, "{2}{W}", 2, 2, "W")
	mine, theirs, bystander := tmcB7Hand(g, me), tmcB7Hand(g, opp), tmcB7Hand(g, other)
	b27Kill(g, splinter)
	g.RunStateChecksForTest()
	if err := tmcB7TryPickPlayer(t, g, me.ID, me.ID); err == nil {
		t.Fatal("\"another target player\" accepted me")
	}
	if err := tmcB7TryPickPlayer(t, g, me.ID, opp.ID); err != nil {
		t.Fatalf("pick the opponent: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := tmcB7Hand(g, me); got != mine+1 {
		t.Errorf("I have %d cards, want %d", got, mine+1)
	}
	if got := tmcB7Hand(g, opp); got != theirs+1 {
		t.Errorf("the target has %d cards, want %d", got, theirs+1)
	}
	if got := tmcB7Hand(g, other); got != bystander {
		t.Errorf("an untargeted player drew: %d cards, want %d", got, bystander)
	}
}

// --- Triceraton Regenta -------------------------------------------------------

func TestTmcB7TriceratonRegentaUntapsAnotherCreatureAndHasVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	regenta := b27Push(g, me.ID, "Triceraton Regenta", "Creature — Dinosaur Noble", tmcB7RegentaOracle, "{2}{W}", 2, 3, "W")
	friend := b12Creature(g, me.ID, "Tapped Friend", "Creature — Bear", 2, 2)
	tmcB7Tap(g, friend)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(regenta, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: regenta}); err == nil {
		t.Fatal("the Regenta accepted itself as the \"other\" target")
	}
	pickTargetRefs(t, g, me.ID, nil, cardRefs(friend)...)
	passPriorityAroundTable(t, g)
	if tappedForTest(t, g, friend) {
		t.Error("the other creature is still tapped")
	}
	if tappedForTest(t, g, regenta) {
		t.Error("the Regenta tapped to attack; it has vigilance")
	}
}

// --- Wooden Cane ----------------------------------------------------------------

func TestTmcB7WoodenCaneMakesAMutantAndEquipsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cane := castCatalogSpell(t, g, "Wooden Cane", "Artifact — Equipment", tmcB7CaneOracle, nil)
	passPriorityAroundTable(t, g)
	mutant := findBattlefieldByName(g, "Mutant")
	if mutant == uuid.Nil {
		t.Fatal("no Mutant token")
	}
	c := s58p6View(t, g, mutant)
	if c.Controller != me.ID || !c.IsToken() || !c.HasSubtype("Mutant") || !c.HasColorWord("R") {
		t.Errorf("the token is %+v, want a red Mutant token under my control", c)
	}
	if c.CurrentPower() != 4 || c.CurrentToughness() != 3 {
		t.Errorf("the Mutant is %d/%d, want 4/3 (2/2 plus +2/+1)", c.CurrentPower(), c.CurrentToughness())
	}
	if s := findBattlefieldCardByID(g, cane); s == nil || !s.IsAttachedTo(mutant) {
		t.Error("the Cane should be attached to the Mutant")
	}
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	equipTo(t, g, me.ID, cane, bear)
	if c := s58p6View(t, g, bear); c.CurrentPower() != 4 || c.CurrentToughness() != 3 {
		t.Errorf("the equipped Bear is %d/%d, want 4/3", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- April O'Neil, Human Element ------------------------------------------------

func TestTmcB7AprilMakesAMutagenForAnyPlayersArtifactInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	b27Push(g, me.ID, "April O'Neil, Human Element", "Legendary Creature — Human Detective", tmcB7AprilOracle, "{3}{U}", 2, 5, "U")

	castFromHandForTest(t, g, me, "Plain Bear", "Creature — Bear", "", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := tmtMuMutagens(g, me.ID); got != 0 {
		t.Fatalf("a creature spell made %d Mutagens", got)
	}

	castFromHandForTest(t, g, me, "Plain Sorcery", "Sorcery", "", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := tmtMuMutagens(g, me.ID); got != 1 {
		t.Fatalf("my sorcery made %d Mutagens, want 1", got)
	}

	castFromHandForTest(t, g, me, "Plain Rock", "Artifact", "", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := tmtMuMutagens(g, me.ID); got != 2 {
		t.Fatalf("my artifact spell: %d Mutagens, want 2", got)
	}

	// An opponent's instant, cast while my spell is on the stack: the
	// Mutagen is mine.
	castFromHandForTest(t, g, me, "Plain Bear", "Creature — Bear", "", "", game.CastSpellParams{})
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	castFromHandForTest(t, g, opp, "Plain Instant", "Instant", "", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := tmtMuMutagens(g, me.ID); got != 3 {
		t.Errorf("an opponent's instant: %d Mutagens, want 3", got)
	}
	if got := tmtMuMutagens(g, opp.ID); got != 0 {
		t.Errorf("the opponent got %d Mutagens; \"you create\" means April's controller", got)
	}
}

// --- Containment Protocol -------------------------------------------------------

func TestTmcB7ContainmentProtocolTapsAndHoldsItsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := pushPermanentForTest(g, me.ID, "Host", "", "Creature — Bear")
	aura := castCatalogSpell(t, g, "Containment Protocol", auraTypeLine, tmcB7ContainmentOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	passPriorityAroundTable(t, g)
	if got := attachmentHostOf(t, g, aura); got.Kind != game.TargetCard || got.ID != host {
		t.Fatalf("Containment Protocol attached to %+v, want %s", got, host)
	}
	if c, _ := battlefieldCard(g, host); !c.Tapped {
		t.Fatal("the enter trigger did not tap the enchanted creature")
	}
	advanceToUpkeepOf(t, g, 0)
	if c, _ := battlefieldCard(g, host); !c.Tapped {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
}

// --- Donatello, Rad Scientist ---------------------------------------------------

func TestTmcB7DonatelloTapsAndStunsUpToThreeOpposingCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	c := b12Creature(g, opp.ID, "Bear C", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	don := castHoldCreature(t, g, "Donatello, Rad Scientist", "Legendary Creature — Mutant Ninja Turtle", tmcB7DonatelloOracle, 5, 6)
	assertKeywords(t, g, don, "vigilance")
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(a, mine)); err == nil {
		t.Fatal("a creature I control was accepted as a target")
	}
	pickTargetRefs(t, g, me.ID, nil, cardRefs(a, b, c)...)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if !tappedForTest(t, g, id) || countersOn(g, id, game.CounterStun) != 1 {
			t.Errorf("target %s: tapped %v, %d stun counters; want tapped with 1", id, tappedForTest(t, g, id), countersOn(g, id, game.CounterStun))
		}
	}
	if tappedForTest(t, g, mine) || countersOn(g, mine, game.CounterStun) != 0 {
		t.Error("my creature was touched")
	}
}

// --- Donatello's Science Lesson -------------------------------------------------

func TestTmcB7ScienceLessonTapsTwoAndTwoPlayersDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	mine, theirs, bystander := tmcB7Hand(g, me), tmcB7Hand(g, opp), tmcB7Hand(g, other)
	castCatalogSpell(t, g, "Donatello's Science Lesson", "Instant", tmcB7LessonOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: a, Slot: 0},
		{Kind: game.TargetCard, ID: b, Slot: 0},
		{Kind: game.TargetPlayer, ID: me.ID, Slot: 1},
		{Kind: game.TargetPlayer, ID: opp.ID, Slot: 1},
	})
	passPriorityAroundTable(t, g)
	if !tappedForTest(t, g, a) || !tappedForTest(t, g, b) {
		t.Error("the two creatures were not tapped")
	}
	// The Lesson was put into my hand and cast after the count; I drew one.
	if got := tmcB7Hand(g, me); got != mine+1 {
		t.Errorf("I have %d cards, want %d", got, mine+1)
	}
	if got := tmcB7Hand(g, opp); got != theirs+1 {
		t.Errorf("the opponent has %d cards, want %d", got, theirs+1)
	}
	if got := tmcB7Hand(g, other); got != bystander {
		t.Errorf("an untargeted player has %d cards, want %d", got, bystander)
	}
}

func TestTmcB7ScienceLessonRefusesThreeCreatures(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	c := b12Creature(g, opp.ID, "Bear C", "Creature — Bear", 2, 2)
	err := castCatalogSpellErr(t, g, "Donatello's Science Lesson", "Instant", tmcB7LessonOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: a, Slot: 0},
		{Kind: game.TargetCard, ID: b, Slot: 0},
		{Kind: game.TargetCard, ID: c, Slot: 0},
	})
	if err == nil {
		t.Error("three creature targets were accepted")
	}
}

// --- Donnie & April, Adorkable Duo ----------------------------------------------

func TestTmcB7DonnieAndAprilDrawAndRegrowForDifferentPlayers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	var bolt, bear, rock uuid.UUID
	g.WithWriteLock(func() {
		bolt = pushGraveyardCardTyped(me, "Old Bolt", "Instant")
		bear = pushGraveyardCardTyped(me, "Old Bear", "Creature — Bear")
		rock = pushGraveyardCardTyped(me, "Old Rock", "Artifact")
	})
	theirs := tmcB7Hand(g, opp)
	castHoldCreature(t, g, "Donnie & April, Adorkable Duo", "Legendary Creature — Mutant Ninja Human Turtle", tmcB7DonnieAprilOracle, 3, 3)
	mine := tmcB7Hand(g, me)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enter trigger did not ask for modes")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0, 1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	if err := tmcB7TryPickPlayer(t, g, me.ID, opp.ID); err != nil {
		t.Fatalf("draw bullet: %v", err)
	}
	if err := tmcB7TryPickPlayer(t, g, me.ID, opp.ID); err == nil {
		t.Fatal("the return bullet accepted the player the draw bullet targets")
	}
	if err := tmcB7TryPickPlayer(t, g, me.ID, me.ID); err != nil {
		t.Fatalf("return bullet: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := tmcB7Hand(g, opp); got != theirs+2 {
		t.Errorf("the opponent has %d cards, want %d", got, theirs+2)
	}
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to return a card")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if !offered[bolt] || !offered[rock] || offered[bear] {
		t.Errorf("offered %v; want the instant and the artifact, not the creature", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if !handHasID(me, rock) {
		t.Error("the artifact did not come back to my hand")
	}
	if got := tmcB7Hand(g, me); got != mine+1 {
		t.Errorf("I have %d cards, want %d", got, mine+1)
	}
}

// --- Utrom Monitor --------------------------------------------------------------

func TestTmcB7UtromMonitorCostsLessPerArtifactAndFlies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	for i := 0; i < 2; i++ {
		b12Permanent(g, me.ID, "Rock", "Artifact")
	}
	// {4}{U} less {2}: {2}{U}, paid exactly.
	g.WithWriteLock(func() {
		me.ManaPool.AddMana(game.ManaToken{Color: "U"})
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	})
	id := castFromHandForTest(t, g, me, "Utrom Monitor", "Artifact Creature — Utrom Scientist", "{4}{U}", tmcB7UtromOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Utrom Monitor did not resolve")
	}
	assertKeywords(t, g, id, "flying")
}

func TestTmcB7UtromMonitorWithoutArtifactsCantBeCastForThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	g.WithWriteLock(func() {
		me.ManaPool.AddMana(game.ManaToken{Color: "U"})
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	})
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Utrom Monitor", TypeLine: "Artifact Creature — Utrom Scientist",
			ManaCost: "{4}{U}", OracleID: tmcB7UtromOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("Utrom Monitor was cast for three with no artifacts")
	}
}
