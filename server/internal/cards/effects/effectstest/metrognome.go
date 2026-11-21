package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// metrognome_test.go — ADR 0153 PR 2 (#3168).

const MetrognomeOracle = "57fe5c94-ee93-45c6-bf7e-89d177d0fe70"

func (d DboTable) Gnomes(p *game.Player) int { return d.CountNamedUnder(p, "Gnome") }
