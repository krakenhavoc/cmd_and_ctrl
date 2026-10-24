package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stickers_cards_test.go pins the sticker sheets and the cards that give
// tickets and put stickers on (ADR 0208, #2138), each through the
// engine's three keyed questions.

const (
	nightBrushwaggOracle   = "1436370f-baf0-4659-804e-90aad871b05b"
	midwayBamboozleOracle  = "dde09abb-e3d3-4c76-b7e8-812949dd67f3"
	ticketomatonOracle     = "dc8290d7-83b0-49ef-b0de-7ebb3633a149"
	nightBrushwaggText     = "{TK}{TK} — Menace\n{TK}{TK}{TK} — Persist (When this permanent dies, if it had no -1/-1 counters on it, return it to the battlefield under its owner's control with a -1/-1 counter on it.)\n{TK}{TK} — 2/3\n{TK}{TK}{TK}{TK}{TK}{TK} — 10/10"
	midwayBamboozleText    = "{TK}{TK} — Whenever this creature attacks, you get {TK}.\n{TK}{TK}{TK} — Undying (When this creature dies, if it had no +1/+1 counters on it, return it to the battlefield under its owner's control with a +1/+1 counter on it.)\n{TK}{TK} — 4/2\n{TK}{TK}{TK}{TK}{TK} — 6/9"
	stickerSheetTypeLine   = "Stickers"
	ticketomatonTypeLine   = "Artifact Creature — Robot"
	stkPersistStickerIndex = 4 // three names, menace, then persist
)

// stkSheet is one of the two sheets as the deck importer reads it.
func stkSheet(name, oracle, text string) game.StickerSheet {
	return deck.StickerSheetOf(cards.Card{
		OracleID: uuid.MustParse(oracle), Name: name, TypeLine: stickerSheetTypeLine, OracleText: text,
	})
}

// stkTable is a catalog game whose seat 0 plays with the two sheets and
// has `tickets` ticket counters.
func stkTable(t *testing.T, tickets int) (*game.Game, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		me.StickerSheets = []game.StickerSheet{
			stkSheet("Night Brushwagg Ringmaster", nightBrushwaggOracle, nightBrushwaggText),
			stkSheet("Phyrexian Midway Bamboozle", midwayBamboozleOracle, midwayBamboozleText),
		}
		if tickets > 0 {
			if err := g.AddPlayerCounterByForEffect(me.ID, me.ID, game.CounterTicket, tickets); err != nil {
				t.Fatalf("tickets: %v", err)
			}
		}
	})
	return g, me
}

func stkTickets(g *game.Game, p *game.Player) int {
	n := 0
	g.ReadSnapshot(func() { n = game.PlayerTickets(p) })
	return n
}

// stkOption is the index of the open option pick's option whose label
// contains `want`, or fails.
func stkOption(t *testing.T, c *game.PendingChoice, want string) int {
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
func stkPutOn(t *testing.T, g *game.Game, me *game.Player, obj uuid.UUID, sticker string, after int) {
	t.Helper()
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no question asks which permanent gets the sticker")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{obj}); err != nil {
		t.Fatalf("choose the permanent: %v", err)
	}
	which := latestOptionPickFor(g, me.ID)
	if which == nil {
		t.Fatal("no question asks which sticker")
	}
	if err := g.ResolveOptionPick(which.ID, me.ID, stkOption(t, which, sticker)); err != nil {
		t.Fatalf("choose the sticker: %v", err)
	}
	if where := latestOptionPickFor(g, me.ID); where != nil && strings.HasPrefix(where.Reason, "Where in the name") {
		if err := g.ResolveOptionPick(where.ID, me.ID, after); err != nil {
			t.Fatalf("choose the position: %v", err)
		}
	}
}

// stkGive puts a sticker on a permanent the way a resolving effect does,
// with the questions answered: the card's own effect, without a card.
func stkGive(t *testing.T, g *game.Game, me *game.Player, obj uuid.UUID, sticker string) {
	t.Helper()
	g.WithWriteLock(func() {
		g.PutStickerForEffect(game.StickerRequest{Player: me.ID, Objects: []uuid.UUID{obj}})
	})
	stkPutOn(t, g, me, obj, sticker, 0)
}

func stkCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
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

