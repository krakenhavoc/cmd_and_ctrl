package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jinx — Instant {1}{U}:
//
//	"Target land becomes the basic land type of your choice until end of
//	 turn.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// ADR 0109 §1 (#1881): CR 305.7 from a resolved spell. The type is chosen
// as it resolves (CR 608.2); until end of turn the land's land types are
// replaced by it (its other subtypes stay, CR 205.1a), it loses the
// abilities its rules text gives it and taps for the chosen colour (CR
// 305.6). The draw is a CR 603.7 delayed trigger at the next upkeep at the
// table, whoever's it is (Gravebind's reading), scheduled after the choice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6e1b51c-ba4b-4997-968a-751b4cba110f",
		Name:         "Jinx",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.TargetLandBecomesChosenTypeThen(ctx, "Jinx", jinxDelayedDraw)
		},
	})
}

// jinxDelayedDraw schedules the draw at the next upkeep, after the land's
// type. An on-disk key.
var jinxDelayedDraw = effects.CardTail("then/jinx-delayed-draw", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.ScheduleDelayedTrigger{
		At:    game.StepUpkeep,
		Label: "Jinx — draw a card",
		Body:  effects.DrawOneBody,
	}.Apply(ctx)
})
