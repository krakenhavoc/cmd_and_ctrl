package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fireglass_mentor.go — Fireglass Mentor test helpers shared across test packages.

func FireglassLands(me *game.Player, n int) []uuid.UUID {
	ids := make([]uuid.UUID, n)
	for i := n - 1; i >= 0; i-- {
		ids[i] = uuid.New()
		me.Library.PushTop(game.Card{InstanceID: ids[i], Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: me.ID, Controller: me.ID})
	}
	return ids
}