// Ticketomaton: "When Ticketomaton enters, you get {TK}, then you may put
// a sticker on a nonland permanent you own." The ticket it gives pays for
// a sticker with it.
func TestStkTicketomatonGetsATicketThenPutsASticker(t *testing.T) {
	g, me := stkTable(t, 1)
	bot := castAndResolveCreature(t, g, "Ticketomaton", ticketomatonTypeLine, ticketomatonOracle)
	passPriorityAroundTable(t, g)
	if got := stkTickets(g, me); got != 2 {
		t.Fatalf("tickets after the trigger = %d, want 2 (1 + the one it gives)", got)
	}
	stkPutOn(t, g, me, bot, "Menace", 0)
	passPriorityAroundTable(t, g)
	c := stkCard(t, g, bot)
	if !game.HasKeyword(&c, game.CounterMenace) {
		t.Error("the menace sticker did not give menace")
	}
	if got := stkTickets(g, me); got != 0 {
		t.Errorf("tickets after the menace sticker = %d, want 0 (it costs {TK}{TK})", got)
	}
	if len(c.Stickers) != 1 || c.Stickers[0].Print.Text != "Menace" {
		t.Errorf("stickers = %+v, want the one menace sticker", c.Stickers)
	}
}

// "You may": answering with no permanent puts nothing on and pays
// nothing; a sticker the player can't afford is never offered (CR
// 123.3c).
func TestStkStickerIsOptionalAndUnaffordableOnesAreNotOffered(t *testing.T) {
	g, me := stkTable(t, 0)
	bot := castAndResolveCreature(t, g, "Ticketomaton", ticketomatonTypeLine, ticketomatonOracle)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the name and art stickers cost nothing, so the question is asked")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bot}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	which := latestOptionPickFor(g, me.ID)
	for _, o := range which.PickOptions {
		if strings.Contains(o.Label, "Persist") || strings.Contains(o.Label, "10/10") {
			t.Errorf("offered %q with one ticket", o.Label)
		}
	}
	if err := g.ResolveOptionPick(which.ID, me.ID, stkOption(t, which, game.DeclineStickerLabel)); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if c := stkCard(t, g, bot); len(c.Stickers) != 0 {
		t.Errorf("declined, yet stickers = %+v", c.Stickers)
	}
	if got := stkTickets(g, me); got != 1 {
		t.Errorf("tickets = %d, want 1", got)
	}
}

// A name sticker adds its word where its controller says (CR 123.6b), and
// the legend rule and the wire read the new name.
func TestStkNameStickerAddsAWord(t *testing.T) {
	g, me := stkTable(t, 0)
	bear := p2Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	stkGive(t, g, me, bear, "“Night”")
	if got := stkCard(t, g, bear).Effective().Name; got != "Night Grizzly Bears" {
		t.Errorf("name = %q, want %q", got, "Night Grizzly Bears")
	}
	g.WithWriteLock(func() {
		g.PutStickerForEffect(game.StickerRequest{Player: me.ID, Objects: []uuid.UUID{bear}})
	})
	stkPutOn(t, g, me, bear, "“Ringmaster”", 3)
	if got := stkCard(t, g, bear).Effective().Name; got != "Night Grizzly Bears Ringmaster" {
		t.Errorf("name = %q, want %q", got, "Night Grizzly Bears Ringmaster")
	}
}

// A power and toughness sticker sets the creature's base power and
// toughness in layer 7b (CR 123.8): a +1/+1 counter still counts on top.
func TestStkPowerToughnessStickerSetsBase(t *testing.T) {
	g, me := stkTable(t, 2)
	bear := p2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	stkGive(t, g, me, bear, "sticker 2/3")
	if p, tough := p2PT(g, bear); p != 2 || tough != 3 {
		t.Fatalf("P/T = %d/%d, want 2/3", p, tough)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	if p, tough := p2PT(g, bear); p != 3 || tough != 4 {
		t.Errorf("P/T with a +1/+1 counter = %d/%d, want 3/4", p, tough)
	}
}

// Night Brushwagg Ringmaster's persist sticker: the creature returns with
// a -1/-1 counter, and the sticker comes back with it (CR 123.5), so it
// does not return a second time.
func TestStkPersistStickerReturnsTheCreatureOnce(t *testing.T) {
	g, me := stkTable(t, 3)
	bear := p2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	stkGive(t, g, me, bear, "Persist")
	if c := stkCard(t, g, bear); !game.HasKeyword(&c, game.KeywordPersist) {
		t.Fatal("the persist sticker did not give persist")
	}
	b25Destroy(g, bear)
	settleProwess(t, g)
	if !onBattlefield(g, bear) || countersOn(g, bear, game.CounterMinusOne) != 1 {
		t.Fatal("persist did not return the creature with a -1/-1 counter")
	}
	c := stkCard(t, g, bear)
	if len(c.Stickers) != 1 || !game.HasKeyword(&c, game.KeywordPersist) {
		t.Fatalf("the returned creature lost its sticker: %+v", c.Stickers)
	}
	b25Destroy(g, bear)
	settleProwess(t, g)
	if onBattlefield(g, bear) {
		t.Error("persist returned a creature that had a -1/-1 counter")
	}
}

// Phyrexian Midway Bamboozle's undying sticker: the creature returns with
// a +1/+1 counter.
func TestStkUndyingStickerReturnsTheCreature(t *testing.T) {
	g, me := stkTable(t, 3)
	bear := p2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	stkGive(t, g, me, bear, "Undying")
	b25Destroy(g, bear)
	settleProwess(t, g)
	if !onBattlefield(g, bear) || countersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Fatal("undying did not return the creature with a +1/+1 counter")
	}
	if c := stkCard(t, g, bear); len(c.Stickers) != 1 {
		t.Errorf("the returned creature lost its sticker: %+v", c.Stickers)
	}
}

