package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func PlayerRefs(ids ...uuid.UUID) []game.TargetRef {
	var out []game.TargetRef
	for _, id := range ids {
		out = append(out, game.TargetRef{Kind: game.TargetPlayer, ID: id})
	}
	return out
}
