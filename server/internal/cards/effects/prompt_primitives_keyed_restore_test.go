package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prompt_primitives_keyed_restore_test.go — #4253 slice 10b, continued: the
// card-set picks (a sacrifice choice, a populate, a put from a library, a
// Ring-bearer, a proliferate) and the pile split, each restored STRICTLY
// while its prompt is open and answered on the restored table.

// applyForTest runs an effect as `controller`'s resolving item, with g.mu
// held.
func applyForTest(t *testing.T, g *game.Game, controller, source uuid.UUID, e Applier) {
	t.Helper()
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{
			ID: uuid.New(), Kind: game.StackItemTriggered, Controller: controller, Owner: controller, SourceCardID: source,
		})
		if err := e.Apply(ctx); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})
}

// A SacrificeChoice with a keyed tail (Ob Nixilis, Unshackled's "sacrifices
// a creature and loses 10 life") asks even over one candidate, is a restore
// point, and the restored answer sacrifices and then runs the tail.
func TestASacrificeChoiceIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	life := opp.Life
	applyForTest(t, g, me.ID, uuid.Nil, SacrificeChoice{
		Player: opp.ID, Candidates: []uuid.UUID{bear}, Question: "sacrifice a creature",
		ThenTail: game.KeyedTail{Ref: thenChangeLife, Player: opp.ID, N: -10},
	})
	c := chooseCardsChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("one candidate asked nothing")
	}
	if c.PickThen != sacrificeChoicePicked.Key() {
		t.Fatalf("the pick's key is %q, want the keyed sacrifice choice", c.PickThen)
	}

	g = restoreRoundTrip(t, g, true)
	opp = g.Seats[1]
	answerChooseCards(t, g, opp.ID, bear)
	if !opp.Graveyard.Contains(bear) {
		t.Error("the restored answer did not sacrifice the creature")
	}
	if opp.Life != life-10 {
		t.Errorf("life %d → %d, want the tail's loss of 10", life, opp.Life)
	}
}

// bolsterProbeTail is a test tail: a shield counter on the creature it is
// told, so a test sees the tail ran and on what.
var bolsterProbeTail = game.RegisterTail("test/bolster-probe", func(g *game.Game, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if err := g.AddCounterForEffect(id, game.CounterShield, 1); err != nil {
			return err
		}
	}
	return nil
})

// A Bolster over two creatures tied for least toughness asks a keyed pick
// (choose-cards/bolster) carrying its tail, so the table is a restore point
// and the restored answer puts the counters on, then runs the tail on the
// creature chosen.
func TestABolsterChoiceIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "Elf A", "Creature — Elf", 1, 1)
	b := b12Creature(g, me.ID, "Elf B", "Creature — Elf", 1, 1)
	applyForTest(t, g, me.ID, uuid.Nil, Bolster{N: 2, ThenTail: game.KeyedTail{Ref: bolsterProbeTail}})
	if chooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("a tie asked nothing")
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	answerChooseCards(t, g, me.ID, b)
	if got := countersOn(g, b, game.CounterPlusOne); got != 2 {
		t.Errorf("the chosen creature has %d +1/+1 counters, want 2", got)
	}
	if got := countersOn(g, b, game.CounterShield); got != 1 {
		t.Errorf("the tail ran %d times on the chosen creature, want once", got)
	}
}

// Full Flowering: each populate's pick is keyed (choose-cards/populate), the
// populates still owed on its tail, so the table is a restore point at both.
func TestFullFloweringsPopulatesAreRestorePoints(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))
	centaur := pushToken(g, me.ID, TokenCard("3/3 green Centaur"))

	castXSpell(t, g, "Full Flowering", "Sorcery", fullFloweringOracle, "{X}{X}{G}", 2, nil)
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	answerChooseCards(t, g, me.ID, centaur)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("the second populate should ask too")
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	answerChooseCards(t, g, me.ID, centaur)
	passPriorityAroundTable(t, g)
	if got := len(populateTokens(g, me.ID, "Centaur")); got != 3 {
		t.Errorf("Centaurs = %d, want 3", got)
	}
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a third populate asked")
	}
}

// The Ring tempts you with two creatures out: the Ring-bearer choice is a
// keyed pick (choose-cards/ring-bearer), and Galadriel, Elven-Queen's
// counter is its tail.
func TestTheRingBearerChoiceIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	applyForTest(t, g, me.ID, uuid.Nil, TheRingTemptsYou{ThenTail: game.KeyedTail{Ref: galadrielElvenQueenBearerCounter, Player: me.ID}})
	if ringPrompt(g, me.ID) == nil {
		t.Fatal("two creatures asked nothing")
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	answerRing(t, g, me.ID, a)
	if ringBearerOf(g, me.ID) != a {
		t.Error("the restored answer did not make the creature the Ring-bearer")
	}
	if got := countersOn(g, a, game.CounterPlusOne); got != 1 {
		t.Errorf("the Ring-bearer has %d +1/+1 counters, want the tail's 1", got)
	}
}

