package testsl

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lengShuffleBoard is lengBoard plus Blightsteel Colossus in hand.
func lengShuffleBoard(t *testing.T, n int) (*game.Game, *game.Player, []uuid.UUID, uuid.UUID) {
	t.Helper()
	g, me, cards := effectstest.LengBoard(t, n)
	var blight uuid.UUID
	g.WithWriteLock(func() {
		blight = effectstest.HandCardFull(me, "Blightsteel Colossus", "Artifact Creature — Phyrexian Golem", "{12}", effectstest.BlightsteelOracle, nil)
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
		effectstest.LengDiscardThen(t, g, me, &ran, order...)
		accept := map[uuid.UUID]bool{kept[0]: true, kept[1]: true}
		if effectstest.LengSettleBatch(t, g, me, order, accept) == 0 {
			t.Fatalf("pos %d: no CR 616 order between Leng and Blightsteel", pos)
		}
		if !me.Library.Contains(blight) {
			t.Fatalf("pos %d: Blightsteel was not shuffled into the library", pos)
		}
		if n := effectstest.BlightsteelShuffleCount(g.Events[cut:], me.ID); n != 1 {
			t.Errorf("pos %d: %d shuffles, want exactly one (CR 701.24f)", pos, n)
		}
		prompt := effectstest.PutInLibraryPrompt(g)
		if prompt == nil {
			t.Fatalf("pos %d: no order offered for the two cards kept on top; pending %d", pos, len(g.PendingChoices))
		}
		if !effectstest.SameIDSet(prompt.ScryCards, kept...) {
			t.Fatalf("pos %d: ordering %v, want exactly the kept cards %v", pos, prompt.ScryCards, kept)
		}
		if ran != 0 {
			t.Fatalf("pos %d: the effect's then ran before the order", pos)
		}
		want := append([]uuid.UUID(nil), kept...)
		if err := g.ResolvePutInLibrary(prompt.ID, me.ID, nil, want); err != nil {
			t.Fatalf("pos %d: ResolvePutInLibrary: %v", pos, err)
		}
		if got := effectstest.LengLibraryTop(me, 2); !effectstest.SameIDs(got, want) {
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
			effectstest.PushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
		}
		keep, drop := cards[0], cards[1]
		order := []uuid.UUID{keep, blight, drop}
		w := &lengWatcher{}
		g.RegisterListener(w)
		ran := 0
		effectstest.LengDiscardThen(t, g, me, &ran, order...)
		for i, ev := range w.discards {
			if i < len(order) && ev.CardID != order[i] {
				t.Fatalf("discard %d was %v, want the picked order %v", i, ev.CardID, order)
			}
		}
		effectstest.LengSettleBatch(t, g, me, order, map[uuid.UUID]bool{keep: true})
		if effectstest.PutInLibraryPrompt(g) != nil {
			t.Fatal("one kept card is not an order to choose")
		}
		if got := effectstest.LengLibraryTop(me, 1); !effectstest.SameIDs(got, []uuid.UUID{keep}) {
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
	effectstest.LengDiscardThen(t, g, me, &ran, order...)
	effectstest.AnswerLengPrompt(t, g, me.ID, true)
	effectstest.AnswerLengPrompt(t, g, me.ID, true)
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceReplacementOrder {
		t.Fatalf("setup: want the CR 616 order on Blightsteel open, got %d choices", len(g.PendingChoices))
	}
	snap := g.Clone()
	accept := map[uuid.UUID]bool{kept[0]: true, kept[1]: true}
	effectstest.LengSettleBatch(t, g, me, order, accept)
	if effectstest.PutInLibraryPrompt(g) == nil {
		t.Fatal("first run: no order offered")
	}

	g.RestoreFrom(snap)
	me = g.Seats[0]
	cut := len(g.Events)
	effectstest.LengSettleBatch(t, g, me, order, accept)
	if n := effectstest.BlightsteelShuffleCount(g.Events[cut:], me.ID); n != 1 {
		t.Errorf("replay: %d shuffles, want one", n)
	}
	prompt := effectstest.PutInLibraryPrompt(g)
	if prompt == nil || !effectstest.SameIDSet(prompt.ScryCards, kept...) {
		t.Fatalf("replay: want an order over exactly the kept cards, got %+v", prompt)
	}
	if err := g.ResolvePutInLibrary(prompt.ID, me.ID, nil, kept); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := effectstest.LengLibraryTop(me, 2); !effectstest.SameIDs(got, kept) {
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
	g.WithWriteLock(func() {
		temper = effectstest.PushCatalogHandCard(me, "Fiery Temper", "Instant", effectstest.FieryTemperOracle)
	})
	keepA, keepB, drop := cards[0], cards[1], cards[2]
	order := []uuid.UUID{keepA, temper, blight, drop, keepB}
	ran := 0
	effectstest.LengDiscardThen(t, g, me, &ran, order...)
	effectstest.LengSettleBatch(t, g, me, order, map[uuid.UUID]bool{keepA: true, keepB: true})
	if !g.Exile.Contains(temper) || !me.Graveyard.Contains(drop) || !me.Library.Contains(blight) {
		t.Fatalf("destinations: temper exiled %v, drop in graveyard %v, Blightsteel in library %v",
			g.Exile.Contains(temper), me.Graveyard.Contains(drop), me.Library.Contains(blight))
	}
	prompt := effectstest.PutInLibraryPrompt(g)
	if prompt == nil || !effectstest.SameIDSet(prompt.ScryCards, keepA, keepB) {
		t.Fatalf("want an order over exactly the two kept cards, got %+v", prompt)
	}
	want := []uuid.UUID{keepA, keepB}
	if err := g.ResolvePutInLibrary(prompt.ID, me.ID, nil, want); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := effectstest.LengLibraryTop(me, 2); !effectstest.SameIDs(got, want) {
		t.Errorf("library top %v, want %v", got, want)
	}
	if ran != 1 {
		t.Errorf("then ran %d times, want once", ran)
	}
}
