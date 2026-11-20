package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stream of Consciousness — Instant — Arcane {1}{U} (Betrayers of
// Kamigawa, #5602):
//
//	"Target player shuffles up to four target cards from their graveyard
//	 into their library."
//
// Krosan Reclamation's two clauses with four cards: the player in slot 0
// and up to four cards from that player's graveyard in slot 1
// (FromTargetPlayersGraveyard), and TargetPlayerShufflesTheirCards for
// the rulings' edges — a legal player shuffles even when every card has
// left, an illegal one does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9013214f-c066-4a09-871f-e9f8396f38a4",
		Name:         "Stream of Consciousness",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets: Clauses(
			TargetPlayer("target player"),
			FromTargetPlayersGraveyard(
				TargetCardInGraveyard("up to four target cards from their graveyard").WithCount(0, 4), 0),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return TargetPlayerShufflesTheirCards(ctx, item.Targets, 0, 1)
		},
	})
}
