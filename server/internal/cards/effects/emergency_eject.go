package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emergency Eject — Instant {2}{W} (Edge of Eternities, #4461):
//
//	"Destroy target nonland permanent. Its controller creates a Lander
//	 token. (It's an artifact with "{2}, {T}, Sacrifice this token:
//	 Search your library for a basic land card, put it onto the
//	 battlefield tapped, then shuffle.")"
//
// Beast Within's shape with a nonland target and Beamsaw Prospector's
// Lander (b16LanderToken) for the 3/3. "Its controller" is read before
// the destroy, so the permanent's controller as it last existed gets the
// Lander (CR 608.2h), and gets it even if the permanent wasn't destroyed
// (an indestructible one). With its target gone the spell doesn't
// resolve, so no one gets a Lander (CR 608.2b), as the printed ruling
// says.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "da363da1-aecf-4d49-8e77-8dc56981a741",
		Name:         "Emergency Eject",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target nonland permanent", Not(Land())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			controller, ok := controllerOfTarget(ctx, target)
			if !ok {
				return nil
			}
			if err := (DestroyTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return CreateToken{Controller: controller, Template: b16LanderToken(), N: 1}.Apply(ctx)
		},
	})
}
