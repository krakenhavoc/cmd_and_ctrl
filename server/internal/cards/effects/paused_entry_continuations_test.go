package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// paused_entry_continuations_test.go — #478. A battlefield ENTRY that
// queues a prompt used to lose the card.
//
// The repro is the issue's: an opponent controls Kismet and Thalia,
// Heretic Cathar, two "enters tapped" replacements that both apply to
// your Guildgate. One of them is a silent modification; TWO of them is
// a CR 616 ordering prompt, and the search bailed on that prompt with
// the card still in the library. The player answered a question that
// did nothing, the permanent never arrived, and the library was never
// shuffled either.
//
// The fix gives the entry the same posture the exit has had since #529:
// nothing moves while the question is open, and the SAME resume
// finishes the move and runs what the effect still owed — the shuffle,
// EventSearchLibrary and the caller's Then for a search, the CR 400.7
// new object for an exile return.

// pe2EntersTapped seats both enters-tapped hatebears under `opp`, so
// any nonbasic land or creature entering under anybody else gets two
// applicable replacements and therefore a CR 616 ordering prompt.
func pe2EntersTapped(g *game.Game, opp uuid.UUID) {
	b12Push(g, opp, "Kismet", "Enchantment", kismetOracle, 0, 0)
	b12Push(g, opp, "Thalia, Heretic Cathar", "Legendary Creature — Human Soldier",
		b13ThaliaHereticCatharOracle, 3, 2)
}

// peGuildgate is the fetch target: a nonbasic Gate land, which is what
// Circuitous Route looks for and what both hatebears tap.
func peGuildgate() game.Card {
	return game.Card{Name: "Azorius Guildgate", TypeLine: "Land — Gate"}
}

// peAnswerOrder answers an open CR 616 ordering prompt for `chooser` in
// the order it was offered.
func peAnswerOrder(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	c := b06ReplacementOrderFor(g, chooser)
	if c == nil {
		t.Fatalf("no replacement-order prompt for %s", chooser)
	}
	if err := g.ResolveReplacementOrder(c.ID, chooser, c.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
}

// peLibraryOrder is the library's instance IDs, bottom-first.
func peLibraryOrder(p *game.Player) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(p.Library.Cards))
	for _, c := range p.Library.Cards {
		out = append(out, c.InstanceID)
	}
	return out
}

