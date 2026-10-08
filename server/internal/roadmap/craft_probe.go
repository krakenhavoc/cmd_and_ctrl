package roadmap

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// declaresCraft is the craft probe (#2124, ADR 0137): an activated
// ability that exiles its source and exiles materials that may come
// from the graveyard (game.ExilePermanentsCost.FromGraveyard) — the
// shape effects.Craft builds and nothing else declares. Kept out of
// registry.go so the seam's entry there stays one hunk.
func declaresCraft(s effects.Spec) bool {
	for _, a := range s.Activated {
		if ec := a.Cost.ExilePermanents; a.Cost.ExileSelf && !ec.Empty() && ec.FromGraveyard {
			return true
		}
	}
	return false
}
