package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #4266 (CR 101.4, CR 616.1): "each player mills", "exile the top card of each
// opponent's library" and the rest do one library at a time, in APNAP order,
// each from the previous one's continuation. Two libraries that each ask for a
// CR 616.1 ordering open one prompt at a time, the active player's first, and
// what follows the instruction waits for the last of them.

// orderMillOf registers two mandatory replacements over `card`'s move to a
// graveyard, so CR 616.1 asks the affected player which goes first and the
// move waits for the answer.
func orderMillOf(g *game.Game, card uuid.UUID) {
	g.WithWriteLock(func() {
		for _, label := range []string{"first order", "second order"} {
			g.RegisterReplacementForTest(game.ReplacementEffect{
				Watches: []game.EventKind{game.EventZoneMove},
				AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
					return ev.Kind == game.RepEventMove && ev.CardID == card && ev.NewZone == game.ZoneGraveyard
				},
				Replace: func(*game.ReplacementEvent, *game.Game, *game.Card) error { return nil },
				Label:   label,
			})
		}
	})
}

// promptingTop puts a card on top of p's library whose move to a graveyard
// (mill) or to exile (exile) asks for an ordering.
func promptingTop(g *game.Game, p *game.Player, mill bool) uuid.UUID {
	var id uuid.UUID
	g.WithWriteLock(func() { id = pushLibraryTopForTest(p, "Prompting Card") })
	if mill {
		orderMillOf(g, id)
	} else {
		orderExileOf(g, id)
	}
	return id
}

func milledInto(p *game.Player, id uuid.UUID) bool { return p.Graveyard.Contains(id) }

func stillInLibraryOf(p *game.Player, id uuid.UUID) bool { return p.Library.Contains(id) }

// expectOnePrompt fails unless exactly one ordering prompt is open.
func expectOnePrompt(t *testing.T, g *game.Game, when string) {
	t.Helper()
	if len(g.PendingChoices) != 1 {
		t.Fatalf("%s: %d prompts open, want exactly one", when, len(g.PendingChoices))
	}
}

func TestMillEachPlayerInTurnOpensOnePromptAtATimeInApnapOrderAndFinishesWithAllResults(t *testing.T) {
	for _, active := range []int{0, 2} {
		g := newCatalogGame(t)
		g.WithWriteLock(func() { g.Turn.ActiveSeat = active })
		first, second := g.Seats[active], g.Seats[(active+2)%4]
		firstCard := promptingTop(g, first, true)
		secondCard := promptingTop(g, second, true)

		// The caller lists the players the other way round: the helper, not the
		// caller, puts them in APNAP order.
		var results []PlayerCards
		var finished int
		g.WithWriteLock(func() {
			ctx := NewContext(g, &game.StackItem{Controller: g.Seats[1].ID})
			err := millEachPlayerInTurn(ctx, []uuid.UUID{second.ID, first.ID}, constantMill(1),
				func(_ *Context, r []PlayerCards) error {
					finished++
					results = r
					return nil
				})
			if err != nil {
				t.Fatalf("millEachPlayerInTurn: %v", err)
			}
		})

		expectOnePrompt(t, g, "after starting")
		if g.PendingChoices[0].Chooser != first.ID {
			t.Fatalf("active seat %d: the first prompt is not the active player's", active)
		}
		if !stillInLibraryOf(first, firstCard) || !stillInLibraryOf(second, secondCard) {
			t.Fatal("a card left its library while the first prompt was open")
		}
		if finished != 0 {
			t.Fatal("the final continuation ran before either player milled")
		}
		answerExileOrder(t, g)

		if !milledInto(first, firstCard) {
			t.Fatal("the first player's card is not milled after their answer")
		}
		expectOnePrompt(t, g, "after the first answer")
		if g.PendingChoices[0].Chooser != second.ID {
			t.Fatal("the second prompt is not the next player's")
		}
		if !stillInLibraryOf(second, secondCard) {
			t.Fatal("the second library was milled before its prompt was answered")
		}
		if finished != 0 {
			t.Fatal("the final continuation ran with the second player still to mill")
		}
		answerExileOrder(t, g)

		if finished != 1 {
			t.Fatalf("the final continuation ran %d times, want once", finished)
		}
		if !milledInto(second, secondCard) {
			t.Fatal("the second player's card is not milled")
		}
		want := []PlayerCards{
			{Player: first.ID, Cards: []uuid.UUID{firstCard}},
			{Player: second.ID, Cards: []uuid.UUID{secondCard}},
		}
		if len(results) != 2 {
			t.Fatalf("results %+v, want %+v", results, want)
		}
		for i := range want {
			if results[i].Player != want[i].Player || len(results[i].Cards) != 1 || results[i].Cards[0] != want[i].Cards[0] {
				t.Errorf("result %d is %+v, want %+v", i, results[i], want[i])
			}
		}
	}
}

