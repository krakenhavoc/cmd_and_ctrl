package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mindwrack Demon — Creature — Demon {2}{B}{B}, 4/5 (Shadows over
// Innistrad, #5543):
//
//	"Flying, trample
//	 When this creature enters, mill four cards.
//	 Delirium — At the beginning of your upkeep, you lose 4 life unless
//	 there are four or more card types among cards in your graveyard."
//
// Not an intervening if: the upkeep trigger always goes on the stack and
// the graveyard is read as it resolves (CR 608.2c), so a graveyard that
// fills or empties in response decides it. The life loss is not damage
// (CR 119.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3dda8d3d-89e2-4600-be8d-9b76b4f36e33",
		Name:            "Mindwrack Demon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Mindwrack Demon — mill four cards", effects.Do(effects.MillCards{N: 4})),
			effects.AtYourUpkeep("Mindwrack Demon — delirium: you lose 4 life unless there are four or more card types among cards in your graveyard",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					if effects.DrcDeliriumMet(ctx.Game, ctx.Controller()) {
						return nil
					}
					return effects.YouLoseLife(4)(ctx)
				}))),
		},
	})
}
