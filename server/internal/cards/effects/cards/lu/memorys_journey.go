package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Memory's Journey — {1}{U} Instant (Innistrad, S82 OLD b448, #5572):
//
//	"Target player shuffles up to three target cards from their graveyard into
//	 their library.
//	 Flashback {G} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Witness the Future's shape: a player clause and up to three cards from THAT
// player's graveyard (FromTargetPlayersGraveyard), shuffled by
// TargetPlayerShufflesTheirCards. An illegal player does nothing at all; a legal
// one shuffles their library even when no card was picked or every one has left.
// Flashback is the shared graveyard cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75e9ac31-5993-4039-93b2-9380ced2499a",
		Name:         "Memory's Journey",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets: effects.Clauses(
			effects.TargetPlayer("target player"),
			effects.FromTargetPlayersGraveyard(
				effects.TargetCardInGraveyard("up to three target cards from their graveyard").WithCount(0, 3), 0),
		),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{G}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.TargetPlayerShufflesTheirCards(ctx, ctx.Item.Targets, 0, 1)
		},
	})
}
