package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blbB1Push puts a catalog creature on the battlefield, able to attack
// and tap this turn.
func BlbB1Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// blbB1Seats is the active player and the next two seats.
func BlbB1Seats(g *game.Game) (me, opp, other *game.Player) {
	n := len(g.Seats)
	a := g.Turn.ActiveSeat
	return g.Seats[a], g.Seats[(a+1)%n], g.Seats[(a+2)%n]
}
