package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_p2c_restore_test.go — #4253's closure sweep, second pass, group C: the
// prompts that carried a Go closure are keyed now, so a table waiting on any of
// them is a restore point. Each test opens the prompt, takes a STRICT restore
// through JSON (restoreRoundTrip with strict, which fails on a table that is
// not a restore point) and answers on the restored table.

// sweepP2COptionKey returns the open option pick of `chooser`, failing the test
// when it names a key other than `want`.
func sweepP2COptionKey(t *testing.T, g *game.Game, chooser uuid.UUID, want string) *game.PendingChoice {
	t.Helper()
	c := latestOptionPickFor(g, chooser)
	if c == nil {
		t.Fatalf("no option pick for %s: %+v", chooser, g.PendingChoices)
	}
	if c.PickThen != want {
		t.Fatalf("the option pick names key %q, want %q", c.PickThen, want)
	}
	return c
}

// sweepP2CDiscardKey returns the open discard prompt of `player`, failing the
// test when it names a key other than `want`.
func sweepP2CDiscardKey(t *testing.T, g *game.Game, player uuid.UUID, want string) *game.PendingChoice {
	t.Helper()
	c := discardChoiceFor(g, player)
	if c == nil {
		t.Fatalf("no discard prompt for %s: %+v", player, g.PendingChoices)
	}
	if c.PickThen != want {
		t.Fatalf("the discard prompt names key %q, want %q", c.PickThen, want)
	}
	return c
}

// Read the Runes: every question of the chain (the option pick, the discard and
// the sacrifice behind it) is a restore point, and the count of questions still
// owed survives the restore.
func TestSweepP2CReadTheRunesChainIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	token := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	castReadTheRunes(t, g, me.ID, 2)
	passPriorityAroundTable(t, g)
	sweepP2COptionKey(t, g, me.ID, "option-pick/read-the-runes-answer")

	// Repetition one: the discard.
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	sweepP2COptionKey(t, g, me.ID, "option-pick/read-the-runes-answer")
	answerOptionPick(t, g, me.ID, 0)
	sweepP2CDiscardKey(t, g, me.ID, "choose-cards/read-the-runes-discarded")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	sweepP2CDiscardKey(t, g, me.ID, "choose-cards/read-the-runes-discarded")
	discardFromHand(t, g, me.ID)

	// Repetition two: the sacrifice, which is the last question.
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	sweepP2COptionKey(t, g, me.ID, "option-pick/read-the-runes-answer")
	answerOptionPick(t, g, me.ID, 1)
	sweep00PickKey(t, g, me.ID, "choose-cards/sacrifice-choice")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := sweep00PickKey(t, g, me.ID, "choose-cards/sacrifice-choice")
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{token}); err != nil {
		t.Fatalf("ResolveChooseCards (sacrifice) on the restored table: %v", err)
	}
	if g.Battlefield.Contains(token) {
		t.Error("the chosen permanent was not sacrificed on the restored table")
	}
	if latestOptionPickFor(g, me.ID) != nil {
		t.Error("a third question was asked for X = 2 after the restores")
	}
}

// Mind Extraction: the discard is a keyed prompt whose set rule still refuses a
// card that shares no color with the sacrificed creature after a restore.
func TestSweepP2CMindExtractionDiscardIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Hand.Cards = nil
	red := uuid.New()
	gold := uuid.New()
	blue := uuid.New()
	for _, c := range []game.Card{
		{InstanceID: red, Name: "Red", TypeLine: "Instant", ManaCost: "{R}"},
		{InstanceID: gold, Name: "Gold", TypeLine: "Instant", ManaCost: "{U}{G}"},
		{InstanceID: blue, Name: "Blue", TypeLine: "Instant", ManaCost: "{U}"},
	} {
		c.Owner, c.Controller = opp.ID, opp.ID
		opp.Hand.PushTop(c)
	}
	paCast(t, g, "Mind Extraction", "Sorcery", "{2}{B}", mindExtractionOracle, game.CastSpellParams{
		Targets: soPlayer(opp.ID), SacrificeIDs: []uuid.UUID{paPermanent(g, me.ID, "Druid", "Creature — Elf", "{R}{G}", 1, 1)},
	})
	passPriorityAroundTable(t, g)
	const key = "choose-cards/mind-extraction-discard-rg"
	sweepP2CDiscardKey(t, g, opp.ID, key)

	g = restoreRoundTrip(t, g, true)
	opp = g.Seats[1]
	c := sweepP2CDiscardKey(t, g, opp.ID, key)
	if c.ChooseMin != 2 || c.ChooseMax != 2 {
		t.Fatalf("discard prompt = %+v, want exactly two", c)
	}
	if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{red, blue}); err == nil {
		t.Fatal("a blue card was accepted for a red and green creature on the restored table")
	}
	if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{red, gold}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	if !opp.Hand.Contains(blue) || opp.Hand.Contains(red) || opp.Hand.Contains(gold) {
		t.Error("the restored hand is not left with only the blue card")
	}
}