// A clean run (nobody asks anything) finishes on the same call, and a player
// whose count is zero is a result with no cards.
func TestMillEachPlayerInTurnFinishesAtOnceWhenNothingAsks(t *testing.T) {
	g := newCatalogGame(t)
	var results []PlayerCards
	ran := false
	before := g.Seats[1].Library.Size()
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: g.Seats[0].ID})
		err := millEachPlayerInTurn(ctx, []uuid.UUID{g.Seats[0].ID, g.Seats[1].ID},
			func(p uuid.UUID) int {
				if p == g.Seats[1].ID {
					return 0
				}
				return 2
			},
			func(_ *Context, r []PlayerCards) error {
				ran = true
				results = r
				return nil
			})
		if err != nil {
			t.Fatalf("millEachPlayerInTurn: %v", err)
		}
	})
	if !ran || len(results) != 2 {
		t.Fatalf("ran=%v results=%+v, want both players", ran, results)
	}
	if len(results[0].Cards) != 2 || len(results[1].Cards) != 0 {
		t.Errorf("milled %d and %d cards, want 2 and 0", len(results[0].Cards), len(results[1].Cards))
	}
	if g.Seats[1].Library.Size() != before {
		t.Error("a player whose count is zero milled")
	}
}

func TestExileTopEachPlayerInTurnOpensOnePromptAtATime(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	other := g.Seats[2]
	myCard := promptingTop(g, me, false)
	otherCard := promptingTop(g, other, false)

	var results []PlayerCards
	finished := 0
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID})
		err := exileTopEachPlayerInTurn(ctx, []uuid.UUID{other.ID, me.ID},
			ExileTopWithPermission{GrantTo: me.ID, N: 1, AnyColor: true},
			func(_ *Context, r []PlayerCards) error {
				finished++
				results = r
				return nil
			})
		if err != nil {
			t.Fatalf("exileTopEachPlayerInTurn: %v", err)
		}
	})
	expectOnePrompt(t, g, "after starting")
	if !stillInLibraryOf(other, otherCard) || !stillInLibraryOf(me, myCard) {
		t.Fatal("a card left its library while the first prompt was open")
	}
	answerExileOrder(t, g)
	expectOnePrompt(t, g, "after the first answer")
	if !stillInLibraryOf(other, otherCard) || finished != 0 {
		t.Fatal("the second library was exiled from, or the continuation ran, before the second answer")
	}
	answerExileOrder(t, g)
	if finished != 1 || len(results) != 2 {
		t.Fatalf("finished %d times with results %+v", finished, results)
	}
	if results[0].Player != me.ID || results[0].Cards[0] != myCard || results[1].Player != other.ID || results[1].Cards[0] != otherCard {
		t.Errorf("results %+v are not the active player's card, then the other's", results)
	}
	for _, id := range []uuid.UUID{myCard, otherCard} {
		if _, ok := findInZone(g.Exile, id); !ok {
			t.Errorf("card %s is not in exile", id)
		}
	}
}

