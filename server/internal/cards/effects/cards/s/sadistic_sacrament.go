package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sadistic Sacrament — Sorcery {B}{B}{B} (Zendikar, S82 OLD batch 491, #5583):
//
//	"Kicker {7}
//	 Search target player's library for up to three cards, exile them, then
//	 that player shuffles. If this spell was kicked, instead search that
//	 player's library for up to fifteen cards, exile them, then that player
//	 shuffles."
//
// Jester's Cap's search (the caster chooses from the target player's library,
// "up to" N with no stated quality, exiled, that library shuffled) with N
// three, or fifteen when the spell was kicked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "e3d08582-85ba-48c4-a453-4a5a385a0c3e",
		Name:          "Sadistic Sacrament",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{7}")},
		Purpose:       game.Purpose{Answers: game.AnswerValue},
		Targets:       effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			n := 3
			if ctx.WasKicked() {
				n = 15
			}
			return effects.SearchLibrary{
				Player:       ctx.Controller(),
				LibraryOwner: item.Targets[0].ID,
				Dest:         game.ZoneExile,
				Limit:        n,
				UpTo:         true,
				Shuffle:      true,
				Source:       ctx.Source(),
				Reason:       "Sadistic Sacrament — search target player's library and exile the cards",
			}.Apply(ctx)
		},
	})
}
