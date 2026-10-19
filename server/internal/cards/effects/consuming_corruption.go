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
		OnResolve: ResolveSteps(
			// The life gain waits for a prompt the damage asks (CR 608.2c, #4155). The
			// Swamps are counted again for it: damage does not change them.
			Instruction(func(ctx *Context) error {
				item := ctx.Item
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				x := swampsControlledBy(ctx.Game, ctx.Controller(), ctx.TextSubtype("Swamp"))
				return DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: x}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				if t := ctx.Item.Targets; len(t) == 0 || t[0].Kind != game.TargetCard {
					return nil
				}
				return GainLife{Amount: swampsControlledBy(ctx.Game, ctx.Controller(), ctx.TextSubtype("Swamp"))}.Apply(ctx)
			}),
		),
	})
}
