package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cloistered Youth // Unholy Fiend — Creature — Human {1}{W}, 1/1
// (Innistrad, S82 OLD batch 450, #5572):
//
//	"At the beginning of your upkeep, you may transform this creature."
//	// Unholy Fiend (back): 3/3
//	"At the beginning of your end step, you lose 1 life."
//
// Search for Azcanta's upkeep transform made optional (Vildin-Pack Alpha's
// confirm), and the back face's end-step life loss.
//
// No simplification.
func init() {
	const oracle = "7552a9b4-b82f-491f-ad52-e271cf730211"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Cloistered Youth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Cloistered Youth — you may transform this creature",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Cloistered Youth — transform it?",
						Then:     cloisteredYouthTransform,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Unholy Fiend",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Unholy Fiend — you lose 1 life",
				effects.Do(effects.GainLife{Amount: -1})),
		},
	})
}

var cloisteredYouthTransform = effects.MayYes("confirm/s82-old-b450-cloistered-youth-transform", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	return effects.TransformThis{}.Apply(ctx)
})
