package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func FcgToken(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1,
		Colors: []string{"R"}, Owner: owner, Controller: owner,
	})
}
