package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hewed Stone Retainers — {3} Artifact Creature — Golem 4/4 (Fate Reforged):
//
//	"Cast this spell only if you've cast another spell this turn."
//
// A cast condition (CR 601.2): the per-turn cast tally is bumped after a spell
// reaches the stack, so when this spell is being announced the tally already
// excludes it and "another spell" is any spell the player cast earlier this turn,
// from any zone, countered or not.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a27938d2-9a15-4a87-9e77-dea37347cf62",
		Name:         "Hewed Stone Retainers",
		Completeness: effects.CompletenessFull,
		CastCondition: func(g *game.Game, you uuid.UUID, _ game.Card) bool {
			return g.CastTallyFor(you).Total > 0
		},
		CastConditionLabel: "Cast this spell only if you've cast another spell this turn.",
	})
}
