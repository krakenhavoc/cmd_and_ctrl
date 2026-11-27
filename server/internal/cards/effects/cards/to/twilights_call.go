package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Twilight's Call — Sorcery {4}{B}{B} (Duel Decks: Izzet vs. Golgari):
//
//	"You may cast this spell as though it had flash if you pay {2} more to cast
//	 it. (You may cast it any time you could cast an instant.)
//	 Each player returns all creature cards from their graveyard to the
//	 battlefield."
//
// Rout's flash-for-{2} (FlashForMore, CR 601.3c), with a reanimation of every
// graveyard at once: the creature cards are read from every player's graveyard
// before any moves, and enter together (CR 603.6a) each under its owner's
// control, so each player returns their own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "ae0a1d9c-19cb-42ee-97c3-464e38e84615",
		Name:          "Twilight's Call",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.FlashForMore("{2}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, p := range effects.ApnapPlayers(ctx.Game) {
				ids = append(ids, effects.GraveyardCardIDs(ctx, p, game.Card.IsCreature)...)
			}
			return effects.ReturnFromGraveyardTogether{Targets: ids}.Apply(ctx)
		},
	})
}
