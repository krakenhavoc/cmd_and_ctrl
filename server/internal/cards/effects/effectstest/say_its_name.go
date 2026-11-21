package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func LibraryOrder(me *game.Player) []uuid.UUID {
	var ids []uuid.UUID
	for _, c := range me.Library.Cards {
		ids = append(ids, c.InstanceID)
	}
	return ids
}
