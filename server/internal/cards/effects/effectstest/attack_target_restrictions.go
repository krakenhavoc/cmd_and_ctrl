package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pushStolenForTest puts a catalog creature under `controller`, owned
// by `owner`, as ADR 0102's entry leaves Xantcha. The zone move a real
// entry emits is what lets the layer pass reach its statics; it has
// been here since before the turn, so it is not summoning sick.
func PushStolenForTest(g *game.Game, owner, controller uuid.UUID, name, oracle string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, OracleID: oracle, TypeLine: "Legendary Creature — Test",
		Power: 4, Toughness: 4, Owner: owner, Controller: controller,
	})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
		FindBattlefieldCardForTest(g, id).SummonedThisTurn = false
	})
	return id
}
