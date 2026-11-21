package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lion Umbra (#3654): enchants a modified creature, +3/+3, reach and
// vigilance, and umbra armor.

// modifiedBear is a 2/2 with a +1/+1 counter on it, which makes it
// modified (CR 700.9) and so a legal host.
func ModifiedBear(g *game.Game, owner uuid.UUID) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: TestCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
		Counters: map[string]int{"+1/+1": 1},
	})
}
