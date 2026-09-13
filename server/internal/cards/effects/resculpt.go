package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Resculpt — Instant, {1}{U} (EDHREC rank 905):
//
//	"Exile target artifact or creature. Its controller creates a 4/4
//	 blue and red Elemental creature token."
//
// Rapid Hybridization's shape with exile in place of destroy and a
// wider clause — an artifact is a legal target, so a Sol Ring can be
// turned into a 4/4. The token goes to the exiled permanent's
// controller, read before the exile (the card carries no controller
// once it is in exile), and "its controller" is honoured for an
// opponent's permanent as for your own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "814e321d-16fd-4f7b-a8d8-2b089be76f2c",
		Name:         "Resculpt",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			controller, ok := controllerOfTarget(ctx, target)
			if !ok {
				return nil
			}
			if err := (ExileTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return CreateToken{
				Controller: controller,
				Template:   b08BlueRedElementalToken(),
				N:          1,
			}.Apply(ctx)
		},
	})
}
