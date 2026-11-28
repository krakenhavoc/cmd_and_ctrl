package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skyshroud Condor — Creature — Bird {1}{U}, 2/2 (Tempest, S82 OLD batch 767,
// #5628):
//
//	"Cast this spell only if you've cast another spell this turn.
//	 Flying"
//
// The restriction is the spell's own cast condition, checked once as the spell
// is announced (CR 601.3), and read off the caster's per-turn spell tally. The
// tally counts a spell as it becomes cast, after this announcement, so a
// positive count is another spell, whether it resolved, was countered or is
// still on the stack. A copy of a spell is not cast and does not count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c7c4c67e-c4db-416b-b2e8-f0e59aa80bba",
		Name:            "Skyshroud Condor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CastCondition: func(g *game.Game, controller uuid.UUID, _ game.Card) bool {
			return g.SpellsCastThisTurn[controller].Total > 0
		},
		CastConditionLabel: "Cast this spell only if you've cast another spell this turn.",
	})
}
