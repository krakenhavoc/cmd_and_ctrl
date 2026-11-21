package testst

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tinybonesTable is a main phase on seat 0's turn with Tinybones on the
// battlefield under seat 0.
func tinybonesTable(t *testing.T) (g *game.Game, me, opp *game.Player, tb uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp = g.Seats[0], g.Seats[1]
	tb = effectstest.PushCatalogPermanent(g, me.ID, "Tinybones, Bauble Burglar", "Legendary Creature — Skeleton Rogue", effectstest.TinybonesOracle, false)
	return
}

// An opponent's discard is exiled from their graveyard with a stash
// counter, and then cast from exile with mana of any type.
func TestTinybonesStashesADiscardAndYouCastIt(t *testing.T) {
	g, me, opp, _ := tinybonesTable(t)
	opp.Hand.Cards = nil
	bear := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bear, Name: "Stashed Bear", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(opp.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) || opp.Graveyard.Contains(bear) {
		t.Fatal("the discarded card was not exiled from the graveyard")
	}
	c, _ := effectstest.ExileCard(g, bear)
	if c.Counters["stash"] != 1 {
		t.Fatalf("stash counters = %v, want 1", c.Counters)
	}
	// Resolving the trigger walked the turn on; the cast needs the main phase back.
	g.Turn.Step = game.StepPrecombatMain
	effectstest.PayAnyType(t, g, me, "{R}{U}")
	if err := effectstest.CastStash(g, me, bear); err != nil {
		t.Fatalf("cast the stashed creature with any-type mana: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	got, ok := effectstest.BattlefieldCard(g, bear)
	if !ok {
		t.Fatal("the stashed creature did not resolve onto the battlefield")
	}
	if got.Counters["stash"] != 0 {
		t.Errorf("the stash counter followed the card out of exile: %v", got.Counters)
	}
	if got.Controller != me.ID {
		t.Error("the cast creature should be under the caster's control")
	}
}

// Your own discards are not "cards you don't own": the opponent's
// Tinybones-stashed cards open for them, yours do not open for you.
func TestTinybonesRefusesYourOwnStashedCard(t *testing.T) {
	g, me, opp, _ := tinybonesTable(t)
	mine := effectstest.Stashed(g, me.ID, "My Own Bear", "Creature — Bear", "{1}{G}", 1)
	theirs := effectstest.Stashed(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}", 1)
	effectstest.PayAnyType(t, g, me, "{R}{U}")
	if err := effectstest.CastStash(g, me, mine); err == nil {
		t.Fatal("a card you own must not be castable through Tinybones")
	}
	if effectstest.EnumeratedFor(g, me.ID, mine) || effectstest.ViewCastable(t, g, me.ID, mine) {
		t.Error("enumerator or view offers a card the holder owns")
	}
	if !effectstest.EnumeratedFor(g, me.ID, theirs) || !effectstest.ViewCastable(t, g, me.ID, theirs) {
		t.Fatal("non-vacuity: the enumerator and the view must offer the opponent's stashed card")
	}
	if err := effectstest.CastStash(g, me, theirs); err != nil {
		t.Fatalf("their stashed card: %v", err)
	}
}

// Only cards carrying the counter qualify.
func TestTinybonesNeedsTheStashCounter(t *testing.T) {
	g, me, opp, _ := tinybonesTable(t)
	plain := effectstest.Stashed(g, opp.ID, "Plain Exile", "Creature — Bear", "{1}{G}", 0)
	effectstest.PayAnyType(t, g, me, "{R}{U}")
	if err := effectstest.CastStash(g, me, plain); err == nil {
		t.Fatal("a card with no stash counter must not be castable")
	}
	if effectstest.EnumeratedFor(g, me.ID, plain) || effectstest.ViewCastable(t, g, me.ID, plain) {
		t.Error("enumerator or view offers a card with no stash counter")
	}
}

// The permission lives on the battlefield: once Tinybones leaves, the
// stash is stranded, in the engine, the enumerator and the view.
func TestTinybonesLeavingEndsThePermission(t *testing.T) {
	g, me, opp, tb := tinybonesTable(t)
	id := effectstest.Stashed(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}", 1)
	effectstest.PayAnyType(t, g, me, "{R}{U}")
	if !effectstest.EnumeratedFor(g, me.ID, id) {
		t.Fatal("non-vacuity: offered while Tinybones is in play")
	}
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(tb); err != nil {
			t.Fatal(err)
		}
	})
	if err := effectstest.CastStash(g, me, id); err == nil {
		t.Fatal("cast succeeded after Tinybones left")
	}
	if effectstest.EnumeratedFor(g, me.ID, id) || effectstest.ViewCastable(t, g, me.ID, id) {
		t.Error("enumerator or view still offers the card after Tinybones left")
	}
}

// "During your turn": off the holder's own turn nothing is castable,
// not even an instant.
func TestTinybonesOnlyWorksDuringYourTurn(t *testing.T) {
	g, me, opp, _ := tinybonesTable(t)
	id := effectstest.Stashed(g, opp.ID, "Their Bolt", "Instant", "{R}", 1)
	effectstest.PayAnyType(t, g, me, "{U}")
	if !effectstest.EnumeratedFor(g, me.ID, id) {
		t.Fatal("non-vacuity: an instant is offered on your own turn")
	}
	g.Turn.ActiveSeat = 1
	if err := effectstest.CastStash(g, me, id); err == nil {
		t.Fatal("cast succeeded off the holder's turn")
	}
	if effectstest.EnumeratedFor(g, me.ID, id) || effectstest.ViewCastable(t, g, me.ID, id) {
		t.Error("enumerator or view offers a cast off the holder's turn")
	}
}

