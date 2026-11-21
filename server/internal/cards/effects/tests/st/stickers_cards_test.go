package testsst

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stkSheet is one of the two sheets as the deck importer reads it.
func stkSheet(name, oracle, text string) game.StickerSheet {
	return deck.StickerSheetOf(cards.Card{
		OracleID: uuid.MustParse(oracle), Name: name, TypeLine: effectstest.StickerSheetTypeLine, OracleText: text,
	})
}

// stkTable is a catalog game whose seat 0 plays with the two sheets and
// has `tickets` ticket counters.
func stkTable(t *testing.T, tickets int) (*game.Game, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		me.StickerSheets = []game.StickerSheet{
			stkSheet("Night Brushwagg Ringmaster", effectstest.NightBrushwaggOracle, effectstest.NightBrushwaggText),
			stkSheet("Phyrexian Midway Bamboozle", effectstest.MidwayBamboozleOracle, effectstest.MidwayBamboozleText),
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

// Ticketomaton: "When Ticketomaton enters, you get {TK}, then you may put
// a sticker on a nonland permanent you own." The ticket it gives pays for
// a sticker with it.
func TestStkTicketomatonGetsATicketThenPutsASticker(t *testing.T) {
	g, me := stkTable(t, 1)
	bot := effectstest.CastAndResolveCreature(t, g, "Ticketomaton", effectstest.TicketomatonTypeLine, effectstest.TicketomatonOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if got := stkTickets(g, me); got != 2 {
		t.Fatalf("tickets after the trigger = %d, want 2 (1 + the one it gives)", got)
	}
	effectstest.StkPutOn(t, g, me, bot, "Menace", 0)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.StkCard(t, g, bot)
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
	bot := effectstest.CastAndResolveCreature(t, g, "Ticketomaton", effectstest.TicketomatonTypeLine, effectstest.TicketomatonOracle)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the name and art stickers cost nothing, so the question is asked")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bot}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	which := effectstest.LatestOptionPickFor(g, me.ID)
	for _, o := range which.PickOptions {
		if strings.Contains(o.Label, "Persist") || strings.Contains(o.Label, "10/10") {
			t.Errorf("offered %q with one ticket", o.Label)
		}
	}
	if err := g.ResolveOptionPick(which.ID, me.ID, effectstest.StkOption(t, which, game.DeclineStickerLabel)); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if c := effectstest.StkCard(t, g, bot); len(c.Stickers) != 0 {
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
	bear := effectstest.P2Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	effectstest.StkGive(t, g, me, bear, "“Night”")
	if got := effectstest.StkCard(t, g, bear).Effective().Name; got != "Night Grizzly Bears" {
		t.Errorf("name = %q, want %q", got, "Night Grizzly Bears")
	}
	g.WithWriteLock(func() {
		g.PutStickerForEffect(game.StickerRequest{Player: me.ID, Objects: []uuid.UUID{bear}})
	})
	effectstest.StkPutOn(t, g, me, bear, "“Ringmaster”", 3)
	if got := effectstest.StkCard(t, g, bear).Effective().Name; got != "Night Grizzly Bears Ringmaster" {
		t.Errorf("name = %q, want %q", got, "Night Grizzly Bears Ringmaster")
	}
}

// A power and toughness sticker sets the creature's base power and
// toughness in layer 7b (CR 123.8): a +1/+1 counter still counts on top.
func TestStkPowerToughnessStickerSetsBase(t *testing.T) {
	g, me := stkTable(t, 2)
	bear := effectstest.P2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.StkGive(t, g, me, bear, "sticker 2/3")
	if p, tough := effectstest.P2PT(g, bear); p != 2 || tough != 3 {
		t.Fatalf("P/T = %d/%d, want 2/3", p, tough)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	if p, tough := effectstest.P2PT(g, bear); p != 3 || tough != 4 {
		t.Errorf("P/T with a +1/+1 counter = %d/%d, want 3/4", p, tough)
	}
}

// Night Brushwagg Ringmaster's persist sticker: the creature returns with
// a -1/-1 counter, and the sticker comes back with it (CR 123.5), so it
// does not return a second time.
func TestStkPersistStickerReturnsTheCreatureOnce(t *testing.T) {
	g, me := stkTable(t, 3)
	bear := effectstest.P2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.StkGive(t, g, me, bear, "Persist")
	if c := effectstest.StkCard(t, g, bear); !game.HasKeyword(&c, game.KeywordPersist) {
		t.Fatal("the persist sticker did not give persist")
	}
	effectstest.B25Destroy(g, bear)
	effectstest.SettleProwess(t, g)
	if !effects.OnBattlefield(g, bear) || effectstest.CountersOn(g, bear, game.CounterMinusOne) != 1 {
		t.Fatal("persist did not return the creature with a -1/-1 counter")
	}
	c := effectstest.StkCard(t, g, bear)
	if len(c.Stickers) != 1 || !game.HasKeyword(&c, game.KeywordPersist) {
		t.Fatalf("the returned creature lost its sticker: %+v", c.Stickers)
	}
	effectstest.B25Destroy(g, bear)
	effectstest.SettleProwess(t, g)
	if effects.OnBattlefield(g, bear) {
		t.Error("persist returned a creature that had a -1/-1 counter")
	}
}

// Phyrexian Midway Bamboozle's undying sticker: the creature returns with
// a +1/+1 counter.
func TestStkUndyingStickerReturnsTheCreature(t *testing.T) {
	g, me := stkTable(t, 3)
	bear := effectstest.P2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.StkGive(t, g, me, bear, "Undying")
	effectstest.B25Destroy(g, bear)
	effectstest.SettleProwess(t, g)
	if !effects.OnBattlefield(g, bear) || effectstest.CountersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Fatal("undying did not return the creature with a +1/+1 counter")
	}
	if c := effectstest.StkCard(t, g, bear); len(c.Stickers) != 1 {
		t.Errorf("the returned creature lost its sticker: %+v", c.Stickers)
	}
}

// Phyrexian Midway Bamboozle's attack sticker: "Whenever this creature
// attacks, you get {TK}."
func TestStkBamboozleAttackStickerGivesATicket(t *testing.T) {
	g, me := stkTable(t, 2)
	bear := effectstest.P2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.StkGive(t, g, me, bear, "Whenever this creature attacks")
	if got := stkTickets(g, me); got != 0 {
		t.Fatalf("tickets after paying = %d, want 0", got)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := stkTickets(g, me); got != 1 {
		t.Errorf("tickets after attacking = %d, want 1", got)
	}
}

// A sticker on an object the player owns can't be put on another, and
// one whose object went to a hidden zone comes off and is free again (CR
// 123.3, 123.5).
func TestStkUsedStickerIsNotOfferedUntilItComesOff(t *testing.T) {
	g, me := stkTable(t, 0)
	bear := effectstest.P2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	other := effectstest.P2Push(g, me.ID, "Other Bear", "Creature — Bear", "", 2, 2)
	effectstest.StkGive(t, g, me, bear, "“Night”")
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
	if c := effectstest.StkCard(t, g, bear); len(c.Stickers) != 0 {
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
	bear := effectstest.P2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		g.PutStickerForEffect(game.StickerRequest{Player: me.ID, Objects: []uuid.UUID{bear}})
	})
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the restored table lost the object question")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	which := effectstest.LatestOptionPickFor(g, me.ID)
	if which == nil {
		t.Fatal("the restored table lost the sticker question")
	}
	if err := g.ResolveOptionPick(which.ID, me.ID, effectstest.StkOption(t, which, "“Bamboozle”")); err != nil {
		t.Fatalf("pick: %v", err)
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	where := effectstest.LatestOptionPickFor(g, me.ID)
	if where == nil {
		t.Fatal("the restored table lost the position question")
	}
	if err := g.ResolveOptionPick(where.ID, me.ID, 1); err != nil {
		t.Fatalf("position: %v", err)
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	if got := effectstest.StkCard(t, g, bear).Effective().Name; got != "Bear Bamboozle" {
		t.Errorf("restored name = %q, want %q", got, "Bear Bamboozle")
	}
}

// The two sheets and Ticketomaton are catalogued, and the sheets say what
// their ability stickers give.
func TestStkStickerSheetsRegistered(t *testing.T) {
	for _, oracle := range []string{effectstest.NightBrushwaggOracle, effectstest.MidwayBamboozleOracle, effectstest.TicketomatonOracle} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: registered=%v completeness=%v", oracle, ok, spec.Completeness)
		}
	}
	for _, tc := range []struct {
		sheet   string
		ordinal int
		kw      string
		grant   bool
	}{
		{effectstest.NightBrushwaggOracle, 0, game.CounterMenace, false},
		{effectstest.NightBrushwaggOracle, 1, game.KeywordPersist, false},
		{effectstest.MidwayBamboozleOracle, 0, "", true},
		{effectstest.MidwayBamboozleOracle, 1, game.KeywordUndying, false},
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
