package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Putrefy — Instant {1}{B}{G}:
//
//	"Destroy target artifact or creature. It can't be regenerated."
//
// Mortify's Golgari counterpart — artifact instead of enchantment.
// Same regeneration note: the clause is inert until regeneration
// exists.
func init() {
	Register(Spec{
		OracleID: "9b271430-f53d-42d6-a547-2f286dd9bcb6",
		Name:     "Putrefy",
		Targets:  TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