// Seize the Spotlight: each opponent's fame-or-fortune answer carries the fame
// seats, the seats still to ask and the fortunes so far as data, so the walk
// goes on across restores and pays out every fortune at the end.
func TestSweepP2CSeizeTheSpotlightWalkIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 4; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	hand := me.Hand.Size()
	g.WithWriteLock(func() {
		item := &game.StackItem{ID: uuid.New(), Kind: game.StackItemSpell, Controller: me.ID, Owner: me.ID}
		ctx := NewContext(g, item)
		if err := seizeTheSpotlightAsk(ctx, ctx.Opponents(), nil, 0); err != nil {
			t.Fatal(err)
		}
	})
	const key = "option-pick/seize-the-spotlight-answer"
	for _, who := range []int{1, 2, 3} {
		g = restoreRoundTrip(t, g, true)
		me = g.Seats[0]
		seat := g.Seats[who]
		sweepP2COptionKey(t, g, seat.ID, key)
		answerOptionPick(t, g, seat.ID, 1)
	}
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d → %d, want +3: one card for each of the three fortunes", hand, got)
	}
	treasures := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Treasure" {
			treasures++
		}
	}
	if treasures != 3 {
		t.Errorf("Treasures = %d, want 3", treasures)
	}
}

// Valakut Awakening: the order prompt for two or more put-back cards has the
// draw as a keyed tail, so the restored answer draws the cards.
func TestSweepP2CValakutAwakeningOrderIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Library One", "Library Two", "Library Three")
	excess1 := pushHandCardWithManaCost(me, "Excess One", "Instant", "{1}")
	excess2 := pushHandCardWithManaCost(me, "Excess Two", "Instant", "{1}")
	handBefore := me.Hand.Size()

	castCatalogSpell(t, g, "Valakut Awakening", "Instant", valakutAwakeningOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no choose_cards prompt over the hand")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{excess1, excess2}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID) == nil {
		t.Fatal("no put_in_library prompt for the bottom order")
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	order := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID)
	if order == nil {
		t.Fatal("the order prompt did not survive the restore")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{excess1, excess2}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary on the restored table: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore-2+3 {
		t.Errorf("hand size %d, want %d (put back 2, draw 3)", got, handBefore-2+3)
	}
}

// Rootborn Defenses: the grant is the populate's keyed tail, so a table waiting
// on the populate's pick restores and still protects both Centaurs.
func TestSweepP2CRootbornDefensesGrantIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))
	centaur := pushToken(g, me.ID, TokenCard("3/3 green Centaur"))

	castCatalogSpell(t, g, "Rootborn Defenses", "Instant", rootbornDefensesOracle, nil)
	passPriorityAroundTable(t, g)
	sweep00PickKey(t, g, me.ID, "choose-cards/populate")

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	answerChooseCards(t, g, me.ID, centaur)
	passPriorityAroundTable(t, g)
	centaurs := populateTokens(g, me.ID, "Centaur")
	if len(centaurs) != 2 {
		t.Fatalf("Centaurs = %d, want 2", len(centaurs))
	}
	for _, id := range centaurs {
		if !effectiveAbilitiesContain(t, g, id, "indestructible") {
			t.Errorf("Centaur %s missed the grant on the restored table", id)
		}
	}
}

