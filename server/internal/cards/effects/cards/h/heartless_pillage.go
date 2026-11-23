package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Heartless Pillage — Sorcery {2}{B} (Breaking News, #5929):
//
//	"Target opponent discards two cards.
//	 Raid — If you attacked this turn, create a Treasure token."
//
// The discard is the opponent's own choice (CR 701.9b), a real prompt, so
// the Treasure waits for it (CR 608.2c). Raid is read as the spell resolves
// (SpellRaid, "if you attacked this turn"), and the Treasure is made even
// when the opponent has no cards to discard: only an illegal target stops
// the whole spell (CR 608.2b), and then nothing happens at all.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8e5e8bca-b52a-4178-ab53-bdd0e0e11dc1",
		Name:         "Heartless Pillage",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 2}),
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				opp := effects.TargetedPlayer(ctx)
				if opp == uuid.Nil {
					return nil
				}
				ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
					Player: opp,
					Source: ctx.Item.SourceCardID,
					N:      2,
				})
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if effects.TargetedPlayer(ctx) == uuid.Nil || !effects.SpellRaid()(ctx.Game, ctx.Item) {
					return nil
				}
				return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(ctx)
			}),
		),
	})
}
