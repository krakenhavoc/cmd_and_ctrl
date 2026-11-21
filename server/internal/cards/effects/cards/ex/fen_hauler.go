package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fen Hauler — Creature — Insect {6}{B}, 5/5:
//
//	"Improvise ... This creature can't be blocked by artifact creatures."
//
// Improvise is Spec.TapCost; the block restriction is a CantBeBlockedBy
// rule on itself (Questing Beast's shape). No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0f909cfd-c643-48d2-983a-da2e38cd2bff",
		Name:         "Fen Hauler",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Improvise(),
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.And(effects.Creature(), effects.Artifact()), "artifact creatures"),
		},
	})
}