// Torment of Hailfire (and the Rottenmouth Viper it shares a shape with): the
// sacrifice branch's pick carries the rest of the queue as a keyed tail, so the
// next opponent is asked after a restore.
func TestSweepP2CPunisherSacrificeIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	castTormentOfHailfire(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, opp.ID, 1)
	sweep00PickKey(t, g, opp.ID, "choose-cards/sacrifice-choice")

	g = restoreRoundTrip(t, g, true)
	c := sweep00PickKey(t, g, opp.ID, "choose-cards/sacrifice-choice")
	if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the chosen permanent was not sacrificed on the restored table")
	}
	next := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	if latestOptionPickFor(g, next.ID) == nil {
		t.Errorf("the next opponent was not asked after the restore: %+v", g.PendingChoices)
	}
	drainRemainingTormentPrompts(t, g)
}

// Wishclaw Talisman: the opponent to hand it to is a keyed player choice, and
// the Talisman changes hands on the restored table.
func TestSweepP2CWishclawHandOverIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tal := castCatalogSpell(t, g, "Wishclaw Talisman", "Artifact", sweepFWishclawOracle, nil)
	passPriorityAroundTable(t, g)
	needle := pushLibraryCardForTest(me, game.Card{Name: "Needle", TypeLine: "Sorcery"})
	sweepFAddMana(g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, tal, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, needle)
	sweepP2COptionKey(t, g, me.ID, "option-pick/wishclaw-talisman-hand-it-over")

	g = restoreRoundTrip(t, g, true)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	sweepP2COptionKey(t, g, me.ID, "option-pick/wishclaw-talisman-hand-it-over")
	answerChoosePlayer(t, g, me.ID, opp)
	if c := layeredCard(t, g, tal); c.Controller != opp.ID {
		t.Errorf("Talisman controller = %v, want the chosen opponent on the restored table", c.Controller)
	}
}

// Mabel, Bitter Recluse: "remove up to three counters" asks again after each
// removal; the permanent, the question and the counters left ride the prompt.
func TestSweepP2CRemoveUpToCountersIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	target := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Counter Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(target, "time", 3); err != nil {
			t.Fatal(err)
		}
		if err := rfRemoveUpToCounters(g, me.ID, uuid.Nil, target, 2, "Remove a counter (up to two)"); err != nil {
			t.Fatal(err)
		}
	})
	const key = "option-pick/reality-fracture-remove-up-to-counters"
	sweepP2COptionKey(t, g, me.ID, key)

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	sweepP2COptionKey(t, g, me.ID, key)
	answerOptionPick(t, g, me.ID, 1)
	if got := counterCount(g, target, "time"); got != 2 {
		t.Fatalf("time counters = %d, want 2 after the first removal", got)
	}

	// The second question is asked from the restored answer, and is a restore
	// point too.
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	sweepP2COptionKey(t, g, me.ID, key)
	answerOptionPick(t, g, me.ID, 1)
	if got := counterCount(g, target, "time"); got != 1 {
		t.Errorf("time counters = %d, want 1 after the second removal", got)
	}
	if latestOptionPickFor(g, me.ID) != nil {
		t.Error("a third question was asked after \"up to two\"")
	}
}

// Something Worth Saving: the pick over the milled cards is a keyed pick that
// reads the asking item as data, so the life follows the pick on the restored
// table.
func TestSweepP2CSomethingWorthSavingPickIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	ids := rfbCastSomethingWorthSaving(t, g,
		mm2Spell("Spell A"), game.Card{Name: "Prize", TypeLine: "Artifact"}, mm2Spell("Spell B"), mm2Spell("Spell C"), mm2Spell("Deep"))
	const key = "choose-cards/mill-then-take-something-worth-saving-pick"
	sweep00PickKey(t, g, me.ID, key)

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := sweep00PickKey(t, g, me.ID, key)
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Error("the permanent card is not in hand on the restored table")
	}
	if me.Life != life+1 {
		t.Errorf("life = %d, want %d on the restored table", me.Life, life+1)
	}
}
