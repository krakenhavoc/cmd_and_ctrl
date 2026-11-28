package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Steel Golem — Artifact Creature — Golem {3}, 3/4 (Weatherlight):
//
//	"You can't cast creature spells."
//
// Only its controller is restricted, and only from casting: creatures that
// arrive any other way are fine.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b5cd9c55-34ad-4e1f-92b2-527c9babc4d4",
		Name:         "Steel Golem",
		Completeness: effects.CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			effects.YouCantCastUnless("You can't cast creature spells.", effects.Creature(), nil),
		},
	})
}
