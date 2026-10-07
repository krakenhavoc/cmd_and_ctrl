package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// tinybones_bauble_burglar_test.go — #2179: a standing permission over
// exiled cards you don't own that carry a stash counter.

const tinybonesOracle = "fd335b2f-e6a2-45ac-949e-a67278b21cb6"

// tinybonesTable is a main phase on seat 0's turn with Tinybones on the
// battlefield under seat 0.
func tinybonesTable(t *testing.T) (g *game.Game, me, opp *game.Player, tb uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	advanceToMain(t, g)
	me, opp = g.Seats[0], g.Seats[1]
	tb = pushCatalogPermanent(g, me.ID, "Tinybones, Bauble Burglar", "Legendary Creature — Skeleton Rogue", tinybonesOracle, false)
	return
}

// stashed puts a card in exile owned by `owner`, carrying `n` stash
// counters, known to every seat.
func stashed(g *game.Game, owner uuid.UUID, name, typeLine, cost string, n int) uuid.UUID {
	id := uuid.New()
	known := map[uuid.UUID]bool{}
	for _, p := range g.Seats {
		known[p.ID] = true
	}
	c := game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Owner: owner, Controller: owner, KnownBy: known}
	if n > 0 {
		c.Counters = map[string]int{"stash": n}
	}
	g.Exile.PushTop(c)
	return id
}

func exileCard(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}

func castStash(g *game.Game, who *game.Player, id uuid.UUID) error {
	return g.CastSpell(who.ID, id, game.CastSpellParams{FromZone: "exile", Strict: true})
}

func payAnyType(t *testing.T, g *game.Game, who *game.Player, mana string) {
	t.Helper()
	if err := g.AddManaForEffect(who.ID, uuid.Nil, mana); err != nil {
		t.Fatal(err)
	}
}

func enumeratedFor(g *game.Game, seat, card uuid.UUID) bool {
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type == legal.TypeCastSpell && m.Source == card {
			return true
		}
	}
	return false
}

func viewCastable(t *testing.T, g *game.Game, viewer, card uuid.UUID) bool {
	t.Helper()
	v := protocol.ViewOfGameFor(g, viewer.String())
	for _, c := range v.Exile.Cards {
		if c.InstanceID == card.String() {
			return c.CastableHere
		}
	}
	t.Fatalf("card %s is not in the exile view", card)
	return false
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
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) || opp.Graveyard.Contains(bear) {
		t.Fatal("the discarded card was not exiled from the graveyard")
	}
	c, _ := exileCard(g, bear)
	if c.Counters["stash"] != 1 {
		t.Fatalf("stash counters = %v, want 1", c.Counters)
	}
	// Resolving the trigger walked the turn on; the cast needs the main phase back.
	g.Turn.Step = game.StepPrecombatMain
	payAnyType(t, g, me, "{R}{U}")
	if err := castStash(g, me, bear); err != nil {
		t.Fatalf("cast the stashed creature with any-type mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	got, ok := battlefieldCard(g, bear)
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
	mine := stashed(g, me.ID, "My Own Bear", "Creature — Bear", "{1}{G}", 1)
	theirs := stashed(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}", 1)
	payAnyType(t, g, me, "{R}{U}")
	if err := castStash(g, me, mine); err == nil {
		t.Fatal("a card you own must not be castable through Tinybones")
	}
	if enumeratedFor(g, me.ID, mine) || viewCastable(t, g, me.ID, mine) {
		t.Error("enumerator or view offers a card the holder owns")
	}
	if !enumeratedFor(g, me.ID, theirs) || !viewCastable(t, g, me.ID, theirs) {
		t.Fatal("non-vacuity: the enumerator and the view must offer the opponent's stashed card")
	}
	if err := castStash(g, me, theirs); err != nil {
		t.Fatalf("their stashed card: %v", err)
	}
}

// Only cards carrying the counter qualify.
func TestTinybonesNeedsTheStashCounter(t *testing.T) {
	g, me, opp, _ := tinybonesTable(t)
	plain := stashed(g, opp.ID, "Plain Exile", "Creature — Bear", "{1}{G}", 0)
	payAnyType(t, g, me, "{R}{U}")
	if err := castStash(g, me, plain); err == nil {
		t.Fatal("a card with no stash counter must not be castable")
	}
	if enumeratedFor(g, me.ID, plain) || viewCastable(t, g, me.ID, plain) {
		t.Error("enumerator or view offers a card with no stash counter")
	}
}

// The permission lives on the battlefield: once Tinybones leaves, the
// stash is stranded, in the engine, the enumerator and the view.
func TestTinybonesLeavingEndsThePermission(t *testing.T) {
	g, me, opp, tb := tinybonesTable(t)
	id := stashed(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}", 1)
	payAnyType(t, g, me, "{R}{U}")
	if !enumeratedFor(g, me.ID, id) {
		t.Fatal("non-vacuity: offered while Tinybones is in play")
	}
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(tb); err != nil {
			t.Fatal(err)
		}
	})
	if err := castStash(g, me, id); err == nil {
		t.Fatal("cast succeeded after Tinybones left")
	}
	if enumeratedFor(g, me.ID, id) || viewCastable(t, g, me.ID, id) {
		t.Error("enumerator or view still offers the card after Tinybones left")
	}
}