// The card's own timing still applies: a creature is not cast at
// instant speed on your own turn.
func TestTinybonesLeavesTheCardsOwnTimingInForce(t *testing.T) {
	g, me, opp, _ := tinybonesTable(t)
	id := effectstest.Stashed(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}", 1)
	effectstest.PayAnyType(t, g, me, "{R}{U}")
	for g.Turn.Step != game.StepUpkeep {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if err := effectstest.CastStash(g, me, id); err == nil {
		t.Fatal("a creature must not be cast outside a main phase")
	}
	if effectstest.EnumeratedFor(g, me.ID, id) {
		t.Error("the enumerator offers a creature outside a main phase")
	}
}

// "Play": a stashed land is played from exile.
func TestTinybonesPlaysAStashedLand(t *testing.T) {
	g, me, opp, _ := tinybonesTable(t)
	land := effectstest.Stashed(g, opp.ID, "Their Island", "Basic Land — Island", "", 1)
	if !effectstest.EnumeratedFor(g, me.ID, land) {
		t.Fatal("the enumerator should offer the stashed land")
	}
	if err := effectstest.CastStash(g, me, land); err != nil {
		t.Fatalf("play the stashed land: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, land); !ok {
		t.Fatal("the land is not on the battlefield")
	}
}

// Two Tinybones at one table: each holder sees the other's stash but
// never their own cards.
func TestTinybonesAnotherHoldersStashIsFairGame(t *testing.T) {
	g, me, opp, _ := tinybonesTable(t)
	effectstest.PushCatalogPermanent(g, opp.ID, "Tinybones, Bauble Burglar", "Legendary Creature — Skeleton Rogue", effectstest.TinybonesOracle, false)
	mineInExile := effectstest.Stashed(g, me.ID, "My Own Bear", "Creature — Bear", "{1}{G}", 1)
	if effectstest.ViewCastable(t, g, me.ID, mineInExile) {
		t.Error("the holder is offered their own card")
	}
	g.Turn.ActiveSeat = 1
	if !effectstest.ViewCastable(t, g, opp.ID, mineInExile) {
		t.Error("the other Tinybones' holder is not offered a card they don't own")
	}
}

// The grant is derived from the battlefield and the marker is a
// counter, so a snapshot round trip keeps both.
func TestTinybonesSurvivesASnapshotRoundTrip(t *testing.T) {
	g, me, opp, _ := tinybonesTable(t)
	id := effectstest.Stashed(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}", 1)
	effectstest.PayAnyType(t, g, me, "{R}{U}")
	blob, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap game.GameSnapshot
	if err := json.Unmarshal(blob, &snap); err != nil {
		t.Fatal(err)
	}
	r, err := snap.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if c, ok := effectstest.ExileCard(r, id); !ok || c.Counters["stash"] != 1 {
		t.Fatalf("restored stash counter = %v (found %v)", c.Counters, ok)
	}
	if err := effectstest.CastStash(r, r.Seats[0], id); err != nil {
		t.Fatalf("restored: cast the stashed card: %v", err)
	}
}

// A stored permission carrying the new filter fields round-trips.
func TestPermissionFilterNewFieldsRoundTripJSON(t *testing.T) {
	in := game.PermissionFilter{NotOwnedByHolder: true, WithCounter: "stash"}
	blob, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out game.PermissionFilter
	if err := json.Unmarshal(blob, &out); err != nil {
		t.Fatal(err)
	}
	if out != in {
		t.Errorf("round trip = %+v, want %+v", out, in)
	}
}

// {3}{B}, {T}: each opponent discards a card, at sorcery speed, and
// every card they discard is stashed.
func TestTinybonesActivatedAbilityMakesEachOpponentDiscardAndStashes(t *testing.T) {
	g, me, _, tb := tinybonesTable(t)
	effectstest.PayAnyType(t, g, me, "{B}{B}{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, tb, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		if effectstest.DiscardOwed(g, p.ID) != 1 {
			t.Fatalf("seat %v owes %d discards, want 1", p.ID, effectstest.DiscardOwed(g, p.ID))
		}
	}
	if effectstest.DiscardOwed(g, me.ID) != 0 {
		t.Error("the controller does not discard")
	}
	for _, p := range g.Seats[1:] {
		effectstest.DiscardFromHand(t, g, p.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	stash := 0
	for _, c := range g.Exile.Cards {
		if c.Counters["stash"] == 1 {
			stash++
		}
	}
	if stash != len(g.Seats)-1 {
		t.Errorf("%d stashed cards in exile, want one per opponent", stash)
	}
}

func TestTinybonesAbilityIsSorcerySpeed(t *testing.T) {
	g, me, _, tb := tinybonesTable(t)
	effectstest.PayAnyType(t, g, me, "{B}{B}{B}{B}")
	g.Turn.ActiveSeat = 1
	if err := g.ActivateCatalogAbility(me.ID, tb, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the ability must be sorcery speed")
	}
}
