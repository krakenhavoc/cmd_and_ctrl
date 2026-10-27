package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_choice_test.go — #568. The option pick and the pile
// split, tested as shapes rather than only through the cards that use
// them: the chosen branch runs, the prompt is addressed to a seat that
// is not the controller, the split chains to a second prompt for a
// DIFFERENT seat, and both rewind.

// answerOptionPick answers the newest open option pick for `chooser`
// with the given index.
func answerOptionPick(t *testing.T, g *game.Game, chooser uuid.UUID, index int) {
	t.Helper()
	c := latestOptionPickFor(g, chooser)
	if c == nil {
		t.Fatalf("no option pick for %s: %+v", chooser, g.PendingChoices)
	}
	if err := g.ResolveOptionPick(c.ID, chooser, index); err != nil {
		t.Fatalf("ResolveOptionPick(%d): %v", index, err)
	}
}

// latestOptionPickFor returns the newest open option pick owed by a
// seat, or nil.
func latestOptionPickFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	var out *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionPick && c.Chooser == chooser {
			out = c
		}
	}
	return out
}

// latestChooseCardsFor returns the newest open choose-cards prompt
// owed by a seat, or nil.
func latestChooseCardsFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	return latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, chooser)
}

// latestRevealPickFor is the same read for #1214's reveal_pick — "an
// opponent picks from a set you revealed", which is what a pile
// split's first leg has been since that kind existed.
func latestRevealPickFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	return latestChoiceOfKindFor(g, game.PendingChoiceRevealPick, chooser)
}

// latestChoiceOfKindFor is the body both of those share.
func latestChoiceOfKindFor(g *game.Game, kind game.PendingChoiceKind, chooser uuid.UUID) *game.PendingChoice {
	var out *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == kind && c.Chooser == chooser {
			out = c
		}
	}
	return out
}

// testPickRan is what the keyed probe answer (testPickProbe) was handed,
// in order (#4253 retired PickOption's closure Then).
var (
	testPickRan   []int
	testPickProbe = OptionPickThen("test/4253-pick-option-probe", func(_ *Context, r game.OptionPicked) error {
		testPickRan = append(testPickRan, r.Index)
		return nil
	})
)

// TestPickOptionRunsTheChosenBranchForAnOpponent is the option pick's
// core contract: the prompt goes to the named seat, only that seat can
// answer it, and exactly the chosen branch runs.
func TestPickOptionRunsTheChosenBranchForAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	testPickRan = nil
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID})
		if err := (PickOption{
			Player:   opp.ID,
			Question: "test — choose one",
			Options: []game.ChoiceOption{
				{Label: "first", LifeCost: 3},
				{Label: "second"},
				{Label: "third"},
			},
			ThenKey: testPickProbe,
		}).Apply(ctx); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})
	c := latestOptionPickFor(g, opp.ID)
	if c == nil {
		t.Fatalf("the prompt is addressed to the opponent: %+v", g.PendingChoices)
	}
	if len(c.PickOptions) != 3 || c.PickOptions[0].LifeCost != 3 {
		t.Errorf("the options ride the prompt with their declared costs: %+v", c.PickOptions)
	}
	if err := g.ResolveOptionPick(c.ID, me.ID, 0); err == nil {
		t.Error("only the chooser may answer")
	}
	if err := g.ResolveOptionPick(c.ID, opp.ID, 7); err == nil {
		t.Error("an out-of-range index is refused")
	}
	if len(g.PendingChoices) != 1 {
		t.Fatal("a refused answer leaves the prompt open")
	}
	answerOptionPick(t, g, opp.ID, 1)
	if ran := testPickRan; len(ran) != 1 || ran[0] != 1 {
		t.Errorf("exactly the chosen branch runs: %v", ran)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("the prompt is dequeued: %+v", g.PendingChoices)
	}
}

// TestPickOptionUndoesAcrossThePrompt — rewinding past the answer puts
// the question back, and the restored prompt answers the same way.
func TestPickOptionUndoesAcrossThePrompt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: uuid.New()})
		if err := (PickOption{
			Player:   opp.ID,
			Question: "test — choose one",
			Options: []game.ChoiceOption{
				{Label: "lose 3 life", LifeCost: 3},
				{Label: "nothing"},
			},
			ThenKey: pickOptionTestBranch,
			Carry:   []uuid.UUID{opp.ID},
		}).Apply(ctx); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})
	snap := g.Clone()
	answerOptionPick(t, g, opp.ID, 0)
	if g.Seats[1].Life != life-3 {
		t.Fatalf("the chosen branch ran: life %d → %d", life, g.Seats[1].Life)
	}
	g.RestoreFrom(snap)
	if g.Seats[1].Life != life {
		t.Errorf("undo rewinds the branch: life %d, want %d", g.Seats[1].Life, life)
	}
	c := latestOptionPickFor(g, opp.ID)
	if c == nil {
		t.Fatalf("undo puts the question back: %+v", g.PendingChoices)
	}
	if len(c.PickOptions) != 2 || c.PickOptions[0].Label != "lose 3 life" {
		t.Errorf("the restored prompt still carries its options: %+v", c.PickOptions)
	}
	if err := g.ResolveOptionPick(c.ID, opp.ID, 0); err != nil {
		t.Fatalf("ResolveOptionPick after undo: %v", err)
	}
	if g.Seats[1].Life != life-3 {
		t.Errorf("answering the restored prompt: life %d, want %d", g.Seats[1].Life, life-3)
	}
}

