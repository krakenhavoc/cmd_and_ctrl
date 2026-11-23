package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// echo_cards_test.go — ADR 0108 §5 (#1888): every echo card the PR lands
// charges the printed echo cost at its controller's first upkeep, and the
// cards with more than echo do the rest of what they print.

// seatEchoCard puts a catalogued card on `owner`'s battlefield with a
// body, so it survives the state checks a payment runs.
func SeatEchoCard(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, OracleID: OracleByName(name), TypeLine: typeLine,
		Power: 3, Toughness: 3, Owner: owner, Controller: owner,
	})
}

func OracleByName(name string) string {
	for _, s := range effects.All() {
		if s.Name == name {
			return s.OracleID
		}
	}
	return ""
}
