package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b411_helpers.go — Gatecrash, batch 411 (#5565).

// S82OldB411XTargetNonlandPermanents is "X target nonland permanents": a
// permanent clause whose count is the announced X (Gridlock), Icy Blast's
// X-target clause over nonland permanents.
func S82OldB411XTargetNonlandPermanents() *game.TargetSpec {
	spec := TargetPermanent("X target nonland permanents", Not(Land()))
	spec.CountFromX = true
	return spec
}
