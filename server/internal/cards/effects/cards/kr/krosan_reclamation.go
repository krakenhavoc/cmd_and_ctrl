package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krosan Reclamation — Instant {1}{G} (Judgment, #5611):
//
//	"Target player shuffles up to two target cards from their graveyard
//	 into their library.
//	 Flashback {1}{G} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Quandrix Command's third bullet at two cards: a player clause, then a
// FromTargetPlayersGraveyard clause tied to it, and the shared
// TargetPlayerShufflesTheirCards body. An illegal player does nothing at
// all; a legal one shuffles even when no card was picked or every pick
// has left (CR 608.2b). Cast with flashback, the card is on the stack,
// not in the graveyard, so it can't pick itself, and it is exiled as it
// leaves the stack (CR 702.34a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e69252a6-8c27-4af9-8cbe-11eb416a6a46",
		Name:         "Krosan Reclamation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets: effects.Clauses(
			effects.TargetPlayer("target player"),
			effects.FromTargetPlayersGraveyard(
				effects.TargetCardInGraveyard("up to two target cards from their graveyard").WithCount(0, 2), 0),
		),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{1}{G}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TargetPlayerShufflesTheirCards(ctx, item.Targets, 0, 1)
		},
	})
}
