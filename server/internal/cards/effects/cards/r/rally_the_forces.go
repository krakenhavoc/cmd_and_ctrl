package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rally the Forces — Instant {2}{R} (Mirrodin Besieged, S82 OLD b467, #5576):
//
//	"Attacking creatures get +1/+0 and gain first strike until end of turn."
//
// The set is every creature attacking as the spell resolves, fixed then (CR
// 611.2c): a creature that attacks later, or joins combat after, is not
// affected. Two effects over the same set, a layer 7c pump and a layer 6
// grant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4145c75d-d052-40d2-8ef4-0a08071582b8",
		Name:         "Rally the Forces",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.BoostUntilEOT{Match: effects.AttackingCreature(), Power: 1, Label: "Rally the Forces — +1/+0"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Match: effects.AttackingCreature(), Keywords: []string{"first strike"}, Label: "Rally the Forces — first strike"}.Apply(ctx)
		},
	})
}
