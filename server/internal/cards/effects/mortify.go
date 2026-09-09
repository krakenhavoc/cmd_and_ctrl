package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mortify — Instant {1}{W}{B}:
//
//	"Destroy target creature or enchantment. It can't be regenerated."
//
// Sandbox note: "can't be regenerated" is a no-op here because
// regeneration isn't implemented — no card in the catalog grants a
// regeneration shield, so there is nothing for the clause to switch
// off. It is recorded rather than dropped so that whoever adds
// regeneration knows this card already promises to beat it.
func init() {
	Register(Spec{
		OracleID: "faa01ed1-ccfa-4e58-951f-cd81f9068027",
		Name:     "Mortify",
		Targets:  TargetPermanent("target creature or enchantment", Or(Creature(), Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
