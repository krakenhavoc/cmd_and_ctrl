package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ozai's Cruelty — Sorcery — Lesson {2}{B} (Avatar: The Last Airbender,
// #4459):
//
//	"Ozai's Cruelty deals 2 damage to target player. That player
//	 discards two cards."
//
// The spell deals the damage from the stack, so a keyword it has there
// (lifelink from Lo and Li, Twin Tutors) applies to it. Then the same
// player discards two cards of their choice (CR 701.9b), Mind Rot's
// prompt; a hand of one or none discards what it has. A target that is
// no longer legal as the spell resolves is neither dealt damage nor made
// to discard (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aba5c046-2f8c-4749-b347-9a7d8cb1c865",
		Name:         "Ozai's Cruelty",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			victim := TargetedPlayer(ctx)
			if victim == uuid.Nil {
				return nil
			}
			if err := (DealDamage{Source: item.SourceCardID, Target: victim, Amount: 2}).Apply(ctx); err != nil {
				return err
			}
			ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
				Player:   victim,
				Source:   item.SourceCardID,
				N:        2,
				Question: "Ozai's Cruelty — discard two cards",
			})
			return nil
		},
	})
}
