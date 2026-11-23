package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func MomB08Library(p *game.Player, names ...string) []uuid.UUID {
	// names are pushed so that names[0] ends up on top.
	ids := make([]uuid.UUID, len(names))
	for i := len(names) - 1; i >= 0; i-- {
		typeLine := "Creature — Test"
		if names[i][0] == '!' {
			typeLine = "Instant"
		}
		id := uuid.New()
		p.Library.PushTop(game.Card{InstanceID: id, Name: names[i], TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
		ids[i] = id
	}
	return ids
}
