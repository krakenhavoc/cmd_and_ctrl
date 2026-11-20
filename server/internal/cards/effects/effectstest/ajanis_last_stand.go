package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ajanis_last_stand_test.go — ADR 0153 PR 2 (#3168).

const AjanisLastStandOracle = "17c4807c-158e-4a86-90f1-d8fb2ebe892a"

func (d DboTable) Avatars(p *game.Player) int { return d.CountNamedUnder(p, "Avatar") }
