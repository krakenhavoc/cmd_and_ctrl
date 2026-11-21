package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guardian of the Great Door — Creature — Angel {W}{W} (#3868):
//
//	"As an additional cost to cast this spell, tap four untapped
//	 artifacts, creatures, and/or lands you control.
//	 Flying"
//
// The additional cost is New Blood's (TapUntappedCost) over any mix of
// the three types. A land tapped to pay it is not tapped for mana, so the
// {W}{W} comes from other sources: the auto-tapper keeps the named
// permanents out of the mana it plans (CR 601.2g before 601.2h).
// Creatures that entered this turn may pay (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "18ba355d-a25f-4a5d-b734-11739b0b86af",
		Name:         "Guardian of the Great Door",
		Completeness: effects.CompletenessFull,
		AdditionalCost: effects.TapUntappedCost(4, "four untapped artifacts, creatures, and/or lands you control",
			effects.Or(effects.Artifact(), effects.Creature(), effects.Land())),
		PrintedKeywords: []string{"flying"},
	})
}
