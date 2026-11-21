package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func PlayerRef(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetPlayer, ID: id} }