func peSameOrder(a, b []uuid.UUID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// peSearchEvents counts EventSearchLibrary since `from`.
func peSearchEvents(g *game.Game, from int) int {
	n := 0
	for _, ev := range g.Events[from:] {
		if ev.Kind == game.EventSearchLibrary {
			n++
		}
	}
	return n
}

// --- the issue's repro ---------------------------------------------

// TestFetchedGuildgateArrivesAfterTheOrderingPromptIsAnswered is #478
// exactly as reported: Kismet + Thalia, Circuitous Route, a Guildgate.
func TestFetchedGuildgateArrivesAfterTheOrderingPromptIsAnswered(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pe2EntersTapped(g, opp.ID)
	// A library with more matches than the "up to two" may take, so the
	// searcher gets a real prompt, and some chaff so the shuffle has
	// something to do.
	gate := pushLibraryCardForTest(me, peGuildgate())
	for i := 0; i < 3; i++ {
		plTop(me, "Forest", "Basic Land — Forest", "")
	}
	for i := 0; i < 4; i++ {
		plTop(me, "Bear", "Creature — Bear", "{1}{G}")
	}
	orderBefore := peLibraryOrder(me)
	before := len(g.Events)

	castCatalogSpell(t, g, "Circuitous Route", "Sorcery", b16CircuitousRouteOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, gate)

	// The prompt is open and NOTHING has happened: this is the half
	// that used to eat the card.
	if b06ReplacementOrderFor(g, me.ID) == nil {
		t.Fatal("two applicable enters-tapped replacements queue the CR 616 ordering prompt")
	}
	if !me.Library.Contains(gate) {
		t.Fatal("nothing moves while the question is open")
	}
	if peSearchEvents(g, before) != 0 {
		t.Error("the search is not finished until the entry is")
	}
	if !peSameOrder(peLibraryOrder(me), orderBefore) {
		t.Error("the library is not shuffled until the search finishes")
	}

	peAnswerOrder(t, g, me.ID)

	if me.Library.Contains(gate) {
		t.Fatal("the Guildgate is still in the library after the prompt was answered — this is the bug")
	}
	if !g.Battlefield.Contains(gate) {
		t.Fatal("the Guildgate is on the battlefield")
	}
	card, ok := searchFetchedCard(g, "Azorius Guildgate")
	if !ok || !card.Tapped {
		t.Error("both replacements applied, so it enters tapped")
	}
	if n := peSearchEvents(g, before); n != 1 {
		t.Errorf("EventSearchLibrary fired %d times, want exactly 1", n)
	}
	if peSameOrder(peLibraryOrder(me), orderBefore) {
		t.Error(`"then shuffle" ran once the entry settled`)
	}
}

// TestFetchedPermanentUndoAcrossTheEntryPromptReplays — the undo
// contract the entry tail signs, the same one the exit route signs:
// rewind into the open prompt, answer it again, and the board follows.
func TestFetchedPermanentUndoAcrossTheEntryPromptReplays(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pe2EntersTapped(g, opp.ID)
	gate := pushLibraryCardForTest(me, peGuildgate())
	for i := 0; i < 3; i++ {
		plTop(me, "Forest", "Basic Land — Forest", "")
	}
	for i := 0; i < 3; i++ {
		plTop(me, "Bear", "Creature — Bear", "{1}{G}")
	}
	meID := me.ID

	castCatalogSpell(t, g, "Circuitous Route", "Sorcery", b16CircuitousRouteOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, meID, gate)
	promptOpen := g.Clone()

	peAnswerOrder(t, g, meID)
	if !g.Battlefield.Contains(gate) {
		t.Fatal("answering fetches the Guildgate")
	}

	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	if !g.Seats[0].Library.Contains(gate) {
		t.Fatal("the rewind puts the Guildgate back in the library, with the prompt open")
	}
	if b06ReplacementOrderFor(g, meID) == nil {
		t.Fatal("the rewound prompt is still there to answer")
	}

	peAnswerOrder(t, g, meID)
	if !g.Battlefield.Contains(gate) {
		t.Error("the replayed answer fetches the Guildgate again")
	}
	if g.Seats[0].Library.Contains(gate) {
		t.Error("and it is in exactly one place")
	}
}

// --- the other two entry sites --------------------------------------

// TestExileReturnFinishesAfterTheEntryPrompt — #909 found that
// ReturnFromExileToBattlefieldForEffect DROPPED a return whose pipeline
// queued a prompt, and Living Death leaned on it ("a battlefield entry
// cannot pause"). Same gap, same fix: the return waits and then
// completes, with the CR 400.7 new object identity intact.
func TestExileReturnFinishesAfterTheEntryPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pe2EntersTapped(g, opp.ID)
	blinked := uuid.New()
	g.Exile.PushTop(game.Card{
		InstanceID: blinked, Name: "Blinked Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})

	var newID uuid.UUID
	g.WithWriteLock(func() {
		newID, _ = g.ReturnFromExileToBattlefieldForEffect(blinked, me.ID, false)
	})
	if newID != uuid.Nil {
		t.Fatal("a paused return has nothing to report yet")
	}
	if !g.Exile.Contains(blinked) {
		t.Fatal("nothing moves while the question is open")
	}

	peAnswerOrder(t, g, me.ID)

	if g.Exile.Contains(blinked) {
		t.Fatal("the return completes once the prompt is answered")
	}
	if countBattlefieldNamed(g, me.ID, "Blinked Bear") != 1 {
		t.Fatal("the Bear is on the battlefield")
	}
	var back game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Blinked Bear" {
			back = c
		}
	}
	if back.InstanceID == blinked {
		t.Error("a returned permanent is a NEW OBJECT with a fresh instance ID (CR 400.7)")
	}
	if !back.Tapped {
		t.Error("both replacements applied, so it enters tapped")
	}
}