// "During your turn": off the holder's own turn nothing is castable,
// not even an instant.
func TestTinybonesOnlyWorksDuringYourTurn(t *testing.T) {
	g, me, opp, _ := tinybonesTable(t)
	id := stashed(g, opp.ID, "Their Bolt", "Instant", "{R}", 1)
	payAnyType(t, g, me, "{U}")
	if !enumeratedFor(g, me.ID, id) {
		t.Fatal("non-vacuity: an instant is offered on your own turn")
	}
	g.Turn.ActiveSeat = 1
	if err := castStash(g, me, id); err == nil {
		t.Fatal("cast succeeded off the holder's turn")
	}
	if enumeratedFor(g, me.ID, id) || viewCastable(t, g, me.ID, id) {
		t.Error("enumerator or view offers a cast off the holder's turn")
	}
}

// The card's own timing still applies: a creature is not cast at
// instant speed on your own turn.
func TestTinybonesLeavesTheCardsOwnTimingInForce(t *testing.T) {
	g, me, opp, _ := tinybonesTable(t)
	id := stashed(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}", 1)
	payAnyType(t, g, me, "{R}{U}")
	for g.Turn.Step != game.StepUpkeep {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if err := castStash(g, me, id); err == nil {
		t.Fatal("a creature must not be cast outside a main phase")
	}
	if enumeratedFor(g, me.ID, id) {
		t.Error("the enumerator offers a creature outside a main phase")
	}
}

// "Play": a stashed land is played from exile.
func TestTinybonesPlaysAStashedLand(t *testing.T) {
	g, me, opp, _ := tinybonesTable(t)
	land := stashed(g, opp.ID, "Their Island", "Basic Land — Island", "", 1)
	if !enumeratedFor(g, me.ID, land) {
		t.Fatal("the enumerator should offer the stashed land")
	}
	if err := castStash(g, me, land); err != nil {
		t.Fatalf("play the stashed land: %v", err)
	}
	if _, ok := battlefieldCard(g, land); !ok {
		t.Fatal("the land is not on the battlefield")
	}
}

// Two Tinybones at one table: each holder sees the other's stash but
// never their own cards.
func TestTinybonesAnotherHoldersStashIsFairGame(t *testing.T) {
	g, me, opp, _ := tinybonesTable(t)
	pushCatalogPermanent(g, opp.ID, "Tinybones, Bauble Burglar", "Legendary Creature — Skeleton Rogue", tinybonesOracle, false)
	mineInExile := stashed(g, me.ID, "My Own Bear", "Creature — Bear", "{1}{G}", 1)
	if viewCastable(t, g, me.ID, mineInExile) {
		t.Error("the holder is offered their own card")
	}
	g.Turn.ActiveSeat = 1
	if !viewCastable(t, g, opp.ID, mineInExile) {
		t.Error("the other Tinybones' holder is not offered a card they don't own")
	}
}

// The grant is derived from the battlefield and the marker is a
// counter, so a snapshot round trip keeps both.
func TestTinybonesSurvivesASnapshotRoundTrip(t *testing.T) {
	g, me, opp, _ := tinybonesTable(t)
	id := stashed(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}", 1)
	payAnyType(t, g, me, "{R}{U}")
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
	if c, ok := exileCard(r, id); !ok || c.Counters["stash"] != 1 {
		t.Fatalf("restored stash counter = %v (found %v)", c.Counters, ok)
	}
	if err := castStash(r, r.Seats[0], id); err != nil {
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
	payAnyType(t, g, me, "{B}{B}{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, tb, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		if discardOwed(g, p.ID) != 1 {
			t.Fatalf("seat %v owes %d discards, want 1", p.ID, discardOwed(g, p.ID))
		}
	}
	if discardOwed(g, me.ID) != 0 {
		t.Error("the controller does not discard")
	}
	for _, p := range g.Seats[1:] {
		discardFromHand(t, g, p.ID)
	}
	passPriorityAroundTable(t, g)
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
	payAnyType(t, g, me, "{B}{B}{B}{B}")
	g.Turn.ActiveSeat = 1
	if err := g.ActivateCatalogAbility(me.ID, tb, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the ability must be sorcery speed")
	}
}
