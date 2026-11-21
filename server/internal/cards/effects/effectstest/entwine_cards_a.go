package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func EwaTarget(id uuid.UUID, occ int) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id, Mode: occ}
}

func EwaPlayer(id uuid.UUID, occ int) game.TargetRef {
	return game.TargetRef{Kind: game.TargetPlayer, ID: id, Mode: occ}
}