func TestExileTopWithPermissionApplyThenRefusesTheFormsWithoutAContinuation(t *testing.T) {
	g := newCatalogGame(t)
	budget := 3
	for _, e := range []ExileTopWithPermission{
		{GrantTo: g.Seats[0].ID, FaceDown: true},
		{GrantTo: g.Seats[0].ID, TotalManaValueAtMost: &budget},
		{GrantTo: g.Seats[0].ID, EachManaValueAtMost: &budget},
	} {
		g.WithWriteLock(func() {
			if err := e.ApplyThen(NewContext(g, &game.StackItem{Controller: g.Seats[0].ID}), nil); err == nil {
				t.Errorf("%+v: want an error, got none", e)
			}
		})
	}
}

// Mezzio Mugger exiles from every library, the active player's first, one
// prompt at a time; each exiled card is playable by the Mugger's controller.
func TestMezzioMuggerExilesEachLibraryInTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	other := g.Seats[3]
	myCard := promptingTop(g, me, false)
	otherCard := promptingTop(g, other, false)
	libs := map[int]int{}
	for i, p := range g.Seats {
		libs[i] = p.Library.Size()
	}

	g.WithWriteLock(func() {
		if err := mezzioMuggerExile(g, &game.StackItem{Controller: me.ID}); err != nil {
			t.Fatalf("mezzioMuggerExile: %v", err)
		}
	})
	expectOnePrompt(t, g, "after the trigger")
	if !stillInLibraryOf(other, otherCard) {
		t.Fatal("the second library was exiled from while the first prompt was open")
	}
	answerExileOrder(t, g)
	expectOnePrompt(t, g, "after the first answer")
	answerExileOrder(t, g)
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts still open", len(g.PendingChoices))
	}
	for _, id := range []uuid.UUID{myCard, otherCard} {
		if _, ok := findInZone(g.Exile, id); !ok {
			t.Errorf("card %s is not exiled", id)
		}
	}
	// Every library lost exactly its top card: Mezzio Mugger exiles from the
	// plain ones as well as the two that asked.
	for i, p := range g.Seats {
		if got, want := p.Library.Size(), libs[i]-1; got != want {
			t.Errorf("seat %d library %d, want %d", i, got, want)
		}
	}
}

// Brainstealer Dragon exiles from each OPPONENT's library, never the
// controller's, one prompt at a time.
func TestBrainstealerDragonExilesEachOpponentInTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	one, three := g.Seats[1], g.Seats[3]
	mine := pushLibraryTopForTest(me, "Mine")
	oneCard := promptingTop(g, one, false)
	threeCard := promptingTop(g, three, false)

	g.WithWriteLock(func() {
		if err := brainstealerDragonExile(g, &game.StackItem{Controller: me.ID}); err != nil {
			t.Fatalf("brainstealerDragonExile: %v", err)
		}
	})
	expectOnePrompt(t, g, "after the trigger")
	if g.PendingChoices[0].Chooser != one.ID {
		t.Fatal("the first prompt is not the first opponent in turn order")
	}
	if !stillInLibraryOf(three, threeCard) {
		t.Fatal("the third opponent was exiled from while the first prompt was open")
	}
	answerExileOrder(t, g)
	expectOnePrompt(t, g, "after the first answer")
	answerExileOrder(t, g)
	for _, id := range []uuid.UUID{oneCard, threeCard} {
		if _, ok := findInZone(g.Exile, id); !ok {
			t.Errorf("card %s is not exiled", id)
		}
	}
	if !stillInLibraryOf(me, mine) {
		t.Error("the controller's own library was exiled from")
	}
}

