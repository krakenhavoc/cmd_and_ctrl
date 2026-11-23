package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Don't Blink — Instant {1}{U} (Doctor Who):
//
//	"Until end of turn, if one or more creatures would enter from exile or
//	 after being cast from exile, their owners shuffle them into their
//	 libraries instead.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// A game-wide replacement record until end of turn (game.
// ModShuffleInsteadOfEnteringFromExile, CR 614.1a, written by
// ShuffleInsteadOfEnteringFromExileThisTurnForEffect): a creature card that
// would move onto the battlefield from exile, or a creature spell cast from
// exile as it resolves, goes to its owner's library, which is shuffled
// (tokens never). Cycling is the shared keyword action.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5733c3fb-c533-456c-b30e-5d2b9e206b6b",
		Name:         "Don't Blink",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ctx.Game.ShuffleInsteadOfEnteringFromExileThisTurnForEffect(item.SourceCardID,
				"Don't Blink — creatures that would enter from exile are shuffled into their owners' libraries instead")
			return nil
		},
	})
}
