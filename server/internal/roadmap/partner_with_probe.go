package roadmap

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// declaresPartnerWith is the partner-with probe (#2142): a triggered
// row built by effects.PartnerWith, the entry search of CR 702.124j.
// The deck-construction half is read off oracle text by internal/deck
// and has nothing to probe. Kept out of registry.go so the seam's entry
// there stays one hunk.
func declaresPartnerWith(s effects.Spec) bool {
	return slices.ContainsFunc(s.Triggered, effects.IsPartnerWithRow)
}
