package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Consuming Corruption — Instant {B}{B}:
//
//	"Consuming Corruption deals X damage to target creature or
//	 planeswalker and you gain X life, where X is the number of Swamps
//	 you control."
//
// X is counted once, as the spell resolves. The life is X, not "that
// much": it is gained even if the damage is prevented. If the target
// is gone the whole spell fizzles and nothing is gained (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3bc0bd36-70be-4180-9fb3-b10ce054107b",
		Name:         "Consuming Corruption",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			x := swampsControlledBy(ctx.Game, ctx.Controller())
			if err := (DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: x}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Player: ctx.Controller(), Amount: x}.Apply(ctx)
		},
	})
}
