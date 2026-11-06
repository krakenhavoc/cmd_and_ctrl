package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// once_ever_skip_draw_cards_test.go — the two cards of #5245 and #5246,
// each on a seam of its own: Acrobatic Cheerleader's "This ability
// triggers only once" (Duskmourn, #4466; the engine half is
// game/trigger_only_once_test.go) and Living Conundrum's "skip that
// draw" (Murders at Karlov Manor, #4469). Each test plays the real
// catalog entry.

const (
	acrobaticCheerleaderOracle = "d8ce3f18-26fe-47e8-be8b-0d89d7de7038"
	livingConundrumOracle      = "378ae023-04d3-44cc-9248-3d787796ed6c"

	acrobaticCheerleaderTypeLine = "Creature — Human Survivor"
	livingConundrumTypeLine      = "Creature — Elemental"
)

func TestOnceEverSkipDrawCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		acrobaticCheerleaderOracle: "Acrobatic Cheerleader",
		livingConundrumOracle:      "Living Conundrum",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	if spec, _ := Lookup(acrobaticCheerleaderOracle); len(spec.Triggered) != 1 || !spec.Triggered[0].OnlyOnce {
		t.Error("Acrobatic Cheerleader: want one trigger that triggers only once")
	}
}

// The trigger row refuses OnlyOnce where the harvester's record cannot
// reach it: no Key, or a state trigger.
func TestCheckTriggersOnlyOnceRefusesWhatTheRecordCannotKeep(t *testing.T) {
	expectRegisterPanic(t, "without a Key", func() {
		checkTriggersOnlyOnce("Fixture", game.TriggeredAbility{Watches: []game.EventKind{game.EventETB}, OnlyOnce: true})
	})
	expectRegisterPanic(t, "state trigger", func() {
		checkTriggersOnlyOnce("Fixture", game.TriggeredAbility{Key: "k", State: "s", OnlyOnce: true})
	})
	checkTriggersOnlyOnce("Fixture", game.TriggeredAbility{Watches: []game.EventKind{game.EventETB}, Key: "k", OnlyOnce: true})
}

// --- Acrobatic Cheerleader ---------------------------------------------------

// cheerleaderSecondMain taps the Cheerleader in seat 0's precombat main
// phase when asked, runs to that turn's second main phase and resolves
// whatever triggered. It reports whether the ability triggered.
func cheerleaderSecondMain(t *testing.T, g *game.Game, id uuid.UUID, tap bool) bool {
	t.Helper()
	advanceToMainOf(t, g, 0)
	if tap {
		dskB1Tap(g, id)
	}
	advanceTo(t, g, game.StepPostcombatMain)
	triggered := triggerOnStack(g, id) != nil
	passPriorityAroundTable(t, g)
	return triggered
}

// cheerleaderRemoveFlying takes the flying counter off, so a second
// trigger would show as a new counter (the ruling: the ability stays
// spent "even if … the flying counter is somehow removed").
func cheerleaderRemoveFlying(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				delete(g.Battlefield.Cards[i].Counters, game.CounterFlying)
			}
		}
	})
}

// The first time it is tapped as the second main phase begins, it gets a
// flying counter; on a later turn, tapped again, it does not trigger.
func TestAcrobaticCheerleaderTriggersOnlyOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Acrobatic Cheerleader", acrobaticCheerleaderTypeLine, acrobaticCheerleaderOracle, false)

	if !cheerleaderSecondMain(t, g, id, true) {
		t.Fatal("the first Survival trigger did not happen")
	}
	if n := countersOn(g, id, game.CounterFlying); n != 1 {
		t.Fatalf("flying counters = %d, want 1", n)
	}
	if !game.HasKeyword(apaLive(g, id), "flying") {
		t.Error("the flying counter does not give flying")
	}

	cheerleaderRemoveFlying(g, id)
	if cheerleaderSecondMain(t, g, id, true) {
		t.Error("the ability triggered a second time, on a later turn")
	}
	if n := countersOn(g, id, game.CounterFlying); n != 0 {
		t.Errorf("flying counters = %d after the later turn, want 0", n)
	}
}

