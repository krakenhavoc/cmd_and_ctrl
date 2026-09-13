package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bedevil — Instant {B}{B}{R} (EDHREC rank 483):
//
//	"Destroy target artifact, creature, or planeswalker."
//
// Rakdos's answer to almost everything that isn't an enchantment or
// a land. One target clause, three types OR-ed, the same shape
// Mortify and Putrefy use.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bceecc64-96f1-4e7b-8904-0aef90377764",
		Name:         "Bedevil",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("target artifact, creature, or planeswalker",
			Or(Artifact(), Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