// Steady Progress under Tekuthal (proliferate twice): both choices are keyed
// (choose-cards/proliferate) with the draw on their tail, so the table is a
// restore point at each, and the draw waits for the second.
func TestSteadyProgressProliferatesAreRestorePoints(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTekuthal(g, me.ID)
	a := pushCounterCreature(g, me.ID, "A", game.CounterPlusOne, 1)
	b := pushCounterCreature(g, me.ID, "B", game.CounterPlusOne, 1)
	handBefore := len(me.Hand.Cards)

	castSteadyProgressToPrompt(t, g)
	g = restoreRoundTrip(t, g, true)
	answerProliferate(t, g, a)
	if proliferatePrompt(g) == nil {
		t.Fatal("the second proliferate asked nothing")
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	if got := len(me.Hand.Cards); got != handBefore {
		t.Errorf("the card was drawn between the two proliferates: hand %d -> %d", handBefore, got)
	}
	answerProliferate(t, g, b)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, a, game.CounterPlusOne); got != 2 {
		t.Errorf("A = %d counters, want 2", got)
	}
	if got := counterCount(g, b, game.CounterPlusOne); got != 2 {
		t.Errorf("B = %d counters, want 2", got)
	}
	if got := len(me.Hand.Cards); got != handBefore+1 {
		t.Errorf("hand = %d, want %d (the draw after both)", got, handBefore+1)
	}
}

// A proliferated counter under Doubling Season and Hardened Scales asks for
// the order: the walk is a keyed tail on the counter (engine/proliferate-
// step), so the order prompt is a restore point, and the draw still follows.
func TestAProliferatedCounterIsARestorePointAtItsOrderPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	// The counter is seeded before the board that would order it.
	a := pushCounterCreature(g, me.ID, "A", game.CounterPlusOne, 1)
	hs, ds := counterOrderBoard(g, me.ID)
	handBefore := len(me.Hand.Cards)

	castSteadyProgressToPrompt(t, g)
	answerProliferate(t, g, a)
	if !orderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}
	g = strictRestoreWithFrame(t, g, game.ReplacementResumeCounterTail)
	me = g.Seats[0]
	// Doubling Season first, then Hardened Scales: 1 → 2 → 3 more.
	answerOrderBySource(t, g, ds, hs)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, a, game.CounterPlusOne); got != 4 {
		t.Errorf("A = %d counters, want 4", got)
	}
	if got := len(me.Hand.Cards); got != handBefore+1 {
		t.Errorf("hand = %d, want %d (the draw after the counters)", got, handBefore+1)
	}
}

// Fact or Fiction: once the opponent is named, the split and the pick are
// keyed (choose-cards/pile-split, option-pick/pile-split-take), the card's
// tail on both, so the table is a restore point at each.
func TestFactOrFictionsSplitAndPickAreRestorePoints(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	revealed := seedFactOrFictionLibrary(me, "FF One", "FF Two", "FF Three", "FF Four", "FF Five")
	hand := me.Hand.Size()

	castCatalogSpell(t, g, "Fact or Fiction", "Instant", factOrFictionOracle, nil)
	passPriorityAroundTable(t, g)
	answerChoosePlayer(t, g, me.ID, opp)
	if latestRevealPickFor(g, opp.ID) == nil {
		t.Fatalf("the opponent separates the piles: %+v", g.PendingChoices)
	}
	g = restoreRoundTrip(t, g, true)
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	split := latestRevealPickFor(g, opp.ID)
	if err := g.ResolveRevealPick(split.ID, opp.ID, revealed[:2]); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	if latestOptionPickFor(g, me.ID) == nil {
		t.Fatalf("the controller picks a pile: %+v", g.PendingChoices)
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	answerOptionPick(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("the taken pile goes to hand: %d → %d, want +3", hand, me.Hand.Size())
	}
	for _, id := range revealed[:2] {
		if !me.Graveyard.Contains(id) {
			t.Errorf("%s, in the pile left, should be in the graveyard", id)
		}
	}
}

// Gluntch, the Bestower: every question of the chain is keyed, and the
// players chosen ride the copy of the trigger each prompt carries, so a
// strict restore at each of the four questions still gives the second and
// third clauses a DIFFERENT player (ctx.ChosenPlayers() after a restore).
func TestGluntchIsARestorePointAtEveryQuestion(t *testing.T) {
	g := newCatalogGame(t)
	me, second := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Gluntch, the Bestower",
		"Legendary Creature — Jellyfish", gluntchOracle, false)
	bear := pushTribalCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	secondHand := len(second.Hand.Cards)

	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	answerChoosePlayer(t, g, me.ID, me)

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatalf("the CHOSEN player picks the creature: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if got := counterCount(g, bear, "+1/+1"); got != 2 {
		t.Errorf("+1/+1 counters %d, want 2", got)
	}

	g = restoreRoundTrip(t, g, true)
	me, second = g.Seats[0], g.Seats[1]
	sec := latestOptionPickFor(g, me.ID)
	if sec == nil {
		t.Fatalf("the second choice is queued: %+v", g.PendingChoices)
	}
	for _, opt := range sec.PickOptions {
		if opt.Player == me.ID {
			t.Errorf("after a restore, a SECOND player still excludes the first: %+v", sec.PickOptions)
		}
	}
	answerChoosePlayer(t, g, me.ID, second)
	if len(second.Hand.Cards) != secondHand+1 {
		t.Errorf("the second player drew: hand %d, want %d", len(second.Hand.Cards), secondHand+1)
	}

	g = restoreRoundTrip(t, g, true)
	me, third := g.Seats[0], g.Seats[2]
	thirdPrompt := latestOptionPickFor(g, me.ID)
	if thirdPrompt == nil {
		t.Fatalf("the third choice is queued: %+v", g.PendingChoices)
	}
	if len(thirdPrompt.PickOptions) != 2 {
		t.Errorf("after a restore, two seats left of four: %+v", thirdPrompt.PickOptions)
	}
	answerChoosePlayer(t, g, me.ID, third)
	if got := countTreasures(g, third.ID); got != 2 {
		t.Errorf("Treasures for the third player: %d, want 2", got)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("the chain is finished: %+v", g.PendingChoices)
	}
}
