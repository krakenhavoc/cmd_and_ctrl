package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Illusory Angel — Creature — Angel Illusion {2}{U}, 4/4 (Magic 2015, #5556):
//
//	"Cast this spell only if you've cast another spell this turn.
//	 Flying"
//
// A cast condition (CR 601.3) read as the spell is announced: the caster's
// cast tally for the turn already holds a spell. Being announced, this spell is
// not yet in it, so it is "another" one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6ee32b5e-3b29-47ff-87f8-913ec05f6b2f",
		Name:            "Illusory Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CastCondition: func(g *game.Game, controller uuid.UUID, _ game.Card) bool {
			return g.CastTallyFor(controller).Total >= 1
		},
		CastConditionLabel: "Cast this spell only if you've cast another spell this turn.",
	})
}