// TestReanimationFinishesAfterTheEntryPrompt — the third entry site,
// on the same resume.
func TestReanimationFinishesAfterTheEntryPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pe2EntersTapped(g, opp.ID)
	dead := pushGraveyardCardForTest(me, "Dead Bear")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == dead {
				me.Graveyard.Cards[i].TypeLine = "Creature — Bear"
				me.Graveyard.Cards[i].Power = 2
				me.Graveyard.Cards[i].Toughness = 2
			}
		}
		_ = g.ReturnFromGraveyardUnderControlForEffect(dead, game.ZoneBattlefield, me.ID)
	})

	if !me.Graveyard.Contains(dead) {
		t.Fatal("nothing moves while the question is open")
	}
	peAnswerOrder(t, g, me.ID)

	if me.Graveyard.Contains(dead) {
		t.Fatal("the reanimation completes once the prompt is answered")
	}
	card, ok := g.LookupCardForEffect(dead)
	if !ok || !g.Battlefield.Contains(dead) {
		t.Fatal("the creature is on the battlefield")
	}
	if !card.Tapped {
		t.Error("both replacements applied, so it enters tapped")
	}
	if card.Controller != me.ID {
		t.Error("reanimated under the named controller")
	}
}

// --- two open prompts ------------------------------------------------

// TestFetchedCommanderEntryWaitsBesideACommandZonePrompt — a paused
// ENTRY and a paused EXIT are two independent frames on one queue, and
// answering them in either order has to leave both cards where their
// answers say. The commander's own CR 903.9 question is about the OTHER
// card; a library → battlefield move is not a CR 903.9 destination, so
// the fetched commander is asked nothing about the command zone.
func TestFetchedCommanderEntryWaitsBesideACommandZonePrompt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pe2EntersTapped(g, opp.ID)
	fetched := pushLibraryCardForTest(me, game.Card{
		Name: "Fetched Commander", TypeLine: "Legendary Creature — Angel", Power: 4, Toughness: 4,
	})
	markCommanderCard(t, g, me, fetched)
	onBoard := b12Creature(g, me.ID, "Board Commander", "Legendary Creature — Angel", 4, 4)
	markCommanderCard(t, g, me, onBoard)

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(me.ID,
			func(c game.Card) bool { return c.Name == "Fetched Commander" },
			game.ZoneBattlefield, 1, false, true)
	})
	if b06ReplacementOrderFor(g, me.ID) == nil {
		t.Fatal("the fetched commander's entry queues the ordering prompt")
	}
	// A second, unrelated question: the other commander is exiled.
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(onBoard) })

	// Answer the CR 903.9 question first.
	b36AcceptCommandZone(t, g, me.ID)
	if !me.Command.Contains(onBoard) {
		t.Fatal("the board commander took the command zone")
	}
	if !me.Library.Contains(fetched) {
		t.Fatal("the fetched commander has not moved: its own question is still open")
	}

	peAnswerOrder(t, g, me.ID)
	if !g.Battlefield.Contains(fetched) {
		t.Fatal("the fetch finishes on its own answer")
	}
	if me.Library.Contains(fetched) || me.Command.Contains(fetched) {
		t.Error("the fetched commander is in exactly one place")
	}
}

// --- #701's prune, on an entry --------------------------------------

