package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const wanShiTongOracle = "2bace3ea-7d28-4a1c-857a-d3c3a1c11b59"

// TestWanShiTongEntersTriggerReadsTheCastX — #1312. "Put X +1/+1
// counters on him. Then draw half X cards, rounded down" is an
// ENTERS TRIGGER, so its Build runs after the permanent has already
// landed and its own StackItem is long gone by the time the trigger
// resolves. Before #1312, CastProvenance carried nothing about X, so
// an enters trigger had no way to answer this at all — Goose Mother
// and friends worked around it by moving the X-read into OnResolve, a
// beat early. This proves the real shape: cast at X=5, resolve the
// spell (which queues the trigger), then resolve the trigger and
// check both halves land — 5 counters, floor(5/2)=2 cards drawn.
func TestWanShiTongEntersTriggerReadsTheCastX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	wst := castXSpell(t, g, "Wan Shi Tong, Librarian", "Legendary Creature — Bird Spirit",
		wanShiTongOracle, "{X}{U}{U}", 5, nil)
	// Wan Shi Tong has already left the hand (onto the stack) by the
	// time castXSpell returns, so the hand is back at its baseline —
	// the draw the trigger owes is the only thing left to move it.
	handBefore := me.Hand.Size()
	passPriorityAroundTable(t, g)

	if got := counterCount(g, wst, game.CounterPlusOne); got != 5 {
		t.Errorf("counters = %d, want 5 (X)", got)
	}
	if got := me.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand = %d, want %d (+floor(5/2)=2 cards drawn)", got, handBefore+2)
	}
}

// TestWanShiTongEntersTriggerAtXZeroDrawsNothing — the "rounded down"
// edge: X=1 is half a card, which is zero, and X=0 (cast for its
// printed minimum) puts on no counters and draws nothing. Neither
// case should error or draw a card it didn't earn.
func TestWanShiTongEntersTriggerAtXZeroDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	wst := castXSpell(t, g, "Wan Shi Tong, Librarian", "Legendary Creature — Bird Spirit",
		wanShiTongOracle, "{X}{U}{U}", 0, nil)
	handBefore := me.Hand.Size()
	passPriorityAroundTable(t, g)

	if got := counterCount(g, wst, game.CounterPlusOne); got != 0 {
		t.Errorf("counters = %d, want 0 at X=0", got)
	}
	if got := me.Hand.Size(); got != handBefore {
		t.Errorf("hand = %d, want %d (no draw at X=0)", got, handBefore)
	}
}

// TestWanShiTongTriggersOnlyOnAnOpponentsOwnSearch — #1335. "Whenever
// an opponent searches THEIR library" must fire when an opponent
// tutors from their own library, and must NOT fire when someone
// searches Wan Shi Tong's controller's own library instead (the
// Bribery shape) — before #1335, EventSearchLibrary carried no
// library-owner field, so those two cases were indistinguishable.
func TestWanShiTongTriggersOnlyOnAnOpponentsOwnSearch(t *testing.T) {
	g := newCatalogGame(t)
	me, opponent := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Wan Shi Tong, Librarian",
		"Legendary Creature — Bird Spirit", wanShiTongOracle, false)

	// An opponent searching their OWN library: the printed trigger.
	pushLibraryCardForTest(opponent, game.Card{
		Name: "Opponent's Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
	})
	handBefore := me.Hand.Size()
	g.WithWriteLock(func() {
		_ = g.SearchLibraryForEffect(opponent.ID, func(game.Card) bool { return true },
			game.ZoneHand, 1, false, false)
	})
	answerSearchByID(t, g, opponent.ID, opponent.Library.Cards[0].InstanceID)
	passPriorityAroundTable(t, g)

	if got := counterCount(g, findBattlefieldByOracle(g, wanShiTongOracle), game.CounterPlusOne); got != 1 {
		t.Errorf("counters after an opponent's own search = %d, want 1", got)
	}
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand after an opponent's own search = %d, want %d", got, handBefore+1)
	}

	// Someone casting Bribery to search MY OWN library instead: must
	// NOT trigger, even though the caster (opponent) is still an
	// opponent of Wan Shi Tong's controller — the search reads MY
	// library, not theirs.
	pushLibraryCardForTest(me, game.Card{
		Name: "My Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
	})
	// Bribery is a sorcery: opponent needs their own main phase, empty
	// stack, to cast it.
	aangAdvanceToMain(t, g, 1)
	countersBefore := counterCount(g, findBattlefieldByOracle(g, wanShiTongOracle), game.CounterPlusOne)
	briberyID := uuid.New()
	opponent.Hand.PushTop(game.Card{
		InstanceID: briberyID, Name: "Bribery", TypeLine: "Sorcery",
		OracleID: briberyOracle, Owner: opponent.ID, Controller: opponent.ID,
	})
	if err := g.CastSpell(opponent.ID, briberyID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err != nil {
		t.Fatalf("CastSpell Bribery: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := counterCount(g, findBattlefieldByOracle(g, wanShiTongOracle), game.CounterPlusOne); got != countersBefore {
		t.Errorf("counters after an opponent's Bribery searched MY library = %d, want unchanged %d", got, countersBefore)
	}
}

// findBattlefieldByOracle finds the (only) battlefield card with the
// given oracle ID.
func findBattlefieldByOracle(g *game.Game, oracleID string) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == oracleID {
			return c.InstanceID
		}
	}
	return uuid.Nil
}