// Phyrexian Midway Bamboozle's attack sticker: "Whenever this creature
// attacks, you get {TK}."
func TestStkBamboozleAttackStickerGivesATicket(t *testing.T) {
	g, me := stkTable(t, 2)
	bear := p2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	stkGive(t, g, me, bear, "Whenever this creature attacks")
	if got := stkTickets(g, me); got != 0 {
		t.Fatalf("tickets after paying = %d, want 0", got)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := stkTickets(g, me); got != 1 {
		t.Errorf("tickets after attacking = %d, want 1", got)
	}
}

// A sticker on an object the player owns can't be put on another, and
// one whose object went to a hidden zone comes off and is free again (CR
// 123.3, 123.5).
func TestStkUsedStickerIsNotOfferedUntilItComesOff(t *testing.T) {
	g, me := stkTable(t, 0)
	bear := p2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	other := p2Push(g, me.ID, "Other Bear", "Creature — Bear", "", 2, 2)
	stkGive(t, g, me, bear, "“Night”")
	offered := func() bool {
		found := false
		g.WithWriteLock(func() {
			for _, s := range g.AvailableStickersForEffect(me.ID, nil) {
				if s.Print.Text == "Night" {
					found = true
				}
			}
		})
		return found
	}
	if offered() {
		t.Fatal("the Night sticker is on a permanent and was still offered")
	}
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(bear) })
	if c := stkCard(t, g, bear); len(c.Stickers) != 0 {
		t.Errorf("a card in a hand kept its stickers: %+v", c.Stickers)
	}
	if !offered() {
		t.Error("the Night sticker came off with the card and is not offered again")
	}
	_ = other
}

// Each of the three questions is a restore point: the table restores
// mid-question and the answer still puts the sticker on.
func TestStkStickerQuestionsAreRestorePoints(t *testing.T) {
	g, me := stkTable(t, 0)
	bear := p2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		g.PutStickerForEffect(game.StickerRequest{Player: me.ID, Objects: []uuid.UUID{bear}})
	})
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the restored table lost the object question")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	which := latestOptionPickFor(g, me.ID)
	if which == nil {
		t.Fatal("the restored table lost the sticker question")
	}
	if err := g.ResolveOptionPick(which.ID, me.ID, stkOption(t, which, "“Bamboozle”")); err != nil {
		t.Fatalf("pick: %v", err)
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	where := latestOptionPickFor(g, me.ID)
	if where == nil {
		t.Fatal("the restored table lost the position question")
	}
	if err := g.ResolveOptionPick(where.ID, me.ID, 1); err != nil {
		t.Fatalf("position: %v", err)
	}
	g = restoreRoundTrip(t, g, true)
	if got := stkCard(t, g, bear).Effective().Name; got != "Bear Bamboozle" {
		t.Errorf("restored name = %q, want %q", got, "Bear Bamboozle")
	}
}

// The two sheets and Ticketomaton are catalogued, and the sheets say what
// their ability stickers give.
func TestStkStickerSheetsRegistered(t *testing.T) {
	for _, oracle := range []string{nightBrushwaggOracle, midwayBamboozleOracle, ticketomatonOracle} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Completeness != CompletenessFull {
			t.Errorf("%s: registered=%v completeness=%v", oracle, ok, spec.Completeness)
		}
	}
	for _, tc := range []struct {
		sheet   string
		ordinal int
		kw      string
		grant   bool
	}{
		{nightBrushwaggOracle, 0, game.CounterMenace, false},
		{nightBrushwaggOracle, 1, game.KeywordPersist, false},
		{midwayBamboozleOracle, 0, "", true},
		{midwayBamboozleOracle, 1, game.KeywordUndying, false},
	} {
		ab, ok := game.CatalogStickerAbility(tc.sheet, tc.ordinal)
		if !ok {
			t.Errorf("%s ability sticker %d: not catalogued", tc.sheet, tc.ordinal)
			continue
		}
		if tc.kw != "" && (len(ab.Keywords) != 1 || ab.Keywords[0] != tc.kw) {
			t.Errorf("%s ability sticker %d keywords = %v, want %q", tc.sheet, tc.ordinal, ab.Keywords, tc.kw)
		}
		if tc.grant != (ab.Grant != "") {
			t.Errorf("%s ability sticker %d grant = %q", tc.sheet, tc.ordinal, ab.Grant)
		}
	}
}
