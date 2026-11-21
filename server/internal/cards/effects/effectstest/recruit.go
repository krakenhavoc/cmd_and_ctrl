package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func RecruitSoldiers(g *game.Game, me uuid.UUID) int {
	n, _ := CountTokensNamed(g, me, effects.RecruitSoldier().Name)
	return n
}
