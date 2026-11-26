package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b506_helpers.go — Conflux, batch 506 (#5587): the search rule behind
// "a Plains card, an Island card, a Swamp card, and a Mountain card".

// OldB506ShardTypes are the four land types Shard Convergence searches for.
var OldB506ShardTypes = []string{"Plains", "Island", "Swamp", "Mountain"}

// OldB506OneOfEachShard is the search's set rule: the picks can be matched to
// distinct slots among Plains, Island, Swamp and Mountain, each pick having
// that subtype (a land with two of them fills only one slot). An on-disk key:
// never renamed, never reused.
var OldB506OneOfEachShard = game.RegisterSearchRule("search/s82-old-b506-plains-island-swamp-mountain", func(picked []game.Card, _ int) bool {
	if len(picked) > len(OldB506ShardTypes) {
		return false
	}
	var fill func(i int, used int) bool
	fill = func(i int, used int) bool {
		if i == len(picked) {
			return true
		}
		for slot, t := range OldB506ShardTypes {
			if used&(1<<slot) == 0 && picked[i].HasSubtypeWord(t) && fill(i+1, used|1<<slot) {
				return true
			}
		}
		return false
	}
	return fill(0, 0)
})
