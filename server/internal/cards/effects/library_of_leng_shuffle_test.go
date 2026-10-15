package effects

import (
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// library_of_leng_shuffle_test.go — #3541: one discard puts some cards
// on top of the library (Library of Leng) and shuffles another into it
// (Blightsteel Colossus). CR 701.24g: the library is shuffled and THEN
// the top cards are put on top, and CR 401.4 lets their owner order
// exactly those cards. The legs settle one after another, so a later
// shuffle used to scatter the earlier top placements and the batch then
// offered no order at all.

// lengSettleBatch answers the discard batch's replacement prompts as it
// walks: each Leng "may" by `accept` for the card being discarded (the
// first of `order` still in hand), and each CR 616 order with Library of
// Leng first. It stops at the first prompt that is neither, and reports
// how many CR 616 order prompts it answered.
func lengSettleBatch(t *testing.T, g *game.Game, me *game.Player, order []uuid.UUID, accept map[uuid.UUID]bool) int {
	t.Helper()
	orders := 0
	for step := 0; step < 4*len(order)+4; step++ {
		var c *game.PendingChoice
		for _, pc := range g.PendingChoices {
			if pc != nil && (pc.Kind == game.PendingChoiceOptionalReplacement || pc.Kind == game.PendingChoiceReplacementOrder) {
				c = pc
				break
			}
		}
		if c == nil {
			return orders
		}
		var cur uuid.UUID
		for _, id := range order {
			if me.Hand.Contains(id) {
				cur = id
				break
			}
		}
		switch c.Kind {
		case game.PendingChoiceReplacementOrder:
			// Library of Leng is a battlefield card's replacement, so
			// its ID is in the catalog range, below every card's own
			// ("reveal it and shuffle it in instead") replacement.
			ids := append([]game.ReplacementEffectID(nil), c.ReplacementEffectIDs...)
			sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
			if err := g.ResolveReplacementOrder(c.ID, me.ID, ids); err != nil {
				t.Fatalf("ResolveReplacementOrder: %v", err)
			}
			orders++
		case game.PendingChoiceOptionalReplacement:
			if err := g.ResolveOptionalReplacement(c.ID, me.ID, accept[cur]); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}
		}
	}
	t.Fatal("the discard batch never stopped asking")
	return orders
}

// lengShuffleBoard is lengBoard plus Blightsteel Colossus in hand.
func lengShuffleBoard(t *testing.T, n int) (*game.Game, *game.Player, []uuid.UUID, uuid.UUID) {
	t.Helper()
	g, me, cards := lengBoard(t, n)
	var blight uuid.UUID
	g.WithWriteLock(func() {
		blight = handCardFull(me, "Blightsteel Colossus", "Artifact Creature — Phyrexian Golem", "{12}", blightsteelOracle, nil)
	})
	return g, me, cards, blight
}

// TestLibraryOfLengBatchOrderSurvivesLaterShuffleIn — the issue's
// reproduction, in every position of the shuffled card: two cards kept
// on top, Blightsteel ordered behind Leng, Leng declined for it, and it
// shuffles itself in. Exactly one shuffle, an order offered over exactly
// the two kept cards, the chosen order on top, and the "then" waiting.
func TestLibraryOfLengBatchOrderSurvivesLaterShuffleIn(t *testing.T) {
	for _, pos := range []int{0, 1, 2} {
		g, me, kept, blight := lengShuffleBoard(t, 2)
		order := append([]uuid.UUID(nil), kept...)
		order = append(order[:pos], append([]uuid.UUID{blight}, order[pos:]...)...)
		w := &lengWatcher{}
		g.RegisterListener(w)
		cut := len(g.Events)
		ran := 0
		lengDiscardThen(t, g, me, &ran, order...)
		accept := map[uuid.UUID]bool{kept[0]: true, kept[1]: true}
		if lengSettleBatch(t, g, me, order, accept) == 0 {
			t.Fatalf("pos %d: no CR 616 order between Leng and Blightsteel", pos)
		}
		if !me.Library.Contains(blight) {
			t.Fatalf("pos %d: Blightsteel was not shuffled into the library", pos)
		}
		if n := blightsteelShuffleCount(g.Events[cut:], me.ID); n != 1 {
			t.Errorf("pos %d: %d shuffles, want exactly one (CR 701.24f)", pos, n)
		}
		prompt := putInLibraryPrompt(g)
		if prompt == nil {
			t.Fatalf("pos %d: no order offered for the two cards kept on top; pending %d", pos, len(g.PendingChoices))
		}
		if !sameIDSet(prompt.ScryCards, kept...) {
			t.Fatalf("pos %d: ordering %v, want exactly the kept cards %v", pos, prompt.ScryCards, kept)
		}
		if ran != 0 {
			t.Fatalf("pos %d: the effect's then ran before the order", pos)
		}
		want := append([]uuid.UUID(nil), kept...)
		if err := g.ResolvePutInLibrary(prompt.ID, me.ID, nil, want); err != nil {
			t.Fatalf("pos %d: ResolvePutInLibrary: %v", pos, err)
		}
		if got := lengLibraryTop(me, 2); !sameIDs(got, want) {
			t.Errorf("pos %d: library top %v, want the chosen %v", pos, got, want)
		}
		if ran != 1 {
			t.Errorf("pos %d: then ran %d times, want once", pos, ran)
		}
		if len(w.discards) != 3 {
			t.Errorf("pos %d: EventDiscardCard x %d, want 3", pos, len(w.discards))
		}
	}
}

