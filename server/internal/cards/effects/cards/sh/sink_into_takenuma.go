package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sink into Takenuma — Sorcery — Arcane {3}{B} (Saviors of Kamigawa, #5601):
//
//	"Sweep — Return any number of Swamps you control to their owner's hand.
//	 Target player discards a card for each Swamp returned this way."
//
// The Swamps are picked as the spell resolves (OldB591AskSweep) and returned
// together; the target player then discards, by their own choice, one card for
// each that reached a hand (CR 701.8). A target that is illegal by resolution
// leaves the whole spell undone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b1d43244-38b3-498d-8522-9b4011998a66",
		Name:         "Sink into Takenuma",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if _, ok := effects.FirstLegalPlayerTarget(ctx); !ok {
				return nil
			}
			return effects.OldB591AskSweep(ctx, "Swamp", "Sink into Takenuma — return any number of Swamps you control to their owner's hand", sinkIntoTakenumaPick)
		},
	})
}

// sinkIntoTakenumaPick returns the Swamps, then the target discards. An
// on-disk key.
var sinkIntoTakenumaPick = effects.AskedPick("choose-cards/sink-into-takenuma-swamps", func(ctx *effects.Context, r game.CardsChosen) error {
	n, err := effects.OldB591ReturnSwept(ctx, r.Picked)
	if err != nil || n <= 0 {
		return err
	}
	player, ok := effects.FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: player, Source: ctx.Source(), N: n})
	return nil
})
