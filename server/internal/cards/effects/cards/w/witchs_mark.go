package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Witch's Mark — Sorcery {1}{R}:
//
//	"You may discard a card. If you do, draw two cards.
//	 Create a Wicked Role token attached to up to one target creature
//	 you control. (If you control another Role on it, put that one into
//	 the graveyard. Enchanted creature gets +1/+0. When this token is put
//	 into a graveyard, each opponent loses 1 life.)"
//
// The optional discard is a real prompt (b39MayDiscardThenDraw) whose
// draw is the continuation, so "if you do" is the count of cards
// actually pitched. The Role does not wait for the answer: nothing it
// does reads the hand.
//
// No simplifications.
func init() {
	const label = "Witch's Mark — you may discard a card; if you do, draw two cards"
	effects.Register(effects.Spec{
		OracleID:     "00807825-88e5-4b88-8527-a49f553a91be",
		Name:         "Witch's Mark",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to one target creature you control", effects.YouControl()).WithCount(0, 1),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
				if err := (effects.CreateRoleToken{Role: effects.RoleWicked, Host: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.MayDiscardACardToDrawTwo(label)(ctx)
		},
	})
}
