package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stream of Thought — Sorcery {U} (Modern Horizons):
//
//	"Target player mills four cards. You shuffle up to four cards from your
//	 graveyard into your library.
//	 Replicate {2}{U}{U}"
//
// Replicate is Shattering Spree's pair (ReplicateMana for the cost and
// Replicate() for the cast trigger, CR 702.56a); each copy has its own
// target player and its own choice. The cards to shuffle are chosen as the
// clause resolves, after the mill (CR 608.2c), so milled cards of your own
// are choosable. The library is shuffled only if a card went into it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "e5b29094-0bb4-4c18-ad47-dce4d4379a4f",
		Name:          "Stream of Thought",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetPlayer("target player"),
		Purpose:       game.Purpose{Answers: game.AnswerValue},
		OptionalCosts: []game.AdditionalCost{effects.ReplicateMana("{2}{U}{U}", 20)},
		Triggered:     []game.TriggeredAbility{effects.Replicate()},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			return effects.S82OldB161MillThenShuffleUpToFour(ctx, item.Targets[0].ID)
		},
	})
}
