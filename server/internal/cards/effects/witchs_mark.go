package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "00807825-88e5-4b88-8527-a49f553a91be",
		Name:         "Witch's Mark",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to one target creature you control", YouControl()).WithCount(0, 1),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if id, ok := b16FirstLegalTargetCard(ctx); ok {
				if err := (CreateRoleToken{Role: RoleWicked, Host: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return b39MayDiscardThenDraw(1, false, label, func(discarded int) int {
				if discarded > 0 {
					return 2
				}
				return 0
			})(ctx)
		},
	})
}