// pickOptionTestBranch is a registered branch — no captured game, no
// captured player pointer: the victim rides the prompt as its Carry.
var pickOptionTestBranch = OptionPickThen("test/4253-pick-option-lose-three", func(ctx *Context, r game.OptionPicked) error {
	if r.Index != 0 || len(r.Carry) == 0 {
		return nil
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), r.Carry[0], -3)
})

// testPileTaken / testPileLeft are the piles the keyed probe tail
// (testPileProbe) was told (#4253 retired PileSplit's closure Then).
var (
	testPileTaken, testPileLeft []uuid.UUID
	testPileProbe               = PileTail("test/4253-pile-split-probe", func(_ *Context, taken, left []uuid.UUID, _ game.TailOutcome) error {
		testPileTaken, testPileLeft = taken, left
		return nil
	})
)

// TestPileSplitChainsToTheControllersPick is the two-seat chain: the
// splitter answers first, the controller's pile pick is queued by that
// answer, and the piles come back in the order the chooser took them.
func TestPileSplitChainsToTheControllersPick(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cards := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	testPileTaken, testPileLeft = nil, nil
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID})
		if err := (PileSplit{
			Splitter:      opp.ID,
			Chooser:       me.ID,
			Owner:         me.ID,
			SplitQuestion: "test — separate them",
			PickQuestion:  "test — take a pile",
			Cards:         cards,
			ThenTail:      ctx.Tail(testPileProbe),
		}).Apply(ctx); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})

	split := latestRevealPickFor(g, opp.ID)
	if split == nil {
		t.Fatalf("the split is asked of the opponent: %+v", g.PendingChoices)
	}
	if split.FromPlayer != me.ID {
		t.Errorf("the pool belongs to the controller: FromPlayer %s", split.FromPlayer)
	}
	if split.ChooseMin != 0 || split.ChooseMax != 3 {
		t.Errorf("an empty pile is a legal split: %d..%d", split.ChooseMin, split.ChooseMax)
	}
	if latestOptionPickFor(g, me.ID) != nil {
		t.Fatal("the controller is not asked until the split is answered")
	}
	if err := g.ResolveRevealPick(split.ID, opp.ID, cards[:1]); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}

	pick := latestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatalf("the split's answer queues the controller's pick: %+v", g.PendingChoices)
	}
	if len(pick.PickOptions) != 2 {
		t.Fatalf("two piles: %+v", pick.PickOptions)
	}
	if len(pick.PickOptions[0].Cards) != 1 || len(pick.PickOptions[1].Cards) != 2 {
		t.Errorf("the piles are the split: %+v", pick.PickOptions)
	}
	if testPileTaken != nil {
		t.Fatal("nothing settles until the pile is picked")
	}
	answerOptionPick(t, g, me.ID, 1)
	taken, left := testPileTaken, testPileLeft
	if len(taken) != 2 || len(left) != 1 {
		t.Errorf("taken %v, left %v — want the second pile taken", taken, left)
	}
	if left[0] != cards[0] {
		t.Errorf("the pile left is the one the splitter put first: %v", left)
	}
}

// TestPileSplitTakesEverythingWhenNobodyCanSplit — CR 800.4a. A
// splitter who has left the game cannot separate anything, and the
// rest of the card still has to resolve.
func TestPileSplitTakesEverythingWhenNobodyCanSplit(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cards := []uuid.UUID{uuid.New(), uuid.New()}
	testPileTaken, testPileLeft = nil, nil
	g.WithWriteLock(func() {
		opp.Eliminated = true
		ctx := NewContext(g, &game.StackItem{Controller: me.ID})
		if err := (PileSplit{
			Splitter: opp.ID,
			Chooser:  me.ID,
			Owner:    me.ID,
			Cards:    cards,
			ThenTail: ctx.Tail(testPileProbe),
		}).Apply(ctx); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})
	if taken, left := testPileTaken, testPileLeft; len(taken) != 2 || len(left) != 0 {
		t.Errorf("taken %v, left %v — want one pile of everything", taken, left)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("no prompt is left for a seat that has gone: %+v", g.PendingChoices)
	}
}

// TestOptionPickBlocksTheTable — the gate classification, asserted
// where a reader will look for it.
func TestOptionPickBlocksTheTable(t *testing.T) {
	if !game.ChoiceBlocksTable(game.PendingChoiceOptionPick) {
		t.Fatal("an option pick is a resolution-time decision and blocks the table")
	}
}
