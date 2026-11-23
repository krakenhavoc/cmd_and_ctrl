package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// as_enters_migration_test.go — #578 (Discussion #560): every printed
// "When ~ enters" that used to run from the direct entry hook now goes
// on the stack. The assertion that earns its keep is the response
// window: before priority passes the trigger is on the stack and its
// effect has NOT happened; after, it has.
//
// The cards enter from hand through MoveCardByID, the path a played
// land or a sandbox drop takes. That path runs the entry pipeline and
// drains harvested triggers onto the stack before it returns, which is
// what lets the test look at the stack between entry and resolution.

// enterFromHand puts a card in owner's hand and moves it onto the
// battlefield.
func EnterFromHand(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p := PlayerByIDForTest(g, owner)
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: 2, Toughness: 1, Owner: owner, Controller: owner,
	})
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneHand, Owner: owner},
		game.ZoneRef{Kind: game.ZoneBattlefield},
		id,
	); err != nil {
		t.Fatalf("enter %s: %v", name, err)
	}
	return id
}

// passUntilSpellResolves passes priority until no spell card is left
// on the stack. Anything the resolution queued — an ETB trigger — has
// been drained onto StackMeta at that boundary and is still waiting,
// which is the moment these tests want to look at.
func PassUntilSpellResolves(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 32; i++ {
		if g.Stack.Size() == 0 {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	t.Fatalf("the spell did not resolve after 32 priority passes")
}
