package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tortoise Formation — Instant {3}{U} (Shards of Alara, S82 OLD batch 518,
// #5588):
//
//	"Creatures you control gain shroud until end of turn. (They can't be the
//	 targets of spells or abilities.)"
//
// Rootborn Defenses' grant: the set of creatures is fixed as the spell
// resolves (CR 611.2c), so one that enters later is not covered, and the keyword
// lasts until the cleanup step. Shroud stops the controller's own spells and
// abilities too (CR 702.18a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3bf84b3-b28a-4531-95d9-d2967ade670b",
		Name:         "Tortoise Formation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerSurvive},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"shroud"},
				Label:    "Tortoise Formation — creatures you control gain shroud",
			}.Apply(ctx)
		},
	})
}
