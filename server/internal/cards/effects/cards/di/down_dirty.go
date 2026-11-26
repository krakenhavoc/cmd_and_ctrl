package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Down // Dirty — split card, oracle eba21e3b-e2b2-4e0b-82e3-f0849943fd89
// (Dragon's Maze, S82 OLD batch 402, #5564):
//
//	Down  — Sorcery {3}{B}: "Target player discards two cards."
//	Dirty — Sorcery {2}{G}: "Return target card from your graveyard to your hand."
//	        Fuse (You may cast one or both halves of this card from your hand.)
//
// Both halves are ADR 0034 faces (Down under the bare oracle ID, Dirty under
// "#1"); the fused cast is the engine's (CR 702.102), Down resolving first, so a
// fused Dirty can take back a card you made yourself discard. Down is Aggressive
// Sabotage's discard of the player's choice (not at random); Dirty is a plain
// return from your own graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eba21e3b-e2b2-4e0b-82e3-f0849943fd89",
		Name:         "Down",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 2}),
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			victim, ok := ctx.ClauseTarget(0)
			if !ok || victim.Kind != game.TargetPlayer {
				return nil
			}
			ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: victim.ID, Source: item.SourceCardID, N: 2})
			return nil
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "eba21e3b-e2b2-4e0b-82e3-f0849943fd89#1",
		Name:         "Dirty",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetCardInGraveyard("target card in your graveyard", effects.YouOwn()),
		OnResolve:    effects.S82OldB402ReturnTargetCardToHand,
	})
}
