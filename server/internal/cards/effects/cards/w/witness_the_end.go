package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Witness the End — {3}{B} Sorcery (Oath of the Gatewatch, S82 OLD b295):
//
//	"Devoid (This card has no color.)
//	 Target opponent exiles two cards from their hand and loses 2 life."
//
// The target opponent chooses which two cards to exile (as many as they hold,
// up to two), and loses the life whether or not they had cards. A target gone
// by resolution does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "340d2261-9454-4ab0-827c-49c7e76c7781",
		Name:            "Witness the End",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Targets:         effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id := effects.TargetedPlayerOf(ctx)
			if id == uuid.Nil {
				return nil
			}
			if err := effects.OldB295TargetOpponentExilesFromHand(ctx.Game, item, 2); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(item.SourceCardID, id, -2)
		},
	})
}
