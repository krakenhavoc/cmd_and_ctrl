package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const secondDoctorOracle = "8a6eda31-2791-4def-b9e7-3adde155a4f0"

func SecondDoctorGame(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	MhTickingClock(t)
	g := NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Second Doctor", TypeLine: "Legendary Creature — Time Lord Doctor",
		OracleID: secondDoctorOracle, Power: 2, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	return g, me
}
