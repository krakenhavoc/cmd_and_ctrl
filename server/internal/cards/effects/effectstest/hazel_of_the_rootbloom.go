package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hazel_of_the_rootbloom.go — Hazel of the Rootbloom test helpers shared across test packages.

func HazelPoolTotal(p *game.Player) int {
	total := 0
	for _, tok := range p.ManaPool {
		total += tok.N()
	}
	return total
}