// An untapped Cheerleader does not trigger at all (the intervening if),
// so it does not use the ability up.
func TestAcrobaticCheerleaderUntappedDoesNotSpendIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Acrobatic Cheerleader", acrobaticCheerleaderTypeLine, acrobaticCheerleaderOracle, false)

	if cheerleaderSecondMain(t, g, id, false) {
		t.Fatal("an untapped Cheerleader triggered")
	}
	if !cheerleaderSecondMain(t, g, id, true) {
		t.Fatal("the first tapped turn did not trigger")
	}
	if n := countersOn(g, id, game.CounterFlying); n != 1 {
		t.Errorf("flying counters = %d, want 1", n)
	}
}

// One that leaves and returns is a new object (CR 400.7), and its
// ability can trigger again.
func TestAcrobaticCheerleaderTriggersAgainAfterAFlicker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Acrobatic Cheerleader", acrobaticCheerleaderTypeLine, acrobaticCheerleaderOracle, false)
	if !cheerleaderSecondMain(t, g, id, true) {
		t.Fatal("the first Survival trigger did not happen")
	}

	flickerForTest(t, g, me.ID, id)
	back := findOnBattlefieldByOracle(g, acrobaticCheerleaderOracle)
	if back == uuid.Nil {
		t.Fatal("the Cheerleader did not come back")
	}
	if n := countersOn(g, back, game.CounterFlying); n != 0 {
		t.Fatalf("the returned Cheerleader kept %d flying counters", n)
	}
	if !cheerleaderSecondMain(t, g, back, true) {
		t.Fatal("the new object did not trigger")
	}
	if n := countersOn(g, back, game.CounterFlying); n != 1 {
		t.Errorf("flying counters = %d, want 1", n)
	}
}

// A restore keeps the ability spent.
func TestAcrobaticCheerleaderStaysSpentAfterARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Acrobatic Cheerleader", acrobaticCheerleaderTypeLine, acrobaticCheerleaderOracle, false)
	if !cheerleaderSecondMain(t, g, id, true) {
		t.Fatal("the first Survival trigger did not happen")
	}
	cheerleaderRemoveFlying(g, id)

	restored := restoreRoundTrip(t, g, true)
	if !restored.TriggeredOnlyOnceSpent(id, "Acrobatic Cheerleader — put a flying counter on it") {
		t.Error("the restored game forgot that the ability triggered")
	}
	if cheerleaderSecondMain(t, restored, id, true) {
		t.Error("after a restore the ability triggered again")
	}
	if n := countersOn(restored, id, game.CounterFlying); n != 0 {
		t.Errorf("flying counters = %d after a restore, want 0", n)
	}
}

// --- Living Conundrum ----------------------------------------------------------

// conundrumPush puts Living Conundrum, printed 2/5, onto the
// battlefield under `owner` and settles the layers once, so the
// library-top watch is armed.
func conundrumPush(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	id := b12Push(g, owner, "Living Conundrum", livingConundrumTypeLine, livingConundrumOracle, 2, 5)
	conundrumView(t, g, id)
	return id
}

// conundrumView brings the layers current, as the next settle point
// would, and returns the permanent.
func conundrumView(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	return layeredCard(t, g, id)
}

// drawsSince counts the EventDrawCard events for `player` after the
// first `from` events of the log.
func drawsSince(g *game.Game, from int, player uuid.UUID) int {
	n := 0
	for _, ev := range g.Events[from:] {
		if ev.Kind == game.EventDrawCard && ev.Actor == player {
			n++
		}
	}
	return n
}

// From an empty library the draw is skipped: no card, no draw event (so
// no "whenever you draw" trigger), no draw counted this turn, and no loss
// for drawing from an empty library.
func TestLivingConundrumSkipsADrawFromAnEmptyLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	conundrumPush(t, g, me.ID)
	emptyLibrary(me)
	hand, logAt := me.Hand.Size(), len(g.Events)
	drawn := g.TurnTallyFor(me.ID).CardsDrawn

	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })

	if me.AttemptedEmptyDraw {
		t.Error("the skipped draw set the empty-library flag")
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d, want no card", hand, me.Hand.Size())
	}
	if n := drawsSince(g, logAt, me.ID); n != 0 {
		t.Errorf("%d draw events, want 0", n)
	}
	if got := g.TurnTallyFor(me.ID).CardsDrawn; got != drawn {
		t.Errorf("cards drawn this turn %d → %d, want unchanged", drawn, got)
	}
	if len(g.CardsDrawnThisTurnFor(me.ID)) != 0 {
		t.Error("the skipped draw is listed as a card drawn this turn")
	}
	checkState(t, g)
	if me.Eliminated {
		t.Error("lost for a draw Living Conundrum skipped")
	}
}