// One card kept, then a shuffle-in: nothing to order, but CR 701.24g
// still puts the kept card back on top after the shuffle. A declined
// card goes to the graveyard and is no part of the placement.
func TestLibraryOfLengOneKeptCardStaysOnTopAcrossAShuffleIn(t *testing.T) {
	// The shuffle is seeded, so each trial grows the library by one
	// card to land the shuffled-away card somewhere else.
	for trial := 0; trial < 6; trial++ {
		g, me, cards, blight := lengShuffleBoard(t, 2)
		for i := 0; i < trial; i++ {
			pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
		}
		keep, drop := cards[0], cards[1]
		order := []uuid.UUID{keep, blight, drop}
		w := &lengWatcher{}
		g.RegisterListener(w)
		ran := 0
		lengDiscardThen(t, g, me, &ran, order...)
		for i, ev := range w.discards {
			if i < len(order) && ev.CardID != order[i] {
				t.Fatalf("discard %d was %v, want the picked order %v", i, ev.CardID, order)
			}
		}
		lengSettleBatch(t, g, me, order, map[uuid.UUID]bool{keep: true})
		if putInLibraryPrompt(g) != nil {
			t.Fatal("one kept card is not an order to choose")
		}
		if got := lengLibraryTop(me, 1); !sameIDs(got, []uuid.UUID{keep}) {
			t.Fatalf("trial %d: library top %v, want the kept card on top after the shuffle", trial, got)
		}
		if !me.Graveyard.Contains(drop) || !me.Library.Contains(blight) {
			t.Error("the declined card is in the graveyard and Blightsteel in the library")
		}
		if ran != 1 {
			t.Errorf("then ran %d times, want once", ran)
		}
	}
}

// Undo (Clone / RestoreFrom) to the open CR 616 order on Blightsteel,
// with both kept cards already on top: the replay shuffles once more
// and still offers the order over exactly the kept cards.
func TestLibraryOfLengShuffleInOrderSurvivesUndo(t *testing.T) {
	g, me, kept, blight := lengShuffleBoard(t, 2)
	order := []uuid.UUID{kept[0], kept[1], blight}
	ran := 0
	lengDiscardThen(t, g, me, &ran, order...)
	answerLengPrompt(t, g, me.ID, true)
	answerLengPrompt(t, g, me.ID, true)
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("setup: want the CR 616 order on Blightsteel open, got %d choices", len(g.PendingChoices))
	}
	snap := g.Clone()
	accept := map[uuid.UUID]bool{kept[0]: true, kept[1]: true}
	lengSettleBatch(t, g, me, order, accept)
	if putInLibraryPrompt(g) == nil {
		t.Fatal("first run: no order offered")
	}

	g.RestoreFrom(snap)
	me = g.Seats[0]
	cut := len(g.Events)
	lengSettleBatch(t, g, me, order, accept)
	if n := blightsteelShuffleCount(g.Events[cut:], me.ID); n != 1 {
		t.Errorf("replay: %d shuffles, want one", n)
	}
	prompt := putInLibraryPrompt(g)
	if prompt == nil || !sameIDSet(prompt.ScryCards, kept...) {
		t.Fatalf("replay: want an order over exactly the kept cards, got %+v", prompt)
	}
	if err := g.ResolvePutInLibrary(prompt.ID, me.ID, nil, kept); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := lengLibraryTop(me, 2); !sameIDs(got, kept) {
		t.Errorf("library top %v, want %v", got, kept)
	}
	if ran != 1 {
		t.Errorf("then ran %d times across the undo, want once", ran)
	}
}

// Library, graveyard and exile in one discard: two kept on top, a
// madness card exiled (Leng declined for it), a plain card declined to
// the graveyard, and Blightsteel shuffled in. The order covers only the
// two kept cards.
func TestLibraryOfLengMixedDestinationsOrderOnlyTheKeptCards(t *testing.T) {
	g, me, cards, blight := lengShuffleBoard(t, 3)
	var temper uuid.UUID
	g.WithWriteLock(func() { temper = pushCatalogHandCard(me, "Fiery Temper", "Instant", fieryTemperOracle) })
	keepA, keepB, drop := cards[0], cards[1], cards[2]
	order := []uuid.UUID{keepA, temper, blight, drop, keepB}
	ran := 0
	lengDiscardThen(t, g, me, &ran, order...)
	lengSettleBatch(t, g, me, order, map[uuid.UUID]bool{keepA: true, keepB: true})
	if !g.Exile.Contains(temper) || !me.Graveyard.Contains(drop) || !me.Library.Contains(blight) {
		t.Fatalf("destinations: temper exiled %v, drop in graveyard %v, Blightsteel in library %v",
			g.Exile.Contains(temper), me.Graveyard.Contains(drop), me.Library.Contains(blight))
	}
	prompt := putInLibraryPrompt(g)
	if prompt == nil || !sameIDSet(prompt.ScryCards, keepA, keepB) {
		t.Fatalf("want an order over exactly the two kept cards, got %+v", prompt)
	}
	want := []uuid.UUID{keepA, keepB}
	if err := g.ResolvePutInLibrary(prompt.ID, me.ID, nil, want); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := lengLibraryTop(me, 2); !sameIDs(got, want) {
		t.Errorf("library top %v, want %v", got, want)
	}
	if ran != 1 {
		t.Errorf("then ran %d times, want once", ran)
	}
}
