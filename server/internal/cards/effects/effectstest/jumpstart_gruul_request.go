package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func JumpTargets(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, len(ids))
	for i, id := range ids {
		out[i] = game.TargetRef{Kind: game.TargetCard, ID: id, Slot: i}
	}
	return out
}
