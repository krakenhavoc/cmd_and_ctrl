package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rise from the Nether — Sorcery — Arcane {6}{R}:
//
//	"Create a 3/1 black Mindless-One creature token for each noncreature,
//	 nonland card in your graveyard."
//
// The count is taken as the spell resolves (CR 608.2h): the cards in the
// caster's graveyard whose printed card types include neither creature
// nor land. The spell itself is still on the stack, so it is not
// counted. An artifact creature is a creature card and does not count;
// a land that is also an artifact does not either. No such card means no
// tokens, with no error.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cc8c0234-a355-435d-b3de-8a6fc024c628",
		Name:         "Rise from the Nether",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := 0
			if p := ctx.Game.PlayerByIDForEffect(ctx.Controller()); p != nil && p.Graveyard != nil {
				for _, c := range p.Graveyard.Cards {
					if !c.IsCreature() && !c.IsLand() {
						n++
					}
				}
			}
			if n == 0 {
				return nil
			}
			return CreateToken{Template: TokenCard("3/1 black Mindless-One"), N: n}.Apply(ctx)
		},
	})
}
