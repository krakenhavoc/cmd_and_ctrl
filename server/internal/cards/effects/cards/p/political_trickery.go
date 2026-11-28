package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Political Trickery — Sorcery {2}{U} (Mirage, S82 OLD b799, #5633):
//
//	"Exchange control of target land you control and target land an opponent
//	 controls. (This effect lasts indefinitely.)"
//
// Trade the Helm's exchange over two lands (ADR 0065): the first yours, the
// second an opponent's. All or nothing (CR 701.12a): if either is illegal at
// resolution nothing changes hands. No duration.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e0eaa43f-1c46-4299-9ecd-5f4a32d292cd",
		Name:         "Political Trickery",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetPermanent("target land you control", effects.Land(), effects.YouControl()),
			effects.TargetPermanent("target land an opponent controls", effects.Land(), effects.OpponentControls()),
		),
		Purpose: game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			mine, okMine := ctx.ClauseTarget(0)
			theirs, okTheirs := ctx.ClauseTarget(1)
			if !okMine || !okTheirs {
				return nil
			}
			return effects.ExchangeControl{A: mine.ID, B: theirs.ID, Label: "Political Trickery — exchange control"}.Apply(ctx)
		},
	})
}
