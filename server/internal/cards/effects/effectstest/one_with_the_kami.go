package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const OneWithTheKamiOracle = "006f4a62-8590-4ca0-9f51-3aea26cd1a54"

func BigBear(g *game.Game, owner uuid.UUID, power int, counters bool) uuid.UUID {
	c := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: TestCreatureTypeLine,
		Power: power, Toughness: 2, Owner: owner, Controller: owner}
	if counters {
		c.Counters = map[string]int{"+1/+1": 1}
	}
	return PushBattlefieldCardWithTimestamp(g, c)
}
