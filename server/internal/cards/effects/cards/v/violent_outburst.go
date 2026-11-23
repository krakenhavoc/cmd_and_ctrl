package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Violent Outburst — Instant {1}{R}{G} (Foundations Jumpstart, #5929):
//
//	"Cascade
//	 Creatures you control get +1/+0 until end of turn."
//
// Cascade is the shared trigger (Bituminous Blast); it resolves first,
// above the spell, as CR 702.85a has it.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c0f69063-9079-4e7a-98dc-c29da8171d33",
		Name:         "Violent Outburst",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		Triggered:    []game.TriggeredAbility{effects.Cascade()},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1,
				Label: "Violent Outburst — +1/+0",
			}.Apply(ctx)
		},
	})
}