// With cards in the library the draw is a draw, and another player's
// draw from an empty library is not this card's to skip.
func TestLivingConundrumLeavesOtherDrawsAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	conundrumPush(t, g, me.ID)
	hand, logAt := me.Hand.Size(), len(g.Events)

	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	if me.Hand.Size() != hand+1 || drawsSince(g, logAt, me.ID) != 1 {
		t.Errorf("a draw from a library with cards: hand %d → %d, %d draw events; want one card and one event",
			hand, me.Hand.Size(), drawsSince(g, logAt, me.ID))
	}

	emptyLibrary(opp)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	if !opp.AttemptedEmptyDraw {
		t.Error("an opponent's draw from an empty library was skipped")
	}
}

// "Draw three" off a two-card library draws two and skips the third
// (the 2024-02-02 ruling): each draw is its own event (CR 121.2).
func TestLivingConundrumSkipsOnlyTheDrawsTheLibraryCannotCover(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	conundrumPush(t, g, me.ID)
	for me.Library.Size() > 2 {
		_, _ = me.Library.PopTop()
	}
	hand, logAt := me.Hand.Size(), len(g.Events)

	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 3) })

	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d → %d, want two cards", hand, me.Hand.Size())
	}
	if n := drawsSince(g, logAt, me.ID); n != 2 {
		t.Errorf("%d draw events, want 2", n)
	}
	if me.AttemptedEmptyDraw {
		t.Error("the third, skipped draw set the empty-library flag")
	}
}

// With an empty library it is a 10/10 with flying and vigilance; with
// cards in it, a 2/5 without them. The draw that empties the library
// turns it on.
func TestLivingConundrumIsATenTenFlierWhileTheLibraryIsEmpty(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	id := conundrumPush(t, g, me.ID)

	c := conundrumView(t, g, id)
	if c.CurrentPower() != 2 || c.CurrentToughness() != 5 || game.HasKeyword(&c, "flying") || game.HasKeyword(&c, "vigilance") {
		t.Errorf("with a library: %d/%d flying=%v vigilance=%v, want a 2/5 without them",
			c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "flying"), game.HasKeyword(&c, "vigilance"))
	}
	if !game.HasKeyword(&c, "hexproof") {
		t.Error("no hexproof")
	}

	for me.Library.Size() > 1 {
		_, _ = me.Library.PopTop()
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	c = conundrumView(t, g, id)
	if c.CurrentPower() != 10 || c.CurrentToughness() != 10 || !game.HasKeyword(&c, "flying") || !game.HasKeyword(&c, "vigilance") {
		t.Errorf("with an empty library: %d/%d flying=%v vigilance=%v, want a 10/10 with both",
			c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "flying"), game.HasKeyword(&c, "vigilance"))
	}

	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Put Back", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	if c := conundrumView(t, g, id); c.CurrentPower() != 2 || game.HasKeyword(&c, "flying") {
		t.Errorf("a card back in the library: %d power, flying=%v, want 2 and none", c.CurrentPower(), game.HasKeyword(&c, "flying"))
	}
}

// A restore keeps both halves: the draw is still skipped and the
// creature is still a 10/10.
func TestLivingConundrumSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	id := conundrumPush(t, g, me.ID)
	emptyLibrary(me)
	conundrumView(t, g, id)

	restored := restoreRoundTrip(t, g, true)
	rme := restored.Seats[1]
	hand, logAt := rme.Hand.Size(), len(restored.Events)
	restored.WithWriteLock(func() { _ = restored.DrawNForEffect(rme.ID, 1) })
	if rme.AttemptedEmptyDraw || rme.Hand.Size() != hand || drawsSince(restored, logAt, rme.ID) != 0 {
		t.Errorf("after a restore the draw was not skipped: flag=%v hand %d → %d", rme.AttemptedEmptyDraw, hand, rme.Hand.Size())
	}
	if c := conundrumView(t, restored, id); c.CurrentPower() != 10 || c.CurrentToughness() != 10 || !game.HasKeyword(&c, "flying") {
		t.Errorf("after a restore: %d/%d flying=%v, want a 10/10 flier", c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "flying"))
	}
}