// TestEntryPromptIsPrunedWhenTheCardLeavesTheLibrary — a paused entry
// moves nothing, so the card sits in the library with the question
// open and something else can take it. #701's prune now covers entries
// for exactly that reason: without it the answer would rip the card out
// of whatever zone it had reached. The search behind it is still told,
// so its shuffle and its Then are not stranded.
func TestEntryPromptIsPrunedWhenTheCardLeavesTheLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pe2EntersTapped(g, opp.ID)
	gate := pushLibraryCardForTest(me, peGuildgate())
	for i := 0; i < 4; i++ {
		plTop(me, "Bear", "Creature — Bear", "{1}{G}")
	}
	before := len(g.Events)
	thenRan := 0

	g.WithWriteLock(func() {
		_ = g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
			Player:  me.ID,
			Pred:    func(c game.Card) bool { return c.Name == "Azorius Guildgate" },
			Dest:    game.ZoneBattlefield,
			Limit:   1,
			Shuffle: true,
			Then:    func(_ *game.Game, _ []uuid.UUID) error { thenRan++; return nil },
		})
	})
	if b06ReplacementOrderFor(g, me.ID) == nil {
		t.Fatal("the entry queues the ordering prompt")
	}
	if thenRan != 0 {
		t.Fatal("the search is not finished while the entry is open")
	}

	// The Guildgate leaves the library by another route entirely.
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(gate) })

	if b06ReplacementOrderFor(g, me.ID) != nil {
		t.Error("the prompt is pruned: the move it is asking about can never happen")
	}
	if !g.Exile.Contains(gate) {
		t.Error("the card is where the other route put it")
	}
	if g.Battlefield.Contains(gate) {
		t.Error("the abandoned entry did not rip the card back out of exile")
	}
	if thenRan != 1 {
		t.Errorf("the search's continuation ran %d times, want exactly 1", thenRan)
	}
	if n := peSearchEvents(g, before); n != 1 {
		t.Errorf("EventSearchLibrary fired %d times, want exactly 1", n)
	}
}

// --- sequencing, and the land drop ----------------------------------

// TestASecondFetchWaitsForTheFirstEntrysPrompt — a search may take
// more than one card ("search your library for up to two … put them
// onto the battlefield"), and a battlefield take is an ENTRY. The takes
// are therefore sequenced through the entry's continuation: the second
// Gate does not arrive over the top of the first one's open question,
// and the shuffle waits for both.
func TestASecondFetchWaitsForTheFirstEntrysPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pe2EntersTapped(g, opp.ID)
	gates := []uuid.UUID{
		pushLibraryCardForTest(me, peGuildgate()),
		pushLibraryCardForTest(me, game.Card{Name: "Boros Guildgate", TypeLine: "Land — Gate"}),
	}
	onBattlefield := func() int {
		n := 0
		for _, id := range gates {
			if g.Battlefield.Contains(id) {
				n++
			}
		}
		return n
	}
	before := len(g.Events)

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(me.ID,
			func(c game.Card) bool { return c.HasSubtype("Gate") },
			game.ZoneBattlefield, 2, false, true)
	})

	if onBattlefield() != 0 {
		t.Fatal("nothing moves while the first entry's question is open")
	}
	peAnswerOrder(t, g, me.ID)
	if n := onBattlefield(); n != 1 {
		t.Fatalf("%d Gates arrived; the second take is its own entry and its own question", n)
	}
	if peSearchEvents(g, before) != 0 {
		t.Error("the search is not finished until every take is")
	}
	peAnswerOrder(t, g, me.ID)

	if n := onBattlefield(); n != 2 {
		t.Fatalf("%d Gates arrived; the second one arrives on its own answer", n)
	}
	if n := peSearchEvents(g, before); n != 1 {
		t.Errorf("EventSearchLibrary fired %d times, want exactly 1", n)
	}
}

// TestAFetchedLandThatPausesIsStillNotALandDrop — CR 305.2. The resume
// bumps the per-turn land tally for the land PLAY it was written for,
// and used to tell a land play from everything else by "this is a land
// and the event carries no stack item". A fetched land that pauses now
// reaches the same resume and satisfies that inference, so the signal
// is declared on the event instead. A land an effect put onto the
// battlefield was not played, and the turn's land drop is still there.
func TestAFetchedLandThatPausesIsStillNotALandDrop(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pe2EntersTapped(g, opp.ID)
	gate := pushLibraryCardForTest(me, peGuildgate())

	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(me.ID,
			func(c game.Card) bool { return c.HasSubtype("Gate") },
			game.ZoneBattlefield, 1, false, true)
	})
	peAnswerOrder(t, g, me.ID)

	if !g.Battlefield.Contains(gate) {
		t.Fatal("the Gate arrived")
	}
	if n := g.LandsPlayedThisTurnFor(me.ID); n != 0 {
		t.Errorf("lands played this turn = %d, want 0 — a fetched land is not a land drop", n)
	}
}
