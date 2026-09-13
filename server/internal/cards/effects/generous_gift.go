package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Generous Gift — Instant {2}{W}:
//
//	"Destroy target permanent. Its controller creates a 3/3 white
//	Elephant creature token."
//
// White's Beast Within, printing-for-printing identical in shape.
// See beast_within.go for why the controller is read before the
// destroy rather than after.
func init() {
	Register(Spec{
		OracleID:     "fae37e28-e137-4177-b973-fa8b4dd8f409",
		Name:         "Generous Gift",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The Elephant token is created colorless instead of white, so anything that cares about a creature's color doesn't see it."},
		Targets:      TargetPermanent("target permanent"),
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
			return CreateToken{
				Controller: controller,
				Template:   WhiteElephantToken(),
				N:          1,
			}.Apply(ctx)
		},
	})
}
