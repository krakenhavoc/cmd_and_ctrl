package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pushRenownCard puts a catalog card (or, with oracle "", a plain
// creature carrying `keywords`) on the battlefield, ready to attack.
func PushRenownCard(g *game.Game, owner uuid.UUID, name, oracle string, power, toughness int, keywords ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Human Soldier", OracleID: oracle,
		Power: power, Toughness: toughness, Keywords: keywords, Owner: owner, Controller: owner,
	})
}

// makeRenowned gives the permanent the designation the way the engine
// does: its counters and the event that switches the gate on.
func MakeRenowned(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		c := FindBattlefieldCardForTest(g, id)
		c.Renowned = true
		g.EmitEvent(game.Event{Kind: game.EventBecameRenowned, CardID: id, Source: id, Target: id, Actor: c.Controller, Amount: 1})
	})
	PassPriorityAroundTable(t, g)
}

func RenownedNow(g *game.Game, id uuid.UUID) bool {
	var r bool
	g.ReadSnapshot(func() { r = g.IsRenowned(id) })
	return r
}
