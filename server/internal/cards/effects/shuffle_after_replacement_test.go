package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shuffle_after_replacement_test.go — #1735 (the Blightsteel Colossus
// item of the slice 296-e retriage). ADR 0013 §5ah gives the
// replacement pipeline a post-move shuffle hook: a replacement can
// mark its redirected library destination as a genuine shuffle-in
// (game.ReplacementEvent.ShuffleDestinationLibrary), consumed once the
// card has actually landed, by executeZoneRouteLocked and
// executeBattlefieldLeaveLocked — the two functions that perform
// every replaced move's physical landing, regardless of which route
// (die, discard, mill, a hand-rolled sandbox move) opened the window.
//
// Before this fix Blightsteel Colossus landed on TOP of its owner's
// library, a real information leak (everyone at the table knows the
// top card) that the printed "reveal it" does not license — the card
// reveals WHICH card, never WHERE it ends up. These tests assert the
// shuffle actually ran, not just that the card is somewhere in the
// library: a shuffle that never fired but still passed the
// `Contains` check is exactly the bug this closes.

// pushBlightsteelOnBattlefield seeds Blightsteel Colossus onto the
// battlefield under `me`'s control, ready to leave it.
func pushBlightsteelOnBattlefield(g *game.Game, me *game.Player) uuid.UUID {
	c := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blightsteel Colossus", TypeLine: "Artifact Creature — Phyrexian Golem",
		OracleID: blightsteelOracle, Power: 11, Toughness: 11, Owner: me.ID, Controller: me.ID,
	})
	return c
}

// blightsteelShuffleCount counts the "shuffle" EventSearchLibrary
// events stamped for `player` in `events` — the same shape
// ShuffleLibraryForEffect emits (effect_api.go, #1335). Position
// alone (`Contains`) cannot tell a shuffle from a landing on top;
// this is the assertion that can.
func blightsteelShuffleCount(events []game.Event, player uuid.UUID) int {
	n := 0
	for _, ev := range events {
		if ev.Kind == game.EventSearchLibrary && ev.Label == "shuffle" && ev.Actor == player {
			n++
		}
	}
	return n
}

// --- destroyed (battlefield exit, via executeBattlefieldLeaveLocked) -----
//
// Blightsteel is indestructible, so an ordinary DestroyTarget never
// reaches the graveyard replacement at all (CR 702.12b filters it out
// of the doomed set before anything is destroyed — ADR 0013 §5d). A
// sacrifice ignores indestructible (CR 701.21a) and is the route the
// card's own doc comment names for exercising this window from the
// battlefield; it goes through the same "put into a graveyard from
// anywhere" replacement and the same executeBattlefieldLeaveLocked
// landing this fix touches.

func TestBlightsteelBattlefieldExitShufflesItsLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	colossus := pushBlightsteelOnBattlefield(g, me)

	cut := len(g.Events)
	var applyErr error
	g.WithWriteLock(func() {
		applyErr = (SacrificePermanent{Target: colossus}).Apply(NewContext(g, &game.StackItem{Controller: me.ID}))
	})
	if applyErr != nil {
		t.Fatalf("SacrificePermanent: %v", applyErr)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(colossus) {
		t.Fatalf("Blightsteel Colossus is still on the battlefield")
	}
	if me.Graveyard.Contains(colossus) {
		t.Errorf("Blightsteel Colossus landed in the graveyard instead of the library")
	}
	if !me.Library.Contains(colossus) {
		t.Fatalf("Blightsteel Colossus is not in its owner's library")
	}
	if n := blightsteelShuffleCount(g.Events[cut:], me.ID); n != 1 {
		t.Errorf("shuffle events for %s after the sacrifice = %d, want 1", me.ID, n)
	}
}

// --- discarded (hand exit, via executeZoneRouteLocked) -------------------

func TestBlightsteelDiscardShufflesItsLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	// DiscardCards{N: 1} picks at random from the WHOLE hand
	// (DiscardRandomForEffect), so Blightsteel has to be the only
	// card there for a one-card discard to name it deterministically.
	me.Hand.Cards = nil
	colossus := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: colossus, Name: "Blightsteel Colossus", TypeLine: "Artifact Creature — Phyrexian Golem",
		OracleID: blightsteelOracle, Power: 11, Toughness: 11, Owner: me.ID, Controller: me.ID,
	})

	cut := len(g.Events)
	g.WithWriteLock(func() {
		if err := (DiscardCards{Player: me.ID, N: 1}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatalf("DiscardCards: %v", err)
		}
	})

	if me.Hand.Contains(colossus) {
		t.Fatalf("Blightsteel Colossus is still in hand")
	}
	if me.Graveyard.Contains(colossus) {
		t.Errorf("Blightsteel Colossus landed in the graveyard instead of the library")
	}
	if !me.Library.Contains(colossus) {
		t.Fatalf("Blightsteel Colossus is not in its owner's library")
	}
	if n := blightsteelShuffleCount(g.Events[cut:], me.ID); n != 1 {
		t.Errorf("shuffle events for %s after the discard = %d, want 1", me.ID, n)
	}
}

// --- milled (library exit, via executeZoneRouteLocked) -------------------

func TestBlightsteelMillShufflesItsLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	colossus := uuid.New()
	me.Library.PushTop(game.Card{
		InstanceID: colossus, Name: "Blightsteel Colossus", TypeLine: "Artifact Creature — Phyrexian Golem",
		OracleID: blightsteelOracle, Power: 11, Toughness: 11, Owner: me.ID, Controller: me.ID,
	})

	cut := len(g.Events)
	g.WithWriteLock(func() {
		if err := (MillCards{Player: me.ID, N: 1}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatalf("MillCards: %v", err)
		}
	})

	if me.Graveyard.Contains(colossus) {
		t.Errorf("Blightsteel Colossus landed in the graveyard instead of the library")
	}
	if !me.Library.Contains(colossus) {
		t.Fatalf("Blightsteel Colossus is not in its owner's library")
	}
	if n := blightsteelShuffleCount(g.Events[cut:], me.ID); n != 1 {
		t.Errorf("shuffle events for %s after the mill = %d, want 1", me.ID, n)
	}
}

// --- undo + snapshot restore ---------------------------------------------
//
// The replacement is unconditional (no CR 616 prompt), so there is no
// in-flight replacementResume frame to worry about here the way the
// prompted cases in ADR 0013 §5b/§5e do. What this pins is simpler and
// still real: RestoreFrom must put the permanent back on the
// battlefield exactly as it was, and replaying the same sacrifice from
// the restored state must reach the library again with a fresh
// shuffle — not a stale one carried over from before the undo.

func TestBlightsteelUndoAndSnapshotRestoreReplaysTheShuffle(t *testing.T) {
	g := newCatalogGame(t)
	meID := g.Seats[0].ID
	colossus := pushBlightsteelOnBattlefield(g, g.Seats[0])

	before := g.Clone()

	g.WithWriteLock(func() {
		if err := (SacrificePermanent{Target: colossus}).Apply(NewContext(g, &game.StackItem{Controller: meID})); err != nil {
			t.Fatalf("SacrificePermanent: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if !g.Seats[0].Library.Contains(colossus) {
		t.Fatalf("setup: Blightsteel Colossus did not land in the library")
	}

	g.WithWriteLock(func() { g.RestoreFrom(before) })
	if !g.Battlefield.Contains(colossus) {
		t.Fatalf("undo did not restore Blightsteel Colossus to the battlefield")
	}
	if g.Seats[0].Library.Contains(colossus) {
		t.Fatalf("undo left Blightsteel Colossus in the library")
	}

	// Redo, from the restored state: same outcome, and a FRESH shuffle
	// — not the one from before the undo.
	cut := len(g.Events)
	g.WithWriteLock(func() {
		if err := (SacrificePermanent{Target: colossus}).Apply(NewContext(g, &game.StackItem{Controller: meID})); err != nil {
			t.Fatalf("SacrificePermanent (redo): %v", err)
		}
	})
	passPriorityAroundTable(t, g)

	if !g.Seats[0].Library.Contains(colossus) {
		t.Fatalf("redo: Blightsteel Colossus did not land in the library")
	}
	if g.Seats[0].Graveyard.Contains(colossus) {
		t.Errorf("redo: Blightsteel Colossus landed in the graveyard instead of the library")
	}
	if n := blightsteelShuffleCount(g.Events[cut:], meID); n != 1 {
		t.Errorf("shuffle events for %s after the redo = %d, want 1", meID, n)
	}
}
