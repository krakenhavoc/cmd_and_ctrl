package effectstest

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stickers_cards_test.go pins the sticker sheets and the cards that give
// tickets and put stickers on (ADR 0208, #2138), each through the
// engine's three keyed questions.

const (
	NightBrushwaggOracle   = "1436370f-baf0-4659-804e-90aad871b05b"
	MidwayBamboozleOracle  = "dde09abb-e3d3-4c76-b7e8-812949dd67f3"
	TicketomatonOracle     = "dc8290d7-83b0-49ef-b0de-7ebb3633a149"
	NightBrushwaggText     = "{TK}{TK} — Menace\n{TK}{TK}{TK} — Persist (When this permanent dies, if it had no -1/-1 counters on it, return it to the battlefield under its owner's control with a -1/-1 counter on it.)\n{TK}{TK} — 2/3\n{TK}{TK}{TK}{TK}{TK}{TK} — 10/10"
	MidwayBamboozleText    = "{TK}{TK} — Whenever this creature attacks, you get {TK}.\n{TK}{TK}{TK} — Undying (When this creature dies, if it had no +1/+1 counters on it, return it to the battlefield under its owner's control with a +1/+1 counter on it.)\n{TK}{TK} — 4/2\n{TK}{TK}{TK}{TK}{TK} — 6/9"
	StickerSheetTypeLine   = "Stickers"
	TicketomatonTypeLine   = "Artifact Creature — Robot"
	stkPersistStickerIndex = 4 // three names, menace, then persist
)

// stkOption is the index of the open option pick's option whose label
// contains `want`, or fails.
func StkOption(t *testing.T, c *game.PendingChoice, want string) int {
	t.Helper()
	for i, o := range c.PickOptions {
		if strings.Contains(o.Label, want) {
			return i
		}
	}
	labels := make([]string, 0, len(c.PickOptions))
	for _, o := range c.PickOptions {
		labels = append(labels, o.Label)
	}
	t.Fatalf("no option contains %q: %q", want, labels)
	return -1
}

// stkPutOn answers the open sticker questions: put the sticker whose
// label contains `sticker` on `obj`, a name sticker after `after` words.
func StkPutOn(t *testing.T, g *game.Game, me *game.Player, obj uuid.UUID, sticker string, after int) {
	t.Helper()
	pick := LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no question asks which permanent gets the sticker")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{obj}); err != nil {
		t.Fatalf("choose the permanent: %v", err)
	}
	which := LatestOptionPickFor(g, me.ID)
	if which == nil {
		t.Fatal("no question asks which sticker")
	}
	if err := g.ResolveOptionPick(which.ID, me.ID, StkOption(t, which, sticker)); err != nil {
		t.Fatalf("choose the sticker: %v", err)
	}
	if where := LatestOptionPickFor(g, me.ID); where != nil && strings.HasPrefix(where.Reason, "Where in the name") {
		if err := g.ResolveOptionPick(where.ID, me.ID, after); err != nil {
			t.Fatalf("choose the position: %v", err)
		}
	}
}

// stkGive puts a sticker on a permanent the way a resolving effect does,
// with the questions answered: the card's own effect, without a card.
func StkGive(t *testing.T, g *game.Game, me *game.Player, obj uuid.UUID, sticker string) {
	t.Helper()
	g.WithWriteLock(func() {
		g.PutStickerForEffect(game.StickerRequest{Player: me.ID, Objects: []uuid.UUID{obj}})
	})
	StkPutOn(t, g, me, obj, sticker, 0)
}

func StkCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var c game.Card
	var ok bool
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, ok = g.LookupCardForEffect(id)
	})
	if !ok {
		t.Fatalf("card %v is nowhere", id)
	}
	return c
}