// Folio of Fancies counts each opponent's hand as the ability resolves and
// mills them one at a time.
func TestFolioOfFanciesMillsEachOpponentByHandSizeInTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spec, _ := Lookup("9030d508-5cfe-46ed-954b-21d970bc3b5a")
	card := promptingTop(g, g.Seats[2], true)
	handSizes := map[int]int{}
	libs := map[int]int{}
	for i, p := range g.Seats {
		handSizes[i] = p.Hand.Size()
		libs[i] = p.Library.Size()
	}
	g.WithWriteLock(func() {
		if err := spec.Activated[1].Effect(g, &game.StackItem{Controller: me.ID}); err != nil {
			t.Fatalf("Folio: %v", err)
		}
	})
	// Seat 1 has no prompt and has milled; seat 2 is waiting on its prompt.
	expectOnePrompt(t, g, "after the ability")
	if g.Seats[1].Library.Size() != libs[1]-handSizes[1] {
		t.Errorf("seat 1 library %d, want %d", g.Seats[1].Library.Size(), libs[1]-handSizes[1])
	}
	if g.Seats[3].Library.Size() != libs[3] {
		t.Error("seat 3 milled while seat 2's prompt was open")
	}
	answerExileOrder(t, g)
	if !milledInto(g.Seats[2], card) {
		t.Error("seat 2's prompting card is not milled")
	}
	for _, i := range []int{1, 2, 3} {
		if got := g.Seats[i].Library.Size(); got != libs[i]-handSizes[i] {
			t.Errorf("seat %d library %d, want %d", i, got, libs[i]-handSizes[i])
		}
	}
	if g.Seats[0].Library.Size() != libs[0] {
		t.Error("the controller milled")
	}
}

// Breach the Multiverse's picks wait for the LAST library's mill: while any
// mill is waiting on its prompt the caster has nothing to choose yet.
func TestBreachTheMultiverseWaitsForEveryMillBeforeThePicks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := pushGraveyardCreature(g, me.ID, "Grave Bear", "{1}{G}")
	promptingTop(g, g.Seats[0], true)
	promptingTop(g, g.Seats[2], true)
	libs := map[int]int{}
	for i, p := range g.Seats {
		libs[i] = p.Library.Size()
	}

	castCatalogSpell(t, g, "Breach the Multiverse", "Sorcery", mcBreachOracle, nil)
	passPriorityAroundTable(t, g)

	expectOnePrompt(t, g, "after the spell resolved")
	if g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("the open prompt is %v, want the first library's ordering", g.PendingChoices[0].Kind)
	}
	answerExileOrder(t, g)
	// The second library's ordering is open; the caster has no pick yet, and
	// no library past the first has been touched.
	expectOnePrompt(t, g, "after the first answer")
	if g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("the open prompt is %v, want the second library's ordering", g.PendingChoices[0].Kind)
	}
	// Seat 1 sits between the two prompting seats: milled after seat 0, before
	// seat 2. Seat 3 comes after the open prompt and is untouched.
	if got := g.Seats[1].Library.Size(); got != libs[1]-10 {
		t.Errorf("seat 1 library %d, want %d", got, libs[1]-10)
	}
	if got := g.Seats[3].Library.Size(); got != libs[3] {
		t.Errorf("seat 3 library %d, want %d: milled while seat 2's prompt was open", got, libs[3])
	}
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("a pick opened before the last library milled")
	}
	answerExileOrder(t, g)
	if got := g.Seats[3].Library.Size(); got != libs[3]-10 {
		t.Errorf("seat 3 library %d, want %d", got, libs[3]-10)
	}
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no pick for the caster after the last mill: %+v", g.PendingChoices)
	}
	for ; pick != nil; pick = chooseCardsChoiceFor(g, me.ID) {
		if len(pick.ChooseCards) == 0 {
			t.Fatal("an empty pick prompt")
		}
		answerChooseCards(t, g, me.ID, pick.ChooseCards[0])
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the caster's pick did not reach the battlefield")
	}
}

// Extract from Darkness's creature choice waits for the last mill (b02EachPlayerMills).
func TestExtractFromDarknessWaitsForEveryMillBeforeTheChoice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	promptingTop(g, g.Seats[0], true)
	promptingTop(g, g.Seats[2], true)

	castCatalogSpell(t, g, "Extract from Darkness", "Sorcery", b30ExtractFromDarknessOracle, nil)
	passPriorityAroundTable(t, g)

	expectOnePrompt(t, g, "after the spell resolved")
	answerExileOrder(t, g)
	expectOnePrompt(t, g, "after the first answer")
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("the creature choice opened before the second library milled")
	}
	answerExileOrder(t, g)
	if len(g.PendingChoices) != 0 && chooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatalf("unexpected prompts after the last mill: %+v", g.PendingChoices)
	}
}
