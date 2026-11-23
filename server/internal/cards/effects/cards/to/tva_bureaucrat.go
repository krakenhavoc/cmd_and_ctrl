package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// TVA Bureaucrat — {1}{U} Creature — Human Advisor 1/3 (Marvel Super
// Heroes Commander, #4455):
//
//	"Whenever you cast a noncreature spell, this creature gets +1/+0 until
//	 end of turn and can't be blocked this turn."
//
// Both halves apply to the Bureaucrat only if it is still on the
// battlefield as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9638ad1-4e1d-4721-91f5-8cf13d0d0d0d",
		Name:         "TVA Bureaucrat",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "TVA Bureaucrat — gets +1/+0 until end of turn and can't be blocked this turn",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					ctx := effects.NewContext(g, item)
					if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: "TVA Bureaucrat — +1/+0"}).Apply(ctx); err != nil {
						return err
					}
					return effects.RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked, Label: "TVA Bureaucrat — can't be blocked"}.Apply(ctx)
				}),
		},
	})
}
