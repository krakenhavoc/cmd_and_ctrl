package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dwarven Catapult — Instant {X}{R} (Fallen Empires, S82 OLD b855, #5637):
//
//	"Dwarven Catapult deals X damage divided evenly, rounded down, among all creatures
//	 target opponent controls."
//
// The creatures counted are those the target opponent controls as the spell resolves
// (CR 608.2), each dealt X divided by their number, rounded down, in one batch; nothing
// is dealt when that share is zero, and the spell is countered for an illegal target
// opponent (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2aae77ed-1a91-41f5-8dc2-e353d63a7586",
		Name:         "Dwarven Catapult",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			theirs := func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsCreature() && c.Controller == player }
			victims := effects.MatchingBattlefield(ctx, theirs)
			if len(victims) == 0 {
				return nil
			}
			share := ctx.X() / len(victims)
			if share <= 0 {
				return nil
			}
			return effects.DamageEachMatching(ctx, theirs, share)
		},
	})
}
