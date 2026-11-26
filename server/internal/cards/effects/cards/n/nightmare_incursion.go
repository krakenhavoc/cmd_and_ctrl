package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nightmare Incursion — Sorcery {5}{B}:
//
//	"Search target player's library for up to X cards, where X is the number
//	 of Swamps you control, and exile them. Then that player shuffles."
//
// Sadistic Sacrament's search: the caster chooses, the library searched and
// shuffled is the target's. X is counted as the spell resolves (CR 608.2h).
// No Swamps means no cards, and the library is still shuffled.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e734759-d43a-4bbe-baff-26d866ff0923",
		Name:         "Nightmare Incursion",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			swamps := effects.CountControlled(ctx.Game, ctx.Controller(), effects.IsLandWithSubtype("Swamp"))
			return effects.SearchLibrary{
				Player:       ctx.Controller(),
				LibraryOwner: item.Targets[0].ID,
				Dest:         game.ZoneExile,
				Limit:        swamps,
				UpTo:         true,
				Shuffle:      true,
				Source:       ctx.Source(),
				Reason:       "Nightmare Incursion — search target player's library and exile the cards",
			}.Apply(ctx)
		},
	})
}
